package transport

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-park-mail-ru/2025_2_Undefined/internal/models/domains"
	ContactDTO "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/dto/contact"
	gen "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/generated/user"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/user-contact/http/mocks"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func TestContactHandler_CreateContact_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserClient := mocks.NewMockUserServiceClient(ctrl)
	handler := NewUserGRPCProxyHandler(mockUserClient)

	userID := uuid.New()
	contactUserID := uuid.New()

	req := ContactDTO.PostContactDTO{
		ContactUserID: contactUserID,
	}

	mockUserClient.EXPECT().
		CreateContact(gomock.Any(), gomock.Any()).
		Return(&emptypb.Empty{}, nil)

	body, _ := json.Marshal(req)
	request := httptest.NewRequest(http.MethodPost, "/contacts", bytes.NewBuffer(body))
	request.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(request.Context(), domains.UserIDKey{}, userID.String())
	request = request.WithContext(ctx)

	recorder := httptest.NewRecorder()
	handler.CreateContact(recorder, request)

	assert.Equal(t, http.StatusCreated, recorder.Code)
}

func TestContactHandler_CreateContact_InvalidJSON(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserClient := mocks.NewMockUserServiceClient(ctrl)
	handler := NewUserGRPCProxyHandler(mockUserClient)

	userID := uuid.New()

	request := httptest.NewRequest(http.MethodPost, "/contacts", bytes.NewBufferString("invalid json"))
	request.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(request.Context(), domains.UserIDKey{}, userID.String())
	request = request.WithContext(ctx)

	recorder := httptest.NewRecorder()
	handler.CreateContact(recorder, request)

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestContactHandler_CreateContact_SelfContact(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserClient := mocks.NewMockUserServiceClient(ctrl)
	handler := NewUserGRPCProxyHandler(mockUserClient)

	userID := uuid.New()

	req := ContactDTO.PostContactDTO{
		ContactUserID: userID,
	}

	body, _ := json.Marshal(req)
	request := httptest.NewRequest(http.MethodPost, "/contacts", bytes.NewBuffer(body))
	request.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(request.Context(), domains.UserIDKey{}, userID.String())
	request = request.WithContext(ctx)

	recorder := httptest.NewRecorder()
	handler.CreateContact(recorder, request)

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestContactHandler_CreateContact_Unauthorized(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserClient := mocks.NewMockUserServiceClient(ctrl)
	handler := NewUserGRPCProxyHandler(mockUserClient)

	contactUserID := uuid.New()
	req := ContactDTO.PostContactDTO{
		ContactUserID: contactUserID,
	}

	body, _ := json.Marshal(req)
	request := httptest.NewRequest(http.MethodPost, "/contacts", bytes.NewBuffer(body))
	request.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()
	handler.CreateContact(recorder, request)

	assert.Equal(t, http.StatusUnauthorized, recorder.Code)
}

func TestContactHandler_CreateContact_UserNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserClient := mocks.NewMockUserServiceClient(ctrl)
	handler := NewUserGRPCProxyHandler(mockUserClient)

	userID := uuid.New()
	contactUserID := uuid.New()

	req := ContactDTO.PostContactDTO{
		ContactUserID: contactUserID,
	}

	mockUserClient.EXPECT().
		CreateContact(gomock.Any(), gomock.Any()).
		Return(nil, status.Error(codes.NotFound, "user not found"))

	body, _ := json.Marshal(req)
	request := httptest.NewRequest(http.MethodPost, "/contacts", bytes.NewBuffer(body))
	request.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(request.Context(), domains.UserIDKey{}, userID.String())
	request = request.WithContext(ctx)

	recorder := httptest.NewRecorder()
	handler.CreateContact(recorder, request)

	assert.Equal(t, http.StatusNotFound, recorder.Code)
}

func TestContactHandler_GetContacts_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserClient := mocks.NewMockUserServiceClient(ctrl)
	handler := NewUserGRPCProxyHandler(mockUserClient)

	userID := uuid.New()
	contactID := uuid.New()

	mockUserClient.EXPECT().
		GetContacts(gomock.Any(), gomock.Any()).
		Return(&gen.GetContactsRes{
			Contacts: []*gen.Contact{
				{
					Id:          contactID.String(),
					PhoneNumber: "+79998887766",
					Name:        "Contact User",
					Username:    "contactuser",
					Bio:         "",
					AvatarUrl:   "",
					AccountType: "personal",
					CreatedAt:   "2024-01-01T00:00:00Z",
					UpdatedAt:   "2024-01-01T00:00:00Z",
				},
			},
		}, nil)

	request := httptest.NewRequest(http.MethodGet, "/contacts", nil)
	ctx := context.WithValue(request.Context(), domains.UserIDKey{}, userID.String())
	request = request.WithContext(ctx)

	recorder := httptest.NewRecorder()
	handler.GetContacts(recorder, request)

	assert.Equal(t, http.StatusOK, recorder.Code)

	var response []ContactDTO.GetContactsDTO
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Len(t, response, 1)
	if len(response) > 0 {
		assert.Equal(t, "Contact User", response[0].ContactUser.Name)
	}
}

func TestContactHandler_GetContacts_EmptyList(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserClient := mocks.NewMockUserServiceClient(ctrl)
	handler := NewUserGRPCProxyHandler(mockUserClient)

	userID := uuid.New()

	mockUserClient.EXPECT().
		GetContacts(gomock.Any(), gomock.Any()).
		Return(&gen.GetContactsRes{
			Contacts: []*gen.Contact{},
		}, nil)

	request := httptest.NewRequest(http.MethodGet, "/contacts", nil)
	ctx := context.WithValue(request.Context(), domains.UserIDKey{}, userID.String())
	request = request.WithContext(ctx)

	recorder := httptest.NewRecorder()
	handler.GetContacts(recorder, request)

	assert.Equal(t, http.StatusOK, recorder.Code)

	var response []ContactDTO.GetContactsDTO
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Len(t, response, 0)
}

func TestContactHandler_GetContacts_Unauthorized(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserClient := mocks.NewMockUserServiceClient(ctrl)
	handler := NewUserGRPCProxyHandler(mockUserClient)

	request := httptest.NewRequest(http.MethodGet, "/contacts", nil)
	recorder := httptest.NewRecorder()
	handler.GetContacts(recorder, request)

	assert.Equal(t, http.StatusUnauthorized, recorder.Code)
}

func TestContactHandler_GetContacts_GRPCError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserClient := mocks.NewMockUserServiceClient(ctrl)
	handler := NewUserGRPCProxyHandler(mockUserClient)

	userID := uuid.New()

	mockUserClient.EXPECT().
		GetContacts(gomock.Any(), gomock.Any()).
		Return(nil, status.Error(codes.Internal, "internal error"))

	request := httptest.NewRequest(http.MethodGet, "/contacts", nil)
	ctx := context.WithValue(request.Context(), domains.UserIDKey{}, userID.String())
	request = request.WithContext(ctx)

	recorder := httptest.NewRecorder()
	handler.GetContacts(recorder, request)

	assert.Equal(t, http.StatusInternalServerError, recorder.Code)
}
