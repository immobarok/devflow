package auth

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	apperrors "github.com/mobarok/devflow/pkg/errors"
	"github.com/mobarok/devflow/pkg/response"
)

// Handler handles HTTP requests for authentication.
type Handler struct {
	service Service
}

// NewHandler creates a new auth HTTP handler.
func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// Register handles user registration (POST /api/v1/auth/register).
func (h *Handler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleValidationError(c, err)
		return
	}

	res, err := h.service.Register(c.Request.Context(), req)
	if err != nil {
		if errors.Is(err, apperrors.ErrConflict) {
			response.Error(c, http.StatusConflict, "CONFLICT", "User with this email already exists")
			return
		}
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to register user")
		return
	}

	response.Success(c, http.StatusCreated, res)
}

// Login handles user login and token issuance (POST /api/v1/auth/login).
func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleValidationError(c, err)
		return
	}

	res, err := h.service.Login(c.Request.Context(), req)
	if err != nil {
		if errors.Is(err, apperrors.ErrUnauthorized) {
			response.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid email or password")
			return
		}
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to process login")
		return
	}

	response.Success(c, http.StatusOK, res)
}

// Me retrieves profile information for the authenticated user (GET /api/v1/auth/me).
func (h *Handler) Me(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "Unauthorized request")
		return
	}

	user, err := h.service.GetMe(c.Request.Context(), userID.(string))
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			response.Error(c, http.StatusNotFound, "NOT_FOUND", "User not found")
			return
		}
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve profile")
		return
	}

	response.Success(c, http.StatusOK, gin.H{"user": user})
}

// ForgotPassword handles initiating a password reset (POST /api/v1/auth/forgot-password).
func (h *Handler) ForgotPassword(c *gin.Context) {
	var req ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleValidationError(c, err)
		return
	}

	otp, err := h.service.ForgotPassword(c.Request.Context(), req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to process request")
		return
	}

	data := gin.H{
		"message":    "If an account with that email exists, an OTP has been sent",
		"expires_in": 600,
	}
	// In development, also return the OTP in response for easy testing
	if otp != "" {
		data["dev_otp"] = otp
	}

	response.Success(c, http.StatusOK, data)
}

// VerifyOTP handles verifying the 6-digit OTP and issuing a reset token (POST /api/v1/auth/verify-otp).
func (h *Handler) VerifyOTP(c *gin.Context) {
	var req VerifyOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleValidationError(c, err)
		return
	}

	res, err := h.service.VerifyOTP(c.Request.Context(), req)
	if err != nil {
		if errors.Is(err, apperrors.ErrUnauthorized) {
			response.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid or expired OTP")
			return
		}
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to verify OTP")
		return
	}

	response.Success(c, http.StatusOK, res)
}

// ResetPassword handles updating the password using a verified reset token (POST /api/v1/auth/reset-password).
func (h *Handler) ResetPassword(c *gin.Context) {
	var req ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleValidationError(c, err)
		return
	}

	err := h.service.ResetPassword(c.Request.Context(), req)
	if err != nil {
		if errors.Is(err, apperrors.ErrUnauthorized) {
			response.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid or expired reset token")
			return
		}
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to reset password")
		return
	}

	response.Success(c, http.StatusOK, gin.H{
		"message": "Password has been reset successfully. Please log in with your new password.",
	})
}

// handleValidationError formats binding errors into clean API error responses.
func (h *Handler) handleValidationError(c *gin.Context, err error) {
	if valErrors, ok := errors.AsType[validator.ValidationErrors](err); ok {
		fieldErrors := make([]response.FieldError, 0, len(valErrors))
		for _, fe := range valErrors {
			fieldErrors = append(fieldErrors, response.FieldError{
				Field:   fe.Field(),
				Message: formatValidationTag(fe),
			})
		}
		response.ValidationError(c, fieldErrors)
		return
	}

	response.Error(c, http.StatusBadRequest, "BAD_REQUEST", "Invalid JSON request body")
}

func formatValidationTag(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "this field is required"
	case "email":
		return "must be a valid email address"
	case "len":
		return "must be exactly " + fe.Param() + " characters"
	case "min":
		return "must be at least " + fe.Param() + " characters"
	default:
		return "failed validation on tag: " + fe.Tag()
	}
}
