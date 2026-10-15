package auth

import "github.com/gin-gonic/gin"

// RegisterRoutes registers authentication endpoints on the Gin router group.
func RegisterRoutes(rg *gin.RouterGroup, h *Handler, authMiddleware gin.HandlerFunc) {
	authGroup := rg.Group("/auth")
	{
		authGroup.POST("/register", h.Register)
		authGroup.POST("/login", h.Login)
		authGroup.GET("/me", authMiddleware, h.Me)
	}
}
