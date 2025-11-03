package usecase

import (
	"context"

	modelsChats "github.com/go-park-mail-ru/2025_2_Undefined/internal/models/chats"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/models/domains"
	modelsMessage "github.com/go-park-mail-ru/2025_2_Undefined/internal/models/message"
	dtoChats "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/dto/chats"
	dtoMessage "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/dto/message"
	interfaceChatsUsecase "github.com/go-park-mail-ru/2025_2_Undefined/internal/usecase/interface/chats"
	interfaceFileStorage "github.com/go-park-mail-ru/2025_2_Undefined/internal/usecase/interface/storage"
	interfaceUserUsecase "github.com/go-park-mail-ru/2025_2_Undefined/internal/usecase/interface/user"
	"github.com/google/uuid"
)

type ChatsUsecase struct {
	chatsRepo   interfaceChatsUsecase.ChatsRepository
	usersRepo   interfaceUserUsecase.UserRepository
	fileStorage interfaceFileStorage.FileStorage
}

func NewChatsUsecase(chatsRepo interfaceChatsUsecase.ChatsRepository, usersRepo interfaceUserUsecase.UserRepository, fileStorage interfaceFileStorage.FileStorage) *ChatsUsecase {
	return &ChatsUsecase{
		chatsRepo:   chatsRepo,
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

	lastMessages, err := uc.chatsRepo.GetLastMessagesOfChats(ctx, userId)
	if err != nil {
		return nil, err
	}

	// Создаем мапу для быстрого поиска последних сообщений по chat_id
	messageMap := make(map[uuid.UUID]modelsMessage.Message)
	for _, msg := range lastMessages {
		messageMap[msg.ChatID] = msg
	}

	result := make([]dtoChats.ChatViewInformationDTO, 0, len(chats))
	for _, chat := range chats {
		chatDTO := dtoChats.ChatViewInformationDTO{
			ID:   chat.ID,
			Name: chat.Name,
			Type: chat.Type,
		}

		if lastMsg, exists := messageMap[chat.ID]; exists {
			avatar_url, err := uc.fileStorage.GetOne(ctx, lastMsg.UserAvatarID)
			if err != nil {
				logger.Warningf("could not get avatar URL for user %s: %v", lastMsg.UserID, err)
				avatar_url = ""
			}

			chatDTO.LastMessage = dtoMessage.MessageDTO{
				SenderName:      lastMsg.UserName,
				Text:            lastMsg.Text,
				CreatedAt:       lastMsg.CreatedAt,
				SenderAvatarURL: avatar_url,
				ChatId:          lastMsg.ChatID,
			}
		}

		result = append(result, chatDTO)
	}

	return result, nil
}

func (uc *ChatsUsecase) GetInformationAboutChat(ctx context.Context, userId, chatId uuid.UUID) (*dtoChats.ChatDetailedInformationDTO, error) {
	const op = "ChatsUsecase.GetInformationAboutChat"

	logger := domains.GetLogger(ctx).WithField("operation", op)

	chat, err := uc.chatsRepo.GetChat(ctx, userId, chatId)
	if err != nil {
		return nil, err
	}

	messages, err := uc.chatsRepo.GetMessagesOfChat(ctx, chatId, 0, 20)
	if err != nil {
		return nil, err
	}

	users, err := uc.chatsRepo.GetUsersOfChat(ctx, chatId)
	if err != nil {
		return nil, err
	}

	userInfo, err := uc.chatsRepo.GetUserInfo(ctx, userId, chatId)
	if err != nil {
		return nil, err
	}

	messagesDTO := make([]dtoMessage.MessageDTO, len(messages))
	for i, message := range messages {
		avatar_url, err := uc.fileStorage.GetOne(ctx, message.UserAvatarID)
		if err != nil {
			logger.Warningf("could not get avatar URL for user %s: %v", message.UserID, err)
			avatar_url = ""
		}

		messagesDTO[i] = dtoMessage.MessageDTO{
			SenderName:      message.UserName,
			Text:            message.Text,
			CreatedAt:       message.CreatedAt,
			SenderAvatarURL: avatar_url,
			ChatId:          message.ChatID,
		}
	}

	usersDTO := make([]dtoChats.UserInfoChatDTO, len(users))
	for i, user := range users {
		avatar_url, err := uc.fileStorage.GetOne(ctx, user.UserAvatarID)
		if err != nil {
			logger.Warningf("could not get avatar URL for user %s: %v", user.UserID, err)
			avatar_url = ""
		}

		usersDTO[i] = dtoChats.UserInfoChatDTO{
			UserId:     user.UserID,
			UserName:   user.UserName,
			UserAvatar: avatar_url,
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

	result := &dtoChats.ChatDetailedInformationDTO{
		ID:        chat.ID,
		Name:      chat.Name,
		IsAdmin:   isAdmin,
		CanChat:   canChat,
		IsMember:  isMember,
		IsPrivate: isPrivate,
		Type:      chat.Type,
		Members:   usersDTO,
		Messages:  messagesDTO,
	}

	return result, nil
}

func (s *ChatsUsecase) CreateChat(ctx context.Context, chatDTO dtoChats.ChatCreateInformationDTO) (uuid.UUID, error) {
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

	usersNames, err := s.usersRepo.GetUsersNames(ctx, usersIds)
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

	err = s.chatsRepo.CreateChat(ctx, chat, usersInfo, usersNames)
	if err != nil {
		return uuid.Nil, err
	}
	return chat.ID, nil
}

func (s *ChatsUsecase) AddUsersToChat(ctx context.Context, chatID uuid.UUID, users []dtoChats.AddChatMemberDTO) error {
	usersInfo := make([]modelsChats.UserInfo, len(users))
	for i, user := range users {
		usersInfo[i] = modelsChats.UserInfo{
			UserID: user.UserId,
			ChatID: chatID,
			Role:   user.Role,
		}
	}

	err := s.chatsRepo.InsertUsersToChat(ctx, chatID, usersInfo)
	if err != nil {
		return err
	}

	return nil
}
