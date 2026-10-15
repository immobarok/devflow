package jwt

import (
	"testing"
	"time"
)

func TestGenerateAndValidateToken(t *testing.T) {
	secret := "super-secret-key-that-is-at-least-32-chars-long"
	userID := "c7a8b418-4b7f-4c57-8107-164a66e4a644"
	email := "test@example.com"
	ttl := 15 * time.Minute

	token, err := GenerateAccessToken(userID, email, secret, ttl)
	if err != nil {
		t.Fatalf("failed to generate access token: %v", err)
	}
	if token == "" {
		t.Fatal("expected non-empty token")
	}

	claims, err := ValidateToken(token, secret)
	if err != nil {
		t.Fatalf("failed to validate token: %v", err)
	}

	if claims.UserID != userID {
		t.Errorf("expected UserID %s, got %s", userID, claims.UserID)
	}
	if claims.Email != email {
		t.Errorf("expected Email %s, got %s", email, claims.Email)
	}
}

func TestExpiredToken(t *testing.T) {
	secret := "super-secret-key-that-is-at-least-32-chars-long"
	userID := "c7a8b418-4b7f-4c57-8107-164a66e4a644"
	email := "test@example.com"
	ttl := -1 * time.Minute // already expired

	token, err := GenerateAccessToken(userID, email, secret, ttl)
	if err != nil {
		t.Fatalf("failed to generate access token: %v", err)
	}

	_, err = ValidateToken(token, secret)
	if err == nil {
		t.Fatal("expected error for expired token, got nil")
	}
	if err != ErrExpiredToken {
		t.Errorf("expected ErrExpiredToken, got %v", err)
	}
}

func TestInvalidSecret(t *testing.T) {
	secret := "secret-one-that-is-at-least-32-chars-long"
	wrongSecret := "secret-two-that-is-at-least-32-chars-long"
	userID := "c7a8b418-4b7f-4c57-8107-164a66e4a644"
	email := "test@example.com"
	ttl := 15 * time.Minute

	token, err := GenerateAccessToken(userID, email, secret, ttl)
	if err != nil {
		t.Fatalf("failed to generate access token: %v", err)
	}

	_, err = ValidateToken(token, wrongSecret)
	if err == nil {
		t.Fatal("expected error for invalid signature, got nil")
	}
}
