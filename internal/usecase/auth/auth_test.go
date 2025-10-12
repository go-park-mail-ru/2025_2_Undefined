package usecase

import (
	"errors"
	"testing"

	"github.com/go-park-mail-ru/2025_2_Undefined/internal/models/errs"
	UserModels "github.com/go-park-mail-ru/2025_2_Undefined/internal/models/user"
	AuthModels "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/dto/auth"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/jwt"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

type MockAuthRepository struct {
	CreateUserFunc        func(name string, phone string, passwordHash string) (*UserModels.User, error)
	GetUserByPhoneFunc    func(phone string) (*UserModels.User, error)
	GetUserByUsernameFunc func(username string) (*UserModels.User, error)
	GetUserByIDFunc       func(id uuid.UUID) (*UserModels.User, error)
}

func (m *MockAuthRepository) CreateUser(name string, phone string, passwordHash string) (*UserModels.User, error) {
	if m.CreateUserFunc != nil {
		return m.CreateUserFunc(name, phone, passwordHash)
	}
	return nil, nil
}

func (m *MockAuthRepository) GetUserByPhone(phone string) (*UserModels.User, error) {
	if m.GetUserByPhoneFunc != nil {
		return m.GetUserByPhoneFunc(phone)
	}
	return nil, errors.New("not found")
}

func (m *MockAuthRepository) GetUserByUsername(username string) (*UserModels.User, error) {
	if m.GetUserByUsernameFunc != nil {
		return m.GetUserByUsernameFunc(username)
	}
	return nil, errors.New("not found")
}

func (m *MockAuthRepository) GetUserByID(id uuid.UUID) (*UserModels.User, error) {
	if m.GetUserByIDFunc != nil {
		return m.GetUserByIDFunc(id)
	}
	return nil, errors.New("not found")
}

type MockTokenRepository struct {
	AddToBlacklistFunc       func(token string) error
	IsInBlacklistFunc        func(token string) bool
	CleanupExpiredTokensFunc func()
}

func (m *MockTokenRepository) AddToBlacklist(token string) error {
	if m.AddToBlacklistFunc != nil {
		return m.AddToBlacklistFunc(token)
	}
	return nil
}

func (m *MockTokenRepository) IsInBlacklist(token string) bool {
	if m.IsInBlacklistFunc != nil {
		return m.IsInBlacklistFunc(token)
	}
	return false
}

func (m *MockTokenRepository) CleanupExpiredTokens() {
	if m.CleanupExpiredTokensFunc != nil {
		m.CleanupExpiredTokensFunc()
	}
}

// FailingTokenator для тестирования ошибок создания токена
type FailingTokenator struct{}

func (ft *FailingTokenator) CreateJWT(userID string) (string, error) {
	return "", errors.New("token creation failed")
}

func (ft *FailingTokenator) ParseJWT(tokenString string) (*jwt.JWTClaims, error) {
	return nil, errors.New("parse failed")
}

func TestRegister_Success(t *testing.T) {
	mockRepo := &MockAuthRepository{
		GetUserByPhoneFunc: func(phone string) (*UserModels.User, error) {
			return nil, errors.New("not found")
		},
		CreateUserFunc: func(name string, phone string, passwordHash string) (*UserModels.User, error) {
			assert.Equal(t, "Test User", name)
			assert.Equal(t, "+79998887766", phone)
			assert.NotEmpty(t, passwordHash)

			err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte("password123"))
			assert.NoError(t, err)

			return &UserModels.User{
				ID:           uuid.New(),
				Name:         name,
				PhoneNumber:  phone,
				PasswordHash: passwordHash,
				AccountType:  UserModels.UserAccount,
			}, nil
		},
	}

	mockTokenator := jwt.NewTokenator()
	mockTokenRepo := &MockTokenRepository{}

	service := New(mockRepo, mockTokenator, mockTokenRepo)

	req := &AuthModels.RegisterRequest{
		PhoneNumber: "+79998887766",
		Password:    "password123",
		Name:        "Test User",
	}

	token, err := service.Register(req)
	assert.Nil(t, err)
	assert.NotEmpty(t, token)
}

func TestRegister_PhoneAlreadyExists(t *testing.T) {
	existingUser := &UserModels.User{
		ID:          uuid.New(),
		PhoneNumber: "+79998887766",
	}

	mockRepo := &MockAuthRepository{
		GetUserByPhoneFunc: func(phone string) (*UserModels.User, error) {
			if phone == "+79998887766" {
				return existingUser, nil
			}
			return nil, errors.New("not found")
		},
	}

	mockTokenator := jwt.NewTokenator()
	mockTokenRepo := &MockTokenRepository{}

	service := New(mockRepo, mockTokenator, mockTokenRepo)

	req := &AuthModels.RegisterRequest{
		PhoneNumber: "+79998887766",
		Password:    "password123",
		Name:        "Test User",
	}

	token, err := service.Register(req)
	assert.NotNil(t, err)
	assert.Equal(t, "Ошибка валидации", err.Message)
	assert.Len(t, err.Errors, 1)
	assert.Equal(t, "phone_number", err.Errors[0].Field)
	assert.Equal(t, "a user with such a phone already exists", err.Errors[0].Message)
	assert.Empty(t, token)
}

func TestRegister_CreateUserError(t *testing.T) {
	mockRepo := &MockAuthRepository{
		GetUserByPhoneFunc: func(phone string) (*UserModels.User, error) {
			return nil, errors.New("not found")
		},
		CreateUserFunc: func(name string, phone string, passwordHash string) (*UserModels.User, error) {
			return nil, errors.New("database error")
		},
	}

	mockTokenator := jwt.NewTokenator()
	mockTokenRepo := &MockTokenRepository{}

	service := New(mockRepo, mockTokenator, mockTokenRepo)

	req := &AuthModels.RegisterRequest{
		PhoneNumber: "+79998887766",
		Password:    "password123",
		Name:        "Test User",
	}

	token, err := service.Register(req)
	assert.NotNil(t, err)
	assert.Equal(t, "database error", err.Message)
	assert.Empty(t, token)
}

func TestRegister_UserNotCreated(t *testing.T) {
	mockRepo := &MockAuthRepository{
		GetUserByPhoneFunc: func(phone string) (*UserModels.User, error) {
			return nil, errors.New("not found")
		},
		CreateUserFunc: func(name string, phone string, passwordHash string) (*UserModels.User, error) {
			return nil, nil 
		},
	}

	mockTokenator := jwt.NewTokenator()
	mockTokenRepo := &MockTokenRepository{}

	service := New(mockRepo, mockTokenator, mockTokenRepo)

	req := &AuthModels.RegisterRequest{
		PhoneNumber: "+79998887766",
		Password:    "password123",
		Name:        "Test User",
	}

	token, err := service.Register(req)
	assert.NotNil(t, err)
	assert.Equal(t, "user not created", err.Message)
	assert.Empty(t, token)
}

func TestLogin_Success(t *testing.T) {
	userID := uuid.New()
	correctPassword := "password123"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(correctPassword), bcrypt.DefaultCost)

	user := &UserModels.User{
		ID:           userID,
		PhoneNumber:  "+79998887766",
		PasswordHash: string(hashedPassword),
	}

	mockRepo := &MockAuthRepository{
		GetUserByPhoneFunc: func(phone string) (*UserModels.User, error) {
			if phone == "+79998887766" {
				return user, nil
			}
			return nil, errors.New("not found")
		},
	}

	mockTokenator := jwt.NewTokenator()
	mockTokenRepo := &MockTokenRepository{}

	service := New(mockRepo, mockTokenator, mockTokenRepo)

	req := &AuthModels.LoginRequest{
		PhoneNumber: "+79998887766",
		Password:    correctPassword,
	}

	token, err := service.Login(req)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestLogin_UserNotFound(t *testing.T) {
	mockRepo := &MockAuthRepository{
		GetUserByPhoneFunc: func(phone string) (*UserModels.User, error) {
			return nil, errors.New("user not found")
		},
	}

	mockTokenator := jwt.NewTokenator()
	mockTokenRepo := &MockTokenRepository{}

	service := New(mockRepo, mockTokenator, mockTokenRepo)

	req := &AuthModels.LoginRequest{
		PhoneNumber: "+79990001122",
		Password:    "password123",
	}

	token, err := service.Login(req)
	assert.Error(t, err)
	assert.Equal(t, errs.ErrInvalidCredentials, err)
	assert.Empty(t, token)
}

func TestLogin_InvalidPassword(t *testing.T) {
	userID := uuid.New()
	correctPassword := "password123"
	wrongPassword := "wrongpassword"

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(correctPassword), bcrypt.DefaultCost)
	user := &UserModels.User{
		ID:           userID,
		PhoneNumber:  "+79998887766",
		PasswordHash: string(hashedPassword),
	}

	mockRepo := &MockAuthRepository{
		GetUserByPhoneFunc: func(phone string) (*UserModels.User, error) {
			if phone == "+79998887766" {
				return user, nil
			}
			return nil, errors.New("not found")
		},
	}

	mockTokenator := jwt.NewTokenator()
	mockTokenRepo := &MockTokenRepository{}

	service := New(mockRepo, mockTokenator, mockTokenRepo)

	req := &AuthModels.LoginRequest{
		PhoneNumber: "+79998887766",
		Password:    wrongPassword,
	}

	token, err := service.Login(req)
	assert.Error(t, err)
	assert.Equal(t, errs.ErrInvalidCredentials, err)
	assert.Empty(t, token)
}

func TestLogin_UserIsNil(t *testing.T) {
	mockRepo := &MockAuthRepository{
		GetUserByPhoneFunc: func(phone string) (*UserModels.User, error) {
			return nil, nil
		},
	}

	mockTokenator := jwt.NewTokenator()
	mockTokenRepo := &MockTokenRepository{}

	service := New(mockRepo, mockTokenator, mockTokenRepo)

	req := &AuthModels.LoginRequest{
		PhoneNumber: "+79998887766",
		Password:    "password123",
	}

	token, err := service.Login(req)
	assert.Error(t, err)
	assert.Equal(t, errs.ErrInvalidCredentials, err)
	assert.Empty(t, token)
}

func TestLogout_Success(t *testing.T) {
	mockRepo := &MockAuthRepository{}
	mockTokenator := jwt.NewTokenator()
	mockTokenRepo := &MockTokenRepository{
		AddToBlacklistFunc: func(token string) error {
			return nil
		},
	}

	service := New(mockRepo, mockTokenator, mockTokenRepo)

	token, _ := mockTokenator.CreateJWT(uuid.New().String())

	err := service.Logout(token)
	assert.NoError(t, err)
}

func TestLogout_InvalidToken(t *testing.T) {
	mockRepo := &MockAuthRepository{}
	mockTokenator := jwt.NewTokenator()
	mockTokenRepo := &MockTokenRepository{}

	service := New(mockRepo, mockTokenator, mockTokenRepo)

	err := service.Logout("invalid.token.string")
	assert.Error(t, err)
	assert.Equal(t, "invalid or expired token", err.Error())
}

func TestLogout_BlacklistError(t *testing.T) {
	mockRepo := &MockAuthRepository{}
	mockTokenator := jwt.NewTokenator()
	mockTokenRepo := &MockTokenRepository{
		AddToBlacklistFunc: func(token string) error {
			return errors.New("blacklist error")
		},
	}

	service := New(mockRepo, mockTokenator, mockTokenRepo)

	token, _ := mockTokenator.CreateJWT(uuid.New().String())

	err := service.Logout(token)
	assert.Error(t, err)
	assert.Equal(t, "blacklist error", err.Error())
}

func TestGetUserById_Success(t *testing.T) {
	userID := uuid.New()
	expectedUser := &UserModels.User{
		ID:          userID,
		Name:        "Test User",
		PhoneNumber: "+79998887766",
	}

	mockRepo := &MockAuthRepository{
		GetUserByIDFunc: func(id uuid.UUID) (*UserModels.User, error) {
			assert.Equal(t, userID, id)
			return expectedUser, nil
		},
	}

	mockTokenator := jwt.NewTokenator()
	mockTokenRepo := &MockTokenRepository{}

	service := New(mockRepo, mockTokenator, mockTokenRepo)

	user, err := service.GetUserById(userID)
	assert.NoError(t, err)
	assert.Equal(t, expectedUser, user)
}

func TestGetUserById_Error(t *testing.T) {
	userID := uuid.New()

	mockRepo := &MockAuthRepository{
		GetUserByIDFunc: func(id uuid.UUID) (*UserModels.User, error) {
			return nil, errors.New("user not found")
		},
	}

	mockTokenator := jwt.NewTokenator()
	mockTokenRepo := &MockTokenRepository{}

	service := New(mockRepo, mockTokenator, mockTokenRepo)

	user, err := service.GetUserById(userID)
	assert.Error(t, err)
	assert.Equal(t, "Error getting user by ID", err.Error())
	assert.Nil(t, user)
}

func TestRegister_TokenCreationError(t *testing.T) {
	mockRepo := &MockAuthRepository{
		GetUserByPhoneFunc: func(phone string) (*UserModels.User, error) {
			return nil, errors.New("not found")
		},
		CreateUserFunc: func(name string, phone string, passwordHash string) (*UserModels.User, error) {
			return &UserModels.User{
				ID:           uuid.New(),
				Name:         name,
				PhoneNumber:  phone,
				PasswordHash: passwordHash,
			}, nil
		},
	}

	failingTokenator := &FailingTokenator{}
	mockTokenRepo := &MockTokenRepository{}

	service := New(mockRepo, failingTokenator, mockTokenRepo)

	req := &AuthModels.RegisterRequest{
		PhoneNumber: "+79998887766",
		Password:    "password123",
		Name:        "Test User",
	}

	token, err := service.Register(req)
	assert.NotNil(t, err)
	assert.Equal(t, "token creation failed", err.Message)
	assert.Empty(t, token)
}

func TestLogin_TokenCreationError(t *testing.T) {
	userID := uuid.New()
	correctPassword := "password123"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(correctPassword), bcrypt.DefaultCost)

	user := &UserModels.User{
		ID:           userID,
		PhoneNumber:  "+79998887766",
		PasswordHash: string(hashedPassword),
	}

	mockRepo := &MockAuthRepository{
		GetUserByPhoneFunc: func(phone string) (*UserModels.User, error) {
			return user, nil
		},
	}

	failingTokenator := &FailingTokenator{}
	mockTokenRepo := &MockTokenRepository{}

	service := New(mockRepo, failingTokenator, mockTokenRepo)

	req := &AuthModels.LoginRequest{
		PhoneNumber: "+79998887766",
		Password:    correctPassword,
	}

	token, err := service.Login(req)
	assert.Error(t, err)
	assert.Equal(t, "token creation failed", err.Error())
	assert.Empty(t, token)
}
