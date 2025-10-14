package cookie

import (
	"log"
	"net/http"
	"time"
)

func Set(w http.ResponseWriter, token, name string) {
	if token == "" {
		log.Println("Warning: empty token for cookie", name)
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
