package tasks

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	apperrors "github.com/mobarok/devflow/pkg/errors"
)

type Repository interface {
	CreateWithNumberTx(ctx context.Context, task *Task) error
	GetByID(ctx context.Context, id string) (*TaskWithDetails, error)
	GetProjectMeta(ctx context.Context, projectID string) (orgID string, key string, err error)
	ListWithFilters(ctx context.Context, projectID string, filters TaskFilterQuery) ([]TaskWithDetails, int, error)
	Update(ctx context.Context, task *Task) error
	Delete(ctx context.Context, id string) error
	IsUserProjectMember(ctx context.Context, projectID, userID string) (bool, error)
}

type repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) Repository {
	return &repository{db: db}
}

func (r *repository) CreateWithNumberTx(ctx context.Context, task *Task) error {
	query := `
		INSERT INTO tasks (
			project_id, organization_id, task_number, title, description,
			status, priority, type, assignee_id, reporter_id, due_date
		)
		VALUES (
			$1, $2,
			(SELECT COALESCE(MAX(task_number), 0) + 1 FROM tasks WHERE project_id = $1),
			$3, $4, $5, $6, $7, $8, $9, $10
		)
		RETURNING id, task_number, created_at, updated_at
	`
	err := r.db.QueryRowContext(
		ctx,
		query,
		task.ProjectID,
		task.OrganizationID,
		task.Title,
		task.Description,
		task.Status,
		task.Priority,
		task.Type,
		task.AssigneeID,
		task.ReporterID,
		task.DueDate,
	).Scan(&task.ID, &task.TaskNumber, &task.CreatedAt, &task.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to insert task: %w", err)
	}

	return nil
}

func (r *repository) GetByID(ctx context.Context, id string) (*TaskWithDetails, error) {
	query := `
		SELECT 
			t.id, t.project_id, t.organization_id, t.task_number, t.title, t.description,
			t.status, t.priority, t.type, t.assignee_id, t.reporter_id, t.due_date,
			t.created_at, t.updated_at,
			p.key AS project_key, p.name AS project_name,
			u_assignee.email AS assignee_email,
			u_assignee.first_name AS assignee_first_name,
			u_assignee.last_name AS assignee_last_name,
			u_assignee.avatar_url AS assignee_avatar_url,
			u_reporter.email AS reporter_email,
			u_reporter.first_name AS reporter_first_name,
			u_reporter.last_name AS reporter_last_name
		FROM tasks t
		JOIN projects p ON p.id = t.project_id
		LEFT JOIN users u_assignee ON u_assignee.id = t.assignee_id
		LEFT JOIN users u_reporter ON u_reporter.id = t.reporter_id
		WHERE t.id = $1
	`
	var task TaskWithDetails
	if err := r.db.GetContext(ctx, &task, query, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, fmt.Errorf("failed to fetch task by id: %w", err)
	}

	return &task, nil
}

func (r *repository) GetProjectMeta(ctx context.Context, projectID string) (string, string, error) {
	query := `SELECT organization_id, key FROM projects WHERE id = $1`
	var meta struct {
		OrganizationID string `db:"organization_id"`
		Key            string `db:"key"`
	}
	if err := r.db.GetContext(ctx, &meta, query, projectID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", "", apperrors.ErrNotFound
		}
		return "", "", fmt.Errorf("failed to get project meta: %w", err)
	}
	return meta.OrganizationID, meta.Key, nil
}

func (r *repository) IsUserProjectMember(ctx context.Context, projectID, userID string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM project_members WHERE project_id = $1 AND user_id = $2)`
	var exists bool
	if err := r.db.GetContext(ctx, &exists, query, projectID, userID); err != nil {
		return false, fmt.Errorf("failed to check project membership: %w", err)
	}
	return exists, nil
}

func (r *repository) ListWithFilters(ctx context.Context, projectID string, filters TaskFilterQuery) ([]TaskWithDetails, int, error) {
	whereClauses := []string{"t.project_id = $1"}
	args := []any{projectID}
	argIdx := 2

	if filters.Status != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("t.status = $%d", argIdx))
		args = append(args, filters.Status)
		argIdx++
	}

	if filters.Priority != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("t.priority = $%d", argIdx))
		args = append(args, filters.Priority)
		argIdx++
	}

	if filters.Type != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("t.type = $%d", argIdx))
		args = append(args, filters.Type)
		argIdx++
	}

	if filters.AssigneeID != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("t.assignee_id = $%d", argIdx))
		args = append(args, filters.AssigneeID)
		argIdx++
	}

	if filters.Search != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("(t.title ILIKE $%d OR t.description ILIKE $%d)", argIdx, argIdx))
		args = append(args, "%"+filters.Search+"%")
		argIdx++
	}

	whereSQL := strings.Join(whereClauses, " AND ")

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM tasks t WHERE %s", whereSQL)
	var total int
	if err := r.db.GetContext(ctx, &total, countQuery, args...); err != nil {
		return nil, 0, fmt.Errorf("failed to count tasks: %w", err)
	}

	limit := filters.Limit
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	offset := filters.Offset
	if offset < 0 {
		offset = 0
	}

	dataQuery := fmt.Sprintf(`
		SELECT 
			t.id, t.project_id, t.organization_id, t.task_number, t.title, t.description,
			t.status, t.priority, t.type, t.assignee_id, t.reporter_id, t.due_date,
			t.created_at, t.updated_at,
			p.key AS project_key, p.name AS project_name,
			u_assignee.email AS assignee_email,
			u_assignee.first_name AS assignee_first_name,
			u_assignee.last_name AS assignee_last_name,
			u_assignee.avatar_url AS assignee_avatar_url,
			u_reporter.email AS reporter_email,
			u_reporter.first_name AS reporter_first_name,
			u_reporter.last_name AS reporter_last_name
		FROM tasks t
		JOIN projects p ON p.id = t.project_id
		LEFT JOIN users u_assignee ON u_assignee.id = t.assignee_id
		LEFT JOIN users u_reporter ON u_reporter.id = t.reporter_id
		WHERE %s
		ORDER BY t.created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereSQL, argIdx, argIdx+1)

	queryArgs := append(args, limit, offset)

	var tasks []TaskWithDetails
	if err := r.db.SelectContext(ctx, &tasks, dataQuery, queryArgs...); err != nil {
		return nil, 0, fmt.Errorf("failed to list tasks: %w", err)
	}

	return tasks, total, nil
}

func (r *repository) Update(ctx context.Context, task *Task) error {
	query := `
		UPDATE tasks
		SET title = $1, description = $2, status = $3, priority = $4,
			type = $5, assignee_id = $6, due_date = $7, updated_at = NOW()
		WHERE id = $8
		RETURNING updated_at
	`
	err := r.db.QueryRowContext(
		ctx,
		query,
		task.Title,
		task.Description,
		task.Status,
		task.Priority,
		task.Type,
		task.AssigneeID,
		task.DueDate,
		task.ID,
	).Scan(&task.UpdatedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return apperrors.ErrNotFound
		}
		return fmt.Errorf("failed to update task: %w", err)
	}

	return nil
}

func (r *repository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM tasks WHERE id = $1`
	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
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
