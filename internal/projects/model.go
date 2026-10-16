package projects

import "time"

type Project struct {
	ID             string    `db:"id" json:"id"`
	OrganizationID string    `db:"organization_id" json:"organization_id"`
	Name           string    `db:"name" json:"name"`
	Key            string    `db:"key" json:"key"`
	Description    *string   `db:"description" json:"description,omitempty"`
	Status         string    `db:"status" json:"status"`
	CreatedBy      *string   `db:"created_by" json:"created_by,omitempty"`
	CreatedAt      time.Time `db:"created_at" json:"created_at"`
	UpdatedAt      time.Time `db:"updated_at" json:"updated_at"`
}

type ProjectMember struct {
	ID        string    `db:"id" json:"id"`
	ProjectID string    `db:"project_id" json:"project_id"`
	UserID    string    `db:"user_id" json:"user_id"`
	Role      string    `db:"role" json:"role"`
	JoinedAt  time.Time `db:"joined_at" json:"joined_at"`
}

type ProjectMemberWithUser struct {
	ID        string    `db:"id" json:"id"`
	ProjectID string    `db:"project_id" json:"project_id"`
	UserID    string    `db:"user_id" json:"user_id"`
	Role      string    `db:"role" json:"role"`
	JoinedAt  time.Time `db:"joined_at" json:"joined_at"`
	Email     string    `db:"email" json:"email"`
	FirstName string    `db:"first_name" json:"first_name"`
	LastName  string    `db:"last_name" json:"last_name"`
	AvatarURL *string   `db:"avatar_url" json:"avatar_url,omitempty"`
}

type ProjectWithRole struct {
	Project
	UserRole *string `db:"user_role" json:"user_role,omitempty"`
}
