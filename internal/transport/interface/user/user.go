package user

import (
	"context"

	UserDTO "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/dto/user"
	"github.com/google/uuid"
)

type UserUsecase interface {
	GetUserById(ctx context.Context, id uuid.UUID) (*UserDTO.User, error)
	GetUserByPhone(ctx context.Context, phone string) (*UserDTO.User, error)
	GetUserByUsername(ctx context.Context, username string) (*UserDTO.User, error)
	UploadUserAvatar(ctx context.Context, userID uuid.UUID, data []byte, filename, contentType string) (string, error)
	UpdateUserInfo(ctx context.Context, userID uuid.UUID, name *string, username *string, bio *string) error
}
