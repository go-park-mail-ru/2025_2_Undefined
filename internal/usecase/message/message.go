package message

import (
	"context"
	"sync"
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
	ChatsRepository   interfaceChatsUsecase.ChatsRepository

	listenerMap       interfaceListenerMap.ListenerMapInterface
	distributeChannel chan dtoMessage.MessageDTO
	ctx               context.Context
	cancel            context.CancelFunc
}

func NewMessageUsecase(messageRepository interfaceMessageUsecase.MessageRepository, userRepository interfaceUserUsecase.UserRepository, chatsRepository interfaceChatsUsecase.ChatsRepository, fileStorage interfaceFileStorage.FileStorage, listenerMap interfaceListenerMap.ListenerMapInterface) *MessageUsecase {
	ctx, cancel := context.WithCancel(context.Background())
	uc := &MessageUsecase{
		listenerMap:       listenerMap,
		messageRepository: messageRepository,
		userRepository:    userRepository,
		ChatsRepository:   chatsRepository,
		fileStorage:       fileStorage,
		distributeChannel: make(chan dtoMessage.MessageDTO, MessagesGLobalBuffer),
		ctx:               ctx,
		cancel:            cancel,
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

	ok, err := uc.ChatsRepository.CheckUserHasRole(ctx, userId, msg.ChatId, modelsChats.RoleViewer)
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

	avatar_url, err := uc.fileStorage.GetOne(ctx, user.AvatarID)
	if err != nil {
		logger.WithError(err).Warningf("could not get avatar URL for user %s: %v", user.ID, err)
		avatar_url = ""
	}

	msgDTO := dtoMessage.MessageDTO{
		SenderName:      user.Name,
		SenderAvatarURL: avatar_url,
		Text:            msg.Text,
		CreatedAt:       msg.CreatedAt,
		ChatId:          msg.ChatId,
		Type:            modelsMessage.MessageTypeUser,
	}

	select {
	case uc.distributeChannel <- msgDTO:
		// Всё ок :-)
	case <-time.After(5 * time.Second):
		return errs.ErrServiceIsOverloaded
	}

	msgCreateModel := modelsMessage.CreateMessage{
		ChatID:    msg.ChatId,
		Text:      msg.Text,
		CreatedAt: msg.CreatedAt,
		Type:      modelsMessage.MessageTypeUser,
		UserID:    userId,
	}

	_, err = uc.messageRepository.InsertMessage(ctx, msgCreateModel)
	if err != nil {
		return err
	}

	return nil
}

func (uc *MessageUsecase) SubscribeUserToChats(ctx context.Context, userId uuid.UUID, chatsViewDTO []dtoChats.ChatViewInformationDTO) <-chan dtoMessage.MessageDTO {
	resultChan := make(chan dtoMessage.MessageDTO, MessagesBufferForAllUserChats)
	var once sync.Once

	for _, chatViewDto := range chatsViewDTO {
		chatChan := uc.listenerMap.SubscribeUserToChat(userId, chatViewDto.ID)
		// Fan-in :)
		go func(chatChan <-chan dtoMessage.MessageDTO) {
			defer once.Do(func() {
				close(resultChan)
			})

			for {
				select {
				case ch := <-chatChan:
					resultChan <- ch

				case <-ctx.Done():
					return
				}
			}
		}(chatChan)
	}

	return resultChan
}

func (uc *MessageUsecase) Stop() {
	uc.cancel()
}
