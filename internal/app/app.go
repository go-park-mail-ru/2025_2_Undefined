package app

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"github.com/go-park-mail-ru/2025_2_Undefined/config"
	_ "github.com/go-park-mail-ru/2025_2_Undefined/docs"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/repository"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/middleware"
	"github.com/gorilla/mux"
	httpSwagger "github.com/swaggo/http-swagger"

	sessionrepo "github.com/go-park-mail-ru/2025_2_Undefined/internal/repository/session"
	sessionutils "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/session"
	sessionuc "github.com/go-park-mail-ru/2025_2_Undefined/internal/usecase/session"

	userrepo "github.com/go-park-mail-ru/2025_2_Undefined/internal/repository/user"
	usert "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/user"
	useruc "github.com/go-park-mail-ru/2025_2_Undefined/internal/usecase/user"

	authrepo "github.com/go-park-mail-ru/2025_2_Undefined/internal/repository/auth"
	autht "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/auth"
	authuc "github.com/go-park-mail-ru/2025_2_Undefined/internal/usecase/auth"

	chatsRepository "github.com/go-park-mail-ru/2025_2_Undefined/internal/repository/chats"
	chatsTransport "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/chats"
	chatsUsecase "github.com/go-park-mail-ru/2025_2_Undefined/internal/usecase/chats"

	messageRepository "github.com/go-park-mail-ru/2025_2_Undefined/internal/repository/message"
	messageTransport "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/message"
	messageUsecase "github.com/go-park-mail-ru/2025_2_Undefined/internal/usecase/message"

	contactRepository "github.com/go-park-mail-ru/2025_2_Undefined/internal/repository/contact"
	contactTransport "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/contact"
	contactUsecase "github.com/go-park-mail-ru/2025_2_Undefined/internal/usecase/contact"
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

	sessionrepo := sessionrepo.New(db)
	sessionuc := sessionuc.New(sessionrepo)
	sessionutils := sessionutils.NewSessionUtils(sessionuc)

	userRepo := userrepo.New(db)
	userUC := useruc.New(userRepo)
	userHandler := usert.New(userUC, sessionutils)

	authRepo := authrepo.New(db)
	authUC := authuc.New(authRepo, userRepo, sessionrepo)
	authHandler := autht.New(authUC, sessionutils)

	chatsRepo := chatsRepository.NewChatsRepository(db)
	chatsUC := chatsUsecase.NewChatsService(chatsRepo, userRepo)
	chatsHandler := chatsTransport.NewChatsHandler(chatsUC, sessionutils)

	messageRepo := messageRepository.NewMessageRepository(db)
	listenerMap := messageUsecase.NewListenerMap()
	messageUC := messageUsecase.NewMessageUsecase(messageRepo, userRepo, listenerMap)
	messageHandler := messageTransport.NewMessageHandler(messageUC, chatsUC, sessionutils)

	contactRepo := contactRepository.New(db)
	contactUC := contactUsecase.New(contactRepo, userRepo)
	contactHandler := contactTransport.New(contactUC, sessionutils)

	// Настройка маршрутищатора
	router := mux.NewRouter()
	router.Use(corsMiddleware)

	apiRouter := router.PathPrefix("/api/v1").Subrouter()

	authRouter := apiRouter.PathPrefix("").Subrouter()
	{
		authRouter.HandleFunc("/login", authHandler.Login).Methods(http.MethodPost)
		authRouter.HandleFunc("/register", authHandler.Register).Methods(http.MethodPost)
		authRouter.Handle("/logout",
			middleware.AuthMiddleware(sessionrepo)(http.HandlerFunc(authHandler.Logout)),
		).Methods(http.MethodPost)
	}

	protectedRouter := apiRouter.NewRoute().Subrouter()
	protectedRouter.Use(middleware.AuthMiddleware(sessionrepo))

	chatRouter := protectedRouter.PathPrefix("/chats").Subrouter()
	{
		chatRouter.HandleFunc("/{chat_id}", chatsHandler.GetInformationAboutChat).Methods(http.MethodGet)
		chatRouter.HandleFunc("", chatsHandler.GetChats).Methods(http.MethodGet)
		chatRouter.HandleFunc("", chatsHandler.PostChats).Methods(http.MethodPost)
	}

	userRouter := protectedRouter.PathPrefix("").Subrouter()
	{
		userRouter.HandleFunc("/me", userHandler.GetCurrentUser).Methods(http.MethodGet)
		userRouter.HandleFunc("/sessions", userHandler.GetSessionsByUser).Methods(http.MethodGet)
	}

	messageRouter := protectedRouter.PathPrefix("").Subrouter()
	{
		messageRouter.HandleFunc("/message/ws", messageHandler.HandleMessages)
	}

	contactRouter := protectedRouter.PathPrefix("/contacts").Subrouter()
	{
		contactRouter.HandleFunc("", contactHandler.CreateContact).Methods(http.MethodPost)
		contactRouter.HandleFunc("", contactHandler.GetContacts).Methods(http.MethodGet)
	}

	// Swagger
	router.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)

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
