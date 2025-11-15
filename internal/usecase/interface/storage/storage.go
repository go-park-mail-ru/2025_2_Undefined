package storage

import (
	"context"

	"github.com/go-park-mail-ru/2025_2_Undefined/internal/repository/minio"
	"github.com/google/uuid"
)

type FileStorage interface {
	CreateOne(ctx context.Context, file minio.FileData, objectID uuid.UUID) (string, error)

	GetOne(ctx context.Context, objectID *uuid.UUID) (string, error)

	DeleteOne(ctx context.Context, objectID uuid.UUID) error
}
