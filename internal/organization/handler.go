package organization

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	apperrors "github.com/mobarok/devflow/pkg/errors"
	"github.com/mobarok/devflow/pkg/response"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// Create godoc
// @Summary Create organization
// @Description Creates a new organization workspace; the creator becomes the Owner
// @Tags Organizations
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreateOrganizationRequest true "Organization details"
// @Success 201 {object} response.SuccessResponse{data=OrganizationResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Router /organizations [post]
func (h *Handler) Create(c *gin.Context) {
	userID := c.GetString("user_id")
	var req CreateOrganizationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleValidationError(c, err)
		return
	}

	org, err := h.service.Create(c.Request.Context(), userID, req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create organization")
		return
	}

	response.Success(c, http.StatusCreated, org)
}

// List godoc
// @Summary List user organizations
// @Description Lists all organizations where the authenticated user is a member
// @Tags Organizations
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.SuccessResponse{data=[]OrganizationResponse}
// @Failure 401 {object} response.ErrorResponse
// @Router /organizations [get]
func (h *Handler) List(c *gin.Context) {
	userID := c.GetString("user_id")
	orgs, err := h.service.ListUserOrganizations(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list organizations")
		return
	}

	response.Success(c, http.StatusOK, orgs)
}

// Get godoc
// @Summary Get organization by ID
// @Description Retrieves details of a specific organization by ID
// @Tags Organizations
// @Produce json
// @Security BearerAuth
// @Param id path string true "Organization ID"
// @Success 200 {object} response.SuccessResponse{data=OrganizationResponse}
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Router /organizations/{id} [get]
func (h *Handler) Get(c *gin.Context) {
	userID := c.GetString("user_id")
	orgID := c.Param("id")

	org, err := h.service.GetByID(c.Request.Context(), orgID, userID)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			response.Error(c, http.StatusNotFound, "NOT_FOUND", "Organization not found")
			return
		}
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to fetch organization")
		return
	}

	response.Success(c, http.StatusOK, org)
}

// GetBySlug godoc
// @Summary Get organization by slug
// @Description Retrieves details of a specific organization by its URL slug
// @Tags Organizations
// @Produce json
// @Security BearerAuth
// @Param slug path string true "Organization slug"
// @Success 200 {object} response.SuccessResponse{data=OrganizationResponse}
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Router /organizations/slug/{slug} [get]
func (h *Handler) GetBySlug(c *gin.Context) {
	userID := c.GetString("user_id")
	slug := c.Param("slug")

	org, err := h.service.GetBySlug(c.Request.Context(), slug, userID)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			response.Error(c, http.StatusNotFound, "NOT_FOUND", "Organization not found")
			return
		}
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to fetch organization")
		return
	}

	response.Success(c, http.StatusOK, org)
}

// Update godoc
// @Summary Update organization
// @Description Updates organization profile (name, logo, etc.)
// @Tags Organizations
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Organization ID"
// @Param request body UpdateOrganizationRequest true "Updated details"
// @Success 200 {object} response.SuccessResponse{data=OrganizationResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Router /organizations/{id} [patch]
func (h *Handler) Update(c *gin.Context) {
	userID := c.GetString("user_id")
	orgID := c.Param("id")

	var req UpdateOrganizationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleValidationError(c, err)
		return
	}

	org, err := h.service.Update(c.Request.Context(), orgID, userID, req)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			response.Error(c, http.StatusNotFound, "NOT_FOUND", "Organization not found")
			return
		}
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update organization")
		return
	}

	response.Success(c, http.StatusOK, org)
}

// Delete godoc
// @Summary Delete organization
// @Description Permanently deletes an organization (owner only)
// @Tags Organizations
// @Produce json
// @Security BearerAuth
// @Param id path string true "Organization ID"
// @Success 200 {object} response.SuccessResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Router /organizations/{id} [delete]
func (h *Handler) Delete(c *gin.Context) {
	userID := c.GetString("user_id")
	orgID := c.Param("id")

	if err := h.service.Delete(c.Request.Context(), orgID, userID); err != nil {
		if errors.Is(err, apperrors.ErrForbidden) {
			response.Error(c, http.StatusForbidden, "FORBIDDEN", "Only organization owner can delete this organization")
			return
		}
		if errors.Is(err, apperrors.ErrNotFound) {
			response.Error(c, http.StatusNotFound, "NOT_FOUND", "Organization not found")
			return
		}
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to delete organization")
		return
	}

	response.Success(c, http.StatusOK, gin.H{"message": "Organization deleted successfully"})
}

// GetMembers godoc
// @Summary List organization members
// @Description Retrieves the list of members and their roles in the organization
// @Tags Organizations
// @Produce json
// @Security BearerAuth
// @Param id path string true "Organization ID"
// @Success 200 {object} response.SuccessResponse{data=[]MemberResponse}
// @Failure 401 {object} response.ErrorResponse
// @Router /organizations/{id}/members [get]
func (h *Handler) GetMembers(c *gin.Context) {
	orgID := c.Param("id")
	members, err := h.service.GetMembers(c.Request.Context(), orgID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to fetch members")
		return
	}

	response.Success(c, http.StatusOK, members)
}

// InviteMember godoc
// @Summary Invite member to organization
// @Description Sends an invitation with a specified role to join the organization
// @Tags Organizations
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Organization ID"
// @Param request body InviteMemberRequest true "Invitation details"
// @Success 201 {object} response.SuccessResponse{data=InvitationResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Router /organizations/{id}/invitations [post]
func (h *Handler) InviteMember(c *gin.Context) {
	userID := c.GetString("user_id")
	orgID := c.Param("id")

	var req InviteMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleValidationError(c, err)
		return
	}

	invite, err := h.service.InviteMember(c.Request.Context(), orgID, userID, req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create invitation")
		return
	}

	response.Success(c, http.StatusCreated, invite)
}

// UpdateMemberRole godoc
// @Summary Update member role in organization
// @Description Updates the role of an existing member in the organization
// @Tags Organizations
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Organization ID"
// @Param member_id path string true "Member User ID"
// @Param request body UpdateMemberRoleRequest true "New role details"
// @Success 200 {object} response.SuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Router /organizations/{id}/members/{member_id} [patch]
func (h *Handler) UpdateMemberRole(c *gin.Context) {
	operatorID := c.GetString("user_id")
	orgID := c.Param("id")
	targetUserID := c.Param("member_id")

	var req UpdateMemberRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleValidationError(c, err)
		return
	}

	if err := h.service.UpdateMemberRole(c.Request.Context(), orgID, operatorID, targetUserID, req); err != nil {
		if errors.Is(err, apperrors.ErrForbidden) {
			response.Error(c, http.StatusForbidden, "FORBIDDEN", "You do not have permission to alter this member's role")
			return
		}
		if errors.Is(err, apperrors.ErrNotFound) {
			response.Error(c, http.StatusNotFound, "NOT_FOUND", "Member not found in organization")
			return
		}
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update member role")
		return
	}

	response.Success(c, http.StatusOK, gin.H{"message": "Member role updated successfully"})
}

// RemoveMember godoc
// @Summary Remove member from organization
// @Description Removes a user from the organization
// @Tags Organizations
// @Produce json
// @Security BearerAuth
// @Param id path string true "Organization ID"
// @Param member_id path string true "Member User ID"
// @Success 200 {object} response.SuccessResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Router /organizations/{id}/members/{member_id} [delete]
func (h *Handler) RemoveMember(c *gin.Context) {
	operatorID := c.GetString("user_id")
	orgID := c.Param("id")
	targetUserID := c.Param("member_id")

	if err := h.service.RemoveMember(c.Request.Context(), orgID, operatorID, targetUserID); err != nil {
		if errors.Is(err, apperrors.ErrForbidden) {
			response.Error(c, http.StatusForbidden, "FORBIDDEN", "Cannot remove this member from organization")
			return
		}
		if errors.Is(err, apperrors.ErrNotFound) {
			response.Error(c, http.StatusNotFound, "NOT_FOUND", "Member not found in organization")
			return
		}
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to remove member")
		return
	}

	response.Success(c, http.StatusOK, gin.H{"message": "Member removed successfully"})
}

// AcceptInvitation godoc
// @Summary Accept organization invitation
// @Description Accepts an invitation token and adds the user to the organization
// @Tags Organizations
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body AcceptInvitationRequest true "Acceptance token"
// @Success 200 {object} response.SuccessResponse{data=OrganizationResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Router /invitations/accept [post]
func (h *Handler) AcceptInvitation(c *gin.Context) {
	userID := c.GetString("user_id")

	var req AcceptInvitationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleValidationError(c, err)
		return
	}

	org, err := h.service.AcceptInvitation(c.Request.Context(), req.Token, userID)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			response.Error(c, http.StatusNotFound, "NOT_FOUND", "Invitation not found")
			return
		}
		if errors.Is(err, apperrors.ErrValidation) {
			response.Error(c, http.StatusBadRequest, "BAD_REQUEST", "Invitation is expired or already accepted")
			return
		}
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to accept invitation")
		return
	}

	response.Success(c, http.StatusOK, org)
}

func (h *Handler) handleValidationError(c *gin.Context, err error) {
	if valErrors, ok := errors.AsType[validator.ValidationErrors](err); ok {
		fieldErrors := make([]response.FieldError, 0, len(valErrors))
		for _, fe := range valErrors {
			fieldErrors = append(fieldErrors, response.FieldError{
				Field:   fe.Field(),
				Message: formatValidationTag(fe),
			})
		}
		response.ValidationError(c, fieldErrors)
		return
	}

	response.Error(c, http.StatusBadRequest, "BAD_REQUEST", "Invalid JSON request body")
}

func formatValidationTag(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "this field is required"
	case "email":
		return "must be a valid email address"
	case "min":
		return "must be at least " + fe.Param() + " characters"
	case "max":
		return "must be at most " + fe.Param() + " characters"
	case "oneof":
		return "must be one of: " + fe.Param()
	default:
		return "failed validation on tag: " + fe.Tag()
	}
}
