package comments

import (
	"context"
	"fmt"
	"strings"

	"github.com/mobarok/devflow/internal/activity"
	apperrors "github.com/mobarok/devflow/pkg/errors"
)

type Service interface {
	Create(ctx context.Context, taskID, userID string, req CreateCommentRequest) (*CommentResponse, error)
	GetByID(ctx context.Context, id string) (*CommentResponse, error)
	ListByTask(ctx context.Context, taskID string) ([]CommentResponse, error)
	Update(ctx context.Context, id, userID string, req UpdateCommentRequest) (*CommentResponse, error)
	Delete(ctx context.Context, id, userID string) error
}

type service struct {
	repo            Repository
	activityService activity.Service
}

func NewService(repo Repository, activityService activity.Service) Service {
	return &service{
		repo:            repo,
		activityService: activityService,
	}
}

func (s *service) Create(ctx context.Context, taskID, userID string, req CreateCommentRequest) (*CommentResponse, error) {
	orgID, projectID, err := s.repo.GetTaskMeta(ctx, taskID)
	if err != nil {
		return nil, err
	}

	content := strings.TrimSpace(req.Content)
	if content == "" {
		return nil, fmt.Errorf("%w: comment content cannot be empty", apperrors.ErrValidation)
	}

	comment := &Comment{
		TaskID:  taskID,
		UserID:  userID,
		Content: content,
	}

	if err := s.repo.Create(ctx, comment); err != nil {
		return nil, fmt.Errorf("failed to create comment: %w", err)
	}

	if s.activityService != nil {
		_ = s.activityService.Log(ctx, orgID, projectID, &taskID, userID, "comment_created", map[string]any{
			"comment_id": comment.ID,
		})
	}

	details, err := s.repo.GetByID(ctx, comment.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch created comment: %w", err)
	}

	res := ToCommentResponse(details)
	return &res, nil
}

func (s *service) GetByID(ctx context.Context, id string) (*CommentResponse, error) {
	details, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	res := ToCommentResponse(details)
	return &res, nil
}

func (s *service) ListByTask(ctx context.Context, taskID string) ([]CommentResponse, error) {
	if _, _, err := s.repo.GetTaskMeta(ctx, taskID); err != nil {
		return nil, err
	}

	comments, err := s.repo.ListByTask(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to list task comments: %w", err)
	}

	responses := make([]CommentResponse, len(comments))
	for i, c := range comments {
		responses[i] = ToCommentResponse(&c)
	}

	return responses, nil
}

func (s *service) Update(ctx context.Context, id, userID string, req UpdateCommentRequest) (*CommentResponse, error) {
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if existing.UserID != userID {
		return nil, apperrors.ErrForbidden
	}

	content := strings.TrimSpace(req.Content)
	if content == "" {
		return nil, fmt.Errorf("%w: comment content cannot be empty", apperrors.ErrValidation)
	}

	if err := s.repo.Update(ctx, id, content); err != nil {
		return nil, fmt.Errorf("failed to update comment: %w", err)
	}

	updated, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch updated comment: %w", err)
	}

	res := ToCommentResponse(updated)
	return &res, nil
}

func (s *service) Delete(ctx context.Context, id, userID string) error {
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if existing.UserID != userID {
		return apperrors.ErrForbidden
	}

	return s.repo.Delete(ctx, id)
}
