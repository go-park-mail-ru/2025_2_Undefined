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
	"github.com/sirupsen/logrus"
	httpSwagger "github.com/swaggo/http-swagger"

	"github.com/go-park-mail-ru/2025_2_Undefined/internal/repository/minio"
	redisClient "github.com/go-park-mail-ru/2025_2_Undefined/internal/repository/redis"
	redisSessionRepo "github.com/go-park-mail-ru/2025_2_Undefined/internal/repository/redis/session"
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

	redisClient, err := redisClient.NewClient(conf.RedisConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %v", err)
	}

	minioClient, err := minio.NewMinioProvider(*conf.MinioConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to minio: %v", err)
	}

	sessionRepo := redisSessionRepo.New(redisClient.Client, conf.SessionConfig.LifeSpan)
	sessionUC := sessionuc.New(sessionRepo)
	sessionUtils := sessionutils.NewSessionUtils(sessionUC, conf.SessionConfig)

	userRepo := userrepo.New(db)
	userUC := useruc.New(userRepo, minioClient)
	userHandler := usert.New(userUC, sessionUC, sessionUtils)

	authRepo := authrepo.New(db)
	authUC := authuc.New(authRepo, userRepo, sessionRepo)
	authHandler := autht.New(authUC, conf.SessionConfig, conf.CSRFConfig, sessionUtils)

	chatsRepo := chatsRepository.NewChatsRepository(db)
	chatsUC := chatsUsecase.NewChatsUsecase(chatsRepo, userRepo, minioClient)
	chatsHandler := chatsTransport.NewChatsHandler(chatsUC, sessionUtils)

	messageRepo := messageRepository.NewMessageRepository(db)
	listenerMap := messageUsecase.NewListenerMap()
	messageUC := messageUsecase.NewMessageUsecase(messageRepo, userRepo, chatsRepo, minioClient, listenerMap)
	messageHandler := messageTransport.NewMessageHandler(messageUC, chatsUC, sessionUtils)

	contactRepo := contactRepository.New(db)
	contactUC := contactUsecase.New(contactRepo, userRepo, minioClient)
	contactHandler := contactTransport.New(contactUC, sessionUtils)

	// Настройка логгера
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetLevel(logrus.WarnLevel)

	// Настройка маршрутищатора
	router := mux.NewRouter()
	router.Use(middleware.CorsMiddleware)
	router.Use(func(next http.Handler) http.Handler {
		return middleware.AccessLogMiddleware(logger, next)
	})

	apiRouter := router.PathPrefix("/api/v1").Subrouter()

	protectedRouter := apiRouter.NewRoute().Subrouter()
	protectedRouter.Use(middleware.AuthMiddleware(conf.SessionConfig, sessionUC))
	protectedRouter.Use(middleware.CSRFMiddleware(conf.SessionConfig, conf.CSRFConfig))

	authRouter := apiRouter.PathPrefix("").Subrouter()
	{
		authRouter.HandleFunc("/login", authHandler.Login).Methods(http.MethodPost)
		authRouter.HandleFunc("/register", authHandler.Register).Methods(http.MethodPost)
		protectedRouter.HandleFunc("/logout", authHandler.Logout).Methods(http.MethodPost)
	}

	chatRouter := protectedRouter.PathPrefix("/chats").Subrouter()
	{
		chatRouter.HandleFunc("/{chat_id}", chatsHandler.GetInformationAboutChat).Methods(http.MethodGet)
		chatRouter.HandleFunc("", chatsHandler.GetChats).Methods(http.MethodGet)
		chatRouter.HandleFunc("", chatsHandler.PostChats).Methods(http.MethodPost)
		chatRouter.HandleFunc("/{chat_id}/members", chatsHandler.AddUsersToChat).Methods(http.MethodPatch)
		chatRouter.HandleFunc("/{chat_id}", chatsHandler.DeleteChat).Methods(http.MethodDelete)
		chatRouter.HandleFunc("/{chat_id}", chatsHandler.UpdateChat).Methods(http.MethodPatch)
		chatRouter.HandleFunc("/dialog/{user_id}", chatsHandler.GetUsersDialog).Methods(http.MethodGet)
	}

	userRouter := protectedRouter.PathPrefix("").Subrouter()
	{
		userRouter.HandleFunc("/me", userHandler.GetCurrentUser).Methods(http.MethodGet)
		userRouter.HandleFunc("/sessions", userHandler.GetSessionsByUser).Methods(http.MethodGet)
		userRouter.HandleFunc("/user/by-phone", userHandler.GetUserByPhone).Methods(http.MethodPost)
		userRouter.HandleFunc("/user/by-username", userHandler.GetUserByUsername).Methods(http.MethodPost)
		userRouter.HandleFunc("/user/avatar", userHandler.UploadUserAvatar)
		userRouter.HandleFunc("/session", userHandler.DeleteSession).Methods(http.MethodDelete)
		userRouter.HandleFunc("/sessions", userHandler.DeleteAllSessionWithoutCurrent).Methods(http.MethodDelete)
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
