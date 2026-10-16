package middleware

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mobarok/devflow/pkg/response"
)

type MemberChecker interface {
	GetMember(ctx context.Context, orgID, userID string) (role string, err error)
}

// Role hierarchy levels for permission comparisons
var roleLevels = map[string]int{
	"viewer":    1,
	"developer": 2,
	"admin":     3,
	"owner":     4,
}

// RequireOrgRole ensures the authenticated user is an organization member with at least minRole permissions.
func RequireOrgRole(minRole string, checker MemberChecker) gin.HandlerFunc {
	minLevel, validRole := roleLevels[minRole]
	if !validRole {
		panic("invalid minRole configured in middleware: " + minRole)
	}

	return func(c *gin.Context) {
		userIDVal, exists := c.Get("user_id")
		if !exists {
			response.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "Unauthorized request")
			c.Abort()
			return
		}
		userID := userIDVal.(string)

		orgID := c.Param("org_id")
		if orgID == "" {
			orgID = c.Param("id")
		}
		if orgID == "" {
			orgID = c.GetHeader("X-Organization-ID")
		}

		if orgID == "" {
			response.Error(c, http.StatusBadRequest, "BAD_REQUEST", "Organization ID must be provided in URL or X-Organization-ID header")
			c.Abort()
			return
		}

		role, err := checker.GetMember(c.Request.Context(), orgID, userID)
		if err != nil {
			response.Error(c, http.StatusForbidden, "FORBIDDEN", "You are not a member of this organization")
			c.Abort()
			return
		}

		userLevel, ok := roleLevels[role]
		if !ok || userLevel < minLevel {
			response.Error(c, http.StatusForbidden, "FORBIDDEN", "Insufficient permissions for this action")
			c.Abort()
			return
		}

		// Store verified role and orgID in context for downstream handlers
		c.Set("org_role", role)
		c.Set("org_id", orgID)

		c.Next()
	}
}
