package auth

import "github.com/gin-gonic/gin"

func RegisterRoutes(rg *gin.RouterGroup, h *Handler, authMiddleware gin.HandlerFunc) {
	authGroup := rg.Group("/auth")
	{
		authGroup.POST("/register", h.Register)
		authGroup.POST("/login", h.Login)
		authGroup.POST("/refresh", h.RefreshToken)
		authGroup.POST("/logout", h.Logout)
		authGroup.POST("/forgot-password", h.ForgotPassword)
		authGroup.POST("/verify-otp", h.VerifyOTP)
		authGroup.POST("/reset-password", h.ResetPassword)
		authGroup.GET("/me", authMiddleware, h.Me)

		// OAuth routes
		authGroup.GET("/google", h.OAuthLogin("google"))
		authGroup.GET("/google/callback", h.OAuthCallback("google"))
		authGroup.GET("/github", h.OAuthLogin("github"))
		authGroup.GET("/github/callback", h.OAuthCallback("github"))
	}
}
