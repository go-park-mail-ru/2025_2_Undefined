package user

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2025_2_Undefined/internal/models/errs"
	UserModels "github.com/go-park-mail-ru/2025_2_Undefined/internal/models/user"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/usecase/mocks"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestUserUsecase_GetUserById_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockUserRepository(ctrl)
	mockFileStorage := mocks.NewMockFileStorage(ctrl)
	uc := New(mockRepo, mockFileStorage)

	ctx := context.Background()
	userID := uuid.New()
	avatarID := uuid.New()

	user := &UserModels.User{
		ID:          userID,
		PhoneNumber: "+79998887766",
		Name:        "Test User",
		Username:    "test_user",
		Bio:         "Test bio",
		AvatarID:    &avatarID,
		AccountType: UserModels.UserAccount,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	mockRepo.EXPECT().GetUserByID(ctx, userID).Return(user, nil)
	mockFileStorage.EXPECT().GetOne(ctx, &avatarID).Return("https://example.com/avatar.jpg", nil)

	result, err := uc.GetUserById(ctx, userID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, userID, result.ID)
	assert.Equal(t, user.PhoneNumber, result.PhoneNumber)
	assert.Equal(t, user.Name, result.Name)
	assert.Equal(t, user.Username, result.Username)
	assert.Equal(t, user.Bio, result.Bio)
	assert.Equal(t, user.AccountType, result.AccountType)
	assert.Equal(t, "https://example.com/avatar.jpg", result.AvatarURL)
}

func TestUserUsecase_GetUserById_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockUserRepository(ctrl)
	mockFileStorage := mocks.NewMockFileStorage(ctrl)
	uc := New(mockRepo, mockFileStorage)

	ctx := context.Background()
	userID := uuid.New()

	mockRepo.EXPECT().GetUserByID(ctx, userID).Return(nil, errors.New("user not found"))

	result, err := uc.GetUserById(ctx, userID)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, errs.ErrUserNotFound, err)
}

func TestUserUsecase_GetUserById_AvatarError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockUserRepository(ctrl)
	mockFileStorage := mocks.NewMockFileStorage(ctrl)
	uc := New(mockRepo, mockFileStorage)

	ctx := context.Background()
	userID := uuid.New()
	avatarID := uuid.New()

	user := &UserModels.User{
		ID:          userID,
		PhoneNumber: "+79998887766",
		Name:        "Test User",
		Username:    "test_user",
		Bio:         "Test bio",
		AvatarID:    &avatarID,
		AccountType: UserModels.UserAccount,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	mockRepo.EXPECT().GetUserByID(ctx, userID).Return(user, nil)
	mockFileStorage.EXPECT().GetOne(ctx, &avatarID).Return("", errors.New("avatar not found"))

	result, err := uc.GetUserById(ctx, userID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, userID, result.ID)
	assert.Equal(t, "", result.AvatarURL) // Should be empty when avatar fetch fails
}

func TestUserUsecase_GetUserByPhone_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockUserRepository(ctrl)
	mockFileStorage := mocks.NewMockFileStorage(ctrl)
	uc := New(mockRepo, mockFileStorage)

	ctx := context.Background()
	phone := "+79998887766"
	avatarID := uuid.New()

	user := &UserModels.User{
		ID:          uuid.New(),
		PhoneNumber: phone,
		Name:        "Test User",
		Username:    "test_user",
		Bio:         "Test bio",
		AvatarID:    &avatarID,
		AccountType: UserModels.UserAccount,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	mockRepo.EXPECT().GetUserByPhone(ctx, phone).Return(user, nil)
	mockFileStorage.EXPECT().GetOne(ctx, &avatarID).Return("https://example.com/avatar.jpg", nil)

	result, err := uc.GetUserByPhone(ctx, phone)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, user.ID, result.ID)
	assert.Equal(t, phone, result.PhoneNumber)
	assert.Equal(t, user.Name, result.Name)
	assert.Equal(t, user.Username, result.Username)
	assert.Equal(t, "https://example.com/avatar.jpg", result.AvatarURL)
}

func TestUserUsecase_GetUserByPhone_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockUserRepository(ctrl)
	mockFileStorage := mocks.NewMockFileStorage(ctrl)
	uc := New(mockRepo, mockFileStorage)

	ctx := context.Background()
	phone := "+79998887766"

	mockRepo.EXPECT().GetUserByPhone(ctx, phone).Return(nil, errors.New("user not found"))

	result, err := uc.GetUserByPhone(ctx, phone)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, errs.ErrUserNotFound, err)
}

func TestUserUsecase_GetUserByUsername_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockUserRepository(ctrl)
	mockFileStorage := mocks.NewMockFileStorage(ctrl)
	uc := New(mockRepo, mockFileStorage)

	ctx := context.Background()
	username := "test_user"
	avatarID := uuid.New()

	user := &UserModels.User{
		ID:          uuid.New(),
		PhoneNumber: "+79998887766",
		Name:        "Test User",
		Username:    username,
		Bio:         "Test bio",
		AvatarID:    &avatarID,
		AccountType: UserModels.UserAccount,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	mockRepo.EXPECT().GetUserByUsername(ctx, username).Return(user, nil)
	mockFileStorage.EXPECT().GetOne(ctx, &avatarID).Return("https://example.com/avatar.jpg", nil)

	result, err := uc.GetUserByUsername(ctx, username)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, user.ID, result.ID)
	assert.Equal(t, user.PhoneNumber, result.PhoneNumber)
	assert.Equal(t, user.Name, result.Name)
	assert.Equal(t, username, result.Username)
	assert.Equal(t, "https://example.com/avatar.jpg", result.AvatarURL)
}

func TestUserUsecase_GetUserByUsername_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockUserRepository(ctrl)
	mockFileStorage := mocks.NewMockFileStorage(ctrl)
	uc := New(mockRepo, mockFileStorage)

	ctx := context.Background()
	username := "nonexistent_user"

	mockRepo.EXPECT().GetUserByUsername(ctx, username).Return(nil, errors.New("user not found"))

	result, err := uc.GetUserByUsername(ctx, username)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, errs.ErrUserNotFound, err)
}

func TestUserUsecase_UploadUserAvatar_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockUserRepository(ctrl)
	mockFileStorage := mocks.NewMockFileStorage(ctrl)
	uc := New(mockRepo, mockFileStorage)

	ctx := context.Background()
	userID := uuid.New()
	data := []byte("fake avatar data")
	filename := "avatar.jpg"
	contentType := "image/jpeg"
	expectedURL := "https://example.com/avatar.jpg"

	mockFileStorage.EXPECT().CreateOne(ctx, gomock.Any(), gomock.Any()).Return(expectedURL, nil)
	mockRepo.EXPECT().UpdateUserAvatar(ctx, userID, gomock.Any(), int64(len(data))).Return(nil)

	result, err := uc.UploadUserAvatar(ctx, userID, data, filename, contentType)

	assert.NoError(t, err)
	assert.Equal(t, expectedURL, result)
}

func TestUserUsecase_UploadUserAvatar_FileStorageError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockUserRepository(ctrl)
	mockFileStorage := mocks.NewMockFileStorage(ctrl)
	uc := New(mockRepo, mockFileStorage)

	ctx := context.Background()
	userID := uuid.New()
	data := []byte("fake avatar data")
	filename := "avatar.jpg"
	contentType := "image/jpeg"

	mockFileStorage.EXPECT().CreateOne(ctx, gomock.Any(), gomock.Any()).Return("", errors.New("storage error"))

	result, err := uc.UploadUserAvatar(ctx, userID, data, filename, contentType)

	assert.Error(t, err)
	assert.Empty(t, result)
	assert.Contains(t, err.Error(), "storage error")
}

func TestUserUsecase_UploadUserAvatar_UpdateError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockUserRepository(ctrl)
	mockFileStorage := mocks.NewMockFileStorage(ctrl)
	uc := New(mockRepo, mockFileStorage)

	ctx := context.Background()
	userID := uuid.New()
	data := []byte("fake avatar data")
	filename := "avatar.jpg"
	contentType := "image/jpeg"
	expectedURL := "https://example.com/avatar.jpg"

	mockFileStorage.EXPECT().CreateOne(ctx, gomock.Any(), gomock.Any()).Return(expectedURL, nil)
	mockRepo.EXPECT().UpdateUserAvatar(ctx, userID, gomock.Any(), int64(len(data))).Return(errors.New("database error"))

	result, err := uc.UploadUserAvatar(ctx, userID, data, filename, contentType)

	assert.Error(t, err)
	assert.Empty(t, result)
	assert.Contains(t, err.Error(), "database error")
}
