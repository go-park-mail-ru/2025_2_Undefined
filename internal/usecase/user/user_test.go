package user

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2025_2_Undefined/internal/models/errs"
	UserModels "github.com/go-park-mail-ru/2025_2_Undefined/internal/models/user"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) GetUserByID(ctx context.Context, id uuid.UUID) (*UserModels.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*UserModels.User), args.Error(1)
}

func (m *MockUserRepository) GetUserByPhone(ctx context.Context, phone string) (*UserModels.User, error) {
	args := m.Called(ctx, phone)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*UserModels.User), args.Error(1)
}

func (m *MockUserRepository) GetUserByUsername(ctx context.Context, username string) (*UserModels.User, error) {
	args := m.Called(ctx, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*UserModels.User), args.Error(1)
}

func TestUserUsecase_GetUserById_Success(t *testing.T) {
	mockRepo := new(MockUserRepository)
	uc := New(mockRepo)

	ctx := context.Background()
	userID := uuid.New()

	user := &UserModels.User{
		ID:          userID,
		PhoneNumber: "+79998887766",
		Name:        "Test User",
		Username:    "test_user",
		Bio:         "Test bio",
		Avatar:      nil,
		AccountType: UserModels.UserAccount,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	mockRepo.On("GetUserByID", ctx, userID).Return(user, nil)

	result, err := uc.GetUserById(ctx, userID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, userID, result.ID)
	assert.Equal(t, user.PhoneNumber, result.PhoneNumber)
	assert.Equal(t, user.Name, result.Name)
	assert.Equal(t, user.Username, result.Username)
	assert.Equal(t, user.Bio, result.Bio)
	assert.Equal(t, user.AccountType, result.AccountType)
	mockRepo.AssertExpectations(t)
}

func TestUserUsecase_GetUserById_NotFound(t *testing.T) {
	mockRepo := new(MockUserRepository)
	uc := New(mockRepo)

	ctx := context.Background()
	userID := uuid.New()

	mockRepo.On("GetUserByID", ctx, userID).Return(nil, errors.New("user not found"))

	result, err := uc.GetUserById(ctx, userID)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, "Error getting user by ID", err.Error())
	mockRepo.AssertExpectations(t)
}

func TestUserUsecase_GetUserByPhone_Success(t *testing.T) {
	mockRepo := new(MockUserRepository)
	uc := New(mockRepo)

	ctx := context.Background()
	phone := "+79998887766"

	user := &UserModels.User{
		ID:          uuid.New(),
		PhoneNumber: phone,
		Name:        "Test User",
		Username:    "test_user",
		Bio:         "Test bio",
		Avatar:      nil,
		AccountType: UserModels.UserAccount,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	mockRepo.On("GetUserByPhone", ctx, phone).Return(user, nil)

	result, err := uc.GetUserByPhone(ctx, phone)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, user.ID, result.ID)
	assert.Equal(t, phone, result.PhoneNumber)
	assert.Equal(t, user.Name, result.Name)
	assert.Equal(t, user.Username, result.Username)
	mockRepo.AssertExpectations(t)
}

func TestUserUsecase_GetUserByPhone_NotFound(t *testing.T) {
	mockRepo := new(MockUserRepository)
	uc := New(mockRepo)

	ctx := context.Background()
	phone := "+79998887766"

	mockRepo.On("GetUserByPhone", ctx, phone).Return(nil, errors.New("user not found"))

	result, err := uc.GetUserByPhone(ctx, phone)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, errs.ErrUserNotFound, err)
	mockRepo.AssertExpectations(t)
}

func TestUserUsecase_GetUserByUsername_Success(t *testing.T) {
	mockRepo := new(MockUserRepository)
	uc := New(mockRepo)

	ctx := context.Background()
	username := "test_user"

	user := &UserModels.User{
		ID:          uuid.New(),
		PhoneNumber: "+79998887766",
		Name:        "Test User",
		Username:    username,
		Bio:         "Test bio",
		Avatar:      nil,
		AccountType: UserModels.UserAccount,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	mockRepo.On("GetUserByUsername", ctx, username).Return(user, nil)

	result, err := uc.GetUserByUsername(ctx, username)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, user.ID, result.ID)
	assert.Equal(t, user.PhoneNumber, result.PhoneNumber)
	assert.Equal(t, user.Name, result.Name)
	assert.Equal(t, username, result.Username)
	mockRepo.AssertExpectations(t)
}

func TestUserUsecase_GetUserByUsername_NotFound(t *testing.T) {
	mockRepo := new(MockUserRepository)
	uc := New(mockRepo)

	ctx := context.Background()
	username := "nonexistent_user"

	mockRepo.On("GetUserByUsername", ctx, username).Return(nil, errors.New("user not found"))

	result, err := uc.GetUserByUsername(ctx, username)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, errs.ErrUserNotFound, err)
	mockRepo.AssertExpectations(t)
}
