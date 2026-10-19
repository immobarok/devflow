package notifications

import (
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(
	rg *gin.RouterGroup,
	h *Handler,
	authMiddleware gin.HandlerFunc,
) {
	notifs := rg.Group("/notifications")
	notifs.Use(authMiddleware)
	{
		notifs.GET("", h.List)
		notifs.GET("/unread-count", h.GetUnreadCount)
		notifs.PATCH("/:id/read", h.MarkAsRead)
		notifs.POST("/read-all", h.MarkAllAsRead)
	}
}
