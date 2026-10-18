package projects

import (
	"context"
	"fmt"
	"strings"

	apperrors "github.com/mobarok/devflow/pkg/errors"
)

type Service interface {
	Create(ctx context.Context, orgID, userID string, req CreateProjectRequest) (*ProjectResponse, error)
	GetByID(ctx context.Context, projectID, userID string) (*ProjectResponse, error)
	GetByIDAndOrg(ctx context.Context, projectID, orgID, userID string) (*ProjectResponse, error)
	ListByOrganization(ctx context.Context, orgID, userID string) ([]ProjectResponse, error)
	Update(ctx context.Context, projectID string, req UpdateProjectRequest) (*ProjectResponse, error)
	Delete(ctx context.Context, projectID string) error

	AddMember(ctx context.Context, projectID string, req AssignMemberRequest) error
	GetMembers(ctx context.Context, projectID string) ([]ProjectMemberResponse, error)
	UpdateMemberRole(ctx context.Context, projectID, userID string, req UpdateMemberRoleRequest) error
	RemoveMember(ctx context.Context, projectID, userID string) error
	GetMemberRole(ctx context.Context, projectID, userID string) (string, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) Create(ctx context.Context, orgID, userID string, req CreateProjectRequest) (*ProjectResponse, error) {
	key := strings.ToUpper(strings.TrimSpace(req.Key))
	name := strings.TrimSpace(req.Name)

	exists, err := s.repo.KeyExists(ctx, orgID, key)
	if err != nil {
		return nil, fmt.Errorf("failed to check project key uniqueness: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("%w: project key '%s' is already in use in this organization", apperrors.ErrConflict, key)
	}

	project := &Project{
		OrganizationID: orgID,
		Name:           name,
		Key:            key,
		Description:    req.Description,
		Status:         "active",
		CreatedBy:      &userID,
	}

	if err := s.repo.CreateWithMemberTx(ctx, project, userID); err != nil {
		return nil, fmt.Errorf("failed to create project: %w", err)
	}

	leadRole := "lead"
	return ToProjectResponse(project, &leadRole), nil
}

func (s *service) GetByID(ctx context.Context, projectID, userID string) (*ProjectResponse, error) {
	project, err := s.repo.GetByID(ctx, projectID)
	if err != nil {
		return nil, err
	}

	var userRole *string
	member, err := s.repo.GetMember(ctx, projectID, userID)
	if err == nil && member != nil {
		userRole = &member.Role
	}

	return ToProjectResponse(project, userRole), nil
}

func (s *service) GetByIDAndOrg(ctx context.Context, projectID, orgID, userID string) (*ProjectResponse, error) {
	project, err := s.repo.GetByIDAndOrg(ctx, projectID, orgID)
	if err != nil {
		return nil, err
	}

	var userRole *string
	member, err := s.repo.GetMember(ctx, projectID, userID)
	if err == nil && member != nil {
		userRole = &member.Role
	}

	return ToProjectResponse(project, userRole), nil
}

func (s *service) ListByOrganization(ctx context.Context, orgID, userID string) ([]ProjectResponse, error) {
	projectsWithRole, err := s.repo.ListUserProjects(ctx, orgID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list projects: %w", err)
	}

	res := make([]ProjectResponse, len(projectsWithRole))
	for i, p := range projectsWithRole {
		res[i] = *ToProjectResponse(&p.Project, p.UserRole)
	}
	return res, nil
}

func (s *service) Update(ctx context.Context, projectID string, req UpdateProjectRequest) (*ProjectResponse, error) {
	project, err := s.repo.GetByID(ctx, projectID)
	if err != nil {
		return nil, err
	}

	if req.Name != nil {
		trimmed := strings.TrimSpace(*req.Name)
		if trimmed != "" {
			project.Name = trimmed
		}
	}
	if req.Description != nil {
		project.Description = req.Description
	}
	if req.Status != nil {
		project.Status = *req.Status
	}

	if err := s.repo.Update(ctx, project); err != nil {
		return nil, fmt.Errorf("failed to update project: %w", err)
	}

	return ToProjectResponse(project, nil), nil
}

func (s *service) Delete(ctx context.Context, projectID string) error {
	if _, err := s.repo.GetByID(ctx, projectID); err != nil {
		return err
	}
	return s.repo.Delete(ctx, projectID)
}

func (s *service) AddMember(ctx context.Context, projectID string, req AssignMemberRequest) error {
	existing, err := s.repo.GetMember(ctx, projectID, req.UserID)
	if err == nil && existing != nil {
		return fmt.Errorf("%w: user is already a member of this project", apperrors.ErrConflict)
	}

	member := &ProjectMember{
		ProjectID: projectID,
		UserID:    req.UserID,
		Role:      req.Role,
	}

	return s.repo.AddMember(ctx, member)
}

func (s *service) GetMembers(ctx context.Context, projectID string) ([]ProjectMemberResponse, error) {
	members, err := s.repo.GetMembers(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get project members: %w", err)
	}

	res := make([]ProjectMemberResponse, len(members))
	for i, m := range members {
		res[i] = *ToProjectMemberResponse(&m)
	}
	return res, nil
}

func (s *service) UpdateMemberRole(ctx context.Context, projectID, userID string, req UpdateMemberRoleRequest) error {
	if _, err := s.repo.GetMember(ctx, projectID, userID); err != nil {
		return err
	}
	return s.repo.UpdateMemberRole(ctx, projectID, userID, req.Role)
}

func (s *service) RemoveMember(ctx context.Context, projectID, userID string) error {
	if _, err := s.repo.GetMember(ctx, projectID, userID); err != nil {
		return err
	}
	return s.repo.RemoveMember(ctx, projectID, userID)
}

func (s *service) GetMemberRole(ctx context.Context, projectID, userID string) (string, error) {
	member, err := s.repo.GetMember(ctx, projectID, userID)
	if err != nil {
		return "", err
	}
	return member.Role, nil
}