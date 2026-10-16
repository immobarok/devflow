package organization

import "time"

type CreateOrganizationRequest struct {
	Name        string  `json:"name" binding:"required,min=2,max=100"`
	Description *string `json:"description"`
	LogoURL     *string `json:"logo_url"`
}

type UpdateOrganizationRequest struct {
	Name        *string `json:"name" binding:"omitempty,min=2,max=100"`
	Description *string `json:"description"`
	LogoURL     *string `json:"logo_url"`
}

type OrganizationResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	Description *string   `json:"description,omitempty"`
	LogoURL     *string   `json:"logo_url,omitempty"`
	UserRole    string    `json:"user_role,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type MemberResponse struct {
	ID             string    `json:"id"`
	OrganizationID string    `json:"organization_id"`
	UserID         string    `json:"user_id"`
	Role           string    `json:"role"`
	JoinedAt       time.Time `json:"joined_at"`
	Email          string    `json:"email"`
	FirstName      string    `json:"first_name"`
	LastName       string    `json:"last_name"`
	AvatarURL      *string   `json:"avatar_url,omitempty"`
}

type InviteMemberRequest struct {
	Email string `json:"email" binding:"required,email"`
	Role  string `json:"role" binding:"required,oneof=admin developer viewer"`
}

type InvitationResponse struct {
	ID             string    `json:"id"`
	OrganizationID string    `json:"organization_id"`
	Email          string    `json:"email"`
	Role           string    `json:"role"`
	Token          string    `json:"token,omitempty"`
	Status         string    `json:"status"`
	ExpiresAt      time.Time `json:"expires_at"`
	CreatedAt      time.Time `json:"created_at"`
}

type UpdateMemberRoleRequest struct {
	Role string `json:"role" binding:"required,oneof=admin developer viewer"`
}

type AcceptInvitationRequest struct {
	Token string `json:"token" binding:"required"`
}

func ToOrganizationResponse(org *Organization, role string) *OrganizationResponse {
	return &OrganizationResponse{
		ID:          org.ID,
		Name:        org.Name,
		Slug:        org.Slug,
		Description: org.Description,
		LogoURL:     org.LogoURL,
		UserRole:    role,
		CreatedAt:   org.CreatedAt,
		UpdatedAt:   org.UpdatedAt,
	}
}

func ToMemberResponse(m *MemberWithUser) *MemberResponse {
	return &MemberResponse{
		ID:             m.ID,
		OrganizationID: m.OrganizationID,
		UserID:         m.UserID,
		Role:           m.Role,
		JoinedAt:       m.JoinedAt,
		Email:          m.Email,
		FirstName:      m.FirstName,
		LastName:       m.LastName,
		AvatarURL:      m.AvatarURL,
	}
}
