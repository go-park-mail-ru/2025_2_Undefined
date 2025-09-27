package service

import (
	"errors"
	"testing"

	"github.com/go-park-mail-ru/2025_2_Undefined/internal/handlers/jwt"
	AuthModels "github.com/go-park-mail-ru/2025_2_Undefined/internal/models/auth"
	UserModels "github.com/go-park-mail-ru/2025_2_Undefined/internal/models/user"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

type MockUserRepository struct {
	CreateFunc        func(user *UserModels.User) error
	GetByIDFunc       func(id uuid.UUID) (*UserModels.User, error)
	GetByPhoneFunc    func(phone string) (*UserModels.User, error)
	GetByEmailFunc    func(email string) (*UserModels.User, error)
	GetByUsernameFunc func(username string) (*UserModels.User, error)
	UpdateFunc        func(user *UserModels.User) error
}

func (m *MockUserRepository) Create(user *UserModels.User) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(user)
	}
	return nil
}

func (m *MockUserRepository) GetByID(id uuid.UUID) (*UserModels.User, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(id)
	}
	return nil, errors.New("not found")
}

func (m *MockUserRepository) GetByPhone(phone string) (*UserModels.User, error) {
	if m.GetByPhoneFunc != nil {
		return m.GetByPhoneFunc(phone)
	}
	return nil, errors.New("not found")
}

func (m *MockUserRepository) GetByEmail(email string) (*UserModels.User, error) {
	if m.GetByEmailFunc != nil {
		return m.GetByEmailFunc(email)
	}
	return nil, errors.New("not found")
}

func (m *MockUserRepository) GetByUsername(username string) (*UserModels.User, error) {
	if m.GetByUsernameFunc != nil {
		return m.GetByUsernameFunc(username)
	}
	return nil, errors.New("not found")
}

func (m *MockUserRepository) Update(user *UserModels.User) error {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(user)
	}
	return nil
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

func TestRegister_Success(t *testing.T) {
	mockUserRepo := &MockUserRepository{
		GetByEmailFunc: func(email string) (*UserModels.User, error) {
			return nil, errors.New("not found")
		},
		GetByPhoneFunc: func(phone string) (*UserModels.User, error) {
			return nil, errors.New("not found")
		},
		GetByUsernameFunc: func(username string) (*UserModels.User, error) {
			return nil, errors.New("not found")
		},
		CreateFunc: func(user *UserModels.User) error {
			assert.Equal(t, "+79998887766", user.PhoneNumber)
			assert.Equal(t, "test@mail.ru", user.Email)
			assert.Equal(t, "testuser", user.Username)
			assert.Equal(t, "Test User", user.Name)
			assert.Equal(t, UserModels.UserAccount, user.AccountType)
			assert.NotEmpty(t, user.PasswordHash)
			return nil
		},
	}

	mockTokenator := jwt.NewTokenator()
	mockTokenRepo := &MockTokenRepository{}

	service := NewAuthService(mockUserRepo, mockTokenator, mockTokenRepo)

	req := &AuthModels.RegisterRequest{
		PhoneNumber: "+79998887766",
		Email:       "test@mail.ru",
		Password:    "password123",
		Name:        "Test User",
		Username:    "testuser",
	}

	token, err := service.Register(req)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestRegister_EmailAlreadyExists(t *testing.T) {
	existingUser := &UserModels.User{
		ID:    uuid.New(),
		Email: "existing@mail.ru",
	}

	mockUserRepo := &MockUserRepository{
		GetByEmailFunc: func(email string) (*UserModels.User, error) {
			if email == "existing@mail.ru" {
				return existingUser, nil
			}
			return nil, errors.New("not found")
		},
	}

	mockTokenator := jwt.NewTokenator()
	mockTokenRepo := &MockTokenRepository{}

	service := NewAuthService(mockUserRepo, mockTokenator, mockTokenRepo)

	req := &AuthModels.RegisterRequest{
		PhoneNumber: "+79998887766",
		Email:       "existing@mail.ru",
		Password:    "password123",
		Name:        "Test User",
		Username:    "testuser",
	}

	token, err := service.Register(req)
	assert.Error(t, err)
	assert.Equal(t, "user with this email already exists", err.Error())
	assert.Empty(t, token)
}

func TestRegister_PhoneAlreadyExists(t *testing.T) {
	existingUser := &UserModels.User{
		ID:          uuid.New(),
		PhoneNumber: "+79998887766",
	}

	mockUserRepo := &MockUserRepository{
		GetByEmailFunc: func(email string) (*UserModels.User, error) {
			return nil, errors.New("not found")
		},
		GetByPhoneFunc: func(phone string) (*UserModels.User, error) {
			if phone == "+79998887766" {
				return existingUser, nil
			}
			return nil, errors.New("not found")
		},
	}

	mockTokenator := jwt.NewTokenator()
	mockTokenRepo := &MockTokenRepository{}

	service := NewAuthService(mockUserRepo, mockTokenator, mockTokenRepo)

	req := &AuthModels.RegisterRequest{
		PhoneNumber: "+79998887766",
		Email:       "test@mail.ru",
		Password:    "password123",
		Name:        "Test User",
		Username:    "testuser",
	}

	token, err := service.Register(req)
	assert.Error(t, err)
	assert.Equal(t, "user with this phone already exists", err.Error())
	assert.Empty(t, token)
}

func TestRegister_UsernameAlreadyExists(t *testing.T) {
	existingUser := &UserModels.User{
		ID:       uuid.New(),
		Username: "existinguser",
	}

	mockUserRepo := &MockUserRepository{
		GetByEmailFunc: func(email string) (*UserModels.User, error) {
			return nil, errors.New("not found")
		},
		GetByPhoneFunc: func(phone string) (*UserModels.User, error) {
			return nil, errors.New("not found")
		},
		GetByUsernameFunc: func(username string) (*UserModels.User, error) {
			if username == "existinguser" {
				return existingUser, nil
			}
			return nil, errors.New("not found")
		},
	}

	mockTokenator := jwt.NewTokenator()
	mockTokenRepo := &MockTokenRepository{}

	service := NewAuthService(mockUserRepo, mockTokenator, mockTokenRepo)

	req := &AuthModels.RegisterRequest{
		PhoneNumber: "+79998887766",
		Email:       "test@mail.ru",
		Password:    "password123",
		Name:        "Test User",
		Username:    "existinguser",
	}

	token, err := service.Register(req)
	assert.Error(t, err)
	assert.Equal(t, "user with this username already exists", err.Error())
	assert.Empty(t, token)
}

func TestRegister_CreateUserError(t *testing.T) {
	mockUserRepo := &MockUserRepository{
		GetByEmailFunc: func(email string) (*UserModels.User, error) {
			return nil, errors.New("not found")
		},
		GetByPhoneFunc: func(phone string) (*UserModels.User, error) {
			return nil, errors.New("not found")
		},
		GetByUsernameFunc: func(username string) (*UserModels.User, error) {
			return nil, errors.New("not found")
		},
		CreateFunc: func(user *UserModels.User) error {
			return errors.New("database error")
		},
	}

	mockTokenator := jwt.NewTokenator()
	mockTokenRepo := &MockTokenRepository{}

	service := NewAuthService(mockUserRepo, mockTokenator, mockTokenRepo)

	req := &AuthModels.RegisterRequest{
		PhoneNumber: "+79998887766",
		Email:       "test@mail.ru",
		Password:    "password123",
		Name:        "Test User",
		Username:    "testuser",
	}

	token, err := service.Register(req)
	assert.Error(t, err)
	assert.Empty(t, token)
}

func TestLogin_Success(t *testing.T) {
	correctPassword := "password123"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(correctPassword), bcrypt.DefaultCost)

	user := &UserModels.User{
		ID:           uuid.New(),
		PhoneNumber:  "+79998887766",
		PasswordHash: string(hashedPassword),
	}

	mockUserRepo := &MockUserRepository{
		GetByPhoneFunc: func(phone string) (*UserModels.User, error) {
			if phone == "+79998887766" {
				return user, nil
			}
			return nil, errors.New("not found")
		},
	}

	mockTokenator := jwt.NewTokenator()
	mockTokenRepo := &MockTokenRepository{}

	service := NewAuthService(mockUserRepo, mockTokenator, mockTokenRepo)

	req := &AuthModels.LoginRequest{
		PhoneNumber: "+79998887766",
		Password:    correctPassword,
	}

	token, err := service.Login(req)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestLogin_UserNotFound(t *testing.T) {
	mockUserRepo := &MockUserRepository{
		GetByPhoneFunc: func(phone string) (*UserModels.User, error) {
			return nil, errors.New("user not found")
		},
	}

	mockTokenator := jwt.NewTokenator()
	mockTokenRepo := &MockTokenRepository{}

	service := NewAuthService(mockUserRepo, mockTokenator, mockTokenRepo)

	req := &AuthModels.LoginRequest{
		PhoneNumber: "+79990001122",
		Password:    "password123",
	}

	token, err := service.Login(req)
	assert.Error(t, err)
	assert.Equal(t, "invalid credentials", err.Error())
	assert.Empty(t, token)
}

func TestLogin_InvalidPassword(t *testing.T) {
	correctPassword := "password123"
	wrongPassword := "wrongpassword"

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(correctPassword), bcrypt.DefaultCost)
	user := &UserModels.User{
		ID:           uuid.New(),
		PhoneNumber:  "+79998887766",
		PasswordHash: string(hashedPassword),
	}

	mockUserRepo := &MockUserRepository{
		GetByPhoneFunc: func(phone string) (*UserModels.User, error) {
			if phone == "+79998887766" {
				return user, nil
			}
			return nil, errors.New("not found")
		},
	}

	mockTokenator := jwt.NewTokenator()
	mockTokenRepo := &MockTokenRepository{}

	service := NewAuthService(mockUserRepo, mockTokenator, mockTokenRepo)

	req := &AuthModels.LoginRequest{
		PhoneNumber: "+79998887766",
		Password:    wrongPassword,
	}

	token, err := service.Login(req)
	assert.Error(t, err)
	assert.Equal(t, "invalid credentials", err.Error())
	assert.Empty(t, token)
}

func TestLogout_Success(t *testing.T) {
	mockUserRepo := &MockUserRepository{}
	mockTokenator := jwt.NewTokenator()
	mockTokenRepo := &MockTokenRepository{
		AddToBlacklistFunc: func(token string) error {
			return nil
		},
	}

	service := NewAuthService(mockUserRepo, mockTokenator, mockTokenRepo)

	token, _ := mockTokenator.CreateJWT(uuid.New().String())

	err := service.Logout(token)
	assert.NoError(t, err)
}

func TestLogout_InvalidToken(t *testing.T) {
	mockUserRepo := &MockUserRepository{}
	mockTokenator := jwt.NewTokenator()
	mockTokenRepo := &MockTokenRepository{}

	service := NewAuthService(mockUserRepo, mockTokenator, mockTokenRepo)

	err := service.Logout("invalid.token.string")
	assert.Error(t, err)
	assert.Equal(t, "invalid or expired token", err.Error())
}

func TestGetUserById_Success(t *testing.T) {
	userID := uuid.New()
	expectedUser := &UserModels.User{
		ID:       userID,
		Username: "testuser",
		Email:    "test@mail.ru",
	}

	mockUserRepo := &MockUserRepository{
		GetByIDFunc: func(id uuid.UUID) (*UserModels.User, error) {
			assert.Equal(t, userID, id)
			return expectedUser, nil
		},
	}

	mockTokenator := jwt.NewTokenator()
	mockTokenRepo := &MockTokenRepository{}

	service := NewAuthService(mockUserRepo, mockTokenator, mockTokenRepo)

	user, err := service.GetUserById(userID)
	assert.NoError(t, err)
	assert.Equal(t, expectedUser, user)
}

func TestGetUserById_Error(t *testing.T) {
	userID := uuid.New()

	mockUserRepo := &MockUserRepository{
		GetByIDFunc: func(id uuid.UUID) (*UserModels.User, error) {
			return nil, errors.New("user not found")
		},
	}

	mockTokenator := jwt.NewTokenator()
	mockTokenRepo := &MockTokenRepository{}

	service := NewAuthService(mockUserRepo, mockTokenator, mockTokenRepo)

	user, err := service.GetUserById(userID)
	assert.Error(t, err)
	assert.Equal(t, "error get user by id", err.Error())
	assert.Nil(t, user)
}

func TestRegister_PasswordHashing(t *testing.T) {
	mockUserRepo := &MockUserRepository{
		GetByEmailFunc: func(email string) (*UserModels.User, error) {
			return nil, errors.New("not found")
		},
		GetByPhoneFunc: func(phone string) (*UserModels.User, error) {
			return nil, errors.New("not found")
		},
		GetByUsernameFunc: func(username string) (*UserModels.User, error) {
			return nil, errors.New("not found")
		},
		CreateFunc: func(user *UserModels.User) error {
			err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte("password123"))
			assert.NoError(t, err)
			return nil
		},
	}

	mockTokenator := jwt.NewTokenator()
	mockTokenRepo := &MockTokenRepository{}

	service := NewAuthService(mockUserRepo, mockTokenator, mockTokenRepo)

	req := &AuthModels.RegisterRequest{
		PhoneNumber: "+79998887766",
		Email:       "test@mail.ru",
		Password:    "password123",
		Name:        "Test User",
		Username:    "testuser",
	}

	_, err := service.Register(req)
	assert.NoError(t, err)
}
