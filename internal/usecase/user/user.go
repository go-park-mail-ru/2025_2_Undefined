package user

import (
	"errors"
	"fmt"
	"log"

	UserModels "github.com/go-park-mail-ru/2025_2_Undefined/internal/models/user"
	"github.com/google/uuid"
)

type UserRepository interface {
	GetUserByID(id uuid.UUID) (*UserModels.User, error)
}

type UserUsecase struct {
	userrepo UserRepository
}

func New(userrepo UserRepository) *UserUsecase {
	return &UserUsecase{
		userrepo: userrepo,
	}
}

func (uc *UserUsecase) GetUserById(id uuid.UUID) (*UserModels.User, error) {
	const op = "AuthUsecase.GetUserById"
	user, err := uc.userrepo.GetUserByID(id)
	if err != nil {
		err = errors.New("Error getting user by ID")
		wrappedErr := fmt.Errorf("%s: %w", op, err)
		log.Printf("Error: %v", wrappedErr)
		return nil, err
	}
	return user, nil
}
