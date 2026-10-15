package auth

import "github.com/gin-gonic/gin"

func RegisterRoutes(rg *gin.RouterGroup, h *Handler, authMiddleware gin.HandlerFunc) {
	authGroup := rg.Group("/auth")
	{
		authGroup.POST("/register", h.Register)
		authGroup.POST("/login", h.Login)
		authGroup.POST("/forgot-password", h.ForgotPassword)
		authGroup.POST("/verify-otp", h.VerifyOTP)
		authGroup.POST("/reset-password", h.ResetPassword)
		authGroup.GET("/me", authMiddleware, h.Me)
	}
}
