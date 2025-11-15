package usecase

import (
	"context"
	"fmt"
	"time"

	modelsChats "github.com/go-park-mail-ru/2025_2_Undefined/internal/models/chats"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/models/domains"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/models/errs"
	modelsMessage "github.com/go-park-mail-ru/2025_2_Undefined/internal/models/message"
	dtoChats "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/dto/chats"
	dtoMessage "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/dto/message"
	dtoUtils "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/dto/utils"
	interfaceChatsRepository "github.com/go-park-mail-ru/2025_2_Undefined/internal/usecase/interface/chats"
	interfaceMessageRepository "github.com/go-park-mail-ru/2025_2_Undefined/internal/usecase/interface/message"
	interfaceFileStorage "github.com/go-park-mail-ru/2025_2_Undefined/internal/usecase/interface/storage"
	interfaceUserRepository "github.com/go-park-mail-ru/2025_2_Undefined/internal/usecase/interface/user"
	"github.com/google/uuid"
)

type ChatsUsecase struct {
	chatsRepo   interfaceChatsRepository.ChatsRepository
	messageRepo interfaceMessageRepository.MessageRepository
	usersRepo   interfaceUserRepository.UserRepository
	fileStorage interfaceFileStorage.FileStorage
}

func NewChatsUsecase(chatsRepo interfaceChatsRepository.ChatsRepository, usersRepo interfaceUserRepository.UserRepository, messageRepo interfaceMessageRepository.MessageRepository, fileStorage interfaceFileStorage.FileStorage) *ChatsUsecase {
	return &ChatsUsecase{
		chatsRepo:   chatsRepo,
		messageRepo: messageRepo,
		usersRepo:   usersRepo,
		fileStorage: fileStorage,
	}
}

func (uc *ChatsUsecase) GetChats(ctx context.Context, userId uuid.UUID) ([]dtoChats.ChatViewInformationDTO, error) {
	const op = "ChatsUsecase.GetChats"

	logger := domains.GetLogger(ctx).WithField("operation", op)

	chats, err := uc.chatsRepo.GetChats(ctx, userId)
	if err != nil {
		return nil, err
	}

	lastMessages, err := uc.messageRepo.GetLastMessagesOfChats(ctx, userId)
	if err != nil {
		return nil, err
	}

	// Создаем мапу для быстрого поиска последних сообщений по chat_id
	messageMap := make(map[uuid.UUID]modelsMessage.Message, len(lastMessages))
	for _, msg := range lastMessages {
		messageMap[msg.ChatID] = msg
	}

	result := make([]dtoChats.ChatViewInformationDTO, 0, len(chats))
	for _, chat := range chats {
		chatName := chat.Name
		chatAvatarUrl := ""

		// Для диалогов определяем имя собеседника и аватарку
		if chat.Type == modelsChats.ChatTypeDialog {
			users, err := uc.chatsRepo.GetUsersOfChat(ctx, chat.ID)
			if err != nil {
				logger.Warningf("could not get users for dialog %s: %v", chat.ID, err)
			} else {
				// Ищем собеседника (не текущего пользователя)
				for _, user := range users {
					if user.UserID != userId {
						chatName = user.UserName
						chatAvatarUrl, err = uc.fileStorage.GetOne(ctx, user.UserAvatarID)
						if err != nil {
							logger.Warningf("could not get user avatar for dialog %s: %v", chat.ID, err)
						}

						break
					}
				}
			}
		}

		chatDTO := dtoChats.ChatViewInformationDTO{
			ID:        chat.ID,
			Name:      chatName,
			Type:      chat.Type,
			AvatarURL: chatAvatarUrl,
		}

		if lastMsg, exists := messageMap[chat.ID]; exists {
			avatarURL, err := uc.fileStorage.GetOne(ctx, lastMsg.UserAvatarID)
			if err != nil {
				logger.Warningf("could not get avatar URL for user %s: %v", lastMsg.UserID, err)
				avatarURL = ""
			}

			chatDTO.LastMessage = dtoMessage.MessageDTO{
				SenderID:        lastMsg.UserID,
				SenderName:      lastMsg.UserName,
				Text:            lastMsg.Text,
				CreatedAt:       lastMsg.CreatedAt,
				SenderAvatarURL: avatarURL,
				ChatId:          lastMsg.ChatID,
				Type:            lastMsg.Type,
			}
		}

		result = append(result, chatDTO)
	}

	return result, nil
}

func (uc *ChatsUsecase) GetInformationAboutChat(ctx context.Context, userID, chatID uuid.UUID) (*dtoChats.ChatDetailedInformationDTO, error) {
	const op = "ChatsUsecase.GetInformationAboutChat"

	logger := domains.GetLogger(ctx).WithField("operation", op)

	chat, err := uc.chatsRepo.GetChat(ctx, chatID)
	if err != nil {
		return nil, err
	}

	messages, err := uc.messageRepo.GetMessagesOfChat(ctx, chatID, 0, 40)
	if err != nil {
		return nil, err
	}

	users, err := uc.chatsRepo.GetUsersOfChat(ctx, chatID)
	if err != nil {
		return nil, err
	}

	userInfo, err := uc.chatsRepo.GetUserInfo(ctx, userID, chatID)
	if err != nil {
		return nil, err
	}

	// Проверяем, что пользователь есть в участниках чата. Иначе - возвращаем ошибку
	isUserInChat := false
	for _, user := range users {
		if user.UserID == userID {
			isUserInChat = true
			break
		}
	}
	if !isUserInChat {
		return nil, errs.ErrNotFound
	}

	messagesDTO := make([]dtoMessage.MessageDTO, len(messages))
	for i, message := range messages {
		avatarURL, err := uc.fileStorage.GetOne(ctx, message.UserAvatarID)
		if err != nil {
			logger.Warningf("could not get avatar URL for user %s: %v", message.UserID, err)
			avatarURL = ""
		}

		messagesDTO[i] = dtoMessage.MessageDTO{
			SenderID:        message.UserID,
			SenderName:      message.UserName,
			Text:            message.Text,
			CreatedAt:       message.CreatedAt,
			SenderAvatarURL: avatarURL,
			ChatId:          message.ChatID,
			Type:            message.Type,
		}
	}

	usersDTO := make([]dtoChats.UserInfoChatDTO, len(users))
	for i, user := range users {
		avatarURL, err := uc.fileStorage.GetOne(ctx, user.UserAvatarID)
		if err != nil {
			logger.Warningf("could not get avatar URL for user %s: %v", user.UserID, err)
			avatarURL = ""
		}

		usersDTO[i] = dtoChats.UserInfoChatDTO{
			UserId:     user.UserID,
			UserName:   user.UserName,
			UserAvatar: avatarURL,
			Role:       user.Role,
		}
	}

	var isAdmin, canChat, isMember, isPrivate bool = false, false, false, false
	switch userInfo.Role {
	case modelsChats.RoleAdmin:
		isAdmin = true
		fallthrough
	case modelsChats.RoleMember:
		isMember = true
		canChat = true
	}

	if chat.Type == modelsChats.ChatTypeDialog {
		isPrivate = true
	}

	avatarURL := ""
	// Определяем название чата
	chatName := chat.Name
	if chat.Type == modelsChats.ChatTypeDialog {
		// Для диалогов название - это имя собеседника
		for _, user := range usersDTO {
			if user.UserId != userID {
				chatName = user.UserName
				avatarURL = user.UserAvatar
				break
			}
		}
	}

	result := &dtoChats.ChatDetailedInformationDTO{
		ID:          chat.ID,
		Name:        chatName,
		IsAdmin:     isAdmin,
		CanChat:     canChat,
		IsMember:    isMember,
		IsPrivate:   isPrivate,
		Type:        chat.Type,
		Members:     usersDTO,
		Messages:    messagesDTO,
		Description: chat.Description,
		AvatarURL:   avatarURL,
	}

	return result, nil
}

func (uc *ChatsUsecase) GetUsersDialog(ctx context.Context, user1ID, user2ID uuid.UUID) (*dtoUtils.IdDTO, error) {
	idDialog, err := uc.chatsRepo.GetUsersDialog(ctx, user1ID, user2ID)
	if err != nil {
		return nil, err
	}

	resultDTO := &dtoUtils.IdDTO{
		ID: idDialog,
	}

	return resultDTO, nil
}

func (uc *ChatsUsecase) CreateChat(ctx context.Context, chatDTO dtoChats.ChatCreateInformationDTO) (uuid.UUID, error) {
	chat := modelsChats.Chat{
		ID:          uuid.New(),
		Name:        chatDTO.Name,
		Type:        chatDTO.Type,
		Description: "",
	}

	usersIds := make([]uuid.UUID, len(chatDTO.Members))
	for i, memberDTO := range chatDTO.Members {
		usersIds[i] = memberDTO.UserId
	}

	usersNames, err := uc.usersRepo.GetUsersNames(ctx, usersIds)
	if err != nil {
		return uuid.Nil, err
	}

	usersInfo := make([]modelsChats.UserInfo, len(chatDTO.Members))
	for i, memberDTO := range chatDTO.Members {
		usersInfo[i] = modelsChats.UserInfo{
			UserID: memberDTO.UserId,
			ChatID: chat.ID,
			Role:   memberDTO.Role,
		}
	}

	err = uc.chatsRepo.CreateChat(ctx, chat, usersInfo, usersNames)
	if err != nil {
		return uuid.Nil, err
	}
	return chat.ID, nil
}

func (uc *ChatsUsecase) AddUsersToChat(ctx context.Context, chatID, userID uuid.UUID, users []dtoChats.AddChatMemberDTO) error {
	ok, err := uc.chatsRepo.CheckUserHasRole(ctx, userID, chatID, modelsChats.RoleAdmin)
	if err != nil {
		return err
	}

	if !ok {
		return errs.ErrNoRights
	}

	usersInfo := make([]modelsChats.UserInfo, len(users))
	usersIDs := make([]uuid.UUID, len(users))
	for i, user := range users {
		usersInfo[i] = modelsChats.UserInfo{
			UserID: user.UserId,
			ChatID: chatID,
			Role:   user.Role,
		}

		usersIDs[i] = user.UserId
	}

	err = uc.chatsRepo.InsertUsersToChat(ctx, chatID, usersInfo)
	if err != nil {
		return err
	}

	chat, err := uc.chatsRepo.GetChat(ctx, chatID)
	if err != nil {
		return err
	}

	if chat.Type == modelsChats.ChatTypeGroup {
		usersNames, err := uc.usersRepo.GetUsersNames(ctx, usersIDs)
		if err != nil {
			return err
		}

		for i := range users {
			_, err = uc.messageRepo.InsertMessage(ctx, modelsMessage.CreateMessage{
				ChatID:    chatID,
				UserID:    &users[i].UserId,
				Text:      fmt.Sprintf("Пользователь %s вступил в группу", usersNames[i]),
				Type:      modelsMessage.MessageTypeSystem,
				CreatedAt: time.Now(),
			})

			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (uc *ChatsUsecase) DeleteChat(ctx context.Context, userId, chatId uuid.UUID) error {
	return uc.chatsRepo.DeleteChat(ctx, userId, chatId)
}

func (uc *ChatsUsecase) UpdateChat(ctx context.Context, userId, chatId uuid.UUID, name, description string) error {
	return uc.chatsRepo.UpdateChat(ctx, userId, chatId, name, description)
}
