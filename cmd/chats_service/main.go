package main

import (
	"context"
	"fmt"
	"net"
	"net/http"

	"github.com/go-park-mail-ru/2025_2_Undefined/config"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/models/domains"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/repository"
	chatsRepo "github.com/go-park-mail-ru/2025_2_Undefined/internal/repository/chats"
	messageRepo "github.com/go-park-mail-ru/2025_2_Undefined/internal/repository/message"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/repository/minio"
	userRepo "github.com/go-park-mail-ru/2025_2_Undefined/internal/repository/user"
	grpcHandler "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/chats-message/grpc"
	gen "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/generated/chats"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/middleware"
	chatsUsecase "github.com/go-park-mail-ru/2025_2_Undefined/internal/usecase/chats"
	messageUsecase "github.com/go-park-mail-ru/2025_2_Undefined/internal/usecase/message"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
)

func main() {
	const op = "chats_service.main"
	ctx := context.Background()
	logger := domains.GetLogger(ctx).WithField("operation", op)

	conf, err := config.NewConfig()
	if err != nil {
		logger.WithError(err).Fatal("config error")
		return
	}

	db, err := repository.NewPgxPool(ctx, conf.DBConfig)
	if err != nil {
		logger.WithError(err).Fatal("failed to connect to database")
		return
	}
	defer db.Close()

	minioClient, err := minio.NewMinioProvider(*conf.MinioConfig)
	if err != nil {
		logger.WithError(err).Fatal("failed to connect to minio")
		return
	}

	chatsRepository := chatsRepo.NewChatsRepository(db)
	messageRepository := messageRepo.NewMessageRepository(db)
	userRepository := userRepo.New(db)
	listenerMap := messageUsecase.NewListenerMap()

	chatsUsecaseInstance := chatsUsecase.NewChatsUsecase(chatsRepository, userRepository, messageRepository, minioClient)
	messageUsecaseInstance := messageUsecase.NewMessageUsecase(messageRepository, userRepository, chatsRepository, minioClient, listenerMap)

	chatsGRPCHandler := grpcHandler.NewChatsGRPCHandler(chatsUsecaseInstance, messageUsecaseInstance)
	messageGRPCHandler := grpcHandler.NewMessageGRPCHandler(messageUsecaseInstance, chatsUsecaseInstance)

	grpcListenAddr := fmt.Sprintf(":%s", conf.GRPCConfig.ChatsServicePort)
	listener, err := net.Listen("tcp", grpcListenAddr)
	if err != nil {
		logger.WithError(err).Fatal("failed to listen")
	}

	go func() {
		http.Handle("/metrics", promhttp.Handler())
		metricsAddr := ":" + conf.MetricsConfig.Port
		logger.Info(fmt.Sprintf("Chats metrics server is running on %s", metricsAddr))
		if err := http.ListenAndServe(metricsAddr, nil); err != nil {
			logger.WithError(err).Error("failed to start metrics server")
		}
	}()

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(middleware.UnaryServerInterceptor()),
	)
	gen.RegisterChatServiceServer(grpcServer, chatsGRPCHandler)
	gen.RegisterMessageServiceServer(grpcServer, messageGRPCHandler)

	logger.Info(fmt.Sprintf("Chats gRPC server is running on %s", grpcListenAddr))
	if err := grpcServer.Serve(listener); err != nil {
		logger.WithError(err).Fatal("failed to serve")
	}
}
