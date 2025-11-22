package message

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	modelsChats "github.com/go-park-mail-ru/2025_2_Undefined/internal/models/chats"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/models/domains"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/models/errs"
	modelsMessage "github.com/go-park-mail-ru/2025_2_Undefined/internal/models/message"
	dtoChats "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/dto/chats"
	dtoMessage "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/dto/message"
	interfaceChatsUsecase "github.com/go-park-mail-ru/2025_2_Undefined/internal/usecase/interface/chats"
	interfaceListenerMap "github.com/go-park-mail-ru/2025_2_Undefined/internal/usecase/interface/listener"
	interfaceMessageUsecase "github.com/go-park-mail-ru/2025_2_Undefined/internal/usecase/interface/message"
	interfaceFileStorage "github.com/go-park-mail-ru/2025_2_Undefined/internal/usecase/interface/storage"
	interfaceUserUsecase "github.com/go-park-mail-ru/2025_2_Undefined/internal/usecase/interface/user"
	"github.com/google/uuid"
)

const (
	MessagesBufferForOneUserChat  = 50
	MessagesBufferForAllUserChats = 500
	MessagesGLobalBuffer          = 1000
	DistributorsCount             = 1
	ClearsCount                   = 1
)

type MessageUsecase struct {
	fileStorage       interfaceFileStorage.FileStorage
	messageRepository interfaceMessageUsecase.MessageRepository
	userRepository    interfaceUserUsecase.UserRepository
	chatsRepository   interfaceChatsUsecase.ChatsRepository

	listenerMap                    interfaceListenerMap.ListenerMapInterface
	distributeChannel              chan dtoMessage.WebSocketMessageDTO
	connectionContext              map[uuid.UUID]context.Context
	connectionContextCount         map[uuid.UUID]int
	mu                             sync.RWMutex
	distributersToOutChannelsCount atomic.Int32

	ctx    context.Context
	cancel context.CancelFunc
}

func NewMessageUsecase(messageRepository interfaceMessageUsecase.MessageRepository, userRepository interfaceUserUsecase.UserRepository, chatsRepository interfaceChatsUsecase.ChatsRepository, fileStorage interfaceFileStorage.FileStorage, listenerMap interfaceListenerMap.ListenerMapInterface) *MessageUsecase {
	ctx, cancel := context.WithCancel(context.Background())
	uc := &MessageUsecase{
		listenerMap:            listenerMap,
		messageRepository:      messageRepository,
		userRepository:         userRepository,
		chatsRepository:        chatsRepository,
		fileStorage:            fileStorage,
		distributeChannel:      make(chan dtoMessage.WebSocketMessageDTO, MessagesGLobalBuffer),
		ctx:                    ctx,
		cancel:                 cancel,
		connectionContext:      make(map[uuid.UUID]context.Context),
		connectionContextCount: make(map[uuid.UUID]int),
	}

	for i := 0; i < DistributorsCount; i++ {
		go uc.distribute(uc.ctx)
	}

	for i := 0; i < ClearsCount; i++ {
		go uc.chatCleaner(uc.ctx)
		go uc.readerCleaner(uc.ctx)
	}

	return uc
}

func (uc *MessageUsecase) AddMessage(ctx context.Context, msg dtoMessage.CreateMessageDTO, userId uuid.UUID) error {
	const op = "MessageUsecase.AddMessage"

	logger := domains.GetLogger(ctx).WithField("operation", op)

	ok, err := uc.chatsRepository.CheckUserHasRole(ctx, userId, msg.ChatId, modelsChats.RoleViewer)
	if err != nil {
		logger.WithError(err).Errorf("could not check user %s role in chat %s", userId, msg.ChatId)
		return err
	}

	// Пользователь не имеет прав на отправку сообщений в этот чат
	if ok {
		logger.WithError(err).Warningf("not enough rights to add message to chat %s by user %s", msg.ChatId, userId)
		return errs.ErrNoRights
	}

	user, err := uc.userRepository.GetUserByID(ctx, userId)
	if err != nil {
		logger.WithError(err).Warningf("could not get user %s", userId)
		return err
	}

	if msg.CreatedAt.IsZero() {
		msg.CreatedAt = time.Now()
	}

	msgCreateModel := modelsMessage.CreateMessage{
		ChatID:    msg.ChatId,
		Text:      msg.Text,
		CreatedAt: msg.CreatedAt,
		Type:      modelsMessage.MessageTypeUser,
		UserID:    &user.ID,
	}

	msgID, err := uc.messageRepository.InsertMessage(ctx, msgCreateModel)
	if err != nil {
		return err
	}

	msgDTO := dtoMessage.MessageDTO{
		ID:         msgID,
		SenderID:   &user.ID,
		SenderName: user.Name,
		Text:       msg.Text,
		CreatedAt:  msg.CreatedAt,
		UpdatedAt:  nil,
		ChatID:     msg.ChatId,
		Type:       modelsMessage.MessageTypeUser,
	}

	select {
	case uc.distributeChannel <- dtoMessage.WebSocketMessageDTO{
		Type:   dtoMessage.WebSocketMessageTypeNewChatMessage,
		ChatID: msg.ChatId,
		Value:  msgDTO,
	}:
		// Всё ок :-)
	case <-time.After(10 * time.Second):
		return errs.ErrServiceIsOverloaded
	}

	return nil
}

func (uc *MessageUsecase) SubscribeConnectionToChats(ctx context.Context, connectionID, userID uuid.UUID, chatsViewDTO []dtoChats.ChatViewInformationDTO) <-chan dtoMessage.WebSocketMessageDTO {
	const op = "MessageUsecase.SubscribeConnectionToChats"

	logger := domains.GetLogger(ctx).WithField("operation", op)
	logger.Debug("Start operation")

	resultChan := uc.listenerMap.GetOutgoingChannel(connectionID)

	uc.mu.Lock()
	uc.connectionContext[connectionID] = ctx
	uc.mu.Unlock()

	for _, chatViewDto := range chatsViewDTO {
		chatChan := uc.listenerMap.SubscribeConnectionToChat(connectionID, chatViewDto.ID, userID)
		// Fan-in :)
		uc.distributeToOutChannel(connectionID, chatChan, resultChan)
	}

	logger.Info("Succesfull completed")

	return resultChan
}

func (uc *MessageUsecase) SubscribeUsersOnChat(ctx context.Context, chatID uuid.UUID, members []dtoChats.AddChatMemberDTO) error {
	const op = "MessageUsecase.SubscribeUsersOnChat"

	logger := domains.GetLogger(ctx).WithField("operation", op)
	logger.Debug("Start operation")

	chatView, err := uc.chatsRepository.GetChat(ctx, chatID)
	if err != nil {
		return fmt.Errorf("error during getting chat: %w", err)
	}

	lastMessage, err := uc.messageRepository.GetMessagesOfChat(ctx, chatID, 0, 1)
	if err != nil {
		return fmt.Errorf("error during getting last message: %w", err)
	}

	for _, member := range members {
		userConnections := uc.listenerMap.AddChatToUserSubscription(member.UserId, chatView.ID)
		for connectionID, connectionChan := range userConnections {
			connectionOutChannel := uc.listenerMap.GetOutgoingChannel(connectionID)
			connectionOutChannel <- dtoMessage.WebSocketMessageDTO{
				Type:   dtoMessage.WebSocketMessageTypeCreatedNewChat,
				ChatID: chatView.ID,
				Value: dtoChats.ChatViewInformationDTO{
					ID:   chatView.ID,
					Name: chatView.Name,
					LastMessage: dtoMessage.MessageDTO{
						ID:         lastMessage[0].ID,
						SenderID:   lastMessage[0].UserID,
						SenderName: lastMessage[0].UserName,
						Text:       lastMessage[0].Text,
						CreatedAt:  lastMessage[0].CreatedAt,
						UpdatedAt:  lastMessage[0].UpdatedAt,
						ChatID:     lastMessage[0].ChatID,
						Type:       lastMessage[0].Type,
					},
					Type: chatView.Type,
				},
			}

			uc.distributeToOutChannel(connectionID, connectionChan, connectionOutChannel)
		}
	}

	logger.Info("Succesfull completed")

	return nil
}

func (uc *MessageUsecase) distributeToOutChannel(connectionID uuid.UUID, in <-chan dtoMessage.WebSocketMessageDTO, out chan<- dtoMessage.WebSocketMessageDTO) {
	uc.mu.Lock()
	ctx := uc.connectionContext[connectionID]
	uc.connectionContextCount[connectionID]++
	uc.mu.Unlock()

	go func(chatChan <-chan dtoMessage.WebSocketMessageDTO) {
		uc.distributersToOutChannelsCount.Add(1)
		defer func() {
			uc.distributersToOutChannelsCount.Add(-1)

			uc.mu.Lock()
			if uc.connectionContextCount[connectionID] == 0 {
				delete(uc.connectionContextCount, connectionID)
			}

			uc.mu.Unlock()
		}()

		for {
			select {
			case ch := <-chatChan:
				out <- ch

			case <-ctx.Done():
				return
			}
		}
	}(in)
}

func (uc *MessageUsecase) Stop() {
	uc.cancel()
}

func (uc *MessageUsecase) EditMessage(ctx context.Context, msg dtoMessage.EditMessageDTO, userID uuid.UUID) error {
	const op = "MessageUsecase.EditMessage"
	logger := domains.GetLogger(ctx).WithField("operation", op)

	message, err := uc.messageRepository.GetMessageByID(ctx, msg.ID)
	if err != nil {
		logger.WithError(err).Error("failed to get message")
		return err
	}

	if message.UserID == nil || *message.UserID != userID {
		logger.Warning("user is not the author of the message")
		return errs.ErrNoRights
	}

	if msg.UpdatedAt.IsZero() {
		msg.UpdatedAt = time.Now()
	}

	if err := uc.messageRepository.UpdateMessage(ctx, msg.ID, msg.Text); err != nil {
		logger.WithError(err).Error("failed to update message")
		return err
	}

	select {
	case uc.distributeChannel <- dtoMessage.WebSocketMessageDTO{
		Type:   dtoMessage.WebSocketMessageTypeEditChatMessage,
		ChatID: message.ChatID,
		Value:  msg,
	}:
		// Всё ок :-)
	case <-time.After(10 * time.Second):
		return errs.ErrServiceIsOverloaded
	}

	return nil
}

func (uc *MessageUsecase) DeleteMessage(ctx context.Context, msg dtoMessage.DeleteMessageDTO, userID uuid.UUID) error {
	const op = "MessageUsecase.DeleteMessage"
	logger := domains.GetLogger(ctx).WithField("operation", op)

	message, err := uc.messageRepository.GetMessageByID(ctx, msg.ID)
	if err != nil {
		logger.WithError(err).Error("failed to get message")
		return err
	}

	isAdmin, err := uc.chatsRepository.CheckUserHasRole(ctx, userID, message.ChatID, modelsChats.RoleAdmin)
	if err != nil {
		logger.WithError(err).Error("failed to check user role")
		return err
	}

	if message.UserID == nil || *message.UserID != userID && !isAdmin {
		logger.Warning("user is not the author or admin")
		return errs.ErrNoRights
	}

	if err := uc.messageRepository.DeleteMessage(ctx, msg.ID); err != nil {
		logger.WithError(err).Error("failed to delete message")
		return err
	}

	select {
	case uc.distributeChannel <- dtoMessage.WebSocketMessageDTO{
		Type:   dtoMessage.WebSocketMessageTypeDeleteChatMessage,
		ChatID: message.ChatID,
		Value:  msg,
	}:
		// Всё ок :-)
	case <-time.After(10 * time.Second):
		return errs.ErrServiceIsOverloaded
	}

	return nil
}
