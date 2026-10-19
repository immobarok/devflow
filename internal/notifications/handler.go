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

func (h *Handler) GetUnreadCount(c *gin.Context) {
	userID := c.GetString("user_id")

	res, err := h.service.GetUnreadCount(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get unread notification count")
		return
	}

	response.Success(c, http.StatusOK, res)
}

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

func (h *Handler) MarkAllAsRead(c *gin.Context) {
	userID := c.GetString("user_id")

	if err := h.service.MarkAllAsRead(c.Request.Context(), userID); err != nil {
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to mark notifications as read")
		return
	}

	response.Success(c, http.StatusOK, gin.H{"message": "All notifications marked as read"})
}
