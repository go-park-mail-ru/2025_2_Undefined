package cookie

import (
	"log"
	"net/http"
	"time"

	"github.com/go-park-mail-ru/2025_2_Undefined/internal/handlers/jwt"
)

func Set(w http.ResponseWriter, token, name string) {
	t1 := jwt.NewTokenator()
	if token == "" {
		log.Println("Warning: empty token for cookie", name)
		return
	}

	tokenLifeSpan := t1.GetTokenLifeSpan()
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    token,
		Path:     "/",
		SameSite: http.SameSiteStrictMode,
		HttpOnly: true,
		Expires:  time.Now().UTC().Add(tokenLifeSpan),
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
