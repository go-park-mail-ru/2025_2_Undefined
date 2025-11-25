package cookie

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSet(t *testing.T) {
	tests := []struct {
		name       string
		token      string
		cookieName string
		wantCookie bool
	}{
		{
			name:       "Valid token",
			token:      "test-token-123",
			cookieName: "session_id",
			wantCookie: true,
		},
		{
			name:       "Empty token",
			token:      "",
			cookieName: "session_id",
			wantCookie: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			Set(w, tt.token, tt.cookieName)

			result := w.Result()
			cookies := result.Cookies()

			if tt.wantCookie {
				assert.Len(t, cookies, 1)
				assert.Equal(t, tt.cookieName, cookies[0].Name)
				assert.Equal(t, tt.token, cookies[0].Value)
				assert.Equal(t, "/", cookies[0].Path)
				assert.True(t, cookies[0].HttpOnly)
				assert.Equal(t, http.SameSiteStrictMode, cookies[0].SameSite)
				assert.True(t, cookies[0].Expires.After(time.Now()))
			} else {
				assert.Len(t, cookies, 0)
			}
		})
	}
}

func TestUnset(t *testing.T) {
	tests := []struct {
		name       string
		cookieName string
	}{
		{
			name:       "Unset existing cookie",
			cookieName: "session_id",
		},
		{
			name:       "Unset different cookie name",
			cookieName: "another_cookie",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			Unset(w, tt.cookieName)

			result := w.Result()
			cookies := result.Cookies()

			assert.Len(t, cookies, 1)
			assert.Equal(t, tt.cookieName, cookies[0].Name)
			assert.Equal(t, "", cookies[0].Value)
			assert.Equal(t, "/", cookies[0].Path)
			assert.True(t, cookies[0].HttpOnly)
			assert.True(t, cookies[0].Expires.Before(time.Now()))
		})
	}
}
