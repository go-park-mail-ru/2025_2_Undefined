package message

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	modelsAttachment "github.com/go-park-mail-ru/2025_2_Undefined/internal/models/attachment"
	modelsChats "github.com/go-park-mail-ru/2025_2_Undefined/internal/models/chats"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/models/domains"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/models/errs"
	modelsMessage "github.com/go-park-mail-ru/2025_2_Undefined/internal/models/message"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/repository/minio"
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
	userClient        interfaceUserUsecase.UserClient
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

func NewMessageUsecase(messageRepository interfaceMessageUsecase.MessageRepository, userClient interfaceUserUsecase.UserClient, chatsRepository interfaceChatsUsecase.ChatsRepository, fileStorage interfaceFileStorage.FileStorage, listenerMap interfaceListenerMap.ListenerMapInterface) *MessageUsecase {
	ctx, cancel := context.WithCancel(context.Background())
	uc := &MessageUsecase{
		listenerMap:            listenerMap,
		messageRepository:      messageRepository,
		userClient:             userClient,
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

	user, err := uc.userClient.GetUserByID(ctx, userId)
	if err != nil {
		logger.WithError(err).Warningf("could not get user %s", userId)
		return err
	}

	if msg.CreatedAt.IsZero() {
		msg.CreatedAt = time.Now()
	}

	// Валидация: сообщения со стикерами, кружками или голосывами не должны содержать текст
	if msg.Attachment != nil && msg.Text != "" && (msg.Attachment.Type == modelsAttachment.AttachmentTypeSticker || msg.Attachment.Type == modelsAttachment.AttachmentTypeVoice || msg.Attachment.Type == modelsAttachment.AttachmentTypeVideoNote) {
		logger.Warningf("message with attachment cannot have text")
		return fmt.Errorf("message with %s attachment can not has text", msg.Attachment.Type)
	}

	msgCreateModel := modelsMessage.CreateMessage{
		ChatID:    msg.ChatId,
		Text:      msg.Text,
		CreatedAt: msg.CreatedAt,
		Type:      modelsMessage.MessageTypeUser,
		UserID:    &user.ID,
	}

	var msgID uuid.UUID
	var attachmentDTO *dtoMessage.AttachmentDTO

	// Обработка вложений
	if msg.Attachment != nil {
		// Валидация типа вложения
		if msg.Attachment.Type == "" {
			logger.Warning("attachment type is required")
			return errors.New("attachment type is required")
		}

		// Проверка, что тип валидный
		validTypes := map[string]bool{
			modelsAttachment.AttachmentTypeSticker:   true,
			modelsAttachment.AttachmentTypeVoice:     true,
			modelsAttachment.AttachmentTypeVideoNote: true,
			modelsAttachment.AttachmentTypeImage:     true,
			modelsAttachment.AttachmentTypeDocument:  true,
			modelsAttachment.AttachmentTypeAudio:     true,
			modelsAttachment.AttachmentTypeVideo:     true,
		}

		if !validTypes[msg.Attachment.Type] {
			logger.Warningf("invalid attachment type: %s", msg.Attachment.Type)
			return errors.New("invalid attachment type")
		}

		if msg.Attachment.Type == modelsAttachment.AttachmentTypeSticker {
			attachmentID := uuid.New()
			msgCreateModel.Attachment = &modelsAttachment.CreateAttachment{
				ID:                 attachmentID,
				Type:               &msg.Attachment.Type,
				FileName:           msg.Attachment.AttachmentID, // ID стикера храним в file_name
				FileSize:           0,
				ContentDisposition: "sticker",
				Duration:           nil,
			}

			msgID, err = uc.messageRepository.InsertMessageWithAttachment(ctx, msgCreateModel)
			if err != nil {
				logger.WithError(err).Error("could not insert message with sticker")
				return err
			}

			attachmentDTO = &dtoMessage.AttachmentDTO{
				ID:       &attachmentID,
				Type:     &msg.Attachment.Type,
				FileURL:  msg.Attachment.AttachmentID,
				Duration: nil,
			}
		} else {
			// В msg.Attachment.FileURL приходит attachment_id (в виде строки)
			attachmentID, err := uuid.Parse(msg.Attachment.AttachmentID)
			if err != nil {
				logger.WithError(err).Warningf("invalid attachment_id: %s", msg.Attachment.AttachmentID)
				return errors.New("invalid attachment_id")
			}

			// Проверяем, что вложение принадлежит текущему пользователю
			isOwner, err := uc.messageRepository.CheckAttachmentOwnership(ctx, attachmentID, userId)
			if err != nil {
				logger.WithError(err).Errorf("could not check attachment %s ownership", attachmentID)
				return err
			}

			if !isOwner {
				logger.Warningf("user %s does not own attachment %s", userId, attachmentID)
				return errs.ErrNoRights
			}

			// Получаем информацию о вложении
			attachment, err := uc.messageRepository.GetAttachmentByID(ctx, attachmentID)
			if err != nil {
				logger.WithError(err).Errorf("could not get attachment %s", attachmentID)
				return err
			}

			// Обновляем тип вложения из сообщения
			err = uc.messageRepository.UpdateAttachmentType(ctx, attachmentID, msg.Attachment.Type)
			if err != nil {
				logger.WithError(err).Errorf("could not update attachment %s type", attachmentID)
				return err
			}

			// Создаём обычное сообщение
			msgID, err = uc.messageRepository.InsertMessage(ctx, msgCreateModel)
			if err != nil {
				logger.WithError(err).Error("could not insert message")
				return err
			}

			err = uc.messageRepository.LinkAttachmentToMessage(ctx, msgID, attachmentID, userId)
			if err != nil {
				logger.WithError(err).Error("could not link attachment to message")
				return err
			}

			attachmentURL, err := uc.fileStorage.GetOne(ctx, &attachment.ID)
			if err != nil {
				logger.Warningf("could not get url of file with id %s", attachment.ID.String())
			}

			attachmentDTO = &dtoMessage.AttachmentDTO{
				ID:       &attachment.ID,
				Type:     &msg.Attachment.Type,
				FileURL:  attachmentURL,
				Duration: attachment.Duration,
			}
		}
	} else {
		// Обычное текстовое сообщение
		msgID, err = uc.messageRepository.InsertMessage(ctx, msgCreateModel)
		if err != nil {
			return err
		}
	}

	msgDTO := dtoMessage.MessageDTO{
		ID:         msgID,
		SenderID:   &user.ID,
		SenderName: &user.Name,
		Text:       msg.Text,
		CreatedAt:  msg.CreatedAt,
		UpdatedAt:  nil,
		ChatID:     msg.ChatId,
		Type:       modelsMessage.MessageTypeUser,
		Attachment: attachmentDTO,
	}

	err = uc.sendWebsocketMessage(dtoMessage.WebSocketMessageDTO{
		Type:   dtoMessage.WebSocketMessageTypeNewChatMessage,
		ChatID: msg.ChatId,
		Value:  msgDTO,
	})
	if err != nil {
		return err
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

	// Регистрируем пользователя в userConnections, даже если у него нет чатов
	// Это необходимо для корректной работы AddChatToUserSubscription при создании нового чата
	uc.listenerMap.RegisterUserConnection(userID, connectionID, resultChan)

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

	err = uc.sendWebsocketMessage(dtoMessage.WebSocketMessageDTO{
		Type:   dtoMessage.WebSocketMessageTypeEditChatMessage,
		ChatID: message.ChatID,
		Value:  msg,
	})
	if err != nil {
		return err
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

	err = uc.sendWebsocketMessage(dtoMessage.WebSocketMessageDTO{
		Type:   dtoMessage.WebSocketMessageTypeDeleteChatMessage,
		ChatID: message.ChatID,
		Value:  msg,
	})
	if err != nil {
		return err
	}

	return nil
}

// convertMessageToDTO преобразует модель Message в MessageDTO с вложениями
func (uc *MessageUsecase) convertMessageToDTO(ctx context.Context, msg modelsMessage.Message) dtoMessage.MessageDTO {
	var attachmentDTO *dtoMessage.AttachmentDTO

	if msg.Attachment != nil {
		attachmentURL, err := uc.fileStorage.GetOne(ctx, &msg.Attachment.ID)
		if err != nil {
			domains.GetLogger(ctx).WithError(err).Warningf("could not get url of file with id %s", msg.Attachment.ID.String())
			attachmentURL = "" // fallback
		}

		attachmentDTO = &dtoMessage.AttachmentDTO{
			ID:       &msg.Attachment.ID,
			Type:     msg.Attachment.Type,
			FileURL:  attachmentURL,
			Duration: msg.Attachment.Duration,
		}
	}

	return dtoMessage.MessageDTO{
		ID:         msg.ID,
		SenderID:   msg.UserID,
		SenderName: msg.UserName,
		Text:       msg.Text,
		CreatedAt:  msg.CreatedAt,
		UpdatedAt:  msg.UpdatedAt,
		ChatID:     msg.ChatID,
		Type:       msg.Type,
		Attachment: attachmentDTO,
	}
}

func (uc *MessageUsecase) GetMessagesBySearch(ctx context.Context, userID, chatID uuid.UUID, text string) ([]dtoMessage.MessageDTO, error) {
	const op = "MessageUsecase.GetMessagesBySearch"
	logger := domains.GetLogger(ctx).WithField("operation", op)

	messages, err := uc.messageRepository.SearchMessagesInChat(ctx, userID, chatID, text)
	if err != nil {
		logger.WithError(err).Error("failed to search messages")
		return nil, err
	}

	messagesDTO := make([]dtoMessage.MessageDTO, 0, len(messages))
	for _, msg := range messages {
		messagesDTO = append(messagesDTO, uc.convertMessageToDTO(ctx, msg))
	}

	return messagesDTO, nil
}

func (uc *MessageUsecase) GetChatMessages(ctx context.Context, userID, chatID uuid.UUID, offset, limit int) ([]dtoMessage.MessageDTO, error) {
	const op = "MessageUsecase.GetChatMessages"
	logger := domains.GetLogger(ctx).WithField("operation", op)

	messages, err := uc.messageRepository.GetMessagesOfChat(ctx, chatID, offset, limit)
	if err != nil {
		logger.WithError(err).Error("failed to get chat messages")
		return nil, err
	}

	messagesDTO := make([]dtoMessage.MessageDTO, 0, len(messages))
	for _, msg := range messages {
		messagesDTO = append(messagesDTO, uc.convertMessageToDTO(ctx, msg))
	}

	return messagesDTO, nil
}

func (uc *MessageUsecase) AddMessageJoinUsers(ctx context.Context, chatID uuid.UUID, users []dtoChats.AddChatMemberDTO) error {
	chat, err := uc.chatsRepository.GetChat(ctx, chatID)
	if err != nil {
		return err
	}

	if chat.Type == modelsChats.ChatTypeGroup {
		usersIDs := make([]uuid.UUID, len(users))
		for i, user := range users {
			usersIDs[i] = user.UserId
		}

		usersNames, err := uc.userClient.GetUsersNames(ctx, usersIDs)
		if err != nil {
			return err
		}

		for i := range users {
			messageID, err := uc.messageRepository.InsertMessage(ctx, modelsMessage.CreateMessage{
				ChatID:    chatID,
				UserID:    &users[i].UserId,
				Text:      fmt.Sprintf("Пользователь %s вступил в группу", usersNames[i]),
				Type:      modelsMessage.MessageTypeSystem,
				CreatedAt: time.Now(),
			})

			if err != nil {
				return err
			}

			now := time.Now()

			uc.sendWebsocketMessage(dtoMessage.WebSocketMessageDTO{
				Type:   dtoMessage.WebSocketMessageTypeNewChatMessage,
				ChatID: chatID,
				Value: dtoMessage.MessageDTO{
					ID:        messageID,
					SenderID:  &users[i].UserId,
					Text:      fmt.Sprintf("Пользователь %s вступил в группу", usersNames[i]),
					CreatedAt: now,
					UpdatedAt: &now,
					ChatID:    chatID,
					Type:      modelsMessage.MessageTypeSystem,
				},
			})
		}
	}
	return nil
}

func (uc *MessageUsecase) sendWebsocketMessage(msg dtoMessage.WebSocketMessageDTO) error {
	select {
	case uc.distributeChannel <- msg:
		// Всё ок :-)
		return nil
	case <-time.After(10 * time.Second):
		return errs.ErrServiceIsOverloaded
	}
}

func (uc *MessageUsecase) UploadAttachment(ctx context.Context, userID, chatID uuid.UUID, contentType string, fileData []byte, filename string, duration *int) (*dtoMessage.AttachmentDTO, error) {
	const op = "MessageUsecase.UploadAttachment"

	logger := domains.GetLogger(ctx).WithField("operation", op)

	// Проверяем права пользователя
	ok, err := uc.chatsRepository.CheckUserHasRole(ctx, userID, chatID, modelsChats.RoleViewer)
	if err != nil {
		logger.WithError(err).Errorf("could not check user %s role in chat %s", userID, chatID)
		return nil, err
	}

	if ok {
		logger.Warningf("not enough rights to upload attachment to chat %s by user %s", chatID, userID)
		return nil, errs.ErrNoRights
	}

	attachmentID := uuid.New()

	// Truncate filename to 255 characters to comply with database constraint
	truncatedFilename := filename
	if len(filename) > 255 {
		truncatedFilename = filename[:255]
		logger.Warningf("filename truncated from %d to 255 characters", len(filename))
	}

	fileURL, err := uc.fileStorage.CreateOne(ctx, minio.FileData{
		Name:        truncatedFilename,
		Data:        fileData,
		ContentType: contentType,
	}, attachmentID)
	if err != nil {
		logger.WithError(err).Error("could not upload file to storage")
		return nil, err
	}

	attachment := modelsAttachment.CreateAttachment{
		ID:                 attachmentID,
		Type:               nil,
		FileName:           truncatedFilename,
		FileSize:           int64(len(fileData)),
		ContentDisposition: contentType,
		Duration:           duration,
	}

	err = uc.messageRepository.InsertAttachment(ctx, attachment, userID)
	if err != nil {
		logger.WithError(err).Error("could not insert attachment to database")
		// TODO: удалить файл из MinIO при ошибке БД (cleanup)
		return nil, err
	}

	return &dtoMessage.AttachmentDTO{
		ID:       &attachmentID,
		Type:     nil,
		FileURL:  fileURL,
		Duration: duration,
	}, nil
}
