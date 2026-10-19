package activity

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/mobarok/devflow/pkg/response"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// ListByProject godoc
// @Summary List project activities
// @Description Retrieves activity audit log events for an entire project
// @Tags Activity
// @Produce json
// @Security BearerAuth
// @Param id path string true "Project ID"
// @Param limit query int false "Pagination limit" default(20)
// @Param offset query int false "Pagination offset" default(0)
// @Success 200 {object} response.SuccessResponse{data=ActivityListResponse}
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /projects/{id}/activities [get]
func (h *Handler) ListByProject(c *gin.Context) {
	projectID := c.Param("id")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	res, err := h.service.ListByProject(c.Request.Context(), projectID, limit, offset)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to fetch activities")
		return
	}

	response.Success(c, http.StatusOK, res)
}

// ListByTask godoc
// @Summary List task activities
// @Description Retrieves activity audit log events for a specific task
// @Tags Activity
// @Produce json
// @Security BearerAuth
// @Param id path string true "Task ID"
// @Param limit query int false "Pagination limit" default(20)
// @Param offset query int false "Pagination offset" default(0)
// @Success 200 {object} response.SuccessResponse{data=ActivityListResponse}
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /tasks/{id}/activities [get]
func (h *Handler) ListByTask(c *gin.Context) {
	taskID := c.Param("id")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	res, err := h.service.ListByTask(c.Request.Context(), taskID, limit, offset)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to fetch activities")
		return
	}

	response.Success(c, http.StatusOK, res)
}
