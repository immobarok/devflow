package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
	apperrors "github.com/mobarok/devflow/pkg/errors"
)

// Repository manages database operations for authentication and users.
type Repository struct {
	db *sqlx.DB
}

// NewRepository creates a new auth repository.
func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

// FindUserByEmail fetches a user record by email address.
// Returns apperrors.ErrNotFound if the user does not exist.
func (r *Repository) FindUserByEmail(ctx context.Context, email string) (*User, error) {
	var user User
	query := `
		SELECT id, email, password, first_name, last_name, avatar_url, is_active, created_at, updated_at
		FROM users
		WHERE email = $1
	`
	if err := r.db.GetContext(ctx, &user, query, email); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, fmt.Errorf("failed to query user by email: %w", err)
	}
	return &user, nil
}

// FindUserByID fetches a user record by their unique UUID.
// Returns apperrors.ErrNotFound if the user does not exist.
func (r *Repository) FindUserByID(ctx context.Context, id string) (*User, error) {
	var user User
	query := `
		SELECT id, email, password, first_name, last_name, avatar_url, is_active, created_at, updated_at
		FROM users
		WHERE id = $1
	`
	if err := r.db.GetContext(ctx, &user, query, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, fmt.Errorf("failed to query user by id: %w", err)
	}
	return &user, nil
}

// CreateUser inserts a new user record into the database and scans generated fields back into the struct.
func (r *Repository) CreateUser(ctx context.Context, user *User) error {
	query := `
		INSERT INTO users (email, password, first_name, last_name, avatar_url, is_active)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at
	`
	err := r.db.QueryRowContext(
		ctx,
		query,
		user.Email,
		user.Password,
		user.FirstName,
		user.LastName,
		user.AvatarURL,
		user.IsActive,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to insert user: %w", err)
	}
	return nil
}

// SaveRefreshToken persists a newly issued refresh token in the database.
func (r *Repository) SaveRefreshToken(ctx context.Context, token *RefreshToken) error {
	query := `
		INSERT INTO refresh_tokens (user_id, token, expires_at, is_revoked)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at
	`
	err := r.db.QueryRowContext(
		ctx,
		query,
		token.UserID,
		token.Token,
		token.ExpiresAt,
		token.IsRevoked,
	).Scan(&token.ID, &token.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to save refresh token: %w", err)
	}
	return nil
}

// FindRefreshToken retrieves a refresh token record by token string.
func (r *Repository) FindRefreshToken(ctx context.Context, tokenString string) (*RefreshToken, error) {
	var token RefreshToken
	query := `
		SELECT id, user_id, token, expires_at, is_revoked, created_at
		FROM refresh_tokens
		WHERE token = $1
	`
	if err := r.db.GetContext(ctx, &token, query, tokenString); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, fmt.Errorf("failed to query refresh token: %w", err)
	}
	return &token, nil
}

// RevokeRefreshToken marks a refresh token as revoked.
func (r *Repository) RevokeRefreshToken(ctx context.Context, tokenString string) error {
	query := `
		UPDATE refresh_tokens
		SET is_revoked = true
		WHERE token = $1
	`
	res, err := r.db.ExecContext(ctx, query, tokenString)
	if err != nil {
		return fmt.Errorf("failed to revoke refresh token: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}
	if rows == 0 {
		return apperrors.ErrNotFound
	}
	return nil
}
