package projects

import (
	"github.com/gin-gonic/gin"
	"github.com/mobarok/devflow/internal/middleware"
)

func RegisterRoutes(
	rg *gin.RouterGroup,
	h *Handler,
	authMiddleware gin.HandlerFunc,
	checker middleware.MemberChecker,
) {
	// Organization-scoped project endpoints
	orgs := rg.Group("/organizations")
	orgs.Use(authMiddleware)
	{
		orgs.POST("/:id/projects", middleware.RequireOrgRole("admin", checker), h.Create)
		orgs.GET("/:id/projects", middleware.RequireOrgRole("viewer", checker), h.ListByOrg)
	}

	// Direct project endpoints
	projects := rg.Group("/projects")
	projects.Use(authMiddleware)
	{
		projects.GET("/:id", h.Get)
		projects.PATCH("/:id", h.Update)
		projects.DELETE("/:id", h.Delete)

		// Project member management
		projects.GET("/:id/members", h.GetMembers)
		projects.POST("/:id/members", h.AddMember)
		projects.PATCH("/:id/members/:user_id", h.UpdateMemberRole)
		projects.DELETE("/:id/members/:user_id", h.RemoveMember)
	}
}
