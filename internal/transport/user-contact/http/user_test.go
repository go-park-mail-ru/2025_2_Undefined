package transport

import (
	"bytes"
	"context"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-park-mail-ru/2025_2_Undefined/internal/models/domains"
	UserDTO "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/dto/user"
	gen "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/generated/user"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/user-contact/http/mocks"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func TestUserHandler_GetUserByPhone_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserClient := mocks.NewMockUserServiceClient(ctrl)
	handler := NewUserGRPCProxyHandler(mockUserClient)

	userID := uuid.New()
	phone := "+79998887766"

	mockUserClient.EXPECT().
		GetUserByPhone(gomock.Any(), gomock.Any()).
		Return(&gen.GetUserByPhoneRes{
			User: &gen.User{
				Id:          userID.String(),
				PhoneNumber: phone,
				Name:        "Test User",
				Username:    "testuser",
				Bio:         "",
				AvatarUrl:   "",
				AccountType: "personal",
				CreatedAt:   "2024-01-01T00:00:00Z",
				UpdatedAt:   "2024-01-01T00:00:00Z",
			},
		}, nil)

	reqBody := UserDTO.GetUserByPhone{PhoneNumber: phone}
	body, _ := json.Marshal(reqBody)
	request := httptest.NewRequest(http.MethodPost, "/user/by-phone", bytes.NewBuffer(body))
	request.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()
	handler.GetUserByPhone(recorder, request)

	assert.Equal(t, http.StatusOK, recorder.Code)

	var response UserDTO.User
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, phone, response.PhoneNumber)
	assert.Equal(t, "Test User", response.Name)
}

func TestUserHandler_GetUserByPhone_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserClient := mocks.NewMockUserServiceClient(ctrl)
	handler := NewUserGRPCProxyHandler(mockUserClient)

	phone := "+79998887766"

	mockUserClient.EXPECT().
		GetUserByPhone(gomock.Any(), gomock.Any()).
		Return(nil, status.Error(codes.NotFound, "user not found"))

	reqBody := UserDTO.GetUserByPhone{PhoneNumber: phone}
	body, _ := json.Marshal(reqBody)
	request := httptest.NewRequest(http.MethodPost, "/user/by-phone", bytes.NewBuffer(body))
	request.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()
	handler.GetUserByPhone(recorder, request)

	assert.Equal(t, http.StatusNotFound, recorder.Code)
}

func TestUserHandler_GetCurrentUser_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserClient := mocks.NewMockUserServiceClient(ctrl)
	handler := NewUserGRPCProxyHandler(mockUserClient)

	userID := uuid.New()

	mockUserClient.EXPECT().
		GetUserById(gomock.Any(), &gen.GetUserByIdReq{UserId: userID.String()}).
		Return(&gen.GetUserByIdRes{
			User: &gen.User{
				Id:          userID.String(),
				PhoneNumber: "+79998887766",
				Name:        "Test User",
				Username:    "testuser",
				AccountType: "personal",
				CreatedAt:   "2024-01-01T00:00:00Z",
				UpdatedAt:   "2024-01-01T00:00:00Z",
			},
		}, nil)

	request := httptest.NewRequest(http.MethodGet, "/user/current", nil)
	ctx := request.Context()
	ctx = context.WithValue(ctx, domains.UserIDKey{}, userID.String())
	request = request.WithContext(ctx)

	recorder := httptest.NewRecorder()
	handler.GetCurrentUser(recorder, request)

	assert.Equal(t, http.StatusOK, recorder.Code)
}

func TestUserHandler_GetCurrentUser_Unauthorized(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserClient := mocks.NewMockUserServiceClient(ctrl)
	handler := NewUserGRPCProxyHandler(mockUserClient)

	request := httptest.NewRequest(http.MethodGet, "/user/current", nil)
	recorder := httptest.NewRecorder()
	handler.GetCurrentUser(recorder, request)

	assert.Equal(t, http.StatusUnauthorized, recorder.Code)
}

func TestUserHandler_GetUserByUsername_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserClient := mocks.NewMockUserServiceClient(ctrl)
	handler := NewUserGRPCProxyHandler(mockUserClient)

	userID := uuid.New()
	username := "testuser"

	mockUserClient.EXPECT().
		GetUserByUsername(gomock.Any(), &gen.GetUserByUsernameReq{Username: username}).
		Return(&gen.GetUserByUsernameRes{
			User: &gen.User{
				Id:          userID.String(),
				PhoneNumber: "+79998887766",
				Name:        "Test User",
				Username:    username,
				AccountType: "personal",
				CreatedAt:   "2024-01-01T00:00:00Z",
				UpdatedAt:   "2024-01-01T00:00:00Z",
			},
		}, nil)

	reqBody := UserDTO.GetUserByUsername{Username: username}
	body, _ := json.Marshal(reqBody)
	request := httptest.NewRequest(http.MethodPost, "/user/by-username", bytes.NewBuffer(body))
	request.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()
	handler.GetUserByUsername(recorder, request)

	assert.Equal(t, http.StatusOK, recorder.Code)
}

func TestUserHandler_UpdateUserInfo_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserClient := mocks.NewMockUserServiceClient(ctrl)
	handler := NewUserGRPCProxyHandler(mockUserClient)

	userID := uuid.New()

	mockUserClient.EXPECT().
		UpdateUserInfo(gomock.Any(), gomock.Any()).
		Return(&emptypb.Empty{}, nil)

	name := "Updated Name"
	username := "newusername"
	reqBody := UserDTO.UpdateUserInfo{
		Name:     &name,
		Username: &username,
	}
	body, _ := json.Marshal(reqBody)
	request := httptest.NewRequest(http.MethodPatch, "/user", bytes.NewBuffer(body))
	request.Header.Set("Content-Type", "application/json")
	ctx := request.Context()
	ctx = context.WithValue(ctx, domains.UserIDKey{}, userID.String())
	request = request.WithContext(ctx)

	recorder := httptest.NewRecorder()
	handler.UpdateUserInfo(recorder, request)

	assert.Equal(t, http.StatusOK, recorder.Code)
}

func TestUserHandler_UpdateUserInfo_Unauthorized(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserClient := mocks.NewMockUserServiceClient(ctrl)
	handler := NewUserGRPCProxyHandler(mockUserClient)

	name := "Updated Name"
	reqBody := UserDTO.UpdateUserInfo{Name: &name}
	body, _ := json.Marshal(reqBody)
	request := httptest.NewRequest(http.MethodPatch, "/user", bytes.NewBuffer(body))
	request.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()
	handler.UpdateUserInfo(recorder, request)

	assert.Equal(t, http.StatusUnauthorized, recorder.Code)
}

func TestUserHandler_GetUserAvatars_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserClient := mocks.NewMockUserServiceClient(ctrl)
	handler := NewUserGRPCProxyHandler(mockUserClient)

	userID := uuid.New()

	mockUserClient.EXPECT().
		GetUserAvatars(gomock.Any(), &gen.GetUserAvatarsReq{UserIds: []string{userID.String()}}).
		Return(&gen.GetUserAvatarsRes{
			Avatars: map[string]string{
				userID.String(): "avatar.jpg",
			},
		}, nil)

	requestBody := map[string]interface{}{
		"ids": []string{userID.String()},
	}
	body, _ := json.Marshal(requestBody)
	request := httptest.NewRequest(http.MethodPost, "/user/avatars", bytes.NewBuffer(body))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	handler.GetUserAvatars(recorder, request)

	assert.Equal(t, http.StatusOK, recorder.Code)
}

func TestUserHandler_UploadUserAvatar_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserClient := mocks.NewMockUserServiceClient(ctrl)
	handler := NewUserGRPCProxyHandler(mockUserClient)

	userID := uuid.New()

	mockUserClient.EXPECT().
		UploadUserAvatar(gomock.Any(), gomock.Any()).
		Return(&gen.UploadUserAvatarRes{
			AvatarUrl: "https://example.com/avatar.jpg",
		}, nil)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("avatar", "avatar.jpg")
	part.Write([]byte("fake image data"))
	writer.Close()

	request := httptest.NewRequest(http.MethodPost, "/user/avatar", body)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	ctx := context.WithValue(request.Context(), domains.UserIDKey{}, userID.String())
	request = request.WithContext(ctx)

	recorder := httptest.NewRecorder()
	handler.UploadUserAvatar(recorder, request)

	assert.Equal(t, http.StatusOK, recorder.Code)
}

func TestUserHandler_UploadUserAvatar_Unauthorized(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserClient := mocks.NewMockUserServiceClient(ctrl)
	handler := NewUserGRPCProxyHandler(mockUserClient)

	request := httptest.NewRequest(http.MethodPost, "/user/avatar", nil)
	recorder := httptest.NewRecorder()
	handler.UploadUserAvatar(recorder, request)

	assert.Equal(t, http.StatusUnauthorized, recorder.Code)
}

func TestUserHandler_UploadUserAvatar_NoFile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserClient := mocks.NewMockUserServiceClient(ctrl)
	handler := NewUserGRPCProxyHandler(mockUserClient)

	userID := uuid.New()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.Close()

	request := httptest.NewRequest(http.MethodPost, "/user/avatar", body)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	ctx := context.WithValue(request.Context(), domains.UserIDKey{}, userID.String())
	request = request.WithContext(ctx)

	recorder := httptest.NewRecorder()
	handler.UploadUserAvatar(recorder, request)

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestUserHandler_UploadUserAvatar_GRPCError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserClient := mocks.NewMockUserServiceClient(ctrl)
	handler := NewUserGRPCProxyHandler(mockUserClient)

	userID := uuid.New()

	mockUserClient.EXPECT().
		UploadUserAvatar(gomock.Any(), gomock.Any()).
		Return(nil, status.Error(codes.Internal, "internal error"))

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("avatar", "avatar.jpg")
	part.Write([]byte("fake image data"))
	writer.Close()

	request := httptest.NewRequest(http.MethodPost, "/user/avatar", body)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	ctx := context.WithValue(request.Context(), domains.UserIDKey{}, userID.String())
	request = request.WithContext(ctx)

	recorder := httptest.NewRecorder()
	handler.UploadUserAvatar(recorder, request)

	assert.Equal(t, http.StatusInternalServerError, recorder.Code)
}
