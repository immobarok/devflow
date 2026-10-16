package organization

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
	apperrors "github.com/mobarok/devflow/pkg/errors"
)

type Repository interface {
	CreateWithTx(ctx context.Context, org *Organization, ownerID string) error
	GetByID(ctx context.Context, id string) (*Organization, error)
	GetBySlug(ctx context.Context, slug string) (*Organization, error)
	SlugExists(ctx context.Context, slug string) (bool, error)
	GetUserOrganizations(ctx context.Context, userID string) ([]OrganizationWithRole, error)
	Update(ctx context.Context, org *Organization) error
	Delete(ctx context.Context, id string) error
	GetMember(ctx context.Context, orgID, userID string) (*OrganizationMember, error)
	GetMembers(ctx context.Context, orgID string) ([]MemberWithUser, error)
	UpdateMemberRole(ctx context.Context, orgID, userID, newRole string) error
	RemoveMember(ctx context.Context, orgID, userID string) error
	CreateInvitation(ctx context.Context, invite *OrganizationInvitation) error
	GetInvitationByToken(ctx context.Context, token string) (*OrganizationInvitation, error)
	AcceptInvitationTx(ctx context.Context, inviteID, orgID, userID, role string) error
}

type repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) Repository {
	return &repository{db: db}
}

func (r *repository) CreateWithTx(ctx context.Context, org *Organization, ownerID string) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	queryOrg := `
		INSERT INTO organizations (name, slug, description, logo_url)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at
	`
	err = tx.QueryRowContext(ctx, queryOrg, org.Name, org.Slug, org.Description, org.LogoURL).
		Scan(&org.ID, &org.CreatedAt, &org.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to insert organization: %w", err)
	}

	queryMember := `
		INSERT INTO organization_members (organization_id, user_id, role)
		VALUES ($1, $2, 'owner')
	`
	if _, err := tx.ExecContext(ctx, queryMember, org.ID, ownerID); err != nil {
		return fmt.Errorf("failed to insert owner member: %w", err)
	}

	return tx.Commit()
}

func (r *repository) GetByID(ctx context.Context, id string) (*Organization, error) {
	var org Organization
	query := `SELECT id, name, slug, description, logo_url, created_at, updated_at FROM organizations WHERE id = $1`
	if err := r.db.GetContext(ctx, &org, query, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, fmt.Errorf("failed to query organization by id: %w", err)
	}
	return &org, nil
}

func (r *repository) GetBySlug(ctx context.Context, slug string) (*Organization, error) {
	var org Organization
	query := `SELECT id, name, slug, description, logo_url, created_at, updated_at FROM organizations WHERE slug = $1`
	if err := r.db.GetContext(ctx, &org, query, slug); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, fmt.Errorf("failed to query organization by slug: %w", err)
	}
	return &org, nil
}

func (r *repository) SlugExists(ctx context.Context, slug string) (bool, error) {
	var count int
	query := `SELECT COUNT(1) FROM organizations WHERE slug = $1`
	if err := r.db.GetContext(ctx, &count, query, slug); err != nil {
		return false, fmt.Errorf("failed to check slug existence: %w", err)
	}
	return count > 0, nil
}

func (r *repository) GetUserOrganizations(ctx context.Context, userID string) ([]OrganizationWithRole, error) {
	var orgs []OrganizationWithRole
	query := `
		SELECT o.id, o.name, o.slug, o.description, o.logo_url, o.created_at, o.updated_at, m.role AS user_role
		FROM organizations o
		JOIN organization_members m ON o.id = m.organization_id
		WHERE m.user_id = $1
		ORDER BY o.created_at DESC
	`
	if err := r.db.SelectContext(ctx, &orgs, query, userID); err != nil {
		return nil, fmt.Errorf("failed to list user organizations: %w", err)
	}
	return orgs, nil
}

func (r *repository) Update(ctx context.Context, org *Organization) error {
	query := `
		UPDATE organizations
		SET name = $1, description = $2, logo_url = $3, updated_at = NOW()
		WHERE id = $4
		RETURNING updated_at
	`
	err := r.db.QueryRowContext(ctx, query, org.Name, org.Description, org.LogoURL, org.ID).Scan(&org.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return apperrors.ErrNotFound
		}
		return fmt.Errorf("failed to update organization: %w", err)
	}
	return nil
}

func (r *repository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM organizations WHERE id = $1`
	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete organization: %w", err)
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

func (r *repository) GetMember(ctx context.Context, orgID, userID string) (*OrganizationMember, error) {
	var member OrganizationMember
	query := `
		SELECT id, organization_id, user_id, role, joined_at
		FROM organization_members
		WHERE organization_id = $1 AND user_id = $2
	`
	if err := r.db.GetContext(ctx, &member, query, orgID, userID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get member: %w", err)
	}
	return &member, nil
}

func (r *repository) GetMembers(ctx context.Context, orgID string) ([]MemberWithUser, error) {
	var members []MemberWithUser
	query := `
		SELECT m.id, m.organization_id, m.user_id, m.role, m.joined_at,
		       u.email, u.first_name, u.last_name, u.avatar_url
		FROM organization_members m
		JOIN users u ON m.user_id = u.id
		WHERE m.organization_id = $1
		ORDER BY m.joined_at ASC
	`
	if err := r.db.SelectContext(ctx, &members, query, orgID); err != nil {
		return nil, fmt.Errorf("failed to list organization members: %w", err)
	}
	return members, nil
}

func (r *repository) UpdateMemberRole(ctx context.Context, orgID, userID, newRole string) error {
	query := `
		UPDATE organization_members
		SET role = $1
		WHERE organization_id = $2 AND user_id = $3
	`
	res, err := r.db.ExecContext(ctx, query, newRole, orgID, userID)
	if err != nil {
		return fmt.Errorf("failed to update member role: %w", err)
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

func (r *repository) RemoveMember(ctx context.Context, orgID, userID string) error {
	query := `
		DELETE FROM organization_members
		WHERE organization_id = $1 AND user_id = $2
	`
	res, err := r.db.ExecContext(ctx, query, orgID, userID)
	if err != nil {
		return fmt.Errorf("failed to remove member: %w", err)
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

func (r *repository) CreateInvitation(ctx context.Context, invite *OrganizationInvitation) error {
	query := `
		INSERT INTO organization_invitations (organization_id, email, role, token, invited_by, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, status, created_at
	`
	err := r.db.QueryRowContext(
		ctx,
		query,
		invite.OrganizationID,
		invite.Email,
		invite.Role,
		invite.Token,
		invite.InvitedBy,
		invite.ExpiresAt,
	).Scan(&invite.ID, &invite.Status, &invite.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to insert invitation: %w", err)
	}
	return nil
}

func (r *repository) GetInvitationByToken(ctx context.Context, token string) (*OrganizationInvitation, error) {
	var invite OrganizationInvitation
	query := `
		SELECT id, organization_id, email, role, token, invited_by, status, expires_at, created_at
		FROM organization_invitations
		WHERE token = $1
	`
	if err := r.db.GetContext(ctx, &invite, query, token); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get invitation by token: %w", err)
	}
	return &invite, nil
}

func (r *repository) AcceptInvitationTx(ctx context.Context, inviteID, orgID, userID, role string) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Update invitation status
	queryInvite := `UPDATE organization_invitations SET status = 'accepted' WHERE id = $1`
	if _, err := tx.ExecContext(ctx, queryInvite, inviteID); err != nil {
		return fmt.Errorf("failed to update invitation: %w", err)
	}

	// Insert into organization_members
	queryMember := `
		INSERT INTO organization_members (organization_id, user_id, role)
		VALUES ($1, $2, $3)
		ON CONFLICT (organization_id, user_id) DO UPDATE SET role = EXCLUDED.role
	`
	if _, err := tx.ExecContext(ctx, queryMember, orgID, userID, role); err != nil {
		return fmt.Errorf("failed to insert accepted member: %w", err)
	}

	return tx.Commit()
}
