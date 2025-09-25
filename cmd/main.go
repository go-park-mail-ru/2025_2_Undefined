package main

import (
	"log"
	"net/http"

	"github.com/go-park-mail-ru/2025_2_Undefined/config"
	authHandlers "github.com/go-park-mail-ru/2025_2_Undefined/internal/handlers/auth"
	chatsHandlers "github.com/go-park-mail-ru/2025_2_Undefined/internal/handlers/chats"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/handlers/jwt"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/middleware"
	inmemory "github.com/go-park-mail-ru/2025_2_Undefined/internal/repository/inmemory"
	blackTokenRep "github.com/go-park-mail-ru/2025_2_Undefined/internal/repository/token"
	authServicePkg "github.com/go-park-mail-ru/2025_2_Undefined/internal/service/auth"
	chatsServicePkg "github.com/go-park-mail-ru/2025_2_Undefined/internal/service/chats"

	"github.com/gorilla/mux"
)

func main() {
	cfg := config.NewConfig()

	userRepo := inmemory.NewUserRepo()
	chatsRepo := inmemory.NewChatsRepo()
	blackTokenRepo := blackTokenRep.NewTokenRepo()
	tokenator := jwt.NewTokenator()

	inmemory.FillWithFakeData(userRepo, chatsRepo)

	authService := authServicePkg.NewAuthService(userRepo, *tokenator, blackTokenRepo)
	chatsService := chatsServicePkg.NewChatsService(chatsRepo)

	authHandler := authHandlers.NewAuthHandler(authService)
	chatsHandler := chatsHandlers.NewChatsHandler(chatsService)

	r := mux.NewRouter()

	// Public routes
	r.HandleFunc("/api/v1/register", authHandler.Register).Methods("POST")
	r.HandleFunc("/api/v1/login", authHandler.Login).Methods("POST")

	// Protected routes
	api := r.PathPrefix("/api/v1").Subrouter()
	api.Use(middleware.AuthMiddleware(tokenator, blackTokenRepo))

	api.HandleFunc("/logout", authHandler.Logout).Methods("POST")
	api.HandleFunc("/me", authHandler.GetCurrentUser).Methods("GET")
	api.HandleFunc("/chats", chatsHandler.GetChats).Methods("GET")
	api.HandleFunc("/chats", chatsHandler.PostChats).Methods("POST")
	api.HandleFunc("/chats/{chatId}", chatsHandler.GetInformationAboutChat).Methods("GET")

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
