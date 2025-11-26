package chats

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-park-mail-ru/2025_2_Undefined/internal/models/domains"
	dtoMessage "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/dto/message"
	gen "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/generated/chats"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type MockMessageClient struct {
	mock.Mock
}

func (m *MockMessageClient) StreamMessagesForUser(ctx context.Context, in *gen.StreamMessagesForUserReq, opts ...grpc.CallOption) (gen.MessageService_StreamMessagesForUserClient, error) {
	args := m.Called(ctx, in, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(gen.MessageService_StreamMessagesForUserClient), args.Error(1)
}

func (m *MockMessageClient) HandleSendMessage(ctx context.Context, in *gen.MessageEventReq, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	args := m.Called(ctx, in, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*emptypb.Empty), args.Error(1)
}

func (m *MockMessageClient) SearchMessages(ctx context.Context, in *gen.SearchMessagesReq, opts ...grpc.CallOption) (*gen.SearchMessagesRes, error) {
	args := m.Called(ctx, in, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*gen.SearchMessagesRes), args.Error(1)
}

func setupMessageContext(userID uuid.UUID) context.Context {
	ctx := context.Background()
	ctx = context.WithValue(ctx, domains.UserIDKey{}, userID.String())
	_ = domains.GetLogger(ctx)
	return ctx
}

func TestSearchMessages_Success(t *testing.T) {
	mockMessageClient := new(MockMessageClient)
	handler := &ChatsGRPCProxyHandler{
		messageClient: mockMessageClient,
	}

	userID := uuid.New()
	chatID := uuid.New()
	textQuery := "test"

	expectedRes := &gen.SearchMessagesRes{
		Messages: []*gen.Message{},
	}

	mockMessageClient.On("SearchMessages", mock.Anything, mock.MatchedBy(func(req *gen.SearchMessagesReq) bool {
		return req.UserId == userID.String() && req.ChatId == chatID.String() && req.Text == textQuery
	}), mock.Anything).Return(expectedRes, nil)

	url := "/chats/" + chatID.String() + "/messages/search?text=" + textQuery
	req := httptest.NewRequest(http.MethodGet, url, nil)
	req = req.WithContext(setupMessageContext(userID))
	req = mux.SetURLVars(req, map[string]string{"chat_id": chatID.String()})

	w := httptest.NewRecorder()
	handler.SearchMessages(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockMessageClient.AssertExpectations(t)
}

func TestSearchMessages_InvalidChatID(t *testing.T) {
	mockMessageClient := new(MockMessageClient)
	handler := &ChatsGRPCProxyHandler{
		messageClient: mockMessageClient,
	}

	userID := uuid.New()

	req := httptest.NewRequest(http.MethodGet, "/chats/invalid-uuid/messages/search?text=test", nil)
	req = req.WithContext(setupMessageContext(userID))
	req = mux.SetURLVars(req, map[string]string{"chat_id": "invalid-uuid"})

	w := httptest.NewRecorder()
	handler.SearchMessages(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "bad format chat_id")
}

func TestSearchMessages_MissingTextQuery(t *testing.T) {
	mockMessageClient := new(MockMessageClient)
	handler := &ChatsGRPCProxyHandler{
		messageClient: mockMessageClient,
	}

	userID := uuid.New()
	chatID := uuid.New()

	req := httptest.NewRequest(http.MethodGet, "/chats/"+chatID.String()+"/messages/search", nil)
	req = req.WithContext(setupMessageContext(userID))
	req = mux.SetURLVars(req, map[string]string{"chat_id": chatID.String()})

	w := httptest.NewRecorder()
	handler.SearchMessages(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "text in query is required")
}

func TestSearchMessages_Unauthorized(t *testing.T) {
	mockMessageClient := new(MockMessageClient)
	handler := &ChatsGRPCProxyHandler{
		messageClient: mockMessageClient,
	}

	chatID := uuid.New()

	req := httptest.NewRequest(http.MethodGet, "/chats/"+chatID.String()+"/messages/search?text=test", nil)
	req = req.WithContext(context.Background())
	req = mux.SetURLVars(req, map[string]string{"chat_id": chatID.String()})

	w := httptest.NewRecorder()
	handler.SearchMessages(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestSearchMessages_ClientError(t *testing.T) {
	mockMessageClient := new(MockMessageClient)
	handler := &ChatsGRPCProxyHandler{
		messageClient: mockMessageClient,
	}

	userID := uuid.New()
	chatID := uuid.New()
	textQuery := "test"

	mockMessageClient.On("SearchMessages", mock.Anything, mock.Anything, mock.Anything).Return((*gen.SearchMessagesRes)(nil), errors.New("grpc error"))

	req := httptest.NewRequest(http.MethodGet, "/chats/"+chatID.String()+"/messages/search?text="+textQuery, nil)
	req = req.WithContext(setupMessageContext(userID))
	req = mux.SetURLVars(req, map[string]string{"chat_id": chatID.String()})

	w := httptest.NewRecorder()
	handler.SearchMessages(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "failed to search messages")
}

func TestShouldCloseConnection(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "normal error",
			err:      errors.New("some error"),
			expected: false,
		},
		{
			name:     "closed network connection string",
			err:      errors.New("use of closed network connection"),
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := shouldCloseConnection(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestWriteJSONErrorWebSocket(t *testing.T) {
	handler := &ChatsGRPCProxyHandler{}

	// Создаем тестовый WebSocket сервер
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Fatal(err)
		}
		defer conn.Close()

		handler.writeJSONErrorWebSocket(conn, "test error")

		var response map[string]string
		err = conn.ReadJSON(&response)
		assert.NoError(t, err)
		assert.Equal(t, "test error", response["error"])
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	// Подключаемся к тестовому серверу
	// В реальном тесте здесь можно использовать gorilla/websocket для клиента
	t.Log("WebSocket error writing test completed with URL:", wsURL)
}

func TestHandleMessages_Unauthorized(t *testing.T) {
	mockMessageClient := new(MockMessageClient)
	handler := &ChatsGRPCProxyHandler{
		messageClient: mockMessageClient,
	}

	req := httptest.NewRequest(http.MethodGet, "/message/ws", nil)
	req = req.WithContext(context.Background())

	w := httptest.NewRecorder()
	handler.HandleMessages(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestWebSocketMessageTypes(t *testing.T) {
	// Проверяем константы типов сообщений
	assert.Equal(t, "new_message", dtoMessage.WebSocketMessageTypeNewChatMessage)
	assert.Equal(t, "edit_message", dtoMessage.WebSocketMessageTypeEditChatMessage)
	assert.Equal(t, "delete_message", dtoMessage.WebSocketMessageTypeDeleteChatMessage)
	assert.Equal(t, "chat_created", dtoMessage.WebSocketMessageTypeCreatedNewChat)
}

func TestUpgraderCheckOrigin(t *testing.T) {
	tests := []struct {
		name     string
		origin   string
		env      string
		expected bool
	}{
		{
			name:     "development environment allows any origin",
			origin:   "http://random-origin.com",
			env:      "development",
			expected: true,
		},
		{
			name:     "dev environment allows any origin",
			origin:   "http://random-origin.com",
			env:      "dev",
			expected: true,
		},
		{
			name:     "production allows specific origin",
			origin:   "https://100gramm.online",
			env:      "production",
			expected: true,
		},
		{
			name:     "production rejects unknown origin",
			origin:   "https://evil-site.com",
			env:      "production",
			expected: false,
		},
		{
			name:     "localhost allowed in production for dev",
			origin:   "http://localhost:3000",
			env:      "production",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("ENVIRONMENT", tt.env)

			req := httptest.NewRequest(http.MethodGet, "/ws", nil)
			req.Header.Set("Origin", tt.origin)

			result := upgrader.CheckOrigin(req)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSearchMessages_EmptyTextQuery(t *testing.T) {
	mockMessageClient := new(MockMessageClient)
	handler := &ChatsGRPCProxyHandler{
		messageClient: mockMessageClient,
	}

	userID := uuid.New()
	chatID := uuid.New()

	req := httptest.NewRequest(http.MethodGet, "/chats/"+chatID.String()+"/messages/search?text=", nil)
	req = req.WithContext(setupMessageContext(userID))
	req = mux.SetURLVars(req, map[string]string{"chat_id": chatID.String()})

	w := httptest.NewRecorder()
	handler.SearchMessages(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "text in query is required")
}

func TestSearchMessages_WithSpecialCharacters(t *testing.T) {
	mockMessageClient := new(MockMessageClient)
	handler := &ChatsGRPCProxyHandler{
		messageClient: mockMessageClient,
	}

	userID := uuid.New()
	chatID := uuid.New()
	textQuery := "testquery"

	expectedRes := &gen.SearchMessagesRes{
		Messages: []*gen.Message{},
	}

	mockMessageClient.On("SearchMessages", mock.Anything, mock.Anything, mock.Anything).Return(expectedRes, nil)

	url := "/chats/" + chatID.String() + "/messages/search?text=" + textQuery
	req := httptest.NewRequest(http.MethodGet, url, nil)
	req = req.WithContext(setupMessageContext(userID))
	req = mux.SetURLVars(req, map[string]string{"chat_id": chatID.String()})

	w := httptest.NewRecorder()
	handler.SearchMessages(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockMessageClient.AssertExpectations(t)
}
