package usecase

import (
	"errors"
	"fmt"
	"log"

	"github.com/go-park-mail-ru/2025_2_Undefined/internal/models/errs"
	UserModels "github.com/go-park-mail-ru/2025_2_Undefined/internal/models/user"
	AuthModels "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/dto/auth"
	dto "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/dto/utils"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/jwt"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/utils/validation"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type ITokenator interface {
	CreateJWT(userID string) (string, error)
	ParseJWT(tokenString string) (*jwt.JWTClaims, error)
}

type IBlackToken interface {
	AddToBlacklist(token string) error
	IsInBlacklist(token string) bool
	CleanupExpiredTokens()
}

type AuthRepository interface {
	CreateUser(name string, phone string, password_hash string) (*UserModels.User, error)
}

type UserRepository interface {
	GetUserByPhone(phone string) (*UserModels.User, error)
	GetUserByID(id uuid.UUID) (*UserModels.User, error)
}

type AuthUsecase struct {
	authrepo   AuthRepository
	userrepo   UserRepository
	tokenator  ITokenator
	blacktoken IBlackToken
}

// GetUserById implements transport.AuthUsecase.
func (uc *AuthUsecase) GetUserById(id uuid.UUID) (*UserModels.User, error) {
	panic("unimplemented")
}

func New(authrepo AuthRepository, userrepo UserRepository, tokenator ITokenator, blacktoken IBlackToken) *AuthUsecase {
	return &AuthUsecase{
		authrepo:   authrepo,
		userrepo:   userrepo,
		tokenator:  tokenator,
		blacktoken: blacktoken,
	}
}

func (uc *AuthUsecase) Register(req *AuthModels.RegisterRequest) (string, *dto.ValidationErrorsDTO) {
	const op = "AuthUsecase.Register"
	errorsValidation := make([]errs.ValidationError, 0)

	existing, _ := uc.userrepo.GetUserByPhone(req.PhoneNumber)
	if existing != nil {
		errorsValidation = append(errorsValidation, errs.ValidationError{
			Field:   "phone_number",
			Message: "a user with such a phone already exists",
		})
	}

	if len(errorsValidation) > 0 {
		wrappedErr := fmt.Errorf("%s: %w", op, errors.New("in validation.ConvertToValidationErrorsDTO"))
		log.Printf("Error: %v", wrappedErr)
		err := validation.ConvertToValidationErrorsDTO(errorsValidation)

		return "", &err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		wrappedErr := fmt.Errorf("%s: %w", op, err)
		log.Printf("Error: %v", wrappedErr)
		return "", &dto.ValidationErrorsDTO{
			Message: err.Error(),
		}
	}

	user, err := uc.authrepo.CreateUser(req.Name, req.PhoneNumber, string(hashedPassword))
	if err != nil {
		wrappedErr := fmt.Errorf("%s: %w", op, err)
		log.Printf("Error: %v", wrappedErr)
		return "", &dto.ValidationErrorsDTO{
			Message: err.Error(),
		}
	}

	if user == nil {
		err = errors.New("user not created")
		wrappedErr := fmt.Errorf("%s: %w", op, err)
		log.Printf("Error: %v", wrappedErr)
		return "", &dto.ValidationErrorsDTO{
			Message: err.Error(),
		}
	}

	// Генерация JWT токена
	token, err := uc.tokenator.CreateJWT(user.ID.String())
	if err != nil {
		wrappedErr := fmt.Errorf("%s: %w", op, err)
		log.Printf("Error: %v", wrappedErr)
		return "", &dto.ValidationErrorsDTO{
			Message: err.Error(),
		}
	}

	return token, nil
}

func (uc *AuthUsecase) Login(req *AuthModels.LoginRequest) (string, error) {
	const op = "AuthUsecase.Login"
	user, err := uc.userrepo.GetUserByPhone(req.PhoneNumber)
	if err != nil || user == nil {
		wrappedErr := fmt.Errorf("%s: %w", op, errs.ErrInvalidCredentials)
		log.Printf("Error: %v", wrappedErr)
		return "", errs.ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		wrappedErr := fmt.Errorf("%s: %w", op, errs.ErrInvalidCredentials)
		log.Printf("Error: %v", wrappedErr)
		return "", errs.ErrInvalidCredentials
	}

	// Генерация JWT токена
	token, err := uc.tokenator.CreateJWT(user.ID.String())
	if err != nil {
		wrappedErr := fmt.Errorf("%s: %w", op, err)
		log.Printf("Error: %v", wrappedErr)
		return "", err
	}

	return token, nil
}

func (uc *AuthUsecase) Logout(tokenString string) error {
	const op = "AuthUsecase.Logout"
	_, err := uc.tokenator.ParseJWT(tokenString)
	if err != nil {
		err = errors.New("invalid or expired token")
		wrappedErr := fmt.Errorf("%s: %w", op, err)
		log.Printf("Error: %v", wrappedErr)
		return err
	}

	return uc.blacktoken.AddToBlacklist(tokenString)
}
