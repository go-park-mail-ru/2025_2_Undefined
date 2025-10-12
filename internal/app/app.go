package app

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"github.com/go-park-mail-ru/2025_2_Undefined/config"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/repository"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/jwt"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/middleware"
	"github.com/gorilla/mux"

	blacktoken "github.com/go-park-mail-ru/2025_2_Undefined/internal/repository/token"

	authrepo "github.com/go-park-mail-ru/2025_2_Undefined/internal/repository/auth"
	autht "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/auth"
	authuc "github.com/go-park-mail-ru/2025_2_Undefined/internal/usecase/auth"
)

type App struct {
	conf   *config.Config
	db     *sql.DB
	router *mux.Router
}

func NewApp(conf *config.Config) (*App, error) {
	dbConn, err := repository.GetConnectionString(conf.DBConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to get connection string: %v", err)
	}

	db, err := sql.Open("postgres", dbConn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}

	tokenator := jwt.NewTokenator()
	blacktoken := blacktoken.NewTokenRepo()

	authRepo := authrepo.New(db)
	authUC := authuc.New(authRepo, tokenator, blacktoken)
	authHandler := autht.New(authUC)

	// Настройка маршрутищатора
	router := mux.NewRouter()
	router.Use(corsMiddleware)

	apiRouter := router.PathPrefix("/api/v1").Subrouter()

	authRouter := apiRouter.PathPrefix("").Subrouter()
	{
		authRouter.HandleFunc("/login", authHandler.Login).Methods(http.MethodPost)
		authRouter.HandleFunc("/register", authHandler.Register).Methods(http.MethodPost)
		authRouter.Handle("/logout",
			middleware.AuthMiddleware(tokenator, blacktoken)(http.HandlerFunc(authHandler.Logout)),
		).Methods(http.MethodPost)
		authRouter.HandleFunc("/me", authHandler.GetCurrentUser).Methods(http.MethodGet)
	}

	return &App{
		conf:   conf,
		db:     db,
		router: router,
	}, nil
}

func (a *App) Run() {
	server := &http.Server{
		Addr:    ":" + a.conf.ServerConfig.Port,
		Handler: a.router,
	}

	log.Printf("Server starting on port %s", a.conf.ServerConfig.Port)
	log.Printf("Swagger UI available at: http://localhost:%s/swagger/", a.conf.ServerConfig.Port)

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server failed to start: %v", err)
	}
}

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