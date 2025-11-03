package user

import (
	"context"

	UserModels "github.com/go-park-mail-ru/2025_2_Undefined/internal/models/user"
	"github.com/google/uuid"
)

type UserRepository interface {
	GetUserByID(ctx context.Context, id uuid.UUID) (*UserModels.User, error)
	GetUsersNames(ctx context.Context, usersIds []uuid.UUID) ([]string, error)
	GetUserByPhone(ctx context.Context, phone string) (*UserModels.User, error)
	GetUserByUsername(ctx context.Context, username string) (*UserModels.User, error)
	UpdateUserAvatar(ctx context.Context, userID uuid.UUID, avatarID uuid.UUID, file_size int64) error
}
