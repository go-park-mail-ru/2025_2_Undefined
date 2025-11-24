package main

import (
	"context"
	"fmt"
	"net"

	"github.com/go-park-mail-ru/2025_2_Undefined/config"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/models/domains"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/repository"
	contactRepo "github.com/go-park-mail-ru/2025_2_Undefined/internal/repository/contact"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/repository/elasticsearch"
	contactES "github.com/go-park-mail-ru/2025_2_Undefined/internal/repository/elasticsearch/contact"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/repository/minio"
	userRepo "github.com/go-park-mail-ru/2025_2_Undefined/internal/repository/user"
	gen "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/generated/user"
	grpcHandler "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/user-contact/grpc"
	contactUsecase "github.com/go-park-mail-ru/2025_2_Undefined/internal/usecase/contact"
	userUsecase "github.com/go-park-mail-ru/2025_2_Undefined/internal/usecase/user"
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

	db, err := repository.NewPgxPool(ctx, conf.DBConfig)
	if err != nil {
		logger.WithError(err).Fatal("failed to connect to database")
	}
	defer db.Close()

	minioClient, err := minio.NewMinioProvider(*conf.MinioConfig)
	if err != nil {
		logger.WithError(err).Fatal("failed to connect to minio")
	}

	var contactSearchRepo *contactES.ContactSearchRepository
	esClient, err := elasticsearch.NewClient(
		conf.ElasticsearchConfig.URL,
		conf.ElasticsearchConfig.ContactsIndex,
		conf.ElasticsearchConfig.Username,
		conf.ElasticsearchConfig.Password,
	)
	if err != nil {
		logger.WithError(err).Warn("failed to connect to elasticsearch, search will be disabled")
		contactSearchRepo = nil
	} else {
		contactSearchRepo = contactES.NewContactSearchRepository(esClient.GetClient(), conf.ElasticsearchConfig.ContactsIndex)
		if err := contactSearchRepo.CreateIndex(ctx); err != nil {
			logger.WithError(err).Warn("failed to create elasticsearch index")
		}
	}

	userRepository := userRepo.New(db)
	contactRepository := contactRepo.New(db)

	userUsecaseInstance := userUsecase.New(userRepository, minioClient)
	contactUsecaseInstance := contactUsecase.New(contactRepository, userRepository, minioClient, contactSearchRepo)

	// Переиндексация существующих контактов в Elasticsearch
	if contactSearchRepo != nil {
		logger.Info("reindexing existing contacts to elasticsearch")
		if err := contactUsecaseInstance.ReindexAllContacts(ctx); err != nil {
			logger.WithError(err).Warn("failed to reindex contacts, search may be incomplete")
		} else {
			logger.Info("contacts reindexed successfully")
		}
	}

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
