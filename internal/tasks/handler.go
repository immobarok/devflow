package tasks

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
// @Summary Create task in project
// @Description Creates a new task within a project (auto-generates project key e.g. DEV-1)
// @Tags Tasks
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Project ID"
// @Param request body CreateTaskRequest true "Task details"
// @Success 201 {object} response.SuccessResponse{data=TaskResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Router /projects/{id}/tasks [post]
func (h *Handler) Create(c *gin.Context) {
	projectID := c.Param("id")
	userID := c.GetString("user_id")

	var req CreateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleValidationError(c, err)
		return
	}

	task, err := h.service.Create(c.Request.Context(), projectID, userID, req)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			response.Error(c, http.StatusNotFound, "NOT_FOUND", "Project not found")
			return
		}
		if errors.Is(err, apperrors.ErrForbidden) {
			response.Error(c, http.StatusForbidden, "FORBIDDEN", err.Error())
			return
		}
		if errors.Is(err, apperrors.ErrValidation) {
			response.Error(c, http.StatusBadRequest, "VALIDATION_FAILED", err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create task")
		return
	}

	response.Success(c, http.StatusCreated, task)
}

// List godoc
// @Summary List and filter tasks
// @Description Retrieves tasks in a project with optional filters (status, priority, search) and pagination
// @Tags Tasks
// @Produce json
// @Security BearerAuth
// @Param id path string true "Project ID"
// @Param status query string false "Filter by status"
// @Param priority query string false "Filter by priority"
// @Param search query string false "Search keyword"
// @Param page query int false "Page number"
// @Param per_page query int false "Page size"
// @Success 200 {object} response.PaginatedResponse{data=[]TaskResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Router /projects/{id}/tasks [get]
func (h *Handler) List(c *gin.Context) {
	projectID := c.Param("id")

	var query TaskFilterQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Error(c, http.StatusBadRequest, "BAD_REQUEST", "Invalid query parameters")
		return
	}

	tasks, err := h.service.ListByProject(c.Request.Context(), projectID, query)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			response.Error(c, http.StatusNotFound, "NOT_FOUND", "Project not found")
			return
		}
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list tasks")
		return
	}

	response.Success(c, http.StatusOK, tasks)
}

// Get godoc
// @Summary Get task by ID
// @Description Retrieves task details including assignee and reporter info
// @Tags Tasks
// @Produce json
// @Security BearerAuth
// @Param id path string true "Task ID"
// @Success 200 {object} response.SuccessResponse{data=TaskResponse}
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Router /tasks/{id} [get]
func (h *Handler) Get(c *gin.Context) {
	taskID := c.Param("id")

	task, err := h.service.GetByID(c.Request.Context(), taskID)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			response.Error(c, http.StatusNotFound, "NOT_FOUND", "Task not found")
			return
		}
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to fetch task")
		return
	}

	response.Success(c, http.StatusOK, task)
}

// Update godoc
// @Summary Update task
// @Description Updates task status, priority, title, description, or assignee
// @Tags Tasks
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Task ID"
// @Param request body UpdateTaskRequest true "Updated task details"
// @Success 200 {object} response.SuccessResponse{data=TaskResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Router /tasks/{id} [patch]
func (h *Handler) Update(c *gin.Context) {
	taskID := c.Param("id")

	var req UpdateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleValidationError(c, err)
		return
	}

	task, err := h.service.Update(c.Request.Context(), taskID, req)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			response.Error(c, http.StatusNotFound, "NOT_FOUND", "Task not found")
			return
		}
		if errors.Is(err, apperrors.ErrValidation) {
			response.Error(c, http.StatusBadRequest, "VALIDATION_FAILED", err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update task")
		return
	}

	response.Success(c, http.StatusOK, task)
}

// Delete godoc
// @Summary Delete task
// @Description Permanently removes a task
// @Tags Tasks
// @Produce json
// @Security BearerAuth
// @Param id path string true "Task ID"
// @Success 200 {object} response.SuccessResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Router /tasks/{id} [delete]
func (h *Handler) Delete(c *gin.Context) {
	taskID := c.Param("id")

	if err := h.service.Delete(c.Request.Context(), taskID); err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			response.Error(c, http.StatusNotFound, "NOT_FOUND", "Task not found")
			return
		}
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to delete task")
		return
	}

	response.Success(c, http.StatusOK, gin.H{"message": "Task deleted successfully"})
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
