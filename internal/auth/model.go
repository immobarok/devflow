package auth

import "time"

type User struct {
	ID        string    `db:"id" json:"id"`
	Email     string    `db:"email" json:"email"`
	Password  string    `db:"password" json:"-"`
	FirstName string    `db:"first_name" json:"first_name"`
	LastName  string    `db:"last_name" json:"last_name"`
	AvatarURL *string   `db:"avatar_url" json:"avatar_url,omitempty"`
	IsActive  bool      `db:"is_active" json:"is_active"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

type RefreshToken struct {
	ID        string    `db:"id" json:"id"`
	UserID    string    `db:"user_id" json:"user_id"`
	Token     string    `db:"token" json:"token"`
	ExpiresAt time.Time `db:"expires_at" json:"expires_at"`
	IsRevoked bool      `db:"is_revoked" json:"is_revoked"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

type PasswordReset struct {
	ID        string    `db:"id" json:"id"`
	UserID    string    `db:"user_id" json:"user_id"`
	OTPHash   string    `db:"otp_hash" json:"-"`
	ExpiresAt time.Time `db:"expires_at" json:"expires_at"`
	IsUsed    bool      `db:"is_used" json:"is_used"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}
