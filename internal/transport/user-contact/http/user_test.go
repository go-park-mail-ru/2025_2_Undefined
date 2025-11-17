package transport

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	UserDTO "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/dto/user"
	gen "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/generated/user"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type MockUserServiceClient struct {
	mock.Mock
}

func (m *MockUserServiceClient) GetUserById(ctx context.Context, in *gen.GetUserByIdReq, opts ...grpc.CallOption) (*gen.GetUserByIdRes, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*gen.GetUserByIdRes), args.Error(1)
}

func (m *MockUserServiceClient) GetUserByPhone(ctx context.Context, in *gen.GetUserByPhoneReq, opts ...grpc.CallOption) (*gen.GetUserByPhoneRes, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*gen.GetUserByPhoneRes), args.Error(1)
}

func (m *MockUserServiceClient) GetUserByUsername(ctx context.Context, in *gen.GetUserByUsernameReq, opts ...grpc.CallOption) (*gen.GetUserByUsernameRes, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*gen.GetUserByUsernameRes), args.Error(1)
}

func (m *MockUserServiceClient) UpdateUserInfo(ctx context.Context, in *gen.UpdateUserInfoReq, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*emptypb.Empty), args.Error(1)
}

func (m *MockUserServiceClient) UploadUserAvatar(ctx context.Context, in *gen.UploadUserAvatarReq, opts ...grpc.CallOption) (*gen.UploadUserAvatarRes, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*gen.UploadUserAvatarRes), args.Error(1)
}

func (m *MockUserServiceClient) CreateContact(ctx context.Context, in *gen.CreateContactReq, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*emptypb.Empty), args.Error(1)
}

func (m *MockUserServiceClient) GetContacts(ctx context.Context, in *gen.GetContactsReq, opts ...grpc.CallOption) (*gen.GetContactsRes, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*gen.GetContactsRes), args.Error(1)
}

func TestUserHandler_GetUserByPhone_Success(t *testing.T) {
	mockUserClient := new(MockUserServiceClient)
	handler := NewUserGRPCProxyHandler(mockUserClient)

	userID := uuid.New()
	phone := "+79998887766"

	mockUserClient.On("GetUserByPhone", mock.Anything, mock.MatchedBy(func(r *gen.GetUserByPhoneReq) bool {
		return r.PhoneNumber == phone
	})).Return(&gen.GetUserByPhoneRes{
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

	mockUserClient.AssertExpectations(t)
}

func TestUserHandler_GetUserByPhone_NotFound(t *testing.T) {
	mockUserClient := new(MockUserServiceClient)
	handler := NewUserGRPCProxyHandler(mockUserClient)

	phone := "+79998887766"

	mockUserClient.On("GetUserByPhone", mock.Anything, mock.Anything).
		Return(nil, status.Error(codes.NotFound, "user not found"))

	reqBody := UserDTO.GetUserByPhone{PhoneNumber: phone}
	body, _ := json.Marshal(reqBody)
	request := httptest.NewRequest(http.MethodPost, "/user/by-phone", bytes.NewBuffer(body))
	request.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()
	handler.GetUserByPhone(recorder, request)

	assert.Equal(t, http.StatusNotFound, recorder.Code)
	mockUserClient.AssertExpectations(t)
}
