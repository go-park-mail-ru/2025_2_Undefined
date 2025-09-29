package main

import (
	"log"
	"net/http"

	"github.com/go-park-mail-ru/2025_2_Undefined/config"
	_ "github.com/go-park-mail-ru/2025_2_Undefined/docs"
	httpSwagger "github.com/swaggo/http-swagger"

	AuthHandlers "github.com/go-park-mail-ru/2025_2_Undefined/internal/handlers/auth"
	ChatHandlers "github.com/go-park-mail-ru/2025_2_Undefined/internal/handlers/chats"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/handlers/jwt"
	utils "github.com/go-park-mail-ru/2025_2_Undefined/internal/handlers/utils/response"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/middleware"
	inmemory "github.com/go-park-mail-ru/2025_2_Undefined/internal/repository/inmemory"
	blackTokenRep "github.com/go-park-mail-ru/2025_2_Undefined/internal/repository/token"
	AuthService "github.com/go-park-mail-ru/2025_2_Undefined/internal/service/auth"
	ChatService "github.com/go-park-mail-ru/2025_2_Undefined/internal/service/chats"
)

// @title           Undefined team API documentation of project Telegram
// @version         1.0
// @description     API сервер для чат-приложения в стиле Telegram. Позволяет регистрировать пользователей, управлять чатами и обмениваться сообщениями.
// @termsOfService  http://swagger.io/terms/

// @contact.name   Undefined Team
// @contact.url    https://github.com/go-park-mail-ru/2025_2_Undefined

// @license.name  MIT
// @license.url   https://opensource.org/licenses/MIT

// @host      localhost:8080
// @BasePath  /api/v1
func main() {

	cfg := config.NewConfig()

	userRepo := inmemory.NewUserRepo()
	chatRepo := inmemory.NewChatsRepo(userRepo)

	blackTokenRepo := blackTokenRep.NewTokenRepo()
	tokenator := jwt.NewTokenator()

	authService := AuthService.NewAuthService(userRepo, tokenator, blackTokenRepo)
	chatService := ChatService.NewChatsService(chatRepo)

	authHandler := AuthHandlers.NewAuthHandler(authService)
	chatsHandler := ChatHandlers.NewChatsHandler(chatService)

	authMiddleware := middleware.AuthMiddleware(tokenator, blackTokenRepo)

	// Создаем хендлеры для каждого маршрута с проверкой методов
	registerHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			utils.SendError(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}
		authHandler.Register(w, r)
	})

	loginHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			utils.SendError(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}
		authHandler.Login(w, r)
	})

	logoutHandlerWithAuth := authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			utils.SendError(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}
		authHandler.Logout(w, r)
	}))

	meHandlerWithAuth := authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			utils.SendError(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}
		authHandler.GetCurrentUser(w, r)
	}))

	chatsUniversalHandler := authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			chatsHandler.GetChats(w, r)
		case "POST":
			chatsHandler.PostChats(w, r)
		default:
			utils.SendError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	}))

	chatInfoHandler := authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			utils.SendError(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}
		chatsHandler.GetInformationAboutChat(w, r)
	}))

	mux := http.NewServeMux()
	mux.Handle("/api/v1/register", registerHandler)
	mux.Handle("/api/v1/login", loginHandler)
	mux.Handle("/api/v1/logout", logoutHandlerWithAuth)
	mux.Handle("/api/v1/me", meHandlerWithAuth)

	mux.Handle("/api/v1/chats", chatsUniversalHandler)
	mux.Handle("/api/v1/chats/", chatInfoHandler)

	mux.Handle("/swagger/", httpSwagger.WrapHandler)

	fs := http.FileServer(http.Dir("frontend-build"))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.ServeFile(w, r, "frontend-build/index.html")
		} else {
			fs.ServeHTTP(w, r)
		}
	})

	handler := corsMiddleware(mux)

	log.Printf("Server starting on port %s", cfg.Port)
	log.Printf("Swagger UI available at: http://localhost:%s/swagger/", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, handler))
}

// настройка CORS для фронта
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
