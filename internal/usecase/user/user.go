package user

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/go-park-mail-ru/2025_2_Undefined/internal/models/errs"
	UserModels "github.com/go-park-mail-ru/2025_2_Undefined/internal/models/user"
	UserDto "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/dto/user"
	"github.com/google/uuid"
)

type UserRepository interface {
	GetUserByID(ctx context.Context, id uuid.UUID) (*UserModels.User, error)
	GetUserByPhone(ctx context.Context, phone string) (*UserModels.User, error)
	GetUserByUsername(ctx context.Context, username string) (*UserModels.User, error)
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

func (uc *UserUsecase) GetUserByPhone(ctx context.Context, phone string) (*UserDto.User, error) {
	const op = "AuthUsecase.GetUserByPhone"
	user, err := uc.userrepo.GetUserByPhone(ctx, phone)
	if err != nil {
		err = errs.ErrUserNotFound
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

func (uc *UserUsecase) GetUserByUsername(ctx context.Context, username string) (*UserDto.User, error) {
	const op = "AuthUsecase.GetUserByPhone"
	user, err := uc.userrepo.GetUserByUsername(ctx, username)
	if err != nil {
		err = errs.ErrUserNotFound
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
