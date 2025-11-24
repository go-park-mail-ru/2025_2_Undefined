package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	ContactModels "github.com/go-park-mail-ru/2025_2_Undefined/internal/models/contact"
	UserModels "github.com/go-park-mail-ru/2025_2_Undefined/internal/models/user"
	ContactDTO "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/dto/contact"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/usecase/mocks"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestContactUsecase_CreateContact_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockContactRepo := mocks.NewMockContactRepository(ctrl)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockFileStorage := mocks.NewMockFileStorage(ctrl)
	uc := New(mockContactRepo, mockUserRepo, mockFileStorage, nil)

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

	mockUserRepo.EXPECT().GetUserByID(ctx, userID).Return(user, nil)
	mockContactRepo.EXPECT().CreateContact(ctx, userID, contactUserID).Return(nil)

	err := uc.CreateContact(ctx, req, userID)

	assert.NoError(t, err)
}

func TestContactUsecase_CreateContact_UserNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockContactRepo := mocks.NewMockContactRepository(ctrl)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockFileStorage := mocks.NewMockFileStorage(ctrl)
	uc := New(mockContactRepo, mockUserRepo, mockFileStorage, nil)

	ctx := context.Background()
	userID := uuid.New()
	contactUserID := uuid.New()
	req := &ContactDTO.PostContactDTO{
		ContactUserID: contactUserID,
	}

	mockUserRepo.EXPECT().GetUserByID(ctx, userID).Return(nil, errors.New("user not found"))

	err := uc.CreateContact(ctx, req, userID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user not found")
}

func TestContactUsecase_CreateContact_RepositoryError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockContactRepo := mocks.NewMockContactRepository(ctrl)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockFileStorage := mocks.NewMockFileStorage(ctrl)
	uc := New(mockContactRepo, mockUserRepo, mockFileStorage, nil)

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

	mockUserRepo.EXPECT().GetUserByID(ctx, userID).Return(user, nil)
	mockContactRepo.EXPECT().CreateContact(ctx, userID, contactUserID).Return(errors.New("database error"))

	err := uc.CreateContact(ctx, req, userID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database error")
}

func TestContactUsecase_GetContacts_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockContactRepo := mocks.NewMockContactRepository(ctrl)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockFileStorage := mocks.NewMockFileStorage(ctrl)
	uc := New(mockContactRepo, mockUserRepo, mockFileStorage, nil)

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

	mockContactRepo.EXPECT().GetContactsByUserID(ctx, userID).Return(contacts, nil)
	mockUserRepo.EXPECT().GetUserByID(ctx, contactUserID).Return(contactUser, nil)

	result, err := uc.GetContacts(ctx, userID)

	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, userID, result[0].UserID)
	assert.Equal(t, contactUserID, result[0].ContactUser.ID)
	assert.Equal(t, "Contact User", result[0].ContactUser.Name)
}

func TestContactUsecase_GetContacts_NoContacts(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockContactRepo := mocks.NewMockContactRepository(ctrl)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockFileStorage := mocks.NewMockFileStorage(ctrl)
	uc := New(mockContactRepo, mockUserRepo, mockFileStorage, nil)

	ctx := context.Background()
	userID := uuid.New()

	mockContactRepo.EXPECT().GetContactsByUserID(ctx, userID).Return(nil, nil)

	result, err := uc.GetContacts(ctx, userID)

	assert.NoError(t, err)
	assert.Empty(t, result)
}

func TestContactUsecase_GetContacts_RepositoryError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockContactRepo := mocks.NewMockContactRepository(ctrl)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockFileStorage := mocks.NewMockFileStorage(ctrl)
	uc := New(mockContactRepo, mockUserRepo, mockFileStorage, nil)

	ctx := context.Background()
	userID := uuid.New()

	mockContactRepo.EXPECT().GetContactsByUserID(ctx, userID).Return(nil, errors.New("database error"))

	result, err := uc.GetContacts(ctx, userID)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "database error")
}

func TestContactUsecase_GetContacts_UserNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockContactRepo := mocks.NewMockContactRepository(ctrl)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockFileStorage := mocks.NewMockFileStorage(ctrl)
	uc := New(mockContactRepo, mockUserRepo, mockFileStorage, nil)

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

	mockContactRepo.EXPECT().GetContactsByUserID(ctx, userID).Return(contacts, nil)
	mockUserRepo.EXPECT().GetUserByID(ctx, contactUserID).Return(nil, errors.New("user not found"))

	result, err := uc.GetContacts(ctx, userID)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "user not found")
}
