package auth

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/mobarok/devflow/pkg/config"
	apperrors "github.com/mobarok/devflow/pkg/errors"
	jwtpkg "github.com/mobarok/devflow/pkg/jwt"
	"github.com/mobarok/devflow/pkg/oauth"
)

type Service interface {
	Register(ctx context.Context, req RegisterRequest) (*AuthResponse, error)
	Login(ctx context.Context, req LoginRequest) (*AuthResponse, error)
	GetMe(ctx context.Context, userID string) (*UserResponse, error)
	ForgotPassword(ctx context.Context, req ForgotPasswordRequest) (string, error)
	VerifyOTP(ctx context.Context, req VerifyOTPRequest) (*VerifyOTPResponse, error)
	ResetPassword(ctx context.Context, req ResetPasswordRequest) error
	RefreshToken(ctx context.Context, req RefreshTokenRequest) (*RefreshTokenResponse, error)
	Logout(ctx context.Context, req LogoutRequest) error
	GetOAuthAuthURL(provider, state string) (string, error)
	HandleOAuthCallback(ctx context.Context, provider, code string) (*AuthResponse, error)
}

type service struct {
	repo     *Repository
	cfg      *config.Config
	oauthMgr *oauth.Manager
}

func NewService(repo *Repository, cfg *config.Config, oauthMgr *oauth.Manager) Service {
	return &service{
		repo:     repo,
		cfg:      cfg,
		oauthMgr: oauthMgr,
	}
}

func (s *service) Register(ctx context.Context, req RegisterRequest) (*AuthResponse, error) {
	existingUser, err := s.repo.FindUserByEmail(ctx, req.Email)
	if err == nil && existingUser != nil {
		return nil, apperrors.ErrConflict
	} else if err != nil && !errors.Is(err, apperrors.ErrNotFound) {
		return nil, fmt.Errorf("failed to check existing user: %w", err)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	passStr := string(hashedPassword)
	user := &User{
		Email:     req.Email,
		Password:  &passStr,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		IsActive:  true,
	}

	if err := s.repo.CreateUser(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to persist user: %w", err)
	}

	return s.generateAuthResponse(ctx, user)
}

func (s *service) Login(ctx context.Context, req LoginRequest) (*AuthResponse, error) {
	user, err := s.repo.FindUserByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			return nil, apperrors.ErrUnauthorized
		}
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	if !user.IsActive {
		return nil, apperrors.ErrUnauthorized
	}

	if user.Password == nil {
		return nil, apperrors.ErrUnauthorized
	}

	if err := bcrypt.CompareHashAndPassword([]byte(*user.Password), []byte(req.Password)); err != nil {
		return nil, apperrors.ErrUnauthorized
	}

	return s.generateAuthResponse(ctx, user)
}

func (s *service) GetMe(ctx context.Context, userID string) (*UserResponse, error) {
	user, err := s.repo.FindUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return ToUserResponse(user), nil
}

func (s *service) generateAuthResponse(ctx context.Context, user *User) (*AuthResponse, error) {
	accessToken, err := jwtpkg.GenerateAccessToken(
		user.ID,
		user.Email,
		s.cfg.JWTSecret,
		s.cfg.AccessTokenExpiry,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := jwtpkg.GenerateRefreshToken(
		user.ID,
		s.cfg.JWTRefreshSecret,
		s.cfg.RefreshTokenExpiry,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	tokenRecord := &RefreshToken{
		UserID:    user.ID,
		Token:     refreshToken,
		ExpiresAt: time.Now().Add(s.cfg.RefreshTokenExpiry),
		IsRevoked: false,
	}
	if err := s.repo.SaveRefreshToken(ctx, tokenRecord); err != nil {
		return nil, fmt.Errorf("failed to save refresh token session: %w", err)
	}

	return &AuthResponse{
		User:         ToUserResponse(user),
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(s.cfg.AccessTokenExpiry.Seconds()),
	}, nil
}

func (s *service) ForgotPassword(ctx context.Context, req ForgotPasswordRequest) (string, error) {
	user, err := s.repo.FindUserByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			slog.Warn("forgot password requested for non-existing email", "email", req.Email)
			return "", nil
		}
		return "", fmt.Errorf("failed to check user: %w", err)
	}

	otp, err := generateSecureOTP(6)
	if err != nil {
		return "", fmt.Errorf("failed to generate secure otp: %w", err)
	}

	if err := s.repo.InvalidateExistingPasswordResets(ctx, user.ID); err != nil {
		return "", fmt.Errorf("failed to invalidate old otp records: %w", err)
	}

	otpHash, err := bcrypt.GenerateFromPassword([]byte(otp), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash otp: %w", err)
	}

	resetRecord := &PasswordReset{
		UserID:    user.ID,
		OTPHash:   string(otpHash),
		ExpiresAt: time.Now().Add(10 * time.Minute),
		IsUsed:    false,
	}

	if err := s.repo.CreatePasswordReset(ctx, resetRecord); err != nil {
		return "", fmt.Errorf("failed to persist password reset: %w", err)
	}

	slog.Info("password reset OTP generated", "email", user.Email, "otp", otp)

	return otp, nil
}

func (s *service) VerifyOTP(ctx context.Context, req VerifyOTPRequest) (*VerifyOTPResponse, error) {
	user, err := s.repo.FindUserByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			return nil, apperrors.ErrUnauthorized
		}
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	reset, err := s.repo.GetLatestActivePasswordReset(ctx, user.ID)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			return nil, apperrors.ErrUnauthorized
		}
		return nil, fmt.Errorf("failed to find active otp: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(reset.OTPHash), []byte(req.OTP)); err != nil {
		return nil, apperrors.ErrUnauthorized
	}

	if err := s.repo.MarkPasswordResetAsUsed(ctx, reset.ID); err != nil {
		return nil, fmt.Errorf("failed to mark otp as used: %w", err)
	}

	resetToken, err := jwtpkg.GenerateResetToken(user.ID, user.Email, s.cfg.JWTSecret, 15*time.Minute)
	if err != nil {
		return nil, fmt.Errorf("failed to generate reset token: %w", err)
	}

	return &VerifyOTPResponse{
		ResetToken: resetToken,
		Message:    "OTP verified successfully",
	}, nil
}

func (s *service) ResetPassword(ctx context.Context, req ResetPasswordRequest) error {
	claims, err := jwtpkg.ValidateToken(req.ResetToken, s.cfg.JWTSecret)
	if err != nil {
		return apperrors.ErrUnauthorized
	}

	if claims.TokenType != "reset" {
		return apperrors.ErrUnauthorized
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash new password: %w", err)
	}

	if err := s.repo.UpdateUserPassword(ctx, claims.UserID, string(hashedPassword)); err != nil {
		return fmt.Errorf("failed to update user password: %w", err)
	}

	if err := s.repo.RevokeAllUserRefreshTokens(ctx, claims.UserID); err != nil {
		slog.Error("failed to revoke user refresh tokens after password reset", "error", err, "user_id", claims.UserID)
	}

	return nil
}

func generateSecureOTP(length int) (string, error) {
	const digits = "0123456789"
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	for i := range b {
		b[i] = digits[int(b[i])%len(digits)]
	}
	return string(b), nil
}

func (s *service) RefreshToken(ctx context.Context, req RefreshTokenRequest) (*RefreshTokenResponse, error) {
	claims, err := jwtpkg.ValidateToken(req.RefreshToken, s.cfg.JWTRefreshSecret)
	if err != nil {
		return nil, apperrors.ErrUnauthorized
	}

	if claims.TokenType != "refresh" {
		return nil, apperrors.ErrUnauthorized
	}

	tokenRecord, err := s.repo.FindRefreshToken(ctx, req.RefreshToken)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			return nil, apperrors.ErrUnauthorized
		}
		return nil, fmt.Errorf("failed to query refresh token: %w", err)
	}

	if tokenRecord.IsRevoked {
		// Token reuse detection: if an already-revoked refresh token is used, suspect breach and revoke all sessions
		_ = s.repo.RevokeAllUserRefreshTokens(ctx, tokenRecord.UserID)
		return nil, apperrors.ErrUnauthorized
	}

	if time.Now().After(tokenRecord.ExpiresAt) {
		return nil, apperrors.ErrUnauthorized
	}

	user, err := s.repo.FindUserByID(ctx, tokenRecord.UserID)
	if err != nil || !user.IsActive {
		return nil, apperrors.ErrUnauthorized
	}

	// Rotate refresh token
	if err := s.repo.RevokeRefreshToken(ctx, req.RefreshToken); err != nil {
		return nil, fmt.Errorf("failed to revoke old refresh token: %w", err)
	}

	newAccessToken, err := jwtpkg.GenerateAccessToken(user.ID, user.Email, s.cfg.JWTSecret, s.cfg.AccessTokenExpiry)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	newRefreshToken, err := jwtpkg.GenerateRefreshToken(user.ID, s.cfg.JWTRefreshSecret, s.cfg.RefreshTokenExpiry)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	newRecord := &RefreshToken{
		UserID:    user.ID,
		Token:     newRefreshToken,
		ExpiresAt: time.Now().Add(s.cfg.RefreshTokenExpiry),
		IsRevoked: false,
	}
	if err := s.repo.SaveRefreshToken(ctx, newRecord); err != nil {
		return nil, fmt.Errorf("failed to save new refresh token: %w", err)
	}

	return &RefreshTokenResponse{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
		ExpiresIn:    int64(s.cfg.AccessTokenExpiry.Seconds()),
	}, nil
}

func (s *service) Logout(ctx context.Context, req LogoutRequest) error {
	_ = s.repo.RevokeRefreshToken(ctx, req.RefreshToken)
	return nil
}

// GetOAuthAuthURL generates the OAuth authorization URL for the specified provider.
func (s *service) GetOAuthAuthURL(provider, state string) (string, error) {
	if s.oauthMgr == nil {
		return "", fmt.Errorf("oauth manager is not configured")
	}

	p, err := s.oauthMgr.GetProvider(provider)
	if err != nil {
		return "", err
	}

	return p.GetAuthURL(state), nil
}

// HandleOAuthCallback handles exchanging the auth code, fetching the user profile,
// provisioning or linking the account, and issuing JWT access and refresh tokens.
func (s *service) HandleOAuthCallback(ctx context.Context, provider, code string) (*AuthResponse, error) {
	if s.oauthMgr == nil {
		return nil, fmt.Errorf("oauth manager is not configured")
	}

	p, err := s.oauthMgr.GetProvider(provider)
	if err != nil {
		return nil, err
	}

	profile, err := p.ExchangeAndFetchProfile(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to authenticate with %s: %w", provider, err)
	}

	// 1. Check if OAuth account link already exists
	oauthAcc, err := s.repo.FindOAuthAccount(ctx, profile.Provider, profile.ProviderID)
	if err == nil && oauthAcc != nil {
		user, err := s.repo.FindUserByID(ctx, oauthAcc.UserID)
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve linked user: %w", err)
		}
		if !user.IsActive {
			return nil, apperrors.ErrUnauthorized
		}
		return s.generateAuthResponse(ctx, user)
	}

	// 2. Check if a user with this email already exists
	existingUser, err := s.repo.FindUserByEmail(ctx, profile.Email)
	if err == nil && existingUser != nil {
		if !existingUser.IsActive {
			return nil, apperrors.ErrUnauthorized
		}

		if err := s.repo.LinkOAuthAccount(ctx, existingUser.ID, profile.Provider, profile.ProviderID); err != nil {
			return nil, fmt.Errorf("failed to link %s account: %w", provider, err)
		}

		if existingUser.AvatarURL == nil && profile.AvatarURL != "" {
			_ = s.repo.UpdateUserAvatar(ctx, existingUser.ID, profile.AvatarURL)
			existingUser.AvatarURL = &profile.AvatarURL
		}

		return s.generateAuthResponse(ctx, existingUser)
	}

	// 3. New user: atomically create user and OAuth link
	var avatarURL *string
	if profile.AvatarURL != "" {
		avatarURL = &profile.AvatarURL
	}

	newUser := &User{
		Email:     profile.Email,
		Password:  nil,
		FirstName: profile.FirstName,
		LastName:  profile.LastName,
		AvatarURL: avatarURL,
		IsActive:  true,
	}

	if err := s.repo.CreateOAuthUserTx(ctx, newUser, profile.Provider, profile.ProviderID); err != nil {
		return nil, fmt.Errorf("failed to create oauth user: %w", err)
	}

	return s.generateAuthResponse(ctx, newUser)
}

