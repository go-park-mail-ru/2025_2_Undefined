package main

import (
	"context"
	"database/sql"
	"fmt"
	"net"

	"github.com/go-park-mail-ru/2025_2_Undefined/config"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/models/domains"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/repository"
	authRepo "github.com/go-park-mail-ru/2025_2_Undefined/internal/repository/auth"
	redisClient "github.com/go-park-mail-ru/2025_2_Undefined/internal/repository/redis"
	redisSession "github.com/go-park-mail-ru/2025_2_Undefined/internal/repository/redis/session"
	userRepo "github.com/go-park-mail-ru/2025_2_Undefined/internal/repository/user"
	grpcHandler "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/auth/grpc"
	gen "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/generated/auth"
	authUsecase "github.com/go-park-mail-ru/2025_2_Undefined/internal/usecase/auth"
	sessionUsecase "github.com/go-park-mail-ru/2025_2_Undefined/internal/usecase/session"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"
)

func main() {
	const op = "auth_service.main"
	ctx := context.Background()
	logger := domains.GetLogger(ctx).WithField("operation", op)

	conf, err := config.NewConfig()
	if err != nil {
		logger.WithError(err).Fatal("config error")
	}

	dbConn, err := repository.GetConnectionString(conf.DBConfig)
	if err != nil {
		logger.WithError(err).Fatal("failed to get connection string")
	}

	db, err := sql.Open("postgres", dbConn)
	if err != nil {
		logger.WithError(err).Fatal("failed to connect to database")
	}
	defer db.Close()

	redisClient, err := redisClient.NewClient(conf.RedisConfig)
	if err != nil {
		logger.WithError(err).Fatal("failed to connect to redis")
	}

	authRepository := authRepo.New(db)
	userRepository := userRepo.New(db)
	sessionRepository := redisSession.New(redisClient.Client, conf.SessionConfig.LifeSpan)

	authUsecaseInstance := authUsecase.New(authRepository, userRepository, sessionRepository)
	sessionUsecaseInstance := sessionUsecase.New(sessionRepository)

	authGRPCHandler := grpcHandler.NewAuthGRPCHandler(authUsecaseInstance, sessionUsecaseInstance, conf.CSRFConfig)

	grpcListenAddr := fmt.Sprintf(":%s", conf.GRPCConfig.AuthServicePort)
	listener, err := net.Listen("tcp", grpcListenAddr)
	if err != nil {
		logger.WithError(err).Fatal("failed to listen")
	}

	grpcServer := grpc.NewServer()
	gen.RegisterAuthServiceServer(grpcServer, authGRPCHandler)

	logger.Info(fmt.Sprintf("Auth gRPC server is running on %s", grpcListenAddr))
	if err := grpcServer.Serve(listener); err != nil {
		logger.WithError(err).Fatal("failed to serve")
	}
}
