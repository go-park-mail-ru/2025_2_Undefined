package minio

import (
	"bytes"
	"context"
	"fmt"

	"github.com/go-park-mail-ru/2025_2_Undefined/config"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/models/domains"
	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinioProvider struct {
	mc         *minio.Client
	bucketName string
	config     config.MinioConfig
}

func NewMinioProvider(cfg config.MinioConfig) (*MinioProvider, error) {
	ctx := context.Background()

	endpoint := fmt.Sprintf("%s:%s", cfg.Host, cfg.PORT)
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.RootUser, cfg.RootPassword, ""),
		Secure: false,
	})

	if err != nil {
		return nil, err
	}

	exists, err := client.BucketExists(ctx, cfg.BucketName)
	if err != nil {
		return nil, err
	}

	if !exists {
		err := client.MakeBucket(ctx, cfg.BucketName, minio.MakeBucketOptions{})
		if err != nil {
			return nil, err
		}
	}

	// Устанавливаем публичную политику для бакета
	policy := fmt.Sprintf(`{
		"Version": "2012-10-17",
		"Statement": [
			{
				"Effect": "Allow",
				"Principal": {
					"AWS": ["*"]
				},
				"Action": ["s3:GetObject"],
				"Resource": ["arn:aws:s3:::%s/*"]
			}
		]
	}`, cfg.BucketName)

	err = client.SetBucketPolicy(ctx, cfg.BucketName, policy)
	if err != nil {
		return nil, fmt.Errorf("failed to set bucket policy: %v", err)
	}

	return &MinioProvider{
		mc:         client,
		bucketName: cfg.BucketName,
		config:     cfg,
	}, nil
}

func (m *MinioProvider) getURL(objectID uuid.UUID) string {
	// Возвращаем прямую публичную URL для объекта
	protocol := "http"
	if m.config.UseSSL {
		protocol = "https"
	}

	host := m.config.PublicHost
	if host == "" {
		host = m.config.Host
	}

	publicURL := fmt.Sprintf("%s://%s/%s/%s",
		protocol,
		host,
		m.bucketName,
		objectID.String())

	return publicURL
}

func (m *MinioProvider) CreateOne(ctx context.Context, file FileData, objectID uuid.UUID) (string, error) {
	const op = "MinioProvider.CreateOne"
	const query = "PUT object"

	logger := domains.GetLogger(ctx).WithField("operation", op).
		WithField("object_id", objectID.String()).
		WithField("file_name", file.Name)

	queryStatus := "success"
	defer func() {
		logger.Debugf("minio query: %s: status: %s", query, queryStatus)
	}()

	logger.Debugf("starting: %s", query)

	reader := bytes.NewReader(file.Data)

	options := minio.PutObjectOptions{}
	if file.ContentType != "" {
		options.ContentType = file.ContentType
	}
	if file.Name != "" {
		options.ContentDisposition = fmt.Sprintf("inline; filename=\"%s\"", file.Name)
	}

	_, err := m.mc.PutObject(ctx, m.bucketName, objectID.String(), reader, int64(len(file.Data)), options)
	if err != nil {
		queryStatus = "fail"
		logger.WithError(err).Errorf("minio query: %s: put object error: status: %s", query, queryStatus)
		return "", fmt.Errorf("error create object in minio %s: %v", file.Name, err)
	}

	return m.getURL(objectID), nil
}

func (m *MinioProvider) GetOne(ctx context.Context, objectID *uuid.UUID) (string, error) {
	if objectID == nil {
		return "", nil
	}

	const op = "MinioProvider.GetOne"
	const query = "GET object URL"

	logger := domains.GetLogger(ctx).WithField("operation", op).WithField("object_id", objectID.String())

	queryStatus := "success"
	defer func() {
		logger.Debugf("minio query: %s: status: %s", query, queryStatus)
	}()

	logger.Debugf("starting: %s", query)

	return m.getURL(*objectID), nil
}

func (m *MinioProvider) DeleteOne(ctx context.Context, objectID uuid.UUID) error {
	const op = "MinioProvider.DeleteOne"
	const query = "DELETE object"

	logger := domains.GetLogger(ctx).WithField("operation", op).WithField("object_id", objectID.String())

	queryStatus := "success"
	defer func() {
		logger.Debugf("minio query: %s: status: %s", query, queryStatus)
	}()

	logger.Debugf("starting: %s", query)

	err := m.mc.RemoveObject(ctx, m.bucketName, objectID.String(), minio.RemoveObjectOptions{})
	if err != nil {
		queryStatus = "fail"
		logger.WithError(err).Errorf("minio query: %s: remove object error: status: %s", query, queryStatus)
		return fmt.Errorf("error deleting object in minio: %v", err)
	}

	return nil
}

type FileData struct {
	Name        string
	Data        []byte
	ContentType string
}

type OperationError struct {
	ObjectID string
	Error    error
}
