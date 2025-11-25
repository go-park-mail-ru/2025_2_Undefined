package context

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-park-mail-ru/2025_2_Undefined/internal/models/domains"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestGetUserIDFromContext_Success(t *testing.T) {
	userID := uuid.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	ctx := context.WithValue(req.Context(), domains.UserIDKey{}, userID.String())
	req = req.WithContext(ctx)

	result, err := GetUserIDFromContext(req)

	assert.NoError(t, err)
	assert.Equal(t, userID, result)
}

func TestGetUserIDFromContext_NotFound(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	result, err := GetUserIDFromContext(req)

	assert.Error(t, err)
	assert.Equal(t, uuid.Nil, result)
	assert.Contains(t, err.Error(), "user_id not found in context")
}

func TestGetUserIDFromContext_InvalidType(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	ctx := context.WithValue(req.Context(), domains.UserIDKey{}, 12345)
	req = req.WithContext(ctx)

	result, err := GetUserIDFromContext(req)

	assert.Error(t, err)
	assert.Equal(t, uuid.Nil, result)
	assert.Contains(t, err.Error(), "user_id has invalid type in context")
}

func TestGetUserIDFromContext_InvalidFormat(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	ctx := context.WithValue(req.Context(), domains.UserIDKey{}, "invalid-uuid")
	req = req.WithContext(ctx)

	result, err := GetUserIDFromContext(req)

	assert.Error(t, err)
	assert.Equal(t, uuid.Nil, result)
	assert.Contains(t, err.Error(), "invalid user_id format")
}

func TestGetSessionIDFromCookie_Success(t *testing.T) {
	sessionID := uuid.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{
		Name:  "session_id",
		Value: sessionID.String(),
	})

	result, err := GetSessionIDFromCookie(req, "session_id")

	assert.NoError(t, err)
	assert.Equal(t, sessionID, result)
}

func TestGetSessionIDFromCookie_NotFound(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	result, err := GetSessionIDFromCookie(req, "session_id")

	assert.Error(t, err)
	assert.Equal(t, uuid.Nil, result)
	assert.Contains(t, err.Error(), "session not found")
}

func TestGetSessionIDFromCookie_InvalidFormat(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{
		Name:  "session_id",
		Value: "invalid-uuid",
	})

	result, err := GetSessionIDFromCookie(req, "session_id")

	assert.Error(t, err)
	assert.Equal(t, uuid.Nil, result)
	assert.Contains(t, err.Error(), "invalid session ID")
}
