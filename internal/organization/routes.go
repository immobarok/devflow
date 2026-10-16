package organization

import (
	"github.com/gin-gonic/gin"
	"github.com/mobarok/devflow/internal/middleware"
)

func RegisterRoutes(rg *gin.RouterGroup, h *Handler, authMiddleware gin.HandlerFunc, checker middleware.MemberChecker) {
	orgs := rg.Group("/organizations")
	orgs.Use(authMiddleware)
	{
		orgs.POST("", h.Create)
		orgs.GET("", h.List)
		orgs.GET("/slug/:slug", h.GetBySlug)

		// Scoped to specific organization with RBAC
		orgs.GET("/:id", middleware.RequireOrgRole("viewer", checker), h.Get)
		orgs.PATCH("/:id", middleware.RequireOrgRole("admin", checker), h.Update)
		orgs.DELETE("/:id", middleware.RequireOrgRole("owner", checker), h.Delete)

		// Membership routes
		orgs.GET("/:id/members", middleware.RequireOrgRole("viewer", checker), h.GetMembers)
		orgs.POST("/:id/invitations", middleware.RequireOrgRole("admin", checker), h.InviteMember)
		orgs.PATCH("/:id/members/:member_id", middleware.RequireOrgRole("admin", checker), h.UpdateMemberRole)
		orgs.DELETE("/:id/members/:member_id", middleware.RequireOrgRole("admin", checker), h.RemoveMember)
	}

	invites := rg.Group("/invitations")
	invites.Use(authMiddleware)
	{
		invites.POST("/accept", h.AcceptInvitation)
	}
}
