package message

import (
	"context"
	"errors"
	"testing"
	"time"

	modelsChats "github.com/go-park-mail-ru/2025_2_Undefined/internal/models/chats"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/models/errs"
	modelsMessage "github.com/go-park-mail-ru/2025_2_Undefined/internal/models/message"
	modelsUser "github.com/go-park-mail-ru/2025_2_Undefined/internal/models/user"
	dtoChats "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/dto/chats"
	dtoMessage "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/dto/message"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/usecase/mocks"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// setupMessageUsecase создает MessageUsecase с настроенными mock-объектами для тестирования
func setupMessageUsecase(t *testing.T) (*MessageUsecase, *mocks.MockMessageRepository, *mocks.MockUserRepository, *mocks.MockChatsRepository, *mocks.MockFileStorage, *mocks.MockListenerMapInterface) {
	ctrl := gomock.NewController(t)

	mockMessageRepo := mocks.NewMockMessageRepository(ctrl)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockChatsRepo := mocks.NewMockChatsRepository(ctrl)
	mockFileStorage := mocks.NewMockFileStorage(ctrl)
	mockListenerMap := mocks.NewMockListenerMapInterface(ctrl)

	// Настраиваем mock для горутин
	mockListenerMap.EXPECT().GetChatListeners(gomock.Any()).Return(make(map[uuid.UUID]chan dtoMessage.WebSocketMessageDTO)).AnyTimes()
	mockListenerMap.EXPECT().CleanInactiveChats().Return(0).AnyTimes()
	mockListenerMap.EXPECT().CleanInactiveReaders().Return(0).AnyTimes()

	uc := NewMessageUsecase(mockMessageRepo, mockUserRepo, mockChatsRepo, mockFileStorage, mockListenerMap)

	return uc, mockMessageRepo, mockUserRepo, mockChatsRepo, mockFileStorage, mockListenerMap
}

func TestMessageUsecase_NewMessageUsecase(t *testing.T) {
	uc, mockMessageRepo, mockUserRepo, mockChatsRepo, mockFileStorage, mockListenerMap := setupMessageUsecase(t)
	defer uc.Stop()

	assert.NotNil(t, uc)
	assert.NotNil(t, uc.distributeChannel)
	assert.NotNil(t, uc.connectionContext)
	assert.NotNil(t, uc.connectionContextCount)
	assert.Equal(t, mockMessageRepo, uc.messageRepository)
	assert.Equal(t, mockUserRepo, uc.userClient)
	assert.Equal(t, mockChatsRepo, uc.chatsRepository)
	assert.Equal(t, mockFileStorage, uc.fileStorage)
	assert.Equal(t, mockListenerMap, uc.listenerMap)
}

func TestMessageUsecase_AddMessage_Success(t *testing.T) {
	uc, mockMessageRepo, mockUserRepo, mockChatsRepo, _, _ := setupMessageUsecase(t)
	defer uc.Stop()

	ctx := context.Background()
	userID := uuid.New()
	chatID := uuid.New()
	messageID := uuid.New()

	msg := dtoMessage.CreateMessageDTO{
		ChatId:    chatID,
		Text:      "Test message",
		CreatedAt: time.Now(),
	}

	user := &modelsUser.User{
		ID:   userID,
		Name: "Test User",
	}

	// Проверяем права пользователя (false означает, что пользователь НЕ viewer, то есть имеет права)
	mockChatsRepo.EXPECT().CheckUserHasRole(ctx, userID, chatID, modelsChats.RoleViewer).Return(false, nil)

	// Получаем пользователя
	mockUserRepo.EXPECT().GetUserByID(ctx, userID).Return(user, nil)

	// Вставляем сообщение в базу данных
	mockMessageRepo.EXPECT().InsertMessage(ctx, gomock.Any()).Return(messageID, nil)

	err := uc.AddMessage(ctx, msg, userID)

	assert.NoError(t, err)
}

func TestMessageUsecase_AddMessage_NoRights(t *testing.T) {
	uc, _, _, mockChatsRepo, _, _ := setupMessageUsecase(t)
	defer uc.Stop()

	ctx := context.Background()
	userID := uuid.New()
	chatID := uuid.New()

	msg := dtoMessage.CreateMessageDTO{
		ChatId:    chatID,
		Text:      "Test message",
		CreatedAt: time.Now(),
	}

	// Пользователь имеет только права viewer (true означает, что он viewer)
	mockChatsRepo.EXPECT().CheckUserHasRole(ctx, userID, chatID, modelsChats.RoleViewer).Return(true, nil)

	err := uc.AddMessage(ctx, msg, userID)

	assert.Error(t, err)
	assert.Equal(t, errs.ErrNoRights, err)
}

func TestMessageUsecase_AddMessage_UserNotFound(t *testing.T) {
	uc, _, mockUserRepo, mockChatsRepo, _, _ := setupMessageUsecase(t)
	defer uc.Stop()

	ctx := context.Background()
	userID := uuid.New()
	chatID := uuid.New()

	msg := dtoMessage.CreateMessageDTO{
		ChatId:    chatID,
		Text:      "Test message",
		CreatedAt: time.Now(),
	}

	// Пользователь имеет права
	mockChatsRepo.EXPECT().CheckUserHasRole(ctx, userID, chatID, modelsChats.RoleViewer).Return(false, nil)

	// Пользователь не найден
	mockUserRepo.EXPECT().GetUserByID(ctx, userID).Return(nil, errors.New("user not found"))

	err := uc.AddMessage(ctx, msg, userID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user not found")
}

func TestMessageUsecase_SubscribeConnectionToChats_Success(t *testing.T) {
	uc, _, _, _, _, mockListenerMap := setupMessageUsecase(t)
	defer uc.Stop()

	ctx := context.Background()
	connectionID := uuid.New()
	userID := uuid.New()
	chatID := uuid.New()

	chatsViewDTO := []dtoChats.ChatViewInformationDTO{
		{
			ID:   chatID,
			Name: "Test Chat",
			Type: "group",
		},
	}

	resultChan := make(chan dtoMessage.WebSocketMessageDTO, 10)
	chatChan := make(<-chan dtoMessage.WebSocketMessageDTO, 10)

	// Инициализируем контекст в словарь перед тестом
	uc.mu.Lock()
	uc.connectionContext[connectionID] = ctx
	uc.mu.Unlock()

	mockListenerMap.EXPECT().GetOutgoingChannel(connectionID).Return(resultChan)
	mockListenerMap.EXPECT().RegisterUserConnection(userID, connectionID, resultChan)
	mockListenerMap.EXPECT().SubscribeConnectionToChat(connectionID, chatID, userID).Return(chatChan)

	outChan := uc.SubscribeConnectionToChats(ctx, connectionID, userID, chatsViewDTO)

	assert.NotNil(t, outChan)
}

func TestMessageUsecase_SubscribeUsersOnChat_Success(t *testing.T) {
	uc, mockMessageRepo, _, mockChatsRepo, _, mockListenerMap := setupMessageUsecase(t)
	defer uc.Stop()

	ctx := context.Background()
	chatID := uuid.New()
	userID := uuid.New()
	messageID := uuid.New()
	connectionID := uuid.New()

	members := []dtoChats.AddChatMemberDTO{
		{UserId: userID},
	}

	chat := &modelsChats.Chat{
		ID:   chatID,
		Type: "group",
		Name: "Test Group",
	}

	userName := "Test User"
	messages := []modelsMessage.Message{
		{
			ID:        messageID,
			ChatID:    chatID,
			UserID:    &userID,
			UserName:  &userName,
			Text:      "Last message",
			CreatedAt: time.Now(),
			Type:      "text",
		},
	}

	userConnections := map[uuid.UUID]chan dtoMessage.WebSocketMessageDTO{
		connectionID: make(chan dtoMessage.WebSocketMessageDTO),
	}

	outChannel := make(chan dtoMessage.WebSocketMessageDTO, 10)

	// Инициализируем контекст в словарь перед тестом
	uc.mu.Lock()
	uc.connectionContext[connectionID] = ctx
	uc.mu.Unlock()

	mockChatsRepo.EXPECT().GetChat(ctx, chatID).Return(chat, nil)
	mockMessageRepo.EXPECT().GetMessagesOfChat(ctx, chatID, 0, 1).Return(messages, nil)
	mockListenerMap.EXPECT().AddChatToUserSubscription(userID, chatID).Return(userConnections)
	mockListenerMap.EXPECT().GetOutgoingChannel(connectionID).Return(outChannel)

	err := uc.SubscribeUsersOnChat(ctx, chatID, members)

	assert.NoError(t, err)
}

func TestMessageUsecase_SubscribeUsersOnChat_ChatNotFound(t *testing.T) {
	uc, _, _, mockChatsRepo, _, _ := setupMessageUsecase(t)
	defer uc.Stop()

	ctx := context.Background()
	chatID := uuid.New()
	userID := uuid.New()

	members := []dtoChats.AddChatMemberDTO{
		{UserId: userID},
	}

	mockChatsRepo.EXPECT().GetChat(ctx, chatID).Return(nil, errors.New("chat not found"))

	err := uc.SubscribeUsersOnChat(ctx, chatID, members)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error during getting chat")
}

func TestMessageUsecase_Stop(t *testing.T) {
	uc, _, _, _, _, _ := setupMessageUsecase(t)

	// Проверяем, что контекст еще не отменен
	select {
	case <-uc.ctx.Done():
		t.Fatal("Context should not be cancelled yet")
	default:
		// OK
	}

	uc.Stop()

	// Проверяем, что контекст отменен
	select {
	case <-uc.ctx.Done():
		// OK
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Context should be cancelled")
	}
}
