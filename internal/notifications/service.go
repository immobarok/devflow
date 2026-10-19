package notifications

import (
	"context"
	"fmt"

	"github.com/mobarok/devflow/pkg/worker"
)

type Service interface {
	Dispatch(n *Notification)
	List(ctx context.Context, userID string, limit, offset int) (*NotificationListResponse, error)
	GetUnreadCount(ctx context.Context, userID string) (*UnreadCountResponse, error)
	MarkAsRead(ctx context.Context, id, userID string) error
	MarkAllAsRead(ctx context.Context, userID string) error
}

type service struct {
	repo Repository
	pool *worker.Pool
}

func NewService(repo Repository, pool *worker.Pool) Service {
	return &service{
		repo: repo,
		pool: pool,
	}
}

func (s *service) Dispatch(n *Notification) {
	job := func(ctx context.Context) error {
		return s.repo.Create(ctx, n)
	}

	if s.pool != nil {
		s.pool.Submit(job)
	} else {
		_ = job(context.Background())
	}
}

func (s *service) List(ctx context.Context, userID string, limit, offset int) (*NotificationListResponse, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	list, total, err := s.repo.ListByUser(ctx, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch notifications: %w", err)
	}

	unreadCount, _ := s.repo.GetUnreadCount(ctx, userID)

	responses := make([]NotificationResponse, len(list))
	for i, n := range list {
		responses[i] = ToNotificationResponse(&n)
	}

	return &NotificationListResponse{
		Notifications: responses,
		Total:         total,
		Unread:        unreadCount,
		Limit:         limit,
		Offset:        offset,
	}, nil
}

func (s *service) GetUnreadCount(ctx context.Context, userID string) (*UnreadCountResponse, error) {
	count, err := s.repo.GetUnreadCount(ctx, userID)
	if err != nil {
		return nil, err
	}
	return &UnreadCountResponse{UnreadCount: count}, nil
}

func (s *service) MarkAsRead(ctx context.Context, id, userID string) error {
	return s.repo.MarkAsRead(ctx, id, userID)
}

func (s *service) MarkAllAsRead(ctx context.Context, userID string) error {
	return s.repo.MarkAllAsRead(ctx, userID)
}
