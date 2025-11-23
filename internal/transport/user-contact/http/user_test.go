package transport

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	UserDTO "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/dto/user"
	gen "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/generated/user"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/user-contact/http/mocks"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
