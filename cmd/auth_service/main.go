package main

import (
	"context"
	"fmt"
	"net"
	"net/http"

	"github.com/go-park-mail-ru/2025_2_Undefined/config"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/models/domains"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/repository"
	authRepo "github.com/go-park-mail-ru/2025_2_Undefined/internal/repository/auth"
	redisClient "github.com/go-park-mail-ru/2025_2_Undefined/internal/repository/redis"
	redisSession "github.com/go-park-mail-ru/2025_2_Undefined/internal/repository/redis/session"
	userRepo "github.com/go-park-mail-ru/2025_2_Undefined/internal/repository/user"
	grpcHandler "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/auth/grpc"
	gen "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/generated/auth"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/middleware"
	authUsecase "github.com/go-park-mail-ru/2025_2_Undefined/internal/usecase/auth"
	sessionUsecase "github.com/go-park-mail-ru/2025_2_Undefined/internal/usecase/session"
	"github.com/prometheus/client_golang/prometheus/promhttp"
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

	db, err := repository.NewPgxPool(ctx, conf.DBConfig)
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

	go func() {
		http.Handle("/metrics", promhttp.Handler())
		metricsAddr := ":" + conf.MetricsConfig.Port
		logger.Info(fmt.Sprintf("Auth metrics server is running on %s", metricsAddr))
		if err := http.ListenAndServe(metricsAddr, nil); err != nil {
			logger.WithError(err).Error("failed to start metrics server")
		}
	}()

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(middleware.UnaryServerInterceptor()),
	)
	gen.RegisterAuthServiceServer(grpcServer, authGRPCHandler)

	logger.Info(fmt.Sprintf("Auth gRPC server is running on %s", grpcListenAddr))
	if err := grpcServer.Serve(listener); err != nil {
		logger.WithError(err).Fatal("failed to serve")
	}
}
