package comments

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
	taskID := c.Param("id")
	userID := c.GetString("user_id")

	var req CreateCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleValidationError(c, err)
		return
	}

	comment, err := h.service.Create(c.Request.Context(), taskID, userID, req)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			response.Error(c, http.StatusNotFound, "NOT_FOUND", "Task not found")
			return
		}
		if errors.Is(err, apperrors.ErrValidation) {
			response.Error(c, http.StatusBadRequest, "VALIDATION_FAILED", err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create comment")
		return
	}

	response.Success(c, http.StatusCreated, comment)
}

func (h *Handler) List(c *gin.Context) {
	taskID := c.Param("id")

	comments, err := h.service.ListByTask(c.Request.Context(), taskID)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			response.Error(c, http.StatusNotFound, "NOT_FOUND", "Task not found")
			return
		}
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list comments")
		return
	}

	response.Success(c, http.StatusOK, comments)
}

func (h *Handler) Update(c *gin.Context) {
	commentID := c.Param("id")
	userID := c.GetString("user_id")

	var req UpdateCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleValidationError(c, err)
		return
	}

	comment, err := h.service.Update(c.Request.Context(), commentID, userID, req)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			response.Error(c, http.StatusNotFound, "NOT_FOUND", "Comment not found")
			return
		}
		if errors.Is(err, apperrors.ErrForbidden) {
			response.Error(c, http.StatusForbidden, "FORBIDDEN", "You can only edit your own comments")
			return
		}
		if errors.Is(err, apperrors.ErrValidation) {
			response.Error(c, http.StatusBadRequest, "VALIDATION_FAILED", err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update comment")
		return
	}

	response.Success(c, http.StatusOK, comment)
}

func (h *Handler) Delete(c *gin.Context) {
	commentID := c.Param("id")
	userID := c.GetString("user_id")

	if err := h.service.Delete(c.Request.Context(), commentID, userID); err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			response.Error(c, http.StatusNotFound, "NOT_FOUND", "Comment not found")
			return
		}
		if errors.Is(err, apperrors.ErrForbidden) {
			response.Error(c, http.StatusForbidden, "FORBIDDEN", "You can only delete your own comments")
			return
		}
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to delete comment")
		return
	}

	response.Success(c, http.StatusOK, gin.H{"message": "Comment deleted successfully"})
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
	default:
		return "failed validation on tag: " + fe.Tag()
	}
}
