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
)

type Service interface {
	Register(ctx context.Context, req RegisterRequest) (*AuthResponse, error)
	Login(ctx context.Context, req LoginRequest) (*AuthResponse, error)
	GetMe(ctx context.Context, userID string) (*UserResponse, error)
	ForgotPassword(ctx context.Context, req ForgotPasswordRequest) (string, error)
	VerifyOTP(ctx context.Context, req VerifyOTPRequest) (*VerifyOTPResponse, error)
	ResetPassword(ctx context.Context, req ResetPasswordRequest) error
}

type service struct {
	repo *Repository
	cfg  *config.Config
}

func NewService(repo *Repository, cfg *config.Config) Service {
	return &service{
		repo: repo,
		cfg:  cfg,
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

	user := &User{
		Email:     req.Email,
		Password:  string(hashedPassword),
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

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
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
