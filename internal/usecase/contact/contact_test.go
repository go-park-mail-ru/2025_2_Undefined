package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	ContactModels "github.com/go-park-mail-ru/2025_2_Undefined/internal/models/contact"
	UserModels "github.com/go-park-mail-ru/2025_2_Undefined/internal/models/user"
	ContactDTO "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/dto/contact"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockContactRepository struct {
	mock.Mock
}

func (m *MockContactRepository) CreateContact(ctx context.Context, userID uuid.UUID, contactUserID uuid.UUID) error {
	args := m.Called(ctx, userID, contactUserID)
	return args.Error(0)
}

func (m *MockContactRepository) GetContactsByUserID(ctx context.Context, userID uuid.UUID) ([]*ContactModels.Contact, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*ContactModels.Contact), args.Error(1)
}

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

func TestContactUsecase_CreateContact_Success(t *testing.T) {
	mockContactRepo := new(MockContactRepository)
	mockUserRepo := new(MockUserRepository)
	uc := New(mockContactRepo, mockUserRepo)

	ctx := context.Background()
	userID := uuid.New()
	contactUserID := uuid.New()
	req := &ContactDTO.PostContactDTO{
		ContactUserID: contactUserID,
	}

	user := &UserModels.User{
		ID:          userID,
		Name:        "Test User",
		PhoneNumber: "+79998887766",
	}

	mockUserRepo.On("GetUserByID", ctx, userID).Return(user, nil)
	mockContactRepo.On("CreateContact", ctx, userID, contactUserID).Return(nil)

	err := uc.CreateContact(ctx, req, userID)

	assert.NoError(t, err)
	mockUserRepo.AssertExpectations(t)
	mockContactRepo.AssertExpectations(t)
}

func TestContactUsecase_CreateContact_UserNotFound(t *testing.T) {
	mockContactRepo := new(MockContactRepository)
	mockUserRepo := new(MockUserRepository)
	uc := New(mockContactRepo, mockUserRepo)

	ctx := context.Background()
	userID := uuid.New()
	contactUserID := uuid.New()
	req := &ContactDTO.PostContactDTO{
		ContactUserID: contactUserID,
	}

	mockUserRepo.On("GetUserByID", ctx, userID).Return(nil, errors.New("user not found"))

	err := uc.CreateContact(ctx, req, userID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user not found")
	mockUserRepo.AssertExpectations(t)
}

func TestContactUsecase_CreateContact_RepositoryError(t *testing.T) {
	mockContactRepo := new(MockContactRepository)
	mockUserRepo := new(MockUserRepository)
	uc := New(mockContactRepo, mockUserRepo)

	ctx := context.Background()
	userID := uuid.New()
	contactUserID := uuid.New()
	req := &ContactDTO.PostContactDTO{
		ContactUserID: contactUserID,
	}

	user := &UserModels.User{
		ID:          userID,
		Name:        "Test User",
		PhoneNumber: "+79998887766",
	}

	mockUserRepo.On("GetUserByID", ctx, userID).Return(user, nil)
	mockContactRepo.On("CreateContact", ctx, userID, contactUserID).Return(errors.New("database error"))

	err := uc.CreateContact(ctx, req, userID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database error")
	mockUserRepo.AssertExpectations(t)
	mockContactRepo.AssertExpectations(t)
}

func TestContactUsecase_GetContacts_Success(t *testing.T) {
	mockContactRepo := new(MockContactRepository)
	mockUserRepo := new(MockUserRepository)
	uc := New(mockContactRepo, mockUserRepo)

	ctx := context.Background()
	userID := uuid.New()
	contactUserID := uuid.New()

	contacts := []*ContactModels.Contact{
		{
			UserID:        userID,
			ContactUserID: contactUserID,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		},
	}

	contactUser := &UserModels.User{
		ID:          contactUserID,
		Name:        "Contact User",
		PhoneNumber: "+79998887777",
		Username:    "contact_user",
		AccountType: UserModels.UserAccount,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	mockContactRepo.On("GetContactsByUserID", ctx, userID).Return(contacts, nil)
	mockUserRepo.On("GetUserByID", ctx, contactUserID).Return(contactUser, nil)

	result, err := uc.GetContacts(ctx, userID)

	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, userID, result[0].UserID)
	assert.Equal(t, contactUserID, result[0].ContactUser.ID)
	assert.Equal(t, "Contact User", result[0].ContactUser.Name)
	mockContactRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}

func TestContactUsecase_GetContacts_NoContacts(t *testing.T) {
	mockContactRepo := new(MockContactRepository)
	mockUserRepo := new(MockUserRepository)
	uc := New(mockContactRepo, mockUserRepo)

	ctx := context.Background()
	userID := uuid.New()

	mockContactRepo.On("GetContactsByUserID", ctx, userID).Return(nil, nil)

	result, err := uc.GetContacts(ctx, userID)

	assert.NoError(t, err)
	assert.Empty(t, result)
	mockContactRepo.AssertExpectations(t)
}

func TestContactUsecase_GetContacts_RepositoryError(t *testing.T) {
	mockContactRepo := new(MockContactRepository)
	mockUserRepo := new(MockUserRepository)
	uc := New(mockContactRepo, mockUserRepo)

	ctx := context.Background()
	userID := uuid.New()

	mockContactRepo.On("GetContactsByUserID", ctx, userID).Return(nil, errors.New("database error"))

	result, err := uc.GetContacts(ctx, userID)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "database error")
	mockContactRepo.AssertExpectations(t)
}

func TestContactUsecase_GetContacts_UserNotFound(t *testing.T) {
	mockContactRepo := new(MockContactRepository)
	mockUserRepo := new(MockUserRepository)
	uc := New(mockContactRepo, mockUserRepo)

	ctx := context.Background()
	userID := uuid.New()
	contactUserID := uuid.New()

	contacts := []*ContactModels.Contact{
		{
			UserID:        userID,
			ContactUserID: contactUserID,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		},
	}

	mockContactRepo.On("GetContactsByUserID", ctx, userID).Return(contacts, nil)
	mockUserRepo.On("GetUserByID", ctx, contactUserID).Return(nil, errors.New("user not found"))

	result, err := uc.GetContacts(ctx, userID)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "user not found")
	mockContactRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}
