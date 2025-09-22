package service

import (
	"errors"
	"time"

	AuthModels "github.com/go-park-mail-ru/2025_2_Undefined/internal/models/auth"
	UserModels "github.com/go-park-mail-ru/2025_2_Undefined/internal/models/user"
	TokenRep "github.com/go-park-mail-ru/2025_2_Undefined/internal/repository/token"
	UserRep "github.com/go-park-mail-ru/2025_2_Undefined/internal/repository/user"
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	userRepo  UserRep.UserRepository
	tokenRepo TokenRep.TokenRepository
	jwtSecret string
	jwtTTL    time.Duration
}

func NewAuthService(userRepo UserRep.UserRepository, tokenRepo TokenRep.TokenRepository, jwtSecret string, jwtTTL time.Duration) *AuthService {
	return &AuthService{
		userRepo:  userRepo,
		tokenRepo: tokenRepo,
		jwtSecret: jwtSecret,
		jwtTTL:    jwtTTL,
	}
}

func (s *AuthService) Register(req *AuthModels.RegisterRequest) (*AuthModels.AuthResponse, error) {
	existing, _ := s.userRepo.GetByEmail(req.Email)
	if existing != nil {
		return nil, errors.New("user with this email already exists")
	}
	existing, _ = s.userRepo.GetByPhone(req.PhoneNumber)
	if existing != nil {
		return nil, errors.New("user with this phone already exists")
	}

	existing, _ = s.userRepo.GetByUsername(req.Username)
	if existing != nil {
		return nil, errors.New("user with this username already exists")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
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
		return nil, err
	}

	// Генерация JWT токена
	token, err := s.generateJWT(user)
	if err != nil {
		return nil, err
	}

	return &AuthModels.AuthResponse{
		Token: token,
		User: UserModels.PublicUser{
			ID:          user.ID,
			Name:        user.Name,
			Username:    user.Username,
			AccountType: user.AccountType,
			CreatedAt:   user.CreatedAt,
		},
	}, nil
}

func (s *AuthService) Login(req *AuthModels.LoginRequest) (*AuthModels.AuthResponse, error) {

	user, err := s.userRepo.GetByEmail(req.Email)
	if err != nil || user == nil {
		return nil, errors.New("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, errors.New("invalid credentials")
	}

	// Генерация JWT токена
	token, err := s.generateJWT(user)
	if err != nil {
		return nil, err
	}

	return &AuthModels.AuthResponse{
		Token: token,
		User: UserModels.PublicUser{
			ID:          user.ID,
			Name:        user.Name,
			Username:    user.Username,
			AccountType: user.AccountType,
			CreatedAt:   user.CreatedAt,
		},
	}, nil
}

func (s *AuthService) Logout(tokenString string) error {
	_, err := s.ValidateToken(tokenString)
	if err != nil {
		return errors.New("invalid or expired token")
	}

	if s.tokenRepo.IsInBlacklist(tokenString) {
		return errors.New("token already invalidated")
	}

	return s.tokenRepo.AddToBlacklist(tokenString)
}

func (s *AuthService) generateJWT(user *UserModels.User) (string, error) {
	expiresAt := time.Now().Add(s.jwtTTL)

	claims := jwt.MapClaims{
		"user_id": user.ID.String(),
		"exp":     expiresAt.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}

func (s *AuthService) ValidateToken(tokenString string) (*UserModels.User, error) {
	if s.tokenRepo.IsInBlacklist(tokenString) {
		return nil, errors.New("token is blacklisted")
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.jwtSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userID := claims["user_id"].(string)
		return s.userRepo.GetByID(userID)
	}

	return nil, errors.New("invalid token")
}
