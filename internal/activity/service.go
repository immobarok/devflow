package activity

import (
	"context"
	"encoding/json"
	"fmt"
)

type Service interface {
	Log(ctx context.Context, orgID, projectID string, taskID *string, actorID, action string, details any) error
	ListByProject(ctx context.Context, projectID string, limit, offset int) (*ActivityListResponse, error)
	ListByTask(ctx context.Context, taskID string, limit, offset int) (*ActivityListResponse, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) Log(ctx context.Context, orgID, projectID string, taskID *string, actorID, action string, details any) error {
	var detailsJSON json.RawMessage = []byte("{}")
	if details != nil {
		b, err := json.Marshal(details)
		if err == nil {
			detailsJSON = b
		}
	}

	log := &ActivityLog{
		OrganizationID: orgID,
		ProjectID:      projectID,
		TaskID:         taskID,
		ActorID:        actorID,
		Action:         action,
		Details:        detailsJSON,
	}

	return s.repo.Record(ctx, log)
}

func (s *service) ListByProject(ctx context.Context, projectID string, limit, offset int) (*ActivityListResponse, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	logs, total, err := s.repo.ListByProject(ctx, projectID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list project activities: %w", err)
	}

	responses := make([]ActivityResponse, len(logs))
	for i, l := range logs {
		responses[i] = ToActivityResponse(&l)
	}

	return &ActivityListResponse{
		Activities: responses,
		Total:      total,
		Limit:      limit,
		Offset:     offset,
	}, nil
}

func (s *service) ListByTask(ctx context.Context, taskID string, limit, offset int) (*ActivityListResponse, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	logs, total, err := s.repo.ListByTask(ctx, taskID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list task activities: %w", err)
	}

	responses := make([]ActivityResponse, len(logs))
	for i, l := range logs {
		responses[i] = ToActivityResponse(&l)
	}

	return &ActivityListResponse{
		Activities: responses,
		Total:      total,
		Limit:      limit,
		Offset:     offset,
	}, nil
}
