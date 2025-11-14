package cookie

import (
	"context"
	"net/http"
	"time"

	"github.com/go-park-mail-ru/2025_2_Undefined/internal/models/domains"
)

func Set(w http.ResponseWriter, token, name string) {
	if token == "" {
		ctx := context.Background()
		logger := domains.GetLogger(ctx).WithField("operation", "cookie.Set")
		logger.Warnf("Warning: empty token for cookie %s", name)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    token,
		Path:     "/",
		SameSite: http.SameSiteStrictMode,
		HttpOnly: true,
		Expires:  time.Now().UTC().Add(90 * 24 * time.Hour),
	})
}

func Unset(w http.ResponseWriter, name string) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    "",
		Path:     "/",
		Expires:  time.Now().UTC().AddDate(0, 0, -1),
		HttpOnly: true,
	})
}
