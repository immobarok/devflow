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

func (h *Handler) List(c *gin.Context) {
	userID := c.GetString("user_id")
	orgs, err := h.service.ListUserOrganizations(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list organizations")
		return
	}

	response.Success(c, http.StatusOK, orgs)
}

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

func (h *Handler) GetMembers(c *gin.Context) {
	orgID := c.Param("id")
	members, err := h.service.GetMembers(c.Request.Context(), orgID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to fetch members")
		return
	}

	response.Success(c, http.StatusOK, members)
}

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
