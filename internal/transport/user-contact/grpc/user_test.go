package grpc

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2025_2_Undefined/internal/models/domains"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/models/errs"
	dtoContact "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/dto/contact"
	dtoUser "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/dto/user"
	gen "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/generated/user"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type MockUserUsecase struct {
	mock.Mock
}

func (m *MockUserUsecase) GetUserById(ctx context.Context, id uuid.UUID) (*dtoUser.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dtoUser.User), args.Error(1)
}

func (m *MockUserUsecase) GetUserByPhone(ctx context.Context, phone string) (*dtoUser.User, error) {
	args := m.Called(ctx, phone)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dtoUser.User), args.Error(1)
}

func (m *MockUserUsecase) GetUserByUsername(ctx context.Context, username string) (*dtoUser.User, error) {
	args := m.Called(ctx, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dtoUser.User), args.Error(1)
}

func (m *MockUserUsecase) UploadUserAvatar(ctx context.Context, userID uuid.UUID, data []byte, filename, contentType string) (string, error) {
	args := m.Called(ctx, userID, data, filename, contentType)
	return args.String(0), args.Error(1)
}

func (m *MockUserUsecase) UpdateUserInfo(ctx context.Context, userID uuid.UUID, name *string, username *string, bio *string) error {
	args := m.Called(ctx, userID, name, username, bio)
	return args.Error(0)
}

func (m *MockUserUsecase) GetUserAvatars(ctx context.Context, userIDs []uuid.UUID) (map[string]*string, error) {
	args := m.Called(ctx, userIDs)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]*string), args.Error(1)
}

type MockContactUsecase struct {
	mock.Mock
}

func (m *MockContactUsecase) CreateContact(ctx context.Context, req *dtoContact.PostContactDTO, userID uuid.UUID) error {
	args := m.Called(ctx, req, userID)
	return args.Error(0)
}

func (m *MockContactUsecase) GetContacts(ctx context.Context, userID uuid.UUID) ([]*dtoContact.GetContactsDTO, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*dtoContact.GetContactsDTO), args.Error(1)
}

func (m *MockContactUsecase) SearchContacts(ctx context.Context, userID uuid.UUID, query string) ([]*dtoContact.GetContactsDTO, error) {
	args := m.Called(ctx, userID, query)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*dtoContact.GetContactsDTO), args.Error(1)
}

func (m *MockContactUsecase) ReindexAllContacts(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func setupContext() context.Context {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)
	ctx := context.Background()
	return context.WithValue(ctx, domains.ContextKeyLogger{}, logrus.NewEntry(logger))
}

func TestGetUserById_Success(t *testing.T) {
	mockUserUC := new(MockUserUsecase)
	mockContactUC := new(MockContactUsecase)
	handler := NewUserGRPCHandler(mockUserUC, mockContactUC)
	ctx := setupContext()

	userID := uuid.New()
	bio := "Test bio"
	testUser := &dtoUser.User{
		ID:          userID,
		PhoneNumber: "+1234567890",
		Name:        "Test User",
		Username:    "testuser",
		Bio:         &bio,
		AccountType: "user",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	mockUserUC.On("GetUserById", ctx, userID).Return(testUser, nil)

	req := &gen.GetUserByIdReq{UserId: userID.String()}
	res, err := handler.GetUserById(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, userID.String(), res.User.Id)
	assert.Equal(t, testUser.PhoneNumber, res.User.PhoneNumber)
	assert.Equal(t, testUser.Name, res.User.Name)
	assert.Equal(t, bio, res.User.Bio)
	mockUserUC.AssertExpectations(t)
}

func TestGetUserById_InvalidUserID(t *testing.T) {
	mockUserUC := new(MockUserUsecase)
	mockContactUC := new(MockContactUsecase)
	handler := NewUserGRPCHandler(mockUserUC, mockContactUC)
	ctx := setupContext()

	req := &gen.GetUserByIdReq{UserId: "invalid-uuid"}
	res, err := handler.GetUserById(ctx, req)

	assert.Error(t, err)
	assert.Nil(t, res)
	st, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.InvalidArgument, st.Code())
}

func TestGetUserById_UserNotFound(t *testing.T) {
	mockUserUC := new(MockUserUsecase)
	mockContactUC := new(MockContactUsecase)
	handler := NewUserGRPCHandler(mockUserUC, mockContactUC)
	ctx := setupContext()

	userID := uuid.New()
	mockUserUC.On("GetUserById", ctx, userID).Return(nil, errs.ErrUserNotFound)

	req := &gen.GetUserByIdReq{UserId: userID.String()}
	res, err := handler.GetUserById(ctx, req)

	assert.Error(t, err)
	assert.Nil(t, res)
	st, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.NotFound, st.Code())
	mockUserUC.AssertExpectations(t)
}

func TestGetUserByPhone_Success(t *testing.T) {
	mockUserUC := new(MockUserUsecase)
	mockContactUC := new(MockContactUsecase)
	handler := NewUserGRPCHandler(mockUserUC, mockContactUC)
	ctx := setupContext()

	phone := "+1234567890"
	testUser := &dtoUser.User{
		ID:          uuid.New(),
		PhoneNumber: phone,
		Name:        "Test User",
		Username:    "testuser",
		Bio:         nil,
		AccountType: "user",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	mockUserUC.On("GetUserByPhone", ctx, phone).Return(testUser, nil)

	req := &gen.GetUserByPhoneReq{PhoneNumber: phone}
	res, err := handler.GetUserByPhone(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, testUser.ID.String(), res.User.Id)
	assert.Equal(t, phone, res.User.PhoneNumber)
	assert.Equal(t, "", res.User.Bio)
	mockUserUC.AssertExpectations(t)
}

func TestGetUserByUsername_Success(t *testing.T) {
	mockUserUC := new(MockUserUsecase)
	mockContactUC := new(MockContactUsecase)
	handler := NewUserGRPCHandler(mockUserUC, mockContactUC)
	ctx := setupContext()

	username := "testuser"
	testUser := &dtoUser.User{
		ID:          uuid.New(),
		PhoneNumber: "+1234567890",
		Name:        "Test User",
		Username:    username,
		Bio:         nil,
		AccountType: "user",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	mockUserUC.On("GetUserByUsername", ctx, username).Return(testUser, nil)

	req := &gen.GetUserByUsernameReq{Username: username}
	res, err := handler.GetUserByUsername(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, username, res.User.Username)
	mockUserUC.AssertExpectations(t)
}

func TestUpdateUserInfo_Success(t *testing.T) {
	mockUserUC := new(MockUserUsecase)
	mockContactUC := new(MockContactUsecase)
	handler := NewUserGRPCHandler(mockUserUC, mockContactUC)
	ctx := setupContext()

	userID := uuid.New()
	name := "New Name"
	username := "newusername"
	bio := "New bio"

	mockUserUC.On("UpdateUserInfo", ctx, userID, &name, &username, &bio).Return(nil)

	req := &gen.UpdateUserInfoReq{
		UserId:   userID.String(),
		Name:     &name,
		Username: &username,
		Bio:      &bio,
	}
	res, err := handler.UpdateUserInfo(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	mockUserUC.AssertExpectations(t)
}

func TestUpdateUserInfo_InvalidUserID(t *testing.T) {
	mockUserUC := new(MockUserUsecase)
	mockContactUC := new(MockContactUsecase)
	handler := NewUserGRPCHandler(mockUserUC, mockContactUC)
	ctx := setupContext()

	req := &gen.UpdateUserInfoReq{UserId: "invalid-uuid"}
	res, err := handler.UpdateUserInfo(ctx, req)

	assert.Error(t, err)
	assert.Nil(t, res)
	st, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.InvalidArgument, st.Code())
}

func TestUpdateUserInfo_InvalidName(t *testing.T) {
	mockUserUC := new(MockUserUsecase)
	mockContactUC := new(MockContactUsecase)
	handler := NewUserGRPCHandler(mockUserUC, mockContactUC)
	ctx := setupContext()

	userID := uuid.New()
	invalidName := "ThisNameIsWayTooLongForValidation"

	req := &gen.UpdateUserInfoReq{
		UserId: userID.String(),
		Name:   &invalidName,
	}
	res, err := handler.UpdateUserInfo(ctx, req)

	assert.Error(t, err)
	assert.Nil(t, res)
	st, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.InvalidArgument, st.Code())
}

func TestUpdateUserInfo_DuplicateUsername(t *testing.T) {
	mockUserUC := new(MockUserUsecase)
	mockContactUC := new(MockContactUsecase)
	handler := NewUserGRPCHandler(mockUserUC, mockContactUC)
	ctx := setupContext()

	userID := uuid.New()
	username := "existinguser"

	mockUserUC.On("UpdateUserInfo", ctx, userID, mock.Anything, &username, mock.Anything).Return(errs.ErrIsDuplicateKey)

	req := &gen.UpdateUserInfoReq{
		UserId:   userID.String(),
		Username: &username,
	}
	res, err := handler.UpdateUserInfo(ctx, req)

	assert.Error(t, err)
	assert.Nil(t, res)
	st, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.AlreadyExists, st.Code())
	mockUserUC.AssertExpectations(t)
}

func TestUploadUserAvatar_Success(t *testing.T) {
	mockUserUC := new(MockUserUsecase)
	mockContactUC := new(MockContactUsecase)
	handler := NewUserGRPCHandler(mockUserUC, mockContactUC)
	ctx := setupContext()

	userID := uuid.New()
	data := []byte("image data")
	filename := "avatar.jpg"
	contentType := "image/jpeg"
	avatarURL := "https://example.com/avatar.jpg"

	mockUserUC.On("UploadUserAvatar", ctx, userID, data, filename, contentType).Return(avatarURL, nil)

	req := &gen.UploadUserAvatarReq{
		UserId:      userID.String(),
		Data:        data,
		Filename:    filename,
		ContentType: contentType,
	}
	res, err := handler.UploadUserAvatar(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, avatarURL, res.AvatarUrl)
	mockUserUC.AssertExpectations(t)
}

func TestUploadUserAvatar_InvalidUserID(t *testing.T) {
	mockUserUC := new(MockUserUsecase)
	mockContactUC := new(MockContactUsecase)
	handler := NewUserGRPCHandler(mockUserUC, mockContactUC)
	ctx := setupContext()

	req := &gen.UploadUserAvatarReq{UserId: "invalid-uuid"}
	res, err := handler.UploadUserAvatar(ctx, req)

	assert.Error(t, err)
	assert.Nil(t, res)
	st, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.InvalidArgument, st.Code())
}

func TestUploadUserAvatar_UploadError(t *testing.T) {
	mockUserUC := new(MockUserUsecase)
	mockContactUC := new(MockContactUsecase)
	handler := NewUserGRPCHandler(mockUserUC, mockContactUC)
	ctx := setupContext()

	userID := uuid.New()
	mockUserUC.On("UploadUserAvatar", ctx, userID, mock.Anything, mock.Anything, mock.Anything).Return("", errors.New("upload failed"))

	req := &gen.UploadUserAvatarReq{
		UserId:      userID.String(),
		Data:        []byte("data"),
		Filename:    "test.jpg",
		ContentType: "image/jpeg",
	}
	res, err := handler.UploadUserAvatar(ctx, req)

	assert.Error(t, err)
	assert.Nil(t, res)
	st, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.Internal, st.Code())
	mockUserUC.AssertExpectations(t)
}

func TestGetUserAvatars_Success(t *testing.T) {
	mockUserUC := new(MockUserUsecase)
	mockContactUC := new(MockContactUsecase)
	handler := NewUserGRPCHandler(mockUserUC, mockContactUC)
	ctx := setupContext()

	userID1 := uuid.New()
	userID2 := uuid.New()
	avatar1 := "https://example.com/avatar1.jpg"
	avatar2 := "https://example.com/avatar2.jpg"

	avatars := map[string]*string{
		userID1.String(): &avatar1,
		userID2.String(): &avatar2,
	}

	mockUserUC.On("GetUserAvatars", ctx, mock.MatchedBy(func(ids []uuid.UUID) bool {
		return len(ids) == 2
	})).Return(avatars, nil)

	req := &gen.GetUserAvatarsReq{
		UserIds: []string{userID1.String(), userID2.String()},
	}
	res, err := handler.GetUserAvatars(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Len(t, res.Avatars, 2)
	assert.Equal(t, avatar1, res.Avatars[userID1.String()])
	assert.Equal(t, avatar2, res.Avatars[userID2.String()])
	mockUserUC.AssertExpectations(t)
}

func TestGetUserAvatars_EmptyRequest(t *testing.T) {
	mockUserUC := new(MockUserUsecase)
	mockContactUC := new(MockContactUsecase)
	handler := NewUserGRPCHandler(mockUserUC, mockContactUC)
	ctx := setupContext()

	req := &gen.GetUserAvatarsReq{UserIds: []string{}}
	res, err := handler.GetUserAvatars(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Empty(t, res.Avatars)
}

func TestGetUserAvatars_InvalidUserID(t *testing.T) {
	mockUserUC := new(MockUserUsecase)
	mockContactUC := new(MockContactUsecase)
	handler := NewUserGRPCHandler(mockUserUC, mockContactUC)
	ctx := setupContext()

	req := &gen.GetUserAvatarsReq{
		UserIds: []string{"invalid-uuid"},
	}
	res, err := handler.GetUserAvatars(ctx, req)

	assert.Error(t, err)
	assert.Nil(t, res)
	st, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.InvalidArgument, st.Code())
}

func TestGetUserAvatars_UsecaseError(t *testing.T) {
	mockUserUC := new(MockUserUsecase)
	mockContactUC := new(MockContactUsecase)
	handler := NewUserGRPCHandler(mockUserUC, mockContactUC)
	ctx := setupContext()

	userID := uuid.New()
	mockUserUC.On("GetUserAvatars", ctx, mock.Anything).Return(nil, errors.New("database error"))

	req := &gen.GetUserAvatarsReq{
		UserIds: []string{userID.String()},
	}
	res, err := handler.GetUserAvatars(ctx, req)

	assert.Error(t, err)
	assert.Nil(t, res)
	st, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.Internal, st.Code())
	mockUserUC.AssertExpectations(t)
}

func TestNewUserGRPCHandler(t *testing.T) {
	mockUserUC := new(MockUserUsecase)
	mockContactUC := new(MockContactUsecase)

	handler := NewUserGRPCHandler(mockUserUC, mockContactUC)

	assert.NotNil(t, handler)
	assert.Equal(t, mockUserUC, handler.userUC)
	assert.Equal(t, mockContactUC, handler.contactUC)
}
