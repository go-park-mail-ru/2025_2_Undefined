package app

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	"github.com/go-park-mail-ru/2025_2_Undefined/config"
	_ "github.com/go-park-mail-ru/2025_2_Undefined/docs"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/models/domains"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/repository"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/middleware"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	httpSwagger "github.com/swaggo/http-swagger"

	"github.com/go-park-mail-ru/2025_2_Undefined/internal/repository/minio"

	userrepo "github.com/go-park-mail-ru/2025_2_Undefined/internal/repository/user"
	usert "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/user"
	useruc "github.com/go-park-mail-ru/2025_2_Undefined/internal/usecase/user"

	autht "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/auth/http"
	gen "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/generated/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

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

	minioClient, err := minio.NewMinioProvider(*conf.MinioConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to minio: %v", err)
	}

	userRepo := userrepo.New(db)
	userUC := useruc.New(userRepo, minioClient)

	// Подключение к gRPC серверу авторизации
	grpcConn, err := grpc.NewClient(
		conf.GRPCConfig.AuthServiceAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to auth gRPC service: %v", err)
	}

	authClient := gen.NewAuthServiceClient(grpcConn)
	authHandler := autht.NewAuthGRPCProxyHandler(authClient, conf.SessionConfig)
	userHandler := usert.New(userUC, authClient, conf.SessionConfig)

	chatsRepo := chatsRepository.NewChatsRepository(db)
	messageRepo := messageRepository.NewMessageRepository(db)
	listenerMap := messageUsecase.NewListenerMap()

	chatsUC := chatsUsecase.NewChatsUsecase(chatsRepo, userRepo, messageRepo, minioClient)
	messageUC := messageUsecase.NewMessageUsecase(messageRepo, userRepo, chatsRepo, minioClient, listenerMap)

	chatsHandler := chatsTransport.NewChatsHandler(messageUC, chatsUC)
	messageHandler := messageTransport.NewMessageHandler(messageUC, chatsUC)

	contactRepo := contactRepository.New(db)
	contactUC := contactUsecase.New(contactRepo, userRepo, minioClient)
	contactHandler := contactTransport.New(contactUC)

	// Настройка логгера
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetLevel(domains.LoggingLevel)

	// Настройка маршрутищатора
	router := mux.NewRouter()
	router.Use(middleware.CorsMiddleware)
	router.Use(func(next http.Handler) http.Handler {
		return middleware.AccessLogMiddleware(logger, next)
	})

	apiRouter := router.PathPrefix("/api/v1").Subrouter()

	protectedRouter := apiRouter.NewRoute().Subrouter()
	protectedRouter.Use(middleware.AuthGRPCMiddleware(conf.SessionConfig, authClient))
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
		userRouter.HandleFunc("/me", userHandler.UpdateUserInfo).Methods(http.MethodPatch)
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
	const op = "App.Run"
	ctx := context.Background()
	logger := domains.GetLogger(ctx).WithField("operation", op)

	server := &http.Server{
		Addr:    ":" + a.conf.ServerConfig.Port,
		Handler: a.router,
	}

	logger.Logger.SetLevel(logrus.InfoLevel)
	logger.Infof("Server starting on port %s", a.conf.ServerConfig.Port)
	logger.Infof("Swagger UI available at: http://localhost:%s/swagger/", a.conf.ServerConfig.Port)
	logger.Logger.SetLevel(domains.LoggingLevel)

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.WithError(err).Fatal("Server failed to start")
	}
}
