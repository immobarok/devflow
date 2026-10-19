package notifications

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
	apperrors "github.com/mobarok/devflow/pkg/errors"
)

type Repository interface {
	Create(ctx context.Context, n *Notification) error
	ListByUser(ctx context.Context, userID string, limit, offset int) ([]NotificationWithActor, int, error)
	GetUnreadCount(ctx context.Context, userID string) (int, error)
	MarkAsRead(ctx context.Context, id, userID string) error
	MarkAllAsRead(ctx context.Context, userID string) error
}

type repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, n *Notification) error {
	query := `
		INSERT INTO notifications (
			user_id, actor_id, organization_id, title, message, type, entity_type, entity_id
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at
	`
	err := r.db.QueryRowContext(
		ctx,
		query,
		n.UserID,
		n.ActorID,
		n.OrganizationID,
		n.Title,
		n.Message,
		n.Type,
		n.EntityType,
		n.EntityID,
	).Scan(&n.ID, &n.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to insert notification: %w", err)
	}

	return nil
}

func (r *repository) ListByUser(ctx context.Context, userID string, limit, offset int) ([]NotificationWithActor, int, error) {
	countQuery := `SELECT COUNT(*) FROM notifications WHERE user_id = $1`
	var total int
	if err := r.db.GetContext(ctx, &total, countQuery, userID); err != nil {
		return nil, 0, fmt.Errorf("failed to count notifications: %w", err)
	}

	query := `
		SELECT 
			n.id, n.user_id, n.actor_id, n.organization_id, n.title, n.message,
			n.type, n.entity_type, n.entity_id, n.read_at, n.created_at,
			u.email AS actor_email, u.first_name AS actor_first_name, u.last_name AS actor_last_name, u.avatar_url AS actor_avatar_url
		FROM notifications n
		LEFT JOIN users u ON u.id = n.actor_id
		WHERE n.user_id = $1
		ORDER BY n.created_at DESC
		LIMIT $2 OFFSET $3
	`
	var list []NotificationWithActor
	if err := r.db.SelectContext(ctx, &list, query, userID, limit, offset); err != nil {
		return nil, 0, fmt.Errorf("failed to list notifications: %w", err)
	}

	return list, total, nil
}

func (r *repository) GetUnreadCount(ctx context.Context, userID string) (int, error) {
	query := `SELECT COUNT(*) FROM notifications WHERE user_id = $1 AND read_at IS NULL`
	var count int
	if err := r.db.GetContext(ctx, &count, query, userID); err != nil {
		return 0, fmt.Errorf("failed to get unread notification count: %w", err)
	}
	return count, nil
}

func (r *repository) MarkAsRead(ctx context.Context, id, userID string) error {
	query := `UPDATE notifications SET read_at = NOW() WHERE id = $1 AND user_id = $2 AND read_at IS NULL`
	res, err := r.db.ExecContext(ctx, query, id, userID)
	if err != nil {
		return fmt.Errorf("failed to mark notification as read: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}
	if rows == 0 {
		return apperrors.ErrNotFound
	}
	return nil
}

func (r *repository) MarkAllAsRead(ctx context.Context, userID string) error {
	query := `UPDATE notifications SET read_at = NOW() WHERE user_id = $1 AND read_at IS NULL`
	_, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to mark all notifications as read: %w", err)
	}
	return nil
}
