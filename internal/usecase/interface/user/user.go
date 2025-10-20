package user

import (
	UserModels "github.com/go-park-mail-ru/2025_2_Undefined/internal/models/user"
	"github.com/google/uuid"
)

type UserRepository interface {
	GetUserByID(id uuid.UUID) (*UserModels.User, error)
	GetUsersNames(usersIds []uuid.UUID) ([]string, error)
}
