package comments

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
	apperrors "github.com/mobarok/devflow/pkg/errors"
)

type Repository interface {
	Create(ctx context.Context, comment *Comment) error
	GetByID(ctx context.Context, id string) (*CommentWithAuthor, error)
	GetTaskMeta(ctx context.Context, taskID string) (orgID, projectID string, err error)
	ListByTask(ctx context.Context, taskID string) ([]CommentWithAuthor, error)
	Update(ctx context.Context, id, content string) error
	Delete(ctx context.Context, id string) error
}

type repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, comment *Comment) error {
	query := `
		INSERT INTO comments (task_id, user_id, content)
		VALUES ($1, $2, $3)
		RETURNING id, created_at, updated_at
	`
	err := r.db.QueryRowContext(
		ctx,
		query,
		comment.TaskID,
		comment.UserID,
		comment.Content,
	).Scan(&comment.ID, &comment.CreatedAt, &comment.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create comment: %w", err)
	}

	return nil
}

func (r *repository) GetByID(ctx context.Context, id string) (*CommentWithAuthor, error) {
	query := `
		SELECT 
			c.id, c.task_id, c.user_id, c.content, c.created_at, c.updated_at,
			u.email AS author_email, u.first_name AS author_first_name, u.last_name AS author_last_name, u.avatar_url AS author_avatar_url
		FROM comments c
		JOIN users u ON u.id = c.user_id
		WHERE c.id = $1
	`
	var comment CommentWithAuthor
	if err := r.db.GetContext(ctx, &comment, query, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get comment by id: %w", err)
	}

	return &comment, nil
}

func (r *repository) GetTaskMeta(ctx context.Context, taskID string) (string, string, error) {
	query := `SELECT organization_id, project_id FROM tasks WHERE id = $1`
	var meta struct {
		OrganizationID string `db:"organization_id"`
		ProjectID      string `db:"project_id"`
	}
	if err := r.db.GetContext(ctx, &meta, query, taskID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", "", apperrors.ErrNotFound
		}
		return "", "", fmt.Errorf("failed to fetch task meta: %w", err)
	}
	return meta.OrganizationID, meta.ProjectID, nil
}

func (r *repository) ListByTask(ctx context.Context, taskID string) ([]CommentWithAuthor, error) {
	query := `
		SELECT 
			c.id, c.task_id, c.user_id, c.content, c.created_at, c.updated_at,
			u.email AS author_email, u.first_name AS author_first_name, u.last_name AS author_last_name, u.avatar_url AS author_avatar_url
		FROM comments c
		JOIN users u ON u.id = c.user_id
		WHERE c.task_id = $1
		ORDER BY c.created_at ASC
	`
	var comments []CommentWithAuthor
	if err := r.db.SelectContext(ctx, &comments, query, taskID); err != nil {
		return nil, fmt.Errorf("failed to list comments by task: %w", err)
	}

	return comments, nil
}

func (r *repository) Update(ctx context.Context, id, content string) error {
	query := `
		UPDATE comments
		SET content = $1, updated_at = NOW()
		WHERE id = $2
	`
	res, err := r.db.ExecContext(ctx, query, content, id)
	if err != nil {
		return fmt.Errorf("failed to update comment: %w", err)
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

func (r *repository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM comments WHERE id = $1`
	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete comment: %w", err)
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
