package notifications

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	apperrors "github.com/mobarok/devflow/pkg/errors"
	"github.com/mobarok/devflow/pkg/response"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// List godoc
// @Summary List notifications
// @Description Retrieves paginated notifications for the authenticated user
// @Tags Notifications
// @Produce json
// @Security BearerAuth
// @Param limit query int false "Pagination limit" default(20)
// @Param offset query int false "Pagination offset" default(0)
// @Success 200 {object} response.SuccessResponse{data=NotificationListResponse}
// @Failure 401 {object} response.ErrorResponse
// @Router /notifications [get]
func (h *Handler) List(c *gin.Context) {
	userID := c.GetString("user_id")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	res, err := h.service.List(c.Request.Context(), userID, limit, offset)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to fetch notifications")
		return
	}

	response.Success(c, http.StatusOK, res)
}

// GetUnreadCount godoc
// @Summary Get unread notification count
// @Description Fast count of unread notifications using partial PostgreSQL index
// @Tags Notifications
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.SuccessResponse{data=UnreadCountResponse}
// @Failure 401 {object} response.ErrorResponse
// @Router /notifications/unread-count [get]
func (h *Handler) GetUnreadCount(c *gin.Context) {
	userID := c.GetString("user_id")

	res, err := h.service.GetUnreadCount(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get unread notification count")
		return
	}

	response.Success(c, http.StatusOK, res)
}

// MarkAsRead godoc
// @Summary Mark notification as read
// @Description Marks an individual notification as read
// @Tags Notifications
// @Produce json
// @Security BearerAuth
// @Param id path string true "Notification ID"
// @Success 200 {object} response.SuccessResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Router /notifications/{id}/read [patch]
func (h *Handler) MarkAsRead(c *gin.Context) {
	notificationID := c.Param("id")
	userID := c.GetString("user_id")

	if err := h.service.MarkAsRead(c.Request.Context(), notificationID, userID); err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			response.Error(c, http.StatusNotFound, "NOT_FOUND", "Notification not found or already read")
			return
		}
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to mark notification as read")
		return
	}

	response.Success(c, http.StatusOK, gin.H{"message": "Notification marked as read"})
}

// MarkAllAsRead godoc
// @Summary Mark all notifications as read
// @Description Marks all pending unread notifications as read for the user
// @Tags Notifications
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.SuccessResponse
// @Failure 401 {object} response.ErrorResponse
// @Router /notifications/read-all [post]
func (h *Handler) MarkAllAsRead(c *gin.Context) {
	userID := c.GetString("user_id")

	if err := h.service.MarkAllAsRead(c.Request.Context(), userID); err != nil {
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to mark notifications as read")
		return
	}

	response.Success(c, http.StatusOK, gin.H{"message": "All notifications marked as read"})
}
