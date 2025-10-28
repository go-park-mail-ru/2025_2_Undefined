package user

import (
	"context"
	"errors"
	"fmt"
	"log"

	UserModels "github.com/go-park-mail-ru/2025_2_Undefined/internal/models/user"
	UserDto "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/dto/user"
	"github.com/google/uuid"
)

type UserRepository interface {
	GetUserByID(ctx context.Context, id uuid.UUID) (*UserModels.User, error)
}

type UserUsecase struct {
	userrepo UserRepository
}

func New(userrepo UserRepository) *UserUsecase {
	return &UserUsecase{
		userrepo: userrepo,
	}
}

func (uc *UserUsecase) GetUserById(ctx context.Context, id uuid.UUID) (*UserDto.User, error) {
	const op = "AuthUsecase.GetUserById"
	user, err := uc.userrepo.GetUserByID(ctx, id)
	if err != nil {
		err = errors.New("Error getting user by ID")
		wrappedErr := fmt.Errorf("%s: %w", op, err)
		log.Printf("Error: %v", wrappedErr)
		return nil, err
	}

	userdto := &UserDto.User{
		ID:          user.ID,
		PhoneNumber: user.PhoneNumber,
		Name:        user.Name,
		Username:    user.Username,
		Bio:         user.Bio,
		Avatar:      user.Avatar,
		AccountType: user.AccountType,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
	}
	
	return userdto, nil
}
