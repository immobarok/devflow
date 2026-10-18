package projects

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
	orgID := getOrgID(c)
	if orgID == "" {
		response.Error(c, http.StatusBadRequest, "BAD_REQUEST", "Organization ID is required")
		return
	}

	var req CreateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleValidationError(c, err)
		return
	}

	project, err := h.service.Create(c.Request.Context(), orgID, userID, req)
	if err != nil {
		if errors.Is(err, apperrors.ErrConflict) {
			response.Error(c, http.StatusConflict, "CONFLICT", err.Error())
			return
		}
		if errors.Is(err, apperrors.ErrForbidden) {
			response.Error(c, http.StatusForbidden, "FORBIDDEN", err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create project")
		return
	}

	response.Success(c, http.StatusCreated, project)
}

func (h *Handler) ListByOrg(c *gin.Context) {
	userID := c.GetString("user_id")
	orgID := getOrgID(c)
	if orgID == "" {
		response.Error(c, http.StatusBadRequest, "BAD_REQUEST", "Organization ID is required")
		return
	}

	projects, err := h.service.ListByOrganization(c.Request.Context(), orgID, userID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list projects")
		return
	}

	response.Success(c, http.StatusOK, projects)
}

func (h *Handler) Get(c *gin.Context) {
	userID := c.GetString("user_id")
	projectID := c.Param("id")

	project, err := h.service.GetByID(c.Request.Context(), projectID, userID)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			response.Error(c, http.StatusNotFound, "NOT_FOUND", "Project not found")
			return
		}
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to fetch project")
		return
	}

	response.Success(c, http.StatusOK, project)
}

func (h *Handler) Update(c *gin.Context) {
	projectID := c.Param("id")

	var req UpdateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleValidationError(c, err)
		return
	}

	project, err := h.service.Update(c.Request.Context(), projectID, req)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			response.Error(c, http.StatusNotFound, "NOT_FOUND", "Project not found")
			return
		}
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update project")
		return
	}

	response.Success(c, http.StatusOK, project)
}

func (h *Handler) Delete(c *gin.Context) {
	projectID := c.Param("id")

	if err := h.service.Delete(c.Request.Context(), projectID); err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			response.Error(c, http.StatusNotFound, "NOT_FOUND", "Project not found")
			return
		}
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to delete project")
		return
	}

	response.Success(c, http.StatusOK, gin.H{"message": "Project deleted successfully"})
}

func (h *Handler) AddMember(c *gin.Context) {
	projectID := c.Param("id")

	var req AssignMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleValidationError(c, err)
		return
	}

	if err := h.service.AddMember(c.Request.Context(), projectID, req); err != nil {
		if errors.Is(err, apperrors.ErrConflict) {
			response.Error(c, http.StatusConflict, "CONFLICT", "User is already a member of this project")
			return
		}
		if errors.Is(err, apperrors.ErrNotFound) {
			response.Error(c, http.StatusNotFound, "NOT_FOUND", "Project not found")
			return
		}
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to add project member")
		return
	}

	response.Success(c, http.StatusCreated, gin.H{"message": "Member added successfully"})
}

func (h *Handler) GetMembers(c *gin.Context) {
	projectID := c.Param("id")

	members, err := h.service.GetMembers(c.Request.Context(), projectID)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			response.Error(c, http.StatusNotFound, "NOT_FOUND", "Project not found")
			return
		}
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get project members")
		return
	}

	response.Success(c, http.StatusOK, members)
}

func (h *Handler) UpdateMemberRole(c *gin.Context) {
	projectID := c.Param("id")
	targetUserID := c.Param("user_id")

	var req UpdateMemberRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleValidationError(c, err)
		return
	}

	if err := h.service.UpdateMemberRole(c.Request.Context(), projectID, targetUserID, req); err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			response.Error(c, http.StatusNotFound, "NOT_FOUND", "Project member not found")
			return
		}
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update member role")
		return
	}

	response.Success(c, http.StatusOK, gin.H{"message": "Member role updated successfully"})
}

func (h *Handler) RemoveMember(c *gin.Context) {
	projectID := c.Param("id")
	targetUserID := c.Param("user_id")

	if err := h.service.RemoveMember(c.Request.Context(), projectID, targetUserID); err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			response.Error(c, http.StatusNotFound, "NOT_FOUND", "Project member not found")
			return
		}
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to remove member")
		return
	}

	response.Success(c, http.StatusOK, gin.H{"message": "Member removed successfully"})
}

func getOrgID(c *gin.Context) string {
	orgID := c.Param("org_id")
	if orgID == "" {
		orgID = c.Param("id")
	}
	if orgID == "" {
		orgID = c.GetHeader("X-Organization-ID")
	}
	return orgID
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
	case "min":
		return "must be at least " + fe.Param() + " characters"
	case "max":
		return "must be at most " + fe.Param() + " characters"
	case "oneof":
		return "must be one of: " + fe.Param()
	case "uuid":
		return "must be a valid UUID"
	default:
		return "failed validation on tag: " + fe.Tag()
	}
}