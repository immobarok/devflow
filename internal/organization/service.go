package organization

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"
	"time"

	apperrors "github.com/mobarok/devflow/pkg/errors"
)

type Service interface {
	Create(ctx context.Context, userID string, req CreateOrganizationRequest) (*OrganizationResponse, error)
	GetByID(ctx context.Context, orgID, userID string) (*OrganizationResponse, error)
	GetBySlug(ctx context.Context, slug, userID string) (*OrganizationResponse, error)
	ListUserOrganizations(ctx context.Context, userID string) ([]OrganizationResponse, error)
	Update(ctx context.Context, orgID, userID string, req UpdateOrganizationRequest) (*OrganizationResponse, error)
	Delete(ctx context.Context, orgID, userID string) error
	GetMembers(ctx context.Context, orgID string) ([]MemberResponse, error)
	InviteMember(ctx context.Context, orgID, inviterID string, req InviteMemberRequest) (*InvitationResponse, error)
	UpdateMemberRole(ctx context.Context, orgID, operatorID, targetUserID string, req UpdateMemberRoleRequest) error
	RemoveMember(ctx context.Context, orgID, operatorID, targetUserID string) error
	AcceptInvitation(ctx context.Context, token, userID string) (*OrganizationResponse, error)
	GetMember(ctx context.Context, orgID, userID string) (string, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) GetMember(ctx context.Context, orgID, userID string) (string, error) {
	m, err := s.repo.GetMember(ctx, orgID, userID)
	if err != nil {
		return "", err
	}
	return m.Role, nil
}

func (s *service) Create(ctx context.Context, userID string, req CreateOrganizationRequest) (*OrganizationResponse, error) {
	slug, err := s.generateUniqueSlug(ctx, req.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to generate unique slug: %w", err)
	}

	org := &Organization{
		Name:        req.Name,
		Slug:        slug,
		Description: req.Description,
		LogoURL:     req.LogoURL,
	}

	if err := s.repo.CreateWithTx(ctx, org, userID); err != nil {
		return nil, fmt.Errorf("failed to create organization: %w", err)
	}

	return ToOrganizationResponse(org, "owner"), nil
}

func (s *service) GetByID(ctx context.Context, orgID, userID string) (*OrganizationResponse, error) {
	org, err := s.repo.GetByID(ctx, orgID)
	if err != nil {
		return nil, err
	}

	role := ""
	if member, err := s.repo.GetMember(ctx, orgID, userID); err == nil && member != nil {
		role = member.Role
	}

	return ToOrganizationResponse(org, role), nil
}

func (s *service) GetBySlug(ctx context.Context, slug, userID string) (*OrganizationResponse, error) {
	org, err := s.repo.GetBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}

	role := ""
	if member, err := s.repo.GetMember(ctx, org.ID, userID); err == nil && member != nil {
		role = member.Role
	}

	return ToOrganizationResponse(org, role), nil
}

func (s *service) ListUserOrganizations(ctx context.Context, userID string) ([]OrganizationResponse, error) {
	orgsWithRole, err := s.repo.GetUserOrganizations(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list organizations: %w", err)
	}

	responses := make([]OrganizationResponse, 0, len(orgsWithRole))
	for _, o := range orgsWithRole {
		resp := ToOrganizationResponse(&o.Organization, o.UserRole)
		responses = append(responses, *resp)
	}

	return responses, nil
}

func (s *service) Update(ctx context.Context, orgID, userID string, req UpdateOrganizationRequest) (*OrganizationResponse, error) {
	org, err := s.repo.GetByID(ctx, orgID)
	if err != nil {
		return nil, err
	}

	if req.Name != nil {
		org.Name = *req.Name
	}
	if req.Description != nil {
		org.Description = req.Description
	}
	if req.LogoURL != nil {
		org.LogoURL = req.LogoURL
	}

	if err := s.repo.Update(ctx, org); err != nil {
		return nil, fmt.Errorf("failed to update organization: %w", err)
	}

	role := ""
	if member, err := s.repo.GetMember(ctx, orgID, userID); err == nil && member != nil {
		role = member.Role
	}

	return ToOrganizationResponse(org, role), nil
}

func (s *service) Delete(ctx context.Context, orgID, userID string) error {
	member, err := s.repo.GetMember(ctx, orgID, userID)
	if err != nil {
		return apperrors.ErrForbidden
	}
	if member.Role != "owner" {
		return apperrors.ErrForbidden
	}

	return s.repo.Delete(ctx, orgID)
}

func (s *service) GetMembers(ctx context.Context, orgID string) ([]MemberResponse, error) {
	members, err := s.repo.GetMembers(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch members: %w", err)
	}

	result := make([]MemberResponse, 0, len(members))
	for _, m := range members {
		result = append(result, *ToMemberResponse(&m))
	}

	return result, nil
}

func (s *service) InviteMember(ctx context.Context, orgID, inviterID string, req InviteMemberRequest) (*InvitationResponse, error) {
	// Generate random 32-byte hex token
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return nil, fmt.Errorf("failed to generate invitation token: %w", err)
	}
	token := hex.EncodeToString(b)

	invite := &OrganizationInvitation{
		OrganizationID: orgID,
		Email:          req.Email,
		Role:           req.Role,
		Token:          token,
		InvitedBy:      &inviterID,
		ExpiresAt:      time.Now().Add(7 * 24 * time.Hour), // 7-day validity
	}

	if err := s.repo.CreateInvitation(ctx, invite); err != nil {
		return nil, fmt.Errorf("failed to create invitation: %w", err)
	}

	return &InvitationResponse{
		ID:             invite.ID,
		OrganizationID: invite.OrganizationID,
		Email:          invite.Email,
		Role:           invite.Role,
		Token:          invite.Token,
		Status:         invite.Status,
		ExpiresAt:      invite.ExpiresAt,
		CreatedAt:      invite.CreatedAt,
	}, nil
}

func (s *service) UpdateMemberRole(ctx context.Context, orgID, operatorID, targetUserID string, req UpdateMemberRoleRequest) error {
	operator, err := s.repo.GetMember(ctx, orgID, operatorID)
	if err != nil {
		return apperrors.ErrForbidden
	}

	target, err := s.repo.GetMember(ctx, orgID, targetUserID)
	if err != nil {
		return err
	}

	// Cannot change owner role directly (requires ownership transfer)
	if target.Role == "owner" {
		return apperrors.ErrForbidden
	}

	// Admins cannot alter other admins or owners
	if operator.Role == "admin" && (target.Role == "admin" || req.Role == "admin") {
		return apperrors.ErrForbidden
	}

	return s.repo.UpdateMemberRole(ctx, orgID, targetUserID, req.Role)
}

func (s *service) RemoveMember(ctx context.Context, orgID, operatorID, targetUserID string) error {
	operator, err := s.repo.GetMember(ctx, orgID, operatorID)
	if err != nil {
		return apperrors.ErrForbidden
	}

	target, err := s.repo.GetMember(ctx, orgID, targetUserID)
	if err != nil {
		return err
	}

	// Owner cannot be removed
	if target.Role == "owner" {
		return apperrors.ErrForbidden
	}

	// Self-removal is allowed unless user is owner
	if operatorID == targetUserID {
		return s.repo.RemoveMember(ctx, orgID, targetUserID)
	}

	// Admins cannot remove other admins
	if operator.Role == "admin" && target.Role == "admin" {
		return apperrors.ErrForbidden
	}

	return s.repo.RemoveMember(ctx, orgID, targetUserID)
}

func (s *service) AcceptInvitation(ctx context.Context, token, userID string) (*OrganizationResponse, error) {
	invite, err := s.repo.GetInvitationByToken(ctx, token)
	if err != nil {
		return nil, err
	}

	if invite.Status != "pending" || time.Now().After(invite.ExpiresAt) {
		return nil, apperrors.ErrValidation
	}

	if err := s.repo.AcceptInvitationTx(ctx, invite.ID, invite.OrganizationID, userID, invite.Role); err != nil {
		return nil, fmt.Errorf("failed to accept invitation: %w", err)
	}

	return s.GetByID(ctx, invite.OrganizationID, userID)
}

var nonAlphanumericRegex = regexp.MustCompile(`[^a-z0-9]+`)

func (s *service) generateUniqueSlug(ctx context.Context, name string) (string, error) {
	baseSlug := strings.ToLower(name)
	baseSlug = nonAlphanumericRegex.ReplaceAllString(baseSlug, "-")
	baseSlug = strings.Trim(baseSlug, "-")
	if baseSlug == "" {
		baseSlug = "workspace"
	}

	slug := baseSlug
	counter := 1
	for {
		exists, err := s.repo.SlugExists(ctx, slug)
		if err != nil {
			return "", err
		}
		if !exists {
			return slug, nil
		}
		slug = fmt.Sprintf("%s-%d", baseSlug, counter)
		counter++
	}
}
