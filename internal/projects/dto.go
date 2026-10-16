package projects

import "time"

type CreateProjectRequest struct {
	Name        string  `json:"name" binding:"required,min=2,max=100"`
	Key         string  `json:"key" binding:"required,min=2,max=10"`
	Description *string `json:"description"`
}

type UpdateProjectRequest struct {
	Name        *string `json:"name" binding:"omitempty,min=2,max=100"`
	Description *string `json:"description"`
	Status      *string `json:"status" binding:"omitempty,oneof=active archived completed"`
}

type AssignMemberRequest struct {
	UserID string `json:"user_id" binding:"required,uuid"`
	Role   string `json:"role" binding:"required,oneof=lead developer viewer"`
}

type UpdateMemberRoleRequest struct {
	Role string `json:"role" binding:"required,oneof=lead developer viewer"`
}

type ProjectResponse struct {
	ID             string    `json:"id"`
	OrganizationID string    `json:"organization_id"`
	Name           string    `json:"name"`
	Key            string    `json:"key"`
	Description    *string   `json:"description,omitempty"`
	Status         string    `json:"status"`
	CreatedBy      *string   `json:"created_by,omitempty"`
	UserRole       *string   `json:"user_role,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type ProjectMemberResponse struct {
	ID        string    `json:"id"`
	ProjectID string    `json:"project_id"`
	UserID    string    `json:"user_id"`
	Role      string    `json:"role"`
	JoinedAt  time.Time `json:"joined_at"`
	Email     string    `json:"email"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	AvatarURL *string   `json:"avatar_url,omitempty"`
}

func ToProjectResponse(p *Project, userRole *string) *ProjectResponse {
	return &ProjectResponse{
		ID:             p.ID,
		OrganizationID: p.OrganizationID,
		Name:           p.Name,
		Key:            p.Key,
		Description:    p.Description,
		Status:         p.Status,
		CreatedBy:      p.CreatedBy,
		UserRole:       userRole,
		CreatedAt:      p.CreatedAt,
		UpdatedAt:      p.UpdatedAt,
	}
}

func ToProjectMemberResponse(m *ProjectMemberWithUser) *ProjectMemberResponse {
	return &ProjectMemberResponse{
		ID:        m.ID,
		ProjectID: m.ProjectID,
		UserID:    m.UserID,
		Role:      m.Role,
		JoinedAt:  m.JoinedAt,
		Email:     m.Email,
		FirstName: m.FirstName,
		LastName:  m.LastName,
		AvatarURL: m.AvatarURL,
	}
}
