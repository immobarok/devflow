package comments

import (
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(
	rg *gin.RouterGroup,
	h *Handler,
	authMiddleware gin.HandlerFunc,
) {
	// Task-scoped comment endpoints
	tasks := rg.Group("/tasks")
	tasks.Use(authMiddleware)
	{
		tasks.POST("/:id/comments", h.Create)
		tasks.GET("/:id/comments", h.List)
	}

	// Direct comment endpoints
	comments := rg.Group("/comments")
	comments.Use(authMiddleware)
	{
		comments.PATCH("/:id", h.Update)
		comments.DELETE("/:id", h.Delete)
	}
}
