package projects

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
	apperrors "github.com/mobarok/devflow/pkg/errors"
)

type Repository interface {
	CreateWithMemberTx(ctx context.Context, project *Project, creatorID string) error
	GetByID(ctx context.Context, id string) (*Project, error)
	GetByIDAndOrg(ctx context.Context, id, orgID string) (*Project, error)
	GetByKey(ctx context.Context, orgID, key string) (*Project, error)
	KeyExists(ctx context.Context, orgID, key string) (bool, error)
	ListByOrganization(ctx context.Context, orgID string) ([]Project, error)
	ListUserProjects(ctx context.Context, orgID, userID string) ([]ProjectWithRole, error)
	Update(ctx context.Context, project *Project) error
	Delete(ctx context.Context, id string) error
	AddMember(ctx context.Context, member *ProjectMember) error
	GetMember(ctx context.Context, projectID, userID string) (*ProjectMember, error)
	GetMembers(ctx context.Context, projectID string) ([]ProjectMemberWithUser, error)
	UpdateMemberRole(ctx context.Context, projectID, userID, newRole string) error
	RemoveMember(ctx context.Context, projectID, userID string) error
}

type repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) Repository {
	return &repository{db: db}
}

func (r *repository) CreateWithMemberTx(ctx context.Context, project *Project, creatorID string) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	queryProject := `
		INSERT INTO projects (organization_id, name, key, description, status, created_by)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at
	`
	err = tx.QueryRowContext(
		ctx,
		queryProject,
		project.OrganizationID,
		project.Name,
		project.Key,
		project.Description,
		project.Status,
		project.CreatedBy,
	).Scan(&project.ID, &project.CreatedAt, &project.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to insert project: %w", err)
	}

	queryMember := `
		INSERT INTO project_members (project_id, user_id, role)
		VALUES ($1, $2, 'lead')
	`
	if _, err := tx.ExecContext(ctx, queryMember, project.ID, creatorID); err != nil {
		return fmt.Errorf("failed to assign project lead: %w", err)
	}

	return tx.Commit()
}

func (r *repository) GetByID(ctx context.Context, id string) (*Project, error) {
	var project Project
	query := `
		SELECT id, organization_id, name, key, description, status, created_by, created_at, updated_at
		FROM projects
		WHERE id = $1
	`
	if err := r.db.GetContext(ctx, &project, query, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get project: %w", err)
	}
	return &project, nil
}

func (r *repository) GetByIDAndOrg(ctx context.Context, id, orgID string) (*Project, error) {
	var project Project
	query := `
		SELECT id, organization_id, name, key, description, status, created_by, created_at, updated_at
		FROM projects
		WHERE id = $1 AND organization_id = $2
	`
	if err := r.db.GetContext(ctx, &project, query, id, orgID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get project in organization: %w", err)
	}
	return &project, nil
}

func (r *repository) GetByKey(ctx context.Context, orgID, key string) (*Project, error) {
	var project Project
	query := `
		SELECT id, organization_id, name, key, description, status, created_by, created_at, updated_at
		FROM projects
		WHERE organization_id = $1 AND key = $2
	`
	if err := r.db.GetContext(ctx, &project, query, orgID, key); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get project by key: %w", err)
	}
	return &project, nil
}

func (r *repository) KeyExists(ctx context.Context, orgID, key string) (bool, error) {
	var count int
	query := `SELECT COUNT(1) FROM projects WHERE organization_id = $1 AND key = $2`
	if err := r.db.GetContext(ctx, &count, query, orgID, key); err != nil {
		return false, fmt.Errorf("failed to check key existence: %w", err)
	}
	return count > 0, nil
}

func (r *repository) ListByOrganization(ctx context.Context, orgID string) ([]Project, error) {
	var projects []Project
	query := `
		SELECT id, organization_id, name, key, description, status, created_by, created_at, updated_at
		FROM projects
		WHERE organization_id = $1
		ORDER BY created_at DESC
	`
	if err := r.db.SelectContext(ctx, &projects, query, orgID); err != nil {
		return nil, fmt.Errorf("failed to list projects: %w", err)
	}
	return projects, nil
}

func (r *repository) ListUserProjects(ctx context.Context, orgID, userID string) ([]ProjectWithRole, error) {
	var projects []ProjectWithRole
	query := `
		SELECT p.id, p.organization_id, p.name, p.key, p.description, p.status, p.created_by, p.created_at, p.updated_at,
		       pm.role AS user_role
		FROM projects p
		LEFT JOIN project_members pm ON p.id = pm.project_id AND pm.user_id = $2
		WHERE p.organization_id = $1
		ORDER BY p.created_at DESC
	`
	if err := r.db.SelectContext(ctx, &projects, query, orgID, userID); err != nil {
		return nil, fmt.Errorf("failed to list user projects: %w", err)
	}
	return projects, nil
}

func (r *repository) Update(ctx context.Context, project *Project) error {
	query := `
		UPDATE projects
		SET name = $1, description = $2, status = $3, updated_at = NOW()
		WHERE id = $4
		RETURNING updated_at
	`
	err := r.db.QueryRowContext(ctx, query, project.Name, project.Description, project.Status, project.ID).
		Scan(&project.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return apperrors.ErrNotFound
		}
		return fmt.Errorf("failed to update project: %w", err)
	}
	return nil
}

func (r *repository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM projects WHERE id = $1`
	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete project: %w", err)
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

func (r *repository) AddMember(ctx context.Context, member *ProjectMember) error {
	query := `
		INSERT INTO project_members (project_id, user_id, role)
		VALUES ($1, $2, $3)
		RETURNING id, joined_at
	`
	err := r.db.QueryRowContext(ctx, query, member.ProjectID, member.UserID, member.Role).
		Scan(&member.ID, &member.JoinedAt)
	if err != nil {
		return fmt.Errorf("failed to add project member: %w", err)
	}
	return nil
}

func (r *repository) GetMember(ctx context.Context, projectID, userID string) (*ProjectMember, error) {
	var member ProjectMember
	query := `
		SELECT id, project_id, user_id, role, joined_at
		FROM project_members
		WHERE project_id = $1 AND user_id = $2
	`
	if err := r.db.GetContext(ctx, &member, query, projectID, userID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get project member: %w", err)
	}
	return &member, nil
}

func (r *repository) GetMembers(ctx context.Context, projectID string) ([]ProjectMemberWithUser, error) {
	var members []ProjectMemberWithUser
	query := `
		SELECT pm.id, pm.project_id, pm.user_id, pm.role, pm.joined_at,
		       u.email, u.first_name, u.last_name, u.avatar_url
		FROM project_members pm
		JOIN users u ON pm.user_id = u.id
		WHERE pm.project_id = $1
		ORDER BY pm.joined_at ASC
	`
	if err := r.db.SelectContext(ctx, &members, query, projectID); err != nil {
		return nil, fmt.Errorf("failed to get project members: %w", err)
	}
	return members, nil
}

func (r *repository) UpdateMemberRole(ctx context.Context, projectID, userID, newRole string) error {
	query := `
		UPDATE project_members
		SET role = $1
		WHERE project_id = $2 AND user_id = $3
	`
	res, err := r.db.ExecContext(ctx, query, newRole, projectID, userID)
	if err != nil {
		return fmt.Errorf("failed to update project member role: %w", err)
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

func (r *repository) RemoveMember(ctx context.Context, projectID, userID string) error {
	query := `
		DELETE FROM project_members
		WHERE project_id = $1 AND user_id = $2
	`
	res, err := r.db.ExecContext(ctx, query, projectID, userID)
	if err != nil {
		return fmt.Errorf("failed to remove project member: %w", err)
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