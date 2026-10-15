package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/mobarok/devflow/pkg/config"
	apperrors "github.com/mobarok/devflow/pkg/errors"
	jwtpkg "github.com/mobarok/devflow/pkg/jwt"
)

// Service defines the contract for authentication operations.
type Service interface {
	Register(ctx context.Context, req RegisterRequest) (*AuthResponse, error)
	Login(ctx context.Context, req LoginRequest) (*AuthResponse, error)
	GetMe(ctx context.Context, userID string) (*UserResponse, error)
}

type service struct {
	repo *Repository
	cfg  *config.Config
}

// NewService creates a new auth service with injected repository and configuration.
func NewService(repo *Repository, cfg *config.Config) Service {
	return &service{
		repo: repo,
		cfg:  cfg,
	}
}

// Register creates a new user account, hashes the password with bcrypt, and issues JWT tokens.
func (s *service) Register(ctx context.Context, req RegisterRequest) (*AuthResponse, error) {
	// Check if a user with this email already exists
	existingUser, err := s.repo.FindUserByEmail(ctx, req.Email)
	if err == nil && existingUser != nil {
		return nil, apperrors.ErrConflict
	} else if err != nil && !errors.Is(err, apperrors.ErrNotFound) {
		return nil, fmt.Errorf("failed to check existing user: %w", err)
	}

	// Hash password using bcrypt
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

// Login verifies user credentials and returns new JWT access and refresh tokens.
func (s *service) Login(ctx context.Context, req LoginRequest) (*AuthResponse, error) {
	user, err := s.repo.FindUserByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			// Avoid revealing whether the email exists
			return nil, apperrors.ErrUnauthorized
		}
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	if !user.IsActive {
		return nil, apperrors.ErrUnauthorized
	}

	// Verify the password against the stored bcrypt hash
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, apperrors.ErrUnauthorized
	}

	return s.generateAuthResponse(ctx, user)
}

// GetMe retrieves the current authenticated user's profile.
func (s *service) GetMe(ctx context.Context, userID string) (*UserResponse, error) {
	user, err := s.repo.FindUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return ToUserResponse(user), nil
}

// generateAuthResponse creates access and refresh tokens and stores the refresh token session.
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

	// Save refresh token in database for session tracking & revocation
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
