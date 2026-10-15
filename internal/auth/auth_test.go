package auth

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"

	"github.com/mobarok/devflow/internal/middleware"
	jwtpkg "github.com/mobarok/devflow/pkg/jwt"
)

func TestPasswordHashingAndVerification(t *testing.T) {
	rawPassword := "SecurePassword123!"

	hashed, err := bcrypt.GenerateFromPassword([]byte(rawPassword), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	// Verify correct password
	if err := bcrypt.CompareHashAndPassword(hashed, []byte(rawPassword)); err != nil {
		t.Errorf("expected password to match, got %v", err)
	}

	// Verify incorrect password fails
	if err := bcrypt.CompareHashAndPassword(hashed, []byte("WrongPassword")); err == nil {
		t.Error("expected error for mismatched password, got nil")
	}
}

func TestAuthMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	secret := "test-secret-at-least-32-characters-long"

	router := gin.New()
	router.GET("/protected", middleware.AuthRequired(secret), func(c *gin.Context) {
		userID, _ := c.Get("user_id")
		c.JSON(http.StatusOK, gin.H{"user_id": userID})
	})

	t.Run("Missing Authorization Header", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("expected 401, got %d", w.Code)
		}
	})

	t.Run("Malformed Authorization Header", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
		req.Header.Set("Authorization", "InvalidFormatToken")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("expected 401, got %d", w.Code)
		}
	})

	t.Run("Valid Token", func(t *testing.T) {
		token, err := jwtpkg.GenerateAccessToken("user-123", "user@example.com", secret, 15*time.Minute)
		if err != nil {
			t.Fatalf("failed to generate token: %v", err)
		}

		req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d. Body: %s", w.Code, w.Body.String())
		}
		if !strings.Contains(w.Body.String(), "user-123") {
			t.Errorf("expected user_id in response, got %s", w.Body.String())
		}
	})
}
