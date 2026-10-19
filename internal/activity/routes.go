package activity

import (
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(
	rg *gin.RouterGroup,
	h *Handler,
	authMiddleware gin.HandlerFunc,
) {
	projects := rg.Group("/projects")
	projects.Use(authMiddleware)
	{
		projects.GET("/:id/activities", h.ListByProject)
	}

	tasks := rg.Group("/tasks")
	tasks.Use(authMiddleware)
	{
		tasks.GET("/:id/activities", h.ListByTask)
	}
}
