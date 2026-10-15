package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	jwtpkg "github.com/mobarok/devflow/pkg/jwt"
	"github.com/mobarok/devflow/pkg/response"
)

// AuthRequired returns a Gin middleware that validates the JWT in the Authorization header.
func AuthRequired(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "Missing authorization header")
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			response.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid authorization header format. Expected 'Bearer <token>'")
			c.Abort()
			return
		}

		tokenString := parts[1]
		claims, err := jwtpkg.ValidateToken(tokenString, jwtSecret)
		if err != nil {
			response.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid or expired token")
			c.Abort()
			return
		}

		// Store user context in Gin context for downstream handlers
		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)

		c.Next()
	}
}
