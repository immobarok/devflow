package organization

import "time"

type Organization struct {
	ID          string    `db:"id" json:"id"`
	Name        string    `db:"name" json:"name"`
	Slug        string    `db:"slug" json:"slug"`
	Description *string   `db:"description" json:"description,omitempty"`
	LogoURL     *string   `db:"logo_url" json:"logo_url,omitempty"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time `db:"updated_at" json:"updated_at"`
}

type OrganizationMember struct {
	ID             string    `db:"id" json:"id"`
	OrganizationID string    `db:"organization_id" json:"organization_id"`
	UserID         string    `db:"user_id" json:"user_id"`
	Role           string    `db:"role" json:"role"`
	JoinedAt       time.Time `db:"joined_at" json:"joined_at"`
}

type MemberWithUser struct {
	ID             string    `db:"id" json:"id"`
	OrganizationID string    `db:"organization_id" json:"organization_id"`
	UserID         string    `db:"user_id" json:"user_id"`
	Role           string    `db:"role" json:"role"`
	JoinedAt       time.Time `db:"joined_at" json:"joined_at"`
	Email          string    `db:"email" json:"email"`
	FirstName      string    `db:"first_name" json:"first_name"`
	LastName       string    `db:"last_name" json:"last_name"`
	AvatarURL      *string   `db:"avatar_url" json:"avatar_url,omitempty"`
}

type OrganizationWithRole struct {
	Organization
	UserRole string `db:"user_role" json:"user_role"`
}
type OrganizationInvitation struct {
	ID             string    `db:"id" json:"id"`
	OrganizationID string    `db:"organization_id" json:"organization_id"`
	Email          string    `db:"email" json:"email"`
	Role           string    `db:"role" json:"role"`
	Token          string    `db:"token" json:"token"`
	InvitedBy      *string   `db:"invited_by" json:"invited_by,omitempty"`
	Status         string    `db:"status" json:"status"`
	ExpiresAt      time.Time `db:"expires_at" json:"expires_at"`
	CreatedAt      time.Time `db:"created_at" json:"created_at"`
}
