package tasks

import (
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(
	rg *gin.RouterGroup,
	h *Handler,
	authMiddleware gin.HandlerFunc,
) {
	// Project-scoped task endpoints
	projects := rg.Group("/projects")
	projects.Use(authMiddleware)
	{
		projects.POST("/:id/tasks", h.Create)
		projects.GET("/:id/tasks", h.List)
	}

	// Direct task endpoints
	tasks := rg.Group("/tasks")
	tasks.Use(authMiddleware)
	{
		tasks.GET("/:id", h.Get)
		tasks.PATCH("/:id", h.Update)
		tasks.DELETE("/:id", h.Delete)
	}
}
