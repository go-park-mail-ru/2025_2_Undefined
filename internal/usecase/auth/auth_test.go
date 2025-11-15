package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/go-park-mail-ru/2025_2_Undefined/internal/models/errs"
	UserModels "github.com/go-park-mail-ru/2025_2_Undefined/internal/models/user"
	AuthDTO "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/dto/auth"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

type MockAuthRepository struct {
	mock.Mock
}

func (m *MockAuthRepository) CreateUser(ctx context.Context, name string, phone string, passwordHash string) (*UserModels.User, error) {
	args := m.Called(ctx, name, phone, passwordHash)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*UserModels.User), args.Error(1)
}

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) GetUserByPhone(ctx context.Context, phone string) (*UserModels.User, error) {
	args := m.Called(ctx, phone)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*UserModels.User), args.Error(1)
}

func (m *MockUserRepository) GetUserByID(ctx context.Context, id uuid.UUID) (*UserModels.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*UserModels.User), args.Error(1)
}

type MockSessionRepository struct {
	mock.Mock
}

func (m *MockSessionRepository) AddSession(ctx context.Context, userID uuid.UUID, device string) (uuid.UUID, error) {
	args := m.Called(ctx, userID, device)
	return args.Get(0).(uuid.UUID), args.Error(1)
}

func (m *MockSessionRepository) DeleteSession(ctx context.Context, sessionID uuid.UUID) error {
	args := m.Called(ctx, sessionID)
	return args.Error(0)
}

func TestAuthUsecase_Register_Success(t *testing.T) {
	ctx := context.Background()
	mockAuthRepo := new(MockAuthRepository)
	mockUserRepo := new(MockUserRepository)
	mockSessionRepo := new(MockSessionRepository)

	uc := New(mockAuthRepo, mockUserRepo, mockSessionRepo)

	req := &AuthDTO.RegisterRequest{
		PhoneNumber: "+79998887766",
		Password:    "password123",
		Name:        "Test User",
	}
	device := "test-device"
	userID := uuid.New()
	sessionID := uuid.New()

	mockUserRepo.On("GetUserByPhone", ctx, req.PhoneNumber).Return(nil, errors.New("not found"))

	mockAuthRepo.On("CreateUser", ctx, req.Name, req.PhoneNumber, mock.AnythingOfType("string")).
		Return(&UserModels.User{
			ID:           userID,
			Name:         req.Name,
			PhoneNumber:  req.PhoneNumber,
			PasswordHash: "hashed_password",
			AccountType:  UserModels.UserAccount,
		}, nil)

	mockSessionRepo.On("AddSession", ctx, userID, device).Return(sessionID, nil)

	result, validationErr := uc.Register(ctx, req, device)

	assert.Nil(t, validationErr)
	assert.Equal(t, sessionID, result)
	mockUserRepo.AssertExpectations(t)
	mockAuthRepo.AssertExpectations(t)
	mockSessionRepo.AssertExpectations(t)
}

func TestAuthUsecase_Register_PhoneAlreadyExists(t *testing.T) {
	ctx := context.Background()
	mockAuthRepo := new(MockAuthRepository)
	mockUserRepo := new(MockUserRepository)
	mockSessionRepo := new(MockSessionRepository)

	uc := New(mockAuthRepo, mockUserRepo, mockSessionRepo)

	req := &AuthDTO.RegisterRequest{
		PhoneNumber: "+79998887766",
		Password:    "password123",
		Name:        "Test User",
	}
	device := "test-device"

	existingUser := &UserModels.User{
		ID:          uuid.New(),
		PhoneNumber: req.PhoneNumber,
	}

	mockUserRepo.On("GetUserByPhone", ctx, req.PhoneNumber).Return(existingUser, nil)

	result, validationErr := uc.Register(ctx, req, device)

	assert.Equal(t, uuid.Nil, result)
	assert.NotNil(t, validationErr)
	assert.Len(t, validationErr.Errors, 1)
	assert.Equal(t, "phone_number", validationErr.Errors[0].Field)
	assert.Equal(t, "a user with such a phone already exists", validationErr.Errors[0].Message)
	mockUserRepo.AssertExpectations(t)
}

func TestAuthUsecase_Register_CreateUserError(t *testing.T) {
	ctx := context.Background()
	mockAuthRepo := new(MockAuthRepository)
	mockUserRepo := new(MockUserRepository)
	mockSessionRepo := new(MockSessionRepository)

	uc := New(mockAuthRepo, mockUserRepo, mockSessionRepo)

	req := &AuthDTO.RegisterRequest{
		PhoneNumber: "+79998887766",
		Password:    "password123",
		Name:        "Test User",
	}

	device := "test-device"
	mockUserRepo.On("GetUserByPhone", ctx, req.PhoneNumber).Return(nil, errors.New("not found"))
	mockAuthRepo.On("CreateUser", ctx, req.Name, req.PhoneNumber, mock.AnythingOfType("string")).
		Return(nil, errors.New("database error"))

	result, validationErr := uc.Register(ctx, req, device)

	assert.Equal(t, uuid.Nil, result)
	assert.NotNil(t, validationErr)
	assert.Equal(t, "database error", validationErr.Message)
	mockUserRepo.AssertExpectations(t)
	mockAuthRepo.AssertExpectations(t)
}

func TestAuthUsecase_Register_UserNotCreated(t *testing.T) {
	ctx := context.Background()
	mockAuthRepo := new(MockAuthRepository)
	mockUserRepo := new(MockUserRepository)
	mockSessionRepo := new(MockSessionRepository)

	uc := New(mockAuthRepo, mockUserRepo, mockSessionRepo)

	req := &AuthDTO.RegisterRequest{
		PhoneNumber: "+79998887766",
		Password:    "password123",
		Name:        "Test User",
	}
	device := "test-device"

	mockUserRepo.On("GetUserByPhone", ctx, req.PhoneNumber).Return(nil, errors.New("not found"))
	mockAuthRepo.On("CreateUser", ctx, req.Name, req.PhoneNumber, mock.AnythingOfType("string")).
		Return(nil, nil)

	result, validationErr := uc.Register(ctx, req, device)

	assert.Equal(t, uuid.Nil, result)
	assert.NotNil(t, validationErr)
	assert.Equal(t, "user not created", validationErr.Message)
	mockUserRepo.AssertExpectations(t)
	mockAuthRepo.AssertExpectations(t)
}

func TestAuthUsecase_Register_SessionCreationError(t *testing.T) {
	ctx := context.Background()
	mockAuthRepo := new(MockAuthRepository)
	mockUserRepo := new(MockUserRepository)
	mockSessionRepo := new(MockSessionRepository)

	uc := New(mockAuthRepo, mockUserRepo, mockSessionRepo)

	req := &AuthDTO.RegisterRequest{
		PhoneNumber: "+79998887766",
		Password:    "password123",
		Name:        "Test User",
	}
	device := "test-device"
	userID := uuid.New()

	mockUserRepo.On("GetUserByPhone", ctx, req.PhoneNumber).Return(nil, errors.New("not found"))

	mockAuthRepo.On("CreateUser", ctx, req.Name, req.PhoneNumber, mock.AnythingOfType("string")).
		Return(&UserModels.User{
			ID:           userID,
			Name:         req.Name,
			PhoneNumber:  req.PhoneNumber,
			PasswordHash: "hashed_password",
			AccountType:  UserModels.UserAccount,
		}, nil)

	mockSessionRepo.On("AddSession", ctx, userID, device).Return(uuid.Nil, errors.New("session creation failed"))

	result, validationErr := uc.Register(ctx, req, device)

	assert.Equal(t, uuid.Nil, result)
	assert.NotNil(t, validationErr)
	assert.Equal(t, "session creation failed", validationErr.Message)
	mockUserRepo.AssertExpectations(t)
	mockAuthRepo.AssertExpectations(t)
	mockSessionRepo.AssertExpectations(t)
}

func TestAuthUsecase_Login_Success(t *testing.T) {
	ctx := context.Background()
	mockAuthRepo := new(MockAuthRepository)
	mockUserRepo := new(MockUserRepository)
	mockSessionRepo := new(MockSessionRepository)

	uc := New(mockAuthRepo, mockUserRepo, mockSessionRepo)

	req := &AuthDTO.LoginRequest{
		PhoneNumber: "+79998887766",
		Password:    "password123",
	}
	device := "test-device"
	userID := uuid.New()
	sessionID := uuid.New()

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)

	user := &UserModels.User{
		ID:           userID,
		PhoneNumber:  req.PhoneNumber,
		PasswordHash: string(hashedPassword),
	}

	mockUserRepo.On("GetUserByPhone", ctx, req.PhoneNumber).Return(user, nil)
	mockSessionRepo.On("AddSession", ctx, userID, device).Return(sessionID, nil)

	result, err := uc.Login(ctx, req, device)

	assert.NoError(t, err)
	assert.Equal(t, sessionID, result)
	mockUserRepo.AssertExpectations(t)
	mockSessionRepo.AssertExpectations(t)
}

func TestAuthUsecase_Login_UserNotFound(t *testing.T) {
	ctx := context.Background()
	mockAuthRepo := new(MockAuthRepository)
	mockUserRepo := new(MockUserRepository)
	mockSessionRepo := new(MockSessionRepository)

	uc := New(mockAuthRepo, mockUserRepo, mockSessionRepo)

	req := &AuthDTO.LoginRequest{
		PhoneNumber: "+79998887766",
		Password:    "password123",
	}
	device := "test-device"

	mockUserRepo.On("GetUserByPhone", ctx, req.PhoneNumber).Return(nil, errors.New("user not found"))

	result, err := uc.Login(ctx, req, device)

	assert.Equal(t, uuid.Nil, result)
	assert.Error(t, err)
	assert.Equal(t, errs.ErrInvalidCredentials, err)
	mockUserRepo.AssertExpectations(t)
}

func TestAuthUsecase_Login_UserIsNil(t *testing.T) {
	ctx := context.Background()
	mockAuthRepo := new(MockAuthRepository)
	mockUserRepo := new(MockUserRepository)
	mockSessionRepo := new(MockSessionRepository)

	uc := New(mockAuthRepo, mockUserRepo, mockSessionRepo)

	req := &AuthDTO.LoginRequest{
		PhoneNumber: "+79998887766",
		Password:    "password123",
	}
	device := "test-device"

	mockUserRepo.On("GetUserByPhone", ctx, req.PhoneNumber).Return(nil, nil)

	result, err := uc.Login(ctx, req, device)

	assert.Equal(t, uuid.Nil, result)
	assert.Error(t, err)
	assert.Equal(t, errs.ErrInvalidCredentials, err)
	mockUserRepo.AssertExpectations(t)
}

func TestAuthUsecase_Login_InvalidPassword(t *testing.T) {
	ctx := context.Background()
	mockAuthRepo := new(MockAuthRepository)
	mockUserRepo := new(MockUserRepository)
	mockSessionRepo := new(MockSessionRepository)

	uc := New(mockAuthRepo, mockUserRepo, mockSessionRepo)

	req := &AuthDTO.LoginRequest{
		PhoneNumber: "+79998887766",
		Password:    "wrongpassword",
	}
	device := "test-device"
	userID := uuid.New()

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("correctpassword"), bcrypt.DefaultCost)

	user := &UserModels.User{
		ID:           userID,
		PhoneNumber:  req.PhoneNumber,
		PasswordHash: string(hashedPassword),
	}

	mockUserRepo.On("GetUserByPhone", ctx, req.PhoneNumber).Return(user, nil)

	result, err := uc.Login(ctx, req, device)

	assert.Equal(t, uuid.Nil, result)
	assert.Error(t, err)
	assert.Equal(t, errs.ErrInvalidCredentials, err)
	mockUserRepo.AssertExpectations(t)
}

func TestAuthUsecase_Login_SessionCreationError(t *testing.T) {
	ctx := context.Background()
	mockAuthRepo := new(MockAuthRepository)
	mockUserRepo := new(MockUserRepository)
	mockSessionRepo := new(MockSessionRepository)

	uc := New(mockAuthRepo, mockUserRepo, mockSessionRepo)

	req := &AuthDTO.LoginRequest{
		PhoneNumber: "+79998887766",
		Password:    "password123",
	}
	device := "test-device"
	userID := uuid.New()

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)

	user := &UserModels.User{
		ID:           userID,
		PhoneNumber:  req.PhoneNumber,
		PasswordHash: string(hashedPassword),
	}

	mockUserRepo.On("GetUserByPhone", ctx, req.PhoneNumber).Return(user, nil)
	mockSessionRepo.On("AddSession", ctx, userID, device).Return(uuid.Nil, errors.New("session creation failed"))

	result, err := uc.Login(ctx, req, device)

	assert.Equal(t, uuid.Nil, result)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "session creation failed")
	mockUserRepo.AssertExpectations(t)
	mockSessionRepo.AssertExpectations(t)
}

func TestAuthUsecase_Logout_Success(t *testing.T) {
	ctx := context.Background()
	mockAuthRepo := new(MockAuthRepository)
	mockUserRepo := new(MockUserRepository)
	mockSessionRepo := new(MockSessionRepository)

	uc := New(mockAuthRepo, mockUserRepo, mockSessionRepo)

	sessionID := uuid.New()

	mockSessionRepo.On("DeleteSession", ctx, sessionID).Return(nil)

	err := uc.Logout(ctx, sessionID)

	assert.NoError(t, err)
	mockSessionRepo.AssertExpectations(t)
}

func TestAuthUsecase_Logout_SessionDeleteError(t *testing.T) {
	ctx := context.Background()
	mockAuthRepo := new(MockAuthRepository)
	mockUserRepo := new(MockUserRepository)
	mockSessionRepo := new(MockSessionRepository)

	uc := New(mockAuthRepo, mockUserRepo, mockSessionRepo)

	sessionID := uuid.New()

	mockSessionRepo.On("DeleteSession", ctx, sessionID).Return(errors.New("delete session failed"))

	err := uc.Logout(ctx, sessionID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "delete session failed")
	mockSessionRepo.AssertExpectations(t)
}
