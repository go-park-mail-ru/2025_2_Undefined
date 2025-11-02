package csrf

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
)

func GenerateCSRFToken(sessionID string, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(sessionID))
	return hex.EncodeToString(h.Sum(nil))
}

func ValidateCSRFToken(csrfToken string, sessionID string, secret string) error {
	expectedToken := GenerateCSRFToken(sessionID, secret)

	if subtle.ConstantTimeCompare([]byte(csrfToken), []byte(expectedToken)) != 1 {
		return fmt.Errorf("invalid CSRF token")
	}

	return nil
}
