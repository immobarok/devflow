package tasks

import (
	"context"
	"fmt"
	"strings"

	apperrors "github.com/mobarok/devflow/pkg/errors"
)

type Service interface {
	Create(ctx context.Context, projectID, userID string, req CreateTaskRequest) (*TaskResponse, error)
	GetByID(ctx context.Context, id string) (*TaskResponse, error)
	ListByProject(ctx context.Context, projectID string, filters TaskFilterQuery) (*TaskListResponse, error)
	Update(ctx context.Context, id string, req UpdateTaskRequest) (*TaskResponse, error)
	Delete(ctx context.Context, id string) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) Create(ctx context.Context, projectID, userID string, req CreateTaskRequest) (*TaskResponse, error) {
	orgID, _, err := s.repo.GetProjectMeta(ctx, projectID)
	if err != nil {
		return nil, err
	}

	isMember, err := s.repo.IsUserProjectMember(ctx, projectID, userID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, fmt.Errorf("%w: you must be a member of this project to create tasks", apperrors.ErrForbidden)
	}

	if req.AssigneeID != nil && *req.AssigneeID != "" {
		isAssigneeMember, err := s.repo.IsUserProjectMember(ctx, projectID, *req.AssigneeID)
		if err != nil {
			return nil, err
		}
		if !isAssigneeMember {
			return nil, fmt.Errorf("%w: assignee must be a member of the project", apperrors.ErrValidation)
		}
	}

	status := "todo"
	if req.Status != nil && *req.Status != "" {
		status = *req.Status
	}

	priority := "medium"
	if req.Priority != nil && *req.Priority != "" {
		priority = *req.Priority
	}

	taskType := "task"
	if req.Type != nil && *req.Type != "" {
		taskType = *req.Type
	}

	task := &Task{
		ProjectID:      projectID,
		OrganizationID: orgID,
		Title:          strings.TrimSpace(req.Title),
		Description:    req.Description,
		Status:         status,
		Priority:       priority,
		Type:           taskType,
		AssigneeID:     req.AssigneeID,
		ReporterID:     &userID,
		DueDate:        req.DueDate,
	}

	if err := s.repo.CreateWithNumberTx(ctx, task); err != nil {
		return nil, fmt.Errorf("failed to create task: %w", err)
	}

	details, err := s.repo.GetByID(ctx, task.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch created task: %w", err)
	}

	return ToTaskResponse(details), nil
}

func (s *service) GetByID(ctx context.Context, id string) (*TaskResponse, error) {
	details, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return ToTaskResponse(details), nil
}

func (s *service) ListByProject(ctx context.Context, projectID string, filters TaskFilterQuery) (*TaskListResponse, error) {
	if _, _, err := s.repo.GetProjectMeta(ctx, projectID); err != nil {
		return nil, err
	}

	if filters.Limit <= 0 || filters.Limit > 100 {
		filters.Limit = 20
	}
	if filters.Offset < 0 {
		filters.Offset = 0
	}

	tasks, total, err := s.repo.ListWithFilters(ctx, projectID, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to list project tasks: %w", err)
	}

	taskResponses := make([]TaskResponse, len(tasks))
	for i, t := range tasks {
		taskResponses[i] = *ToTaskResponse(&t)
	}

	return &TaskListResponse{
		Tasks:  taskResponses,
		Total:  total,
		Limit:  filters.Limit,
		Offset: filters.Offset,
	}, nil
}

func (s *service) Update(ctx context.Context, id string, req UpdateTaskRequest) (*TaskResponse, error) {
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.AssigneeID != nil && *req.AssigneeID != "" {
		isMember, err := s.repo.IsUserProjectMember(ctx, existing.ProjectID, *req.AssigneeID)
		if err != nil {
			return nil, err
		}
		if !isMember {
			return nil, fmt.Errorf("%w: assignee must be a member of the project", apperrors.ErrValidation)
		}
	}

	task := existing.Task
	if req.Title != nil {
		trimmed := strings.TrimSpace(*req.Title)
		if trimmed != "" {
			task.Title = trimmed
		}
	}
	if req.Description != nil {
		task.Description = req.Description
	}
	if req.Status != nil {
		task.Status = *req.Status
	}
	if req.Priority != nil {
		task.Priority = *req.Priority
	}
	if req.Type != nil {
		task.Type = *req.Type
	}
	if req.AssigneeID != nil {
		task.AssigneeID = req.AssigneeID
	}
	if req.DueDate != nil {
		task.DueDate = req.DueDate
	}

	if err := s.repo.Update(ctx, &task); err != nil {
		return nil, fmt.Errorf("failed to update task: %w", err)
	}

	updated, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch updated task: %w", err)
	}

	return ToTaskResponse(updated), nil
}

func (s *service) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}
