package app

import (
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

	usert "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/user"
	useruc "github.com/go-park-mail-ru/2025_2_Undefined/internal/usecase/user"

	autht "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/auth"
	authuc "github.com/go-park-mail-ru/2025_2_Undefined/internal/usecase/auth"

	chatsTransport "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/chats"
	chatsUsecase "github.com/go-park-mail-ru/2025_2_Undefined/internal/usecase/chats"

	messageTransport "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/message"
	messageUsecase "github.com/go-park-mail-ru/2025_2_Undefined/internal/usecase/message"

	contactTransport "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/contact"
	contactUsecase "github.com/go-park-mail-ru/2025_2_Undefined/internal/usecase/contact"
)

type App struct {
	conf   *config.Config
	repos  *repository.Repositories
	router *mux.Router
}

func NewApp(conf *config.Config) (*App, error) {
	// Подключение к PostgreSQL через pgx v5
	dbConfig := &repository.Config{
		Host:     conf.DBConfig.Host,
		Port:     conf.DBConfig.Port,
		User:     conf.DBConfig.User,
		Password: conf.DBConfig.Password,
		DBName:   conf.DBConfig.DBName,
		SSLMode:  conf.DBConfig.SSLMode,
	}

	repos, err := repository.NewRepositories(dbConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize repositories: %w", err)
	}

	// Подключение к Redis
	redisClient, err := redisClient.NewClient(conf.RedisConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %v", err)
	}

	// Подключение к MinIO
	minioClient, err := minio.NewMinioProvider(*conf.MinioConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to minio: %v", err)
	}

	// Инициализация Use Cases и Handlers
	sessionRepo := redisSessionRepo.New(redisClient.Pool, conf.SessionConfig.LifeSpan)
	sessionUC := sessionuc.New(sessionRepo)
	sessionUtils := sessionutils.NewSessionUtils(sessionUC, conf.SessionConfig)

	userUC := useruc.New(repos.User, minioClient)
	userHandler := usert.New(userUC, sessionUC, sessionUtils)

	authUC := authuc.New(repos.Auth, repos.User, sessionRepo)
	authHandler := autht.New(authUC, conf.SessionConfig, conf.CSRFConfig, sessionUtils)

	chatsUC := chatsUsecase.NewChatsUsecase(repos.Chats, repos.User, minioClient)
	chatsHandler := chatsTransport.NewChatsHandler(chatsUC, sessionUtils)

	listenerMap := messageUsecase.NewListenerMap()
	messageUC := messageUsecase.NewMessageUsecase(repos.Message, repos.User, repos.Chats, minioClient, listenerMap)
	messageHandler := messageTransport.NewMessageHandler(messageUC, chatsUC, sessionUtils)

	contactUC := contactUsecase.New(repos.Contact, repos.User, minioClient)
	contactHandler := contactTransport.New(contactUC, sessionUtils)

	// Настройка логгера
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetLevel(logrus.WarnLevel)

	// Настройка маршрутизатора
	router := mux.NewRouter()
	router.Use(middleware.CorsMiddleware)
	router.Use(func(next http.Handler) http.Handler {
		return middleware.AccessLogMiddleware(logger, next)
	})

	apiRouter := router.PathPrefix("/api/v1").Subrouter()

	authRouter := apiRouter.PathPrefix("").Subrouter()
	{
		authRouter.HandleFunc("/login", authHandler.Login).Methods(http.MethodPost)
		authRouter.HandleFunc("/register", authHandler.Register).Methods(http.MethodPost)
		authRouter.Handle("/logout",
			middleware.AuthMiddleware(conf.SessionConfig, sessionUC)(http.HandlerFunc(authHandler.Logout)),
		).Methods(http.MethodPost)
	}

	protectedRouter := apiRouter.NewRoute().Subrouter()
	protectedRouter.Use(middleware.AuthMiddleware(conf.SessionConfig, sessionUC))
	protectedRouter.Use(middleware.CSRFMiddleware(conf.SessionConfig, conf.CSRFConfig.Secret))

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
		repos:  repos,
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

func (a *App) Close() {
	if a.repos != nil {
		a.repos.Close()
	}
}
