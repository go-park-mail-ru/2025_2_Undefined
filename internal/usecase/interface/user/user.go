package user

import (
	"context"

	UserModels "github.com/go-park-mail-ru/2025_2_Undefined/internal/models/user"
	"github.com/google/uuid"
)

type UserRepository interface {
	GetUserByID(ctx context.Context, id uuid.UUID) (*UserModels.User, error)
	GetUsersNames(ctx context.Context, usersIds []uuid.UUID) ([]string, error)
}
