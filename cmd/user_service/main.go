package main

import (
	"context"
	"database/sql"
	"fmt"
	"net"

	"github.com/go-park-mail-ru/2025_2_Undefined/config"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/models/domains"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/repository"
	contactRepo "github.com/go-park-mail-ru/2025_2_Undefined/internal/repository/contact"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/repository/minio"
	userRepo "github.com/go-park-mail-ru/2025_2_Undefined/internal/repository/user"
	gen "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/generated/user"
	grpcHandler "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/user-contact/grpc"
	contactUsecase "github.com/go-park-mail-ru/2025_2_Undefined/internal/usecase/contact"
	userUsecase "github.com/go-park-mail-ru/2025_2_Undefined/internal/usecase/user"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"
)

func main() {
	const op = "user_service.main"
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

	minioClient, err := minio.NewMinioProvider(*conf.MinioConfig)
	if err != nil {
		logger.WithError(err).Fatal("failed to connect to minio")
	}

	userRepository := userRepo.New(db)
	contactRepository := contactRepo.New(db)

	userUsecaseInstance := userUsecase.New(userRepository, minioClient)
	contactUsecaseInstance := contactUsecase.New(contactRepository, userRepository, minioClient)

	userGRPCHandler := grpcHandler.NewUserGRPCHandler(userUsecaseInstance, contactUsecaseInstance)

	grpcListenAddr := fmt.Sprintf(":%s", conf.GRPCConfig.UserServicePort)
	listener, err := net.Listen("tcp", grpcListenAddr)
	if err != nil {
		logger.WithError(err).Fatal("failed to listen")
	}

	grpcServer := grpc.NewServer()
	gen.RegisterUserServiceServer(grpcServer, userGRPCHandler)

	logger.Info(fmt.Sprintf("User gRPC server is running on %s", grpcListenAddr))
	if err := grpcServer.Serve(listener); err != nil {
		logger.WithError(err).Fatal("failed to serve")
	}
}
