package service

import (
	"errors"
	"time"

	Token "github.com/go-park-mail-ru/2025_2_Undefined/internal/handlers/jwt"
	AuthModels "github.com/go-park-mail-ru/2025_2_Undefined/internal/models/auth"
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

func NewAuthService(userRepo UserRep.UserRepository, tokenRepo Token.Tokenator, blacktokenRepo BlackToken.TokenRepository) *AuthService {
	return &AuthService{
		userRepo:   userRepo,
		token:      tokenRepo,
		blacktoken: blacktokenRepo,
	}
}

func (s *AuthService) Register(req *AuthModels.RegisterRequest) (string, error) {
	existing, _ := s.userRepo.GetByEmail(req.Email)
	if existing != nil {
		return "", errors.New("user with this email already exists")
	}
	existing, _ = s.userRepo.GetByPhone(req.PhoneNumber)
	if existing != nil {
		return "", errors.New("user with this phone already exists")
	}

	existing, _ = s.userRepo.GetByUsername(req.Username)
	if existing != nil {
		return "", errors.New("user with this username already exists")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
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
		return "", err
	}

	// Генерация JWT токена
	token, err := s.token.CreateJWT(user.ID.String())
	if err != nil {
		return "", err
	}

	return token, nil
}

func (s *AuthService) Login(req *AuthModels.LoginRequest) (string, error) {

	user, err := s.userRepo.GetByPhone(req.PhoneNumber)
	if err != nil || user == nil {
		return "", errors.New("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return "", errors.New("invalid credentials")
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
		return errors.New("invalid or expired token")
	}

	return s.blacktoken.AddToBlacklist(tokenString)
}

func (s *AuthService) GetUserById(id uuid.UUID) (*UserModels.User, error) {
	user, err := s.userRepo.GetByID(id)
	if err != nil {
		return nil, errors.New("error get user by id")
	}
	return user, nil
}
