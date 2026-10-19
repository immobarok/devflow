package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
	apperrors "github.com/mobarok/devflow/pkg/errors"
)

type Repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

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

func (r *Repository) CreatePasswordReset(ctx context.Context, reset *PasswordReset) error {
	query := `
		INSERT INTO password_resets (user_id, otp_hash, expires_at, is_used)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at
	`
	err := r.db.QueryRowContext(
		ctx,
		query,
		reset.UserID,
		reset.OTPHash,
		reset.ExpiresAt,
		reset.IsUsed,
	).Scan(&reset.ID, &reset.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to save password reset record: %w", err)
	}
	return nil
}

func (r *Repository) GetLatestActivePasswordReset(ctx context.Context, userID string) (*PasswordReset, error) {
	var reset PasswordReset
	query := `
		SELECT id, user_id, otp_hash, expires_at, is_used, created_at
		FROM password_resets
		WHERE user_id = $1 AND is_used = false AND expires_at > NOW()
		ORDER BY created_at DESC
		LIMIT 1
	`
	if err := r.db.GetContext(ctx, &reset, query, userID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, fmt.Errorf("failed to query active password reset: %w", err)
	}
	return &reset, nil
}

func (r *Repository) MarkPasswordResetAsUsed(ctx context.Context, id string) error {
	query := `UPDATE password_resets SET is_used = true WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to mark password reset as used: %w", err)
	}
	return nil
}

func (r *Repository) InvalidateExistingPasswordResets(ctx context.Context, userID string) error {
	query := `UPDATE password_resets SET is_used = true WHERE user_id = $1 AND is_used = false`
	_, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to invalidate prior password resets: %w", err)
	}
	return nil
}

func (r *Repository) UpdateUserPassword(ctx context.Context, userID string, hashedPassword string) error {
	query := `
		UPDATE users
		SET password = $1, updated_at = NOW()
		WHERE id = $2
	`
	res, err := r.db.ExecContext(ctx, query, hashedPassword, userID)
	if err != nil {
		return fmt.Errorf("failed to update user password: %w", err)
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

func (r *Repository) RevokeAllUserRefreshTokens(ctx context.Context, userID string) error {
	query := `
		UPDATE refresh_tokens
		SET is_revoked = true
		WHERE user_id = $1 AND is_revoked = false
	`
	_, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to revoke user refresh tokens: %w", err)
	}
	return nil
}

// FindOAuthAccount queries an OAuth record by provider name and provider-assigned user ID.
func (r *Repository) FindOAuthAccount(ctx context.Context, provider, providerID string) (*OAuthAccount, error) {
	var account OAuthAccount
	query := `
		SELECT id, user_id, provider, provider_id, created_at
		FROM oauth_accounts
		WHERE provider = $1 AND provider_id = $2
	`
	if err := r.db.GetContext(ctx, &account, query, provider, providerID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, fmt.Errorf("failed to query oauth account: %w", err)
	}
	return &account, nil
}

// LinkOAuthAccount links an existing user account to an OAuth provider.
func (r *Repository) LinkOAuthAccount(ctx context.Context, userID, provider, providerID string) error {
	query := `
		INSERT INTO oauth_accounts (user_id, provider, provider_id)
		VALUES ($1, $2, $3)
		ON CONFLICT (provider, provider_id) DO NOTHING
	`
	if _, err := r.db.ExecContext(ctx, query, userID, provider, providerID); err != nil {
		return fmt.Errorf("failed to link oauth account: %w", err)
	}
	return nil
}

// CreateOAuthUserTx creates a new user without password and links their OAuth account inside a single transaction.
func (r *Repository) CreateOAuthUserTx(ctx context.Context, user *User, provider, providerID string) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	userQuery := `
		INSERT INTO users (email, password, first_name, last_name, avatar_url, is_active)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at
	`
	err = tx.QueryRowContext(
		ctx,
		userQuery,
		user.Email,
		user.Password,
		user.FirstName,
		user.LastName,
		user.AvatarURL,
		user.IsActive,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create oauth user: %w", err)
	}

	oauthQuery := `
		INSERT INTO oauth_accounts (user_id, provider, provider_id)
		VALUES ($1, $2, $3)
	`
	if _, err := tx.ExecContext(ctx, oauthQuery, user.ID, provider, providerID); err != nil {
		return fmt.Errorf("failed to link oauth account in tx: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit oauth transaction: %w", err)
	}
	return nil
}

// UpdateUserAvatar updates the avatar URL of a user.
func (r *Repository) UpdateUserAvatar(ctx context.Context, userID, avatarURL string) error {
	query := `UPDATE users SET avatar_url = $1, updated_at = NOW() WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, avatarURL, userID)
	if err != nil {
		return fmt.Errorf("failed to update user avatar: %w", err)
	}
	return nil
}

