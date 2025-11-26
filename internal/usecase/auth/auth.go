package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-park-mail-ru/2025_2_Undefined/internal/models/domains"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/models/errs"
	UserModels "github.com/go-park-mail-ru/2025_2_Undefined/internal/models/user"
	AuthDTO "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/dto/auth"
	dto "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/dto/utils"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/utils/validation"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthRepository interface {
	CreateUser(ctx context.Context, name string, phone string, password_hash string) (*UserModels.User, error)
}

type UserRepository interface {
	GetUserByPhone(ctx context.Context, phone string) (*UserModels.User, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*UserModels.User, error)
}

type SessionRepository interface {
	AddSession(ctx context.Context, UserID uuid.UUID, device string) (uuid.UUID, error)
	DeleteSession(ctx context.Context, SessionID uuid.UUID) error
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

func (uc *AuthUsecase) Register(ctx context.Context, req *AuthDTO.RegisterRequest, device string) (uuid.UUID, *dto.ValidationErrorsDTO) {
	const op = "AuthUsecase.Register"

	logger := domains.GetLogger(ctx).WithField("operation", op)
	errorsValidation := make([]errs.ValidationError, 0)

	existing, _ := uc.userrepo.GetUserByPhone(ctx, req.PhoneNumber)
	if existing != nil {
		errorsValidation = append(errorsValidation, errs.ValidationError{
			Field:   "phone_number",
			Message: errs.ValidateUserAlreadyExists,
		})
	}

	if len(errorsValidation) > 0 {
		wrappedErr := fmt.Errorf("%s: %w", op, errors.New("in validation.ConvertToValidationErrorsDTO"))
		logger.WithError(wrappedErr).Error("validation errors found")
		err := validation.ConvertToValidationErrorsDTO(errorsValidation)

		return uuid.Nil, &err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		wrappedErr := fmt.Errorf("%s: %w", op, err)
		logger.WithError(wrappedErr).Error("failed to hash password")
		return uuid.Nil, &dto.ValidationErrorsDTO{
			Message: err.Error(),
		}
	}

	user, err := uc.authrepo.CreateUser(ctx, req.Name, req.PhoneNumber, string(hashedPassword))
	if err != nil {
		wrappedErr := fmt.Errorf("%s: %w", op, err)
		logger.WithError(wrappedErr).Error("failed to create user")
		return uuid.Nil, &dto.ValidationErrorsDTO{
			Message: err.Error(),
		}
	}

	if user == nil {
		err = errors.New("user not created")
		wrappedErr := fmt.Errorf("%s: %w", op, err)
		logger.WithError(wrappedErr).Error("user not created")
		return uuid.Nil, &dto.ValidationErrorsDTO{
			Message: err.Error(),
		}
	}

	newsSession, err := uc.sessionrepo.AddSession(ctx, user.ID, device)
	if err != nil {
		wrappedErr := fmt.Errorf("%s: %w", op, err)
		logger.WithError(wrappedErr).Error("failed to create session")
		return uuid.Nil, &dto.ValidationErrorsDTO{
			Message: err.Error(),
		}
	}

	return newsSession, nil
}

func (uc *AuthUsecase) Login(ctx context.Context, req *AuthDTO.LoginRequest, device string) (uuid.UUID, error) {
	const op = "AuthUsecase.Login"

	logger := domains.GetLogger(ctx).WithField("operation", op)
	user, err := uc.userrepo.GetUserByPhone(ctx, req.PhoneNumber)
	if err != nil || user == nil {
		wrappedErr := fmt.Errorf("%s: %w", op, errs.ErrInvalidCredentials)
		logger.WithError(wrappedErr).Error("user not found or database error")
		return uuid.Nil, errs.ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		wrappedErr := fmt.Errorf("%s: %w", op, errs.ErrInvalidCredentials)
		logger.WithError(wrappedErr).Error("invalid password")
		return uuid.Nil, errs.ErrInvalidCredentials
	}

	newSession, err := uc.sessionrepo.AddSession(ctx, user.ID, device)
	if err != nil {
		wrappedErr := fmt.Errorf("%s: %w", op, err)
		logger.WithError(wrappedErr).Error("failed to create session")
		return uuid.Nil, wrappedErr
	}

	return newSession, nil
}

func (uc *AuthUsecase) Logout(ctx context.Context, SessionID uuid.UUID) error {
	const op = "AuthUsecase.Logout"

	logger := domains.GetLogger(ctx)
	logger.WithField("session_id", SessionID).Debug("Starting session logout")

	err := uc.sessionrepo.DeleteSession(ctx, SessionID)
	if err != nil {
		wrappedErr := fmt.Errorf("%s: %w", op, err)
		logger.WithError(wrappedErr).Error("Failed to delete session")
		return wrappedErr
	}

	logger.Info("Session logout successful")
	return nil
}
