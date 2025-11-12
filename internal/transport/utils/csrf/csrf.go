package csrf

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"
)

func GenerateCSRFToken(sessionID string, secret string) string {
	timestamp := time.Now().Unix()
	data := fmt.Sprintf("%s:%d", sessionID, timestamp)
	
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(data))
	signature := hex.EncodeToString(h.Sum(nil))
	
	return fmt.Sprintf("%d.%s", timestamp, signature)
}

func ValidateCSRFToken(csrfToken string, sessionID string, secret string, timeout time.Duration) error {
	parts := strings.Split(csrfToken, ".")
	if len(parts) != 2 {
		return fmt.Errorf("invalid CSRF token format")
	}
	
	timestampStr, signature := parts[0], parts[1]
	timestamp, err := strconv.ParseInt(timestampStr, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid CSRF token timestamp")
	}
	
	tokenTime := time.Unix(timestamp, 0)
	if time.Since(tokenTime) > timeout {
		return fmt.Errorf("CSRF token has expired")
	}
	
	data := fmt.Sprintf("%s:%d", sessionID, timestamp)
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(data))
	expectedSignature := hex.EncodeToString(h.Sum(nil))
	
	if subtle.ConstantTimeCompare([]byte(signature), []byte(expectedSignature)) != 1 {
		return fmt.Errorf("invalid CSRF token signature")
	}
	
	return nil
}

// GetCSRFTokenTimeLeft возвращает оставшееся время жизни CSRF токена
func GetCSRFTokenTimeLeft(csrfToken string, timeout time.Duration) (time.Duration, error) {
	parts := strings.Split(csrfToken, ".")
	if len(parts) != 2 {
		return 0, fmt.Errorf("invalid CSRF token format")
	}
	
	timestampStr := parts[0]
	timestamp, err := strconv.ParseInt(timestampStr, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid CSRF token timestamp")
	}
	
	tokenTime := time.Unix(timestamp, 0)
	elapsed := time.Since(tokenTime)
	
	if elapsed > timeout {
		return 0, fmt.Errorf("CSRF token has expired")
	}
	
	return timeout - elapsed, nil
}
