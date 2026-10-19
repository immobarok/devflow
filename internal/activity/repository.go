package activity

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
)

type Repository interface {
	Record(ctx context.Context, log *ActivityLog) error
	ListByProject(ctx context.Context, projectID string, limit, offset int) ([]ActivityLogWithActor, int, error)
	ListByTask(ctx context.Context, taskID string, limit, offset int) ([]ActivityLogWithActor, int, error)
}

type repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Record(ctx context.Context, log *ActivityLog) error {
	query := `
		INSERT INTO activity_logs (
			organization_id, project_id, task_id, actor_id, action, details
		)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at
	`
	err := r.db.QueryRowContext(
		ctx,
		query,
		log.OrganizationID,
		log.ProjectID,
		log.TaskID,
		log.ActorID,
		log.Action,
		log.Details,
	).Scan(&log.ID, &log.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to record activity log: %w", err)
	}

	return nil
}

func (r *repository) ListByProject(ctx context.Context, projectID string, limit, offset int) ([]ActivityLogWithActor, int, error) {
	countQuery := `SELECT COUNT(*) FROM activity_logs WHERE project_id = $1`
	var total int
	if err := r.db.GetContext(ctx, &total, countQuery, projectID); err != nil {
		return nil, 0, fmt.Errorf("failed to count project activities: %w", err)
	}

	query := `
		SELECT 
			a.id, a.organization_id, a.project_id, a.task_id, a.actor_id, a.action, a.details, a.created_at,
			u.email AS actor_email, u.first_name AS actor_first_name, u.last_name AS actor_last_name, u.avatar_url AS actor_avatar_url
		FROM activity_logs a
		JOIN users u ON u.id = a.actor_id
		WHERE a.project_id = $1
		ORDER BY a.created_at DESC
		LIMIT $2 OFFSET $3
	`
	var logs []ActivityLogWithActor
	if err := r.db.SelectContext(ctx, &logs, query, projectID, limit, offset); err != nil {
		return nil, 0, fmt.Errorf("failed to fetch project activities: %w", err)
	}

	return logs, total, nil
}

func (r *repository) ListByTask(ctx context.Context, taskID string, limit, offset int) ([]ActivityLogWithActor, int, error) {
	countQuery := `SELECT COUNT(*) FROM activity_logs WHERE task_id = $1`
	var total int
	if err := r.db.GetContext(ctx, &total, countQuery, taskID); err != nil {
		return nil, 0, fmt.Errorf("failed to count task activities: %w", err)
	}

	query := `
		SELECT 
			a.id, a.organization_id, a.project_id, a.task_id, a.actor_id, a.action, a.details, a.created_at,
			u.email AS actor_email, u.first_name AS actor_first_name, u.last_name AS actor_last_name, u.avatar_url AS actor_avatar_url
		FROM activity_logs a
		JOIN users u ON u.id = a.actor_id
		WHERE a.task_id = $1
		ORDER BY a.created_at DESC
		LIMIT $2 OFFSET $3
	`
	var logs []ActivityLogWithActor
	if err := r.db.SelectContext(ctx, &logs, query, taskID, limit, offset); err != nil {
		return nil, 0, fmt.Errorf("failed to fetch task activities: %w", err)
	}

	return logs, total, nil
}
