package usecase

import (
	"errors"
	"fmt"
	"log"

	"github.com/go-park-mail-ru/2025_2_Undefined/internal/models/errs"
	UserModels "github.com/go-park-mail-ru/2025_2_Undefined/internal/models/user"
	AuthModels "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/dto/auth"
	dto "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/dto/utils"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/utils/validation"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthRepository interface {
	CreateUser(name string, phone string, password_hash string) (*UserModels.User, error)
}

type UserRepository interface {
	GetUserByPhone(phone string) (*UserModels.User, error)
	GetUserByID(id uuid.UUID) (*UserModels.User, error)
}

type SessionRepository interface {
	AddSession(UserID uuid.UUID, device string) (uuid.UUID, error)
	DeleteSession(SessionID uuid.UUID) error
}

type AuthUsecase struct {
	authrepo    AuthRepository
	userrepo    UserRepository
	sessionrepo SessionRepository
}

func New(authrepo AuthRepository, userrepo UserRepository, sessionrepo SessionRepository) *AuthUsecase {
	return &AuthUsecase{
		authrepo:    authrepo,
		userrepo:    userrepo,
		sessionrepo: sessionrepo,
	}
}

func (uc *AuthUsecase) Register(req *AuthModels.RegisterRequest, device string) (uuid.UUID, *dto.ValidationErrorsDTO) {
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

		return uuid.Nil, &err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		wrappedErr := fmt.Errorf("%s: %w", op, err)
		log.Printf("Error: %v", wrappedErr)
		return uuid.Nil, &dto.ValidationErrorsDTO{
			Message: err.Error(),
		}
	}

	user, err := uc.authrepo.CreateUser(req.Name, req.PhoneNumber, string(hashedPassword))
	if err != nil {
		wrappedErr := fmt.Errorf("%s: %w", op, err)
		log.Printf("Error: %v", wrappedErr)
		return uuid.Nil, &dto.ValidationErrorsDTO{
			Message: err.Error(),
		}
	}

	if user == nil {
		err = errors.New("user not created")
		wrappedErr := fmt.Errorf("%s: %w", op, err)
		log.Printf("Error: %v", wrappedErr)
		return uuid.Nil, &dto.ValidationErrorsDTO{
			Message: err.Error(),
		}
	}

	newsSession, err := uc.sessionrepo.AddSession(user.ID, device)
	if err != nil {
		wrappedErr := fmt.Errorf("%s: %w", op, err)
		log.Printf("Error: %v", wrappedErr)
		return uuid.Nil, &dto.ValidationErrorsDTO{
			Message: err.Error(),
		}
	}

	return newsSession, nil
}

func (uc *AuthUsecase) Login(req *AuthModels.LoginRequest, device string) (uuid.UUID, error) {
	const op = "AuthUsecase.Login"
	user, err := uc.userrepo.GetUserByPhone(req.PhoneNumber)
	if err != nil || user == nil {
		wrappedErr := fmt.Errorf("%s: %w", op, errs.ErrInvalidCredentials)
		log.Printf("Error: %v", wrappedErr)
		return uuid.Nil, errs.ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		wrappedErr := fmt.Errorf("%s: %w", op, errs.ErrInvalidCredentials)
		log.Printf("Error: %v", wrappedErr)
		return uuid.Nil, errs.ErrInvalidCredentials
	}

	newSession, err := uc.sessionrepo.AddSession(user.ID, device)
	if err != nil {
		wrappedErr := fmt.Errorf("%s: %w", op, err)
		log.Printf("Error: %v", wrappedErr)
		return uuid.Nil, wrappedErr
	}

	return newSession, nil
}

func (uc *AuthUsecase) Logout(SessionID uuid.UUID) error {
	const op = "AuthUsecase.Logout"
	err := uc.sessionrepo.DeleteSession(SessionID)
	if err != nil {
		wrappedErr := fmt.Errorf("%s: %w", op, err)
		log.Printf("Error: %v", wrappedErr)
		return wrappedErr
	}

	return nil
}
