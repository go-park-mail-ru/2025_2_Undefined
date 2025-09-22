package main

import (
	"log"
	"net/http"

	"github.com/go-park-mail-ru/2025_2_Undefined/config"
	handlers "github.com/go-park-mail-ru/2025_2_Undefined/internal/handlers/auth"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/middleware"
	inmemory "github.com/go-park-mail-ru/2025_2_Undefined/internal/repository/inmemory"
	tokenRep "github.com/go-park-mail-ru/2025_2_Undefined/internal/repository/token"
	service "github.com/go-park-mail-ru/2025_2_Undefined/internal/service/auth"
	"github.com/gorilla/mux"
)

func main() {
	cfg := config.NewConfig()

	userRepo := inmemory.NewUserRepo()
	tokenRepo := tokenRep.NewTokenRepo()

	authService := service.NewAuthService(userRepo, tokenRepo, cfg.JWTSecret, cfg.JWTTTL)

	authHandler := handlers.NewAuthHandler(authService)

	r := mux.NewRouter()

	// Public routes
	r.HandleFunc("/api/v1/register", authHandler.Register).Methods("POST")
	r.HandleFunc("/api/v1/login", authHandler.Login).Methods("POST")

	// Protected routes
	api := r.PathPrefix("/api/v1").Subrouter()
	api.Use(middleware.AuthMiddleware(authService))

	api.HandleFunc("/logout", authHandler.Logout).Methods("POST")
	api.HandleFunc("/me", authHandler.GetCurrentUser).Methods("GET")

	// CORS middleware для фронтенда
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	})

	log.Printf("Server starting on port %s", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, r))
}
