package app

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	"github.com/go-park-mail-ru/2025_2_Undefined/config"
	_ "github.com/go-park-mail-ru/2025_2_Undefined/docs"
	_ "github.com/lib/pq"

	"github.com/go-park-mail-ru/2025_2_Undefined/internal/models/domains"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/repository"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/middleware"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	httpSwagger "github.com/swaggo/http-swagger"

	userHttpProxy "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/user-contact/http"

	autht "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/auth/http"
	authGen "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/generated/auth"
	userGen "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/generated/user"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	chatsHTTTPProxy "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/chats-message/http"
	chatsGen "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/generated/chats"
)

type App struct {
	conf   *config.Config
	db     *sql.DB
	router *mux.Router
}

func NewApp(conf *config.Config) (*App, error) {
	dbConn := repository.GetConnectionString(conf.DBConfig)

	db, err := sql.Open("postgres", dbConn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}

	// Подключение к gRPC серверу авторизации
	authGrpcConn, err := grpc.NewClient(
		conf.GRPCConfig.AuthServiceAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to auth gRPC service: %v", err)
	}

	// Подключение к gRPC серверу user+contacts
	userGrpcConn, err := grpc.NewClient(
		conf.GRPCConfig.UserServiceAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to user gRPC service: %v", err)
	}

	chatsGrpcConn, err := grpc.NewClient(
		conf.GRPCConfig.ChatsServiceAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to chats gRPC service: %v", err)
	}

	authClient := authGen.NewAuthServiceClient(authGrpcConn)
	authHandler := autht.NewAuthGRPCProxyHandler(authClient, conf.SessionConfig)

	userClient := userGen.NewUserServiceClient(userGrpcConn)
	userHandler := userHttpProxy.NewUserGRPCProxyHandler(userClient)

	chatsClient := chatsGen.NewChatServiceClient(chatsGrpcConn)
	messageClient := chatsGen.NewMessageServiceClient(chatsGrpcConn)
	chatsHandler := chatsHTTTPProxy.NewChatsGRPCProxyHandler(chatsClient, messageClient)

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
		chatRouter.HandleFunc("/avatars/query", chatsHandler.GetChatAvatars).Methods(http.MethodPost)
		chatRouter.HandleFunc("/{chat_id}/avatar", chatsHandler.UploadChatAvatar).Methods(http.MethodPost)
	}

	userRouter := protectedRouter.PathPrefix("").Subrouter()
	{
		userRouter.HandleFunc("/me", userHandler.GetCurrentUser).Methods(http.MethodGet)
		userRouter.HandleFunc("/me", userHandler.UpdateUserInfo).Methods(http.MethodPatch)
		userRouter.HandleFunc("/user/by-phone", userHandler.GetUserByPhone).Methods(http.MethodPost)
		userRouter.HandleFunc("/user/by-username", userHandler.GetUserByUsername).Methods(http.MethodPost)
		userRouter.HandleFunc("/users/avatar", userHandler.UploadUserAvatar).Methods(http.MethodPost)
		userRouter.HandleFunc("/users/avatars/query", userHandler.GetUserAvatars).Methods(http.MethodPost)
	}

	sessionRouter := protectedRouter.PathPrefix("").Subrouter()
	{
		sessionRouter.HandleFunc("/sessions", authHandler.GetSessionsByUser).Methods(http.MethodGet)
		sessionRouter.HandleFunc("/session", authHandler.DeleteSession).Methods(http.MethodDelete)
		sessionRouter.HandleFunc("/sessions", authHandler.DeleteAllSessionsExceptCurrent).Methods(http.MethodDelete)
	}

	messageRouter := protectedRouter.PathPrefix("").Subrouter()
	{
		messageRouter.HandleFunc("/message/ws", chatsHandler.HandleMessages)
	}

	contactRouter := protectedRouter.PathPrefix("/contacts").Subrouter()
	{
		contactRouter.HandleFunc("", userHandler.CreateContact).Methods(http.MethodPost)
		contactRouter.HandleFunc("", userHandler.GetContacts).Methods(http.MethodGet)
		contactRouter.HandleFunc("/search", userHandler.SearchContacts).Methods(http.MethodGet)
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
