package csrf

import (
	"testing"
	"time"
)

func TestGenerateCSRFToken(t *testing.T) {
	sessionID := "test-session-id"
	secret := "test-secret"

	token := GenerateCSRFToken(sessionID, secret)

	if token == "" {
		t.Error("Generated token should not be empty")
	}

	// Token should contain timestamp and signature separated by dot
	if len(token) < 10 {
		t.Error("Token should be longer than 10 characters")
	}
}

func TestValidateCSRFToken_ValidToken(t *testing.T) {
	sessionID := "test-session-id"
	secret := "test-secret"
	timeout := 1 * time.Hour

	token := GenerateCSRFToken(sessionID, secret)

	err := ValidateCSRFToken(token, sessionID, secret, timeout)
	if err != nil {
		t.Errorf("Valid token should pass validation, got error: %v", err)
	}
}

func TestValidateCSRFToken_ExpiredToken(t *testing.T) {
	sessionID := "test-session-id"
	secret := "test-secret"
	timeout := 1 * time.Millisecond // Very short timeout

	token := GenerateCSRFToken(sessionID, secret)

	// Wait for token to expire
	time.Sleep(2 * time.Millisecond)

	err := ValidateCSRFToken(token, sessionID, secret, timeout)
	if err == nil {
		t.Error("Expired token should fail validation")
	}

	if err.Error() != "CSRF token has expired" {
		t.Errorf("Expected 'CSRF token has expired' error, got: %v", err)
	}
}

func TestValidateCSRFToken_InvalidFormat(t *testing.T) {
	sessionID := "test-session-id"
	secret := "test-secret"
	timeout := 1 * time.Hour

	invalidToken := "invalid-token-format"

	err := ValidateCSRFToken(invalidToken, sessionID, secret, timeout)
	if err == nil {
		t.Error("Invalid token format should fail validation")
	}

	if err.Error() != "invalid CSRF token format" {
		t.Errorf("Expected 'invalid CSRF token format' error, got: %v", err)
	}
}

func TestValidateCSRFToken_WrongSecret(t *testing.T) {
	sessionID := "test-session-id"
	secret := "test-secret"
	wrongSecret := "wrong-secret"
	timeout := 1 * time.Hour

	token := GenerateCSRFToken(sessionID, secret)

	err := ValidateCSRFToken(token, sessionID, wrongSecret, timeout)
	if err == nil {
		t.Error("Token with wrong secret should fail validation")
	}

	if err.Error() != "invalid CSRF token signature" {
		t.Errorf("Expected 'invalid CSRF token signature' error, got: %v", err)
	}
}

func TestValidateCSRFToken_WrongSessionID(t *testing.T) {
	sessionID := "test-session-id"
	wrongSessionID := "wrong-session-id"
	secret := "test-secret"
	timeout := 1 * time.Hour

	token := GenerateCSRFToken(sessionID, secret)

	err := ValidateCSRFToken(token, wrongSessionID, secret, timeout)
	if err == nil {
		t.Error("Token with wrong session ID should fail validation")
	}

	if err.Error() != "invalid CSRF token signature" {
		t.Errorf("Expected 'invalid CSRF token signature' error, got: %v", err)
	}
}
