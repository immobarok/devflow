package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"github.com/mobarok/devflow/pkg/config"
	apperrors "github.com/mobarok/devflow/pkg/errors"
	"github.com/mobarok/devflow/pkg/response"
)

// Handler handles HTTP requests for authentication.
type Handler struct {
	service Service
	cfg     *config.Config
}

// NewHandler creates a new auth HTTP handler.
func NewHandler(service Service, cfg *config.Config) *Handler {
	return &Handler{
		service: service,
		cfg:     cfg,
	}
}

// Register handles user registration (POST /api/v1/auth/register).
// @Summary Register a new user
// @Description Creates a new user account and returns JWT access and refresh tokens
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body RegisterRequest true "User registration details"
// @Success 201 {object} response.SuccessResponse{data=AuthResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Router /auth/register [post]
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
// @Summary Login user
// @Description Authenticates email and password credentials and returns JWT access and refresh tokens
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Login credentials"
// @Success 200 {object} response.SuccessResponse{data=AuthResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Router /auth/login [post]
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
// @Summary Current user profile
// @Description Fetches detailed profile of the currently authenticated user
// @Tags Auth
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.SuccessResponse{data=UserResponse}
// @Failure 401 {object} response.ErrorResponse
// @Router /auth/me [get]
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
// @Summary Forgot password
// @Description Initiates password reset by issuing a 6-digit OTP to the user's email
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body ForgotPasswordRequest true "Account email"
// @Success 200 {object} response.SuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Router /auth/forgot-password [post]
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
// @Summary Verify OTP
// @Description Verifies 6-digit OTP and issues a short-lived reset token
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body VerifyOTPRequest true "Email and 6-digit OTP"
// @Success 200 {object} response.SuccessResponse{data=VerifyOTPResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Router /auth/verify-otp [post]
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
// @Summary Reset password
// @Description Sets a new password using a verified reset token
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body ResetPasswordRequest true "Verified reset token and new password"
// @Success 200 {object} response.SuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Router /auth/reset-password [post]
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

// RefreshToken handles issuing new access and refresh tokens (POST /api/v1/auth/refresh).
// @Summary Refresh tokens
// @Description Rotates and issues fresh access and refresh tokens
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body RefreshTokenRequest true "Valid refresh token"
// @Success 200 {object} response.SuccessResponse{data=RefreshTokenResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Router /auth/refresh [post]
func (h *Handler) RefreshToken(c *gin.Context) {
	var req RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleValidationError(c, err)
		return
	}

	res, err := h.service.RefreshToken(c.Request.Context(), req)
	if err != nil {
		if errors.Is(err, apperrors.ErrUnauthorized) {
			response.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid, expired, or revoked refresh token")
			return
		}
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to refresh token")
		return
	}

	response.Success(c, http.StatusOK, res)
}

// Logout handles terminating a session by revoking the refresh token (POST /api/v1/auth/logout).
// @Summary Logout
// @Description Revokes refresh token to terminate session
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body LogoutRequest true "Refresh token to revoke"
// @Success 200 {object} response.SuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Router /auth/logout [post]
func (h *Handler) Logout(c *gin.Context) {
	var req LogoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleValidationError(c, err)
		return
	}

	_ = h.service.Logout(c.Request.Context(), req)
	response.Success(c, http.StatusOK, gin.H{
		"message": "Logged out successfully",
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

// OAuthLogin initiates the OAuth authorization flow for a specific provider.
func (h *Handler) OAuthLogin(provider string) gin.HandlerFunc {
	return func(c *gin.Context) {
		state, err := generateRandomHex(16)
		if err != nil {
			response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to generate security state")
			return
		}

		// Store state in an HTTP-only cookie for CSRF verification (5 minutes expiry)
		c.SetCookie("oauth_state", state, 300, "/", "", false, true)

		authURL, err := h.service.GetOAuthAuthURL(provider, state)
		if err != nil {
			response.Error(c, http.StatusBadRequest, "BAD_REQUEST", err.Error())
			return
		}

		c.Redirect(http.StatusTemporaryRedirect, authURL)
	}
}

// OAuthCallback processes the callback from an OAuth provider with the authorization code.
func (h *Handler) OAuthCallback(provider string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if oauthErr := c.Query("error"); oauthErr != "" {
			errDesc := c.DefaultQuery("error_description", oauthErr)
			h.handleOAuthError(c, fmt.Sprintf("OAuth error: %s", errDesc))
			return
		}

		cookieState, _ := c.Cookie("oauth_state")
		queryState := c.Query("state")
		c.SetCookie("oauth_state", "", -1, "/", "", false, true)

		if cookieState != "" && queryState != cookieState {
			h.handleOAuthError(c, "Invalid or expired OAuth state")
			return
		}

		code := c.Query("code")
		if code == "" {
			h.handleOAuthError(c, "Missing authorization code")
			return
		}

		authResp, err := h.service.HandleOAuthCallback(c.Request.Context(), provider, code)
		if err != nil {
			h.handleOAuthError(c, fmt.Sprintf("Authentication failed: %v", err))
			return
		}

		if c.Query("format") == "json" || c.GetHeader("Accept") == "application/json" {
			response.Success(c, http.StatusOK, authResp)
			return
		}

		redirectURL := fmt.Sprintf("%s/auth/callback?access_token=%s&refresh_token=%s",
			h.cfg.FrontendURL,
			url.QueryEscape(authResp.AccessToken),
			url.QueryEscape(authResp.RefreshToken),
		)
		c.Redirect(http.StatusTemporaryRedirect, redirectURL)
	}
}

func (h *Handler) handleOAuthError(c *gin.Context, errMsg string) {
	if c.Query("format") == "json" || c.GetHeader("Accept") == "application/json" {
		response.Error(c, http.StatusUnauthorized, "OAUTH_FAILED", errMsg)
		return
	}

	redirectURL := fmt.Sprintf("%s/auth/callback?error=%s",
		h.cfg.FrontendURL,
		url.QueryEscape(errMsg),
	)
	c.Redirect(http.StatusTemporaryRedirect, redirectURL)
}

func generateRandomHex(byteCount int) (string, error) {
	b := make([]byte, byteCount)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
