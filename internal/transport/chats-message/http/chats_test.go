package chats

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/go-park-mail-ru/2025_2_Undefined/internal/models/domains"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/models/errs"
	dtoChats "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/dto/chats"
	dtoMessage "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/dto/message"
	dtoUtils "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/dto/utils"
	gen "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/generated/chats"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// MockChatServiceClient is a mock of ChatServiceClient interface
type MockChatServiceClient struct {
	ctrl     *gomock.Controller
	recorder *MockChatServiceClientMockRecorder
}

// MockChatServiceClientMockRecorder is the mock recorder for MockChatServiceClient
type MockChatServiceClientMockRecorder struct {
	mock *MockChatServiceClient
}

// NewMockChatServiceClient creates a new mock instance
func NewMockChatServiceClient(ctrl *gomock.Controller) *MockChatServiceClient {
	mock := &MockChatServiceClient{ctrl: ctrl}
	mock.recorder = &MockChatServiceClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockChatServiceClient) EXPECT() *MockChatServiceClientMockRecorder {
	return m.recorder
}

// GetChats mocks base method
func (m *MockChatServiceClient) GetChats(ctx context.Context, in *gen.GetChatsReq, opts ...grpc.CallOption) (*gen.GetChatsRes, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, in}
	for _, a := range opts {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "GetChats", varargs...)
	ret0, _ := ret[0].(*gen.GetChatsRes)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetChats indicates an expected call of GetChats
func (mr *MockChatServiceClientMockRecorder) GetChats(ctx, in interface{}, opts ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, in}, opts...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetChats", reflect.TypeOf((*MockChatServiceClient)(nil).GetChats), varargs...)
}

// CreateChat mocks base method
func (m *MockChatServiceClient) CreateChat(ctx context.Context, in *gen.CreateChatReq, opts ...grpc.CallOption) (*gen.IdRes, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, in}
	for _, a := range opts {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "CreateChat", varargs...)
	ret0, _ := ret[0].(*gen.IdRes)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateChat indicates an expected call of CreateChat
func (mr *MockChatServiceClientMockRecorder) CreateChat(ctx, in interface{}, opts ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, in}, opts...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateChat", reflect.TypeOf((*MockChatServiceClient)(nil).CreateChat), varargs...)
}

// GetChat mocks base method
func (m *MockChatServiceClient) GetChat(ctx context.Context, in *gen.GetChatReq, opts ...grpc.CallOption) (*gen.ChatDetailedInformation, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, in}
	for _, a := range opts {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "GetChat", varargs...)
	ret0, _ := ret[0].(*gen.ChatDetailedInformation)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetChat indicates an expected call of GetChat
func (mr *MockChatServiceClientMockRecorder) GetChat(ctx, in interface{}, opts ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, in}, opts...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetChat", reflect.TypeOf((*MockChatServiceClient)(nil).GetChat), varargs...)
}

// GetUsersDialog mocks base method
func (m *MockChatServiceClient) GetUsersDialog(ctx context.Context, in *gen.GetUsersDialogReq, opts ...grpc.CallOption) (*gen.IdRes, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, in}
	for _, a := range opts {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "GetUsersDialog", varargs...)
	ret0, _ := ret[0].(*gen.IdRes)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetUsersDialog indicates an expected call of GetUsersDialog
func (mr *MockChatServiceClientMockRecorder) GetUsersDialog(ctx, in interface{}, opts ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, in}, opts...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetUsersDialog", reflect.TypeOf((*MockChatServiceClient)(nil).GetUsersDialog), varargs...)
}

// UpdateChat mocks base method
func (m *MockChatServiceClient) UpdateChat(ctx context.Context, in *gen.UpdateChatReq, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, in}
	for _, a := range opts {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "UpdateChat", varargs...)
	ret0, _ := ret[0].(*emptypb.Empty)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// UpdateChat indicates an expected call of UpdateChat
func (mr *MockChatServiceClientMockRecorder) UpdateChat(ctx, in interface{}, opts ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, in}, opts...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateChat", reflect.TypeOf((*MockChatServiceClient)(nil).UpdateChat), varargs...)
}

// DeleteChat mocks base method
func (m *MockChatServiceClient) DeleteChat(ctx context.Context, in *gen.GetChatReq, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, in}
	for _, a := range opts {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "DeleteChat", varargs...)
	ret0, _ := ret[0].(*emptypb.Empty)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// DeleteChat indicates an expected call of DeleteChat
func (mr *MockChatServiceClientMockRecorder) DeleteChat(ctx, in interface{}, opts ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, in}, opts...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteChat", reflect.TypeOf((*MockChatServiceClient)(nil).DeleteChat), varargs...)
}

// AddUserToChat mocks base method
func (m *MockChatServiceClient) AddUserToChat(ctx context.Context, in *gen.AddUserToChatReq, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, in}
	for _, a := range opts {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "AddUserToChat", varargs...)
	ret0, _ := ret[0].(*emptypb.Empty)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// AddUserToChat indicates an expected call of AddUserToChat
func (mr *MockChatServiceClientMockRecorder) AddUserToChat(ctx, in interface{}, opts ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, in}, opts...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddUserToChat", reflect.TypeOf((*MockChatServiceClient)(nil).AddUserToChat), varargs...)
}

// RemoveUserFromChat mocks base method
func (m *MockChatServiceClient) RemoveUserFromChat(ctx context.Context, in *gen.RemoveUserFromChatReq, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, in}
	for _, a := range opts {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "RemoveUserFromChat", varargs...)
	ret0, _ := ret[0].(*emptypb.Empty)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// RemoveUserFromChat indicates an expected call of RemoveUserFromChat
func (mr *MockChatServiceClientMockRecorder) RemoveUserFromChat(ctx, in interface{}, opts ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, in}, opts...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RemoveUserFromChat", reflect.TypeOf((*MockChatServiceClient)(nil).RemoveUserFromChat), varargs...)
}

func TestGRPCGetChats_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := NewMockChatServiceClient(ctrl)
	mockMessageClient := NewMockMessageServiceClient(ctrl)
	handler := NewChatsGRPCProxyHandler(mockClient, mockMessageClient)

	userID := uuid.New()
	chatID := uuid.New()

	expectedResponse := &gen.GetChatsRes{
		Chats: []*gen.Chat{
			{
				Id:   chatID.String(),
				Name: "Test Chat",
				Type: "private",
			},
		},
	}

	mockClient.EXPECT().
		GetChats(gomock.Any(), &gen.GetChatsReq{UserId: userID.String()}).
		Return(expectedResponse, nil)

	request := httptest.NewRequest(http.MethodGet, "/chats", nil)
	ctx := context.WithValue(request.Context(), domains.UserIDKey{}, userID.String())
	request = request.WithContext(ctx)

	recorder := httptest.NewRecorder()
	handler.GetChats(recorder, request)

	assert.Equal(t, http.StatusOK, recorder.Code)
}

func TestGRPCGetChats_Unauthorized(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := NewMockChatServiceClient(ctrl)
	mockMessageClient := NewMockMessageServiceClient(ctrl)
	handler := NewChatsGRPCProxyHandler(mockClient, mockMessageClient)

	request := httptest.NewRequest(http.MethodGet, "/chats", nil)
	recorder := httptest.NewRecorder()
	handler.GetChats(recorder, request)

	assert.Equal(t, http.StatusUnauthorized, recorder.Code)
}

func TestGRPCPostChats_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := NewMockChatServiceClient(ctrl)
	mockMessageClient := NewMockMessageServiceClient(ctrl)
	handler := NewChatsGRPCProxyHandler(mockClient, mockMessageClient)

	userID := uuid.New()
	participantID := uuid.New()
	chatID := uuid.New()

	reqBody := dtoChats.ChatCreateInformationDTO{
		Name: "New Chat",
		Type: "private",
		Members: []dtoChats.AddChatMemberDTO{
			{
				UserId: participantID,
				Role:   "writer",
			},
		},
	}

	mockClient.EXPECT().
		CreateChat(gomock.Any(), gomock.Any()).
		Return(&gen.IdRes{Id: chatID.String()}, nil)

	body, _ := json.Marshal(reqBody)
	request := httptest.NewRequest(http.MethodPost, "/chats", bytes.NewBuffer(body))
	request.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(request.Context(), domains.UserIDKey{}, userID.String())
	request = request.WithContext(ctx)

	recorder := httptest.NewRecorder()
	handler.PostChats(recorder, request)

	assert.Equal(t, http.StatusCreated, recorder.Code)

	var response dtoUtils.IdDTO
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, chatID, response.ID)
}

func TestGRPCPostChats_InvalidJSON(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := NewMockChatServiceClient(ctrl)
	mockMessageClient := NewMockMessageServiceClient(ctrl)
	handler := NewChatsGRPCProxyHandler(mockClient, mockMessageClient)

	userID := uuid.New()

	request := httptest.NewRequest(http.MethodPost, "/chats", bytes.NewBufferString("invalid json"))
	request.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(request.Context(), domains.UserIDKey{}, userID.String())
	request = request.WithContext(ctx)

	recorder := httptest.NewRecorder()
	handler.PostChats(recorder, request)

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestGRPCGetInformationAboutChat_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := NewMockChatServiceClient(ctrl)
	mockMessageClient := NewMockMessageServiceClient(ctrl)
	handler := NewChatsGRPCProxyHandler(mockClient, mockMessageClient)

	userID := uuid.New()
	chatID := uuid.New()

	expectedResponse := &gen.ChatDetailedInformation{
		Id:   chatID.String(),
		Name: "Test Chat",
		Type: "private",
	}

	mockClient.EXPECT().
		GetChat(gomock.Any(), &gen.GetChatReq{
			ChatId: chatID.String(),
			UserId: userID.String(),
		}).
		Return(expectedResponse, nil)

	request := httptest.NewRequest(http.MethodGet, "/chats/"+chatID.String(), nil)
	ctx := context.WithValue(request.Context(), domains.UserIDKey{}, userID.String())
	request = request.WithContext(ctx)

	recorder := httptest.NewRecorder()
	handler.GetInformationAboutChat(recorder, request)

	assert.Equal(t, http.StatusOK, recorder.Code)
}

func TestGRPCGetInformationAboutChat_InvalidUUID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := NewMockChatServiceClient(ctrl)
	mockMessageClient := NewMockMessageServiceClient(ctrl)
	handler := NewChatsGRPCProxyHandler(mockClient, mockMessageClient)

	userID := uuid.New()

	request := httptest.NewRequest(http.MethodGet, "/chats/invalid-uuid", nil)
	ctx := context.WithValue(request.Context(), domains.UserIDKey{}, userID.String())
	request = request.WithContext(ctx)

	recorder := httptest.NewRecorder()
	handler.GetInformationAboutChat(recorder, request)

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestGRPCGetInformationAboutChat_Unauthorized(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := NewMockChatServiceClient(ctrl)
	mockMessageClient := NewMockMessageServiceClient(ctrl)
	handler := NewChatsGRPCProxyHandler(mockClient, mockMessageClient)

	chatID := uuid.New()

	request := httptest.NewRequest(http.MethodGet, "/chats/"+chatID.String(), nil)
	recorder := httptest.NewRecorder()
	handler.GetInformationAboutChat(recorder, request)

	assert.Equal(t, http.StatusUnauthorized, recorder.Code)
}

func TestGRPCGetUsersDialog_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := NewMockChatServiceClient(ctrl)
	mockMessageClient := NewMockMessageServiceClient(ctrl)
	handler := NewChatsGRPCProxyHandler(mockClient, mockMessageClient)

	userID := uuid.New()
	otherUserID := uuid.New()
	dialogID := uuid.New()

	mockClient.EXPECT().
		GetUsersDialog(gomock.Any(), &gen.GetUsersDialogReq{
			User1Id: userID.String(),
			User2Id: otherUserID.String(),
		}).
		Return(&gen.IdRes{Id: dialogID.String()}, nil)

	request := httptest.NewRequest(http.MethodGet, "/chats/dialog/"+otherUserID.String(), nil)
	ctx := context.WithValue(request.Context(), domains.UserIDKey{}, userID.String())
	request = request.WithContext(ctx)

	recorder := httptest.NewRecorder()
	handler.GetUsersDialog(recorder, request)

	assert.Equal(t, http.StatusOK, recorder.Code)

	var response dtoUtils.IdDTO
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, dialogID, response.ID)
}

func TestGRPCUpdateChat_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := NewMockChatServiceClient(ctrl)
	mockMessageClient := NewMockMessageServiceClient(ctrl)
	handler := NewChatsGRPCProxyHandler(mockClient, mockMessageClient)

	userID := uuid.New()
	chatID := uuid.New()

	reqBody := dtoChats.ChatUpdateDTO{
		Name:        "Updated Chat",
		Description: "Updated description",
	}

	mockClient.EXPECT().
		UpdateChat(gomock.Any(), gomock.Any()).
		Return(&emptypb.Empty{}, nil)

	body, _ := json.Marshal(reqBody)
	request := httptest.NewRequest(http.MethodPatch, "/chats/"+chatID.String(), bytes.NewBuffer(body))
	request.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(request.Context(), domains.UserIDKey{}, userID.String())
	request = request.WithContext(ctx)

	recorder := httptest.NewRecorder()
	handler.UpdateChat(recorder, request)

	assert.Equal(t, http.StatusOK, recorder.Code)
}

func TestGRPCUpdateChat_GRPCError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := NewMockChatServiceClient(ctrl)
	mockMessageClient := NewMockMessageServiceClient(ctrl)
	handler := NewChatsGRPCProxyHandler(mockClient, mockMessageClient)

	userID := uuid.New()
	chatID := uuid.New()

	reqBody := dtoChats.ChatUpdateDTO{
		Name: "Updated Chat",
	}

	mockClient.EXPECT().
		UpdateChat(gomock.Any(), gomock.Any()).
		Return(nil, status.Error(codes.PermissionDenied, "user is not admin"))

	body, _ := json.Marshal(reqBody)
	request := httptest.NewRequest(http.MethodPatch, "/chats/"+chatID.String(), bytes.NewBuffer(body))
	request.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(request.Context(), domains.UserIDKey{}, userID.String())
	request = request.WithContext(ctx)

	recorder := httptest.NewRecorder()
	handler.UpdateChat(recorder, request)

	assert.NotEqual(t, http.StatusOK, recorder.Code)
}

func TestGRPCDeleteChat_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := NewMockChatServiceClient(ctrl)
	mockMessageClient := NewMockMessageServiceClient(ctrl)
	handler := NewChatsGRPCProxyHandler(mockClient, mockMessageClient)

	userID := uuid.New()
	chatID := uuid.New()

	mockClient.EXPECT().
		DeleteChat(gomock.Any(), &gen.GetChatReq{
			ChatId: chatID.String(),
			UserId: userID.String(),
		}).
		Return(&emptypb.Empty{}, nil)

	request := httptest.NewRequest(http.MethodDelete, "/chats/"+chatID.String(), nil)
	ctx := context.WithValue(request.Context(), domains.UserIDKey{}, userID.String())
	request = request.WithContext(ctx)

	recorder := httptest.NewRecorder()
	handler.DeleteChat(recorder, request)

	assert.Equal(t, http.StatusOK, recorder.Code)
}

func TestGRPCDeleteChat_GRPCError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := NewMockChatServiceClient(ctrl)
	mockMessageClient := NewMockMessageServiceClient(ctrl)
	handler := NewChatsGRPCProxyHandler(mockClient, mockMessageClient)

	userID := uuid.New()
	chatID := uuid.New()

	mockClient.EXPECT().
		DeleteChat(gomock.Any(), &gen.GetChatReq{
			ChatId: chatID.String(),
			UserId: userID.String(),
		}).
		Return(nil, status.Error(codes.PermissionDenied, "user is not admin"))

	request := httptest.NewRequest(http.MethodDelete, "/chats/"+chatID.String(), nil)
	ctx := context.WithValue(request.Context(), domains.UserIDKey{}, userID.String())
	request = request.WithContext(ctx)

	recorder := httptest.NewRecorder()
	handler.DeleteChat(recorder, request)

	assert.NotEqual(t, http.StatusOK, recorder.Code)
}

func TestGRPCAddUsersToChat_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := NewMockChatServiceClient(ctrl)
	mockMessageClient := NewMockMessageServiceClient(ctrl)
	handler := NewChatsGRPCProxyHandler(mockClient, mockMessageClient)

	userID := uuid.New()
	chatID := uuid.New()
	participantID := uuid.New()

	reqBody := dtoChats.AddUsersToChatDTO{
		Users: []dtoChats.AddChatMemberDTO{
			{
				UserId: participantID,
				Role:   "writer",
			},
		},
	}

	mockClient.EXPECT().
		AddUserToChat(gomock.Any(), gomock.Any()).
		Return(&emptypb.Empty{}, nil)

	body, _ := json.Marshal(reqBody)
	request := httptest.NewRequest(http.MethodPatch, "/chats/"+chatID.String()+"/members", bytes.NewBuffer(body))
	request.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(request.Context(), domains.UserIDKey{}, userID.String())
	request = request.WithContext(ctx)

	recorder := httptest.NewRecorder()
	handler.AddUsersToChat(recorder, request)

	assert.Equal(t, http.StatusOK, recorder.Code)
}

func TestGRPCAddUsersToChat_InvalidJSON(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := NewMockChatServiceClient(ctrl)
	mockMessageClient := NewMockMessageServiceClient(ctrl)
	handler := NewChatsGRPCProxyHandler(mockClient, mockMessageClient)

	userID := uuid.New()
	chatID := uuid.New()

	request := httptest.NewRequest(http.MethodPatch, "/chats/"+chatID.String()+"/members", bytes.NewBufferString("invalid json"))
	request.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(request.Context(), domains.UserIDKey{}, userID.String())
	request = request.WithContext(ctx)

	recorder := httptest.NewRecorder()
	handler.AddUsersToChat(recorder, request)

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestGRPCAddUsersToChat_Unauthorized(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := NewMockChatServiceClient(ctrl)
	mockMessageClient := NewMockMessageServiceClient(ctrl)
	handler := NewChatsGRPCProxyHandler(mockClient, mockMessageClient)

	chatID := uuid.New()
	participantID := uuid.New()

	reqBody := dtoChats.AddUsersToChatDTO{
		Users: []dtoChats.AddChatMemberDTO{
			{
				UserId: participantID,
				Role:   "writer",
			},
		},
	}

	body, _ := json.Marshal(reqBody)
	request := httptest.NewRequest(http.MethodPatch, "/chats/"+chatID.String()+"/members", bytes.NewBuffer(body))
	request.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()
	handler.AddUsersToChat(recorder, request)

	assert.Equal(t, http.StatusUnauthorized, recorder.Code)
}

func TestEditMessage_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockChatClient := mocks.NewMockChatServiceClient(ctrl)
	mockMessageClient := mocks.NewMockMessageServiceClient(ctrl)
	handler := NewChatsGRPCProxyHandler(mockChatClient, mockMessageClient)

	userID := uuid.New()
	messageID := uuid.New()

	reqBody := dtoMessage.EditMessageDTO{
		ID:   messageID,
		Text: "Updated message text",
	}

	mockMessageClient.EXPECT().
		EditMessage(gomock.Any(), reqBody, userID).
		Return(nil)

	body, _ := json.Marshal(reqBody)
	request := httptest.NewRequest(http.MethodPut, "/messages", bytes.NewBuffer(body))
	request.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(request.Context(), domains.UserIDKey{}, userID.String())
	request = request.WithContext(ctx)

	recorder := httptest.NewRecorder()
	handler.EditMessage(recorder, request)

	assert.Equal(t, http.StatusOK, recorder.Code)
}

func TestEditMessage_NoRights(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockChatClient := mocks.NewMockChatServiceClient(ctrl)
	mockMessageClient := mocks.NewMockMessageServiceClient(ctrl)
	handler := NewChatsGRPCProxyHandler(mockChatClient, mockMessageClient)

	userID := uuid.New()
	messageID := uuid.New()

	reqBody := dtoMessage.EditMessageDTO{
		ID:   messageID,
		Text: "Updated message text",
	}

	mockMessageClient.EXPECT().
		EditMessage(gomock.Any(), reqBody, userID).
		Return(errs.ErrNoRights)

	body, _ := json.Marshal(reqBody)
	request := httptest.NewRequest(http.MethodPut, "/messages", bytes.NewBuffer(body))
	request.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(request.Context(), domains.UserIDKey{}, userID.String())
	request = request.WithContext(ctx)

	recorder := httptest.NewRecorder()
	handler.EditMessage(recorder, request)

	assert.Equal(t, http.StatusForbidden, recorder.Code)
}

func TestDeleteMessage_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockChatClient := mocks.NewMockChatServiceClient(ctrl)
	mockMessageClient := mocks.NewMockMessageServiceClient(ctrl)
	handler := NewChatsGRPCProxyHandler(mockChatClient, mockMessageClient)

	userID := uuid.New()
	messageID := uuid.New()

	reqBody := dtoMessage.DeleteMessageDTO{
		ID: messageID,
	}

	mockMessageClient.EXPECT().
		DeleteMessage(gomock.Any(), reqBody, userID).
		Return(nil)

	body, _ := json.Marshal(reqBody)
	request := httptest.NewRequest(http.MethodDelete, "/messages", bytes.NewBuffer(body))
	request.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(request.Context(), domains.UserIDKey{}, userID.String())
	request = request.WithContext(ctx)

	recorder := httptest.NewRecorder()
	handler.DeleteMessage(recorder, request)

	assert.Equal(t, http.StatusOK, recorder.Code)
}

func TestDeleteMessage_NoRights(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockChatClient := mocks.NewMockChatServiceClient(ctrl)
	mockMessageClient := mocks.NewMockMessageServiceClient(ctrl)
	handler := NewChatsGRPCProxyHandler(mockChatClient, mockMessageClient)

	userID := uuid.New()
	messageID := uuid.New()

	reqBody := dtoMessage.DeleteMessageDTO{
		ID: messageID,
	}

	mockMessageClient.EXPECT().
		DeleteMessage(gomock.Any(), reqBody, userID).
		Return(errs.ErrNoRights)

	body, _ := json.Marshal(reqBody)
	request := httptest.NewRequest(http.MethodDelete, "/messages", bytes.NewBuffer(body))
	request.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(request.Context(), domains.UserIDKey{}, userID.String())
	request = request.WithContext(ctx)

	recorder := httptest.NewRecorder()
	handler.DeleteMessage(recorder, request)

	assert.Equal(t, http.StatusForbidden, recorder.Code)
}
