package usecase

import (
	"errors"
	"time"

	"github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/dto"
	Token "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/jwt"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/utils/validation"
	AuthModels "github.com/go-park-mail-ru/2025_2_Undefined/internal/models/auth"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/models/errs"
	UserModels "github.com/go-park-mail-ru/2025_2_Undefined/internal/models/user"
	BlackToken "github.com/go-park-mail-ru/2025_2_Undefined/internal/repository/token"
	UserRep "github.com/go-park-mail-ru/2025_2_Undefined/internal/repository/user"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	userRepo   UserRep.UserRepository
	token      Token.Tokenator
	blacktoken BlackToken.TokenRepository
}

func NewAuthService(userRepo UserRep.UserRepository, tokenRepo *Token.Tokenator, blacktokenRepo BlackToken.TokenRepository) *AuthService {
	return &AuthService{
		userRepo:   userRepo,
		token:      *tokenRepo,
		blacktoken: blacktokenRepo,
	}
}

func (s *AuthService) Register(req *AuthModels.RegisterRequest) (string, *dto.ValidationErrorsDTO) {
	errorsValidation := make([]errs.ValidationError, 0)

	existing, _ := s.userRepo.GetByEmail(req.Email)
	if existing != nil {
		errorsValidation = append(errorsValidation, errs.ValidationError{
			Field:   "email",
			Message: "пользователь с таким email уже существует",
		})
	}
	existing, _ = s.userRepo.GetByPhone(req.PhoneNumber)
	if existing != nil {
		errorsValidation = append(errorsValidation, errs.ValidationError{
			Field:   "phone_number",
			Message: "пользователь с таким телефоном уже существует",
		})
	}

	existing, _ = s.userRepo.GetByUsername(req.Username)
	if existing != nil {
		errorsValidation = append(errorsValidation, errs.ValidationError{
			Field:   "username",
			Message: "пользователь с таким именем пользователя уже существует",
		})
	}

	if len(errorsValidation) > 0 {
		err := validation.ConvertToValidationErrorsDTO(errorsValidation)
		return "", &err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return "", &dto.ValidationErrorsDTO{
			Message: err.Error(),
		}
	}

	user := &UserModels.User{
		ID:           uuid.New(),
		PhoneNumber:  req.PhoneNumber,
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		Name:         req.Name,
		Username:     req.Username,
		AccountType:  UserModels.UserAccount,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := s.userRepo.Create(user); err != nil {
		return "", &dto.ValidationErrorsDTO{
			Message: err.Error(),
		}
	}

	// Генерация JWT токена
	token, err := s.token.CreateJWT(user.ID.String())
	if err != nil {
		return "", &dto.ValidationErrorsDTO{
			Message: err.Error(),
		}
	}

	return token, nil
}

func (s *AuthService) Login(req *AuthModels.LoginRequest) (string, error) {

	user, err := s.userRepo.GetByPhone(req.PhoneNumber)
	if err != nil || user == nil {
		return "", errors.New("неверные учетные данные")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return "", errors.New("неверные учетные данные")
	}

	// Генерация JWT токена
	token, err := s.token.CreateJWT(user.ID.String())
	if err != nil {
		return "", err
	}

	return token, nil
}

func (s *AuthService) Logout(tokenString string) error {
	_, err := s.token.ParseJWT(tokenString)
	if err != nil {
		return errors.New("недействительный или истекший токен")
	}

	return s.blacktoken.AddToBlacklist(tokenString)
}

func (s *AuthService) GetUserById(id uuid.UUID) (*UserModels.User, error) {
	user, err := s.userRepo.GetByID(id)
	if err != nil {
		return nil, errors.New("ошибка получения пользователя по id")
	}
	return user, nil
}
