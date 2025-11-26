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

	user := &UserModels.User{
		ID:          userID,
		PhoneNumber: "+79998887766",
		Name:        "Test User",
		Username:    "test_user",
		Bio:         &[]string{"Test bio"}[0],
		AccountType: UserModels.UserAccount,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	mockRepo.EXPECT().GetUserByID(ctx, userID).Return(user, nil)

	result, err := uc.GetUserById(ctx, userID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, userID, result.ID)
	assert.Equal(t, user.PhoneNumber, result.PhoneNumber)
	assert.Equal(t, user.Name, result.Name)
	assert.Equal(t, user.Username, result.Username)
	assert.Equal(t, user.Bio, result.Bio)
	assert.Equal(t, user.AccountType, result.AccountType)
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

	user := &UserModels.User{
		ID:          userID,
		PhoneNumber: "+79998887766",
		Name:        "Test User",
		Username:    "test_user",
		Bio:         &[]string{"Test bio"}[0],
		AccountType: UserModels.UserAccount,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	mockRepo.EXPECT().GetUserByID(ctx, userID).Return(user, nil)

	result, err := uc.GetUserById(ctx, userID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, userID, result.ID)
}

func TestUserUsecase_GetUserByPhone_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockUserRepository(ctrl)
	mockFileStorage := mocks.NewMockFileStorage(ctrl)
	uc := New(mockRepo, mockFileStorage)

	ctx := context.Background()
	phone := "+79998887766"

	user := &UserModels.User{
		ID:          uuid.New(),
		PhoneNumber: phone,
		Name:        "Test User",
		Username:    "test_user",
		Bio:         &[]string{"Test bio"}[0],
		AccountType: UserModels.UserAccount,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	mockRepo.EXPECT().GetUserByPhone(ctx, phone).Return(user, nil)

	result, err := uc.GetUserByPhone(ctx, phone)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, user.ID, result.ID)
	assert.Equal(t, phone, result.PhoneNumber)
	assert.Equal(t, user.Name, result.Name)
	assert.Equal(t, user.Username, result.Username)
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

	user := &UserModels.User{
		ID:          uuid.New(),
		PhoneNumber: "+79998887766",
		Name:        "Test User",
		Username:    username,
		Bio:         &[]string{"Test bio"}[0],
		AccountType: UserModels.UserAccount,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	mockRepo.EXPECT().GetUserByUsername(ctx, username).Return(user, nil)

	result, err := uc.GetUserByUsername(ctx, username)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, user.ID, result.ID)
	assert.Equal(t, user.PhoneNumber, result.PhoneNumber)
	assert.Equal(t, user.Name, result.Name)
	assert.Equal(t, username, result.Username)
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

func TestUserUsecase_UpdateUserInfo_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockUserRepository(ctrl)
	mockFileStorage := mocks.NewMockFileStorage(ctrl)
	uc := New(mockRepo, mockFileStorage)

	ctx := context.Background()
	userID := uuid.New()
	name := "New Name"
	username := "new_username"
	bio := "New bio"

	mockRepo.EXPECT().UpdateUserInfo(ctx, userID, &name, &username, &bio).Return(nil)

	err := uc.UpdateUserInfo(ctx, userID, &name, &username, &bio)

	assert.NoError(t, err)
}

func TestUserUsecase_UpdateUserInfo_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockUserRepository(ctrl)
	mockFileStorage := mocks.NewMockFileStorage(ctrl)
	uc := New(mockRepo, mockFileStorage)

	ctx := context.Background()
	userID := uuid.New()
	name := "New Name"

	mockRepo.EXPECT().UpdateUserInfo(ctx, userID, &name, nil, nil).Return(errors.New("database error"))

	err := uc.UpdateUserInfo(ctx, userID, &name, nil, nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database error")
}

func TestUserUsecase_GetUserAvatars_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockUserRepository(ctrl)
	mockFileStorage := mocks.NewMockFileStorage(ctrl)
	uc := New(mockRepo, mockFileStorage)

	ctx := context.Background()
	userID1 := uuid.New()
	userID2 := uuid.New()
	userIDs := []uuid.UUID{userID1, userID2}

	attachmentID1 := uuid.New()
	attachmentID2 := uuid.New()
	avatarsIDs := map[string]uuid.UUID{
		userID1.String(): attachmentID1,
		userID2.String(): attachmentID2,
	}

	url1 := "https://example.com/avatar1.jpg"
	url2 := "https://example.com/avatar2.jpg"

	mockRepo.EXPECT().GetUserAvatars(ctx, userIDs).Return(avatarsIDs, nil)
	mockFileStorage.EXPECT().GetOne(ctx, &attachmentID1).Return(url1, nil)
	mockFileStorage.EXPECT().GetOne(ctx, &attachmentID2).Return(url2, nil)

	result, err := uc.GetUserAvatars(ctx, userIDs)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 2)
	assert.NotNil(t, result[userID1.String()])
	assert.NotNil(t, result[userID2.String()])
	assert.Equal(t, url1, *result[userID1.String()])
	assert.Equal(t, url2, *result[userID2.String()])
}

func TestUserUsecase_GetUserAvatars_RepositoryError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockUserRepository(ctrl)
	mockFileStorage := mocks.NewMockFileStorage(ctrl)
	uc := New(mockRepo, mockFileStorage)

	ctx := context.Background()
	userIDs := []uuid.UUID{uuid.New(), uuid.New()}

	mockRepo.EXPECT().GetUserAvatars(ctx, userIDs).Return(nil, errors.New("database error"))

	result, err := uc.GetUserAvatars(ctx, userIDs)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "database error")
}

func TestUserUsecase_GetUserAvatars_FileStorageError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockUserRepository(ctrl)
	mockFileStorage := mocks.NewMockFileStorage(ctrl)
	uc := New(mockRepo, mockFileStorage)

	ctx := context.Background()
	userID := uuid.New()
	userIDs := []uuid.UUID{userID}

	attachmentID := uuid.New()
	avatarsIDs := map[string]uuid.UUID{
		userID.String(): attachmentID,
	}

	mockRepo.EXPECT().GetUserAvatars(ctx, userIDs).Return(avatarsIDs, nil)
	mockFileStorage.EXPECT().GetOne(ctx, &attachmentID).Return("", errors.New("storage error"))

	result, err := uc.GetUserAvatars(ctx, userIDs)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Nil(t, result[userID.String()])
}
