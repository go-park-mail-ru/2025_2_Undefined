package usecase

import (
	"context"
	"fmt"

	modelsChats "github.com/go-park-mail-ru/2025_2_Undefined/internal/models/chats"
	modelsMessage "github.com/go-park-mail-ru/2025_2_Undefined/internal/models/message"
	dtoChats "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/dto/chats"
	dtoMessage "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/dto/message"
	interfaceChatsUsecase "github.com/go-park-mail-ru/2025_2_Undefined/internal/usecase/interface/chats"
	interfaceUserUsecase "github.com/go-park-mail-ru/2025_2_Undefined/internal/usecase/interface/user"
	"github.com/google/uuid"
)

type ChatsService struct {
	chatsRepo interfaceChatsUsecase.ChatsRepository
	usersRepo interfaceUserUsecase.UserRepository
}

func NewChatsService(chatsRepo interfaceChatsUsecase.ChatsRepository, usersRepo interfaceUserUsecase.UserRepository) *ChatsService {
	return &ChatsService{
		chatsRepo: chatsRepo,
		usersRepo: usersRepo,
	}
}

func (s *ChatsService) GetChats(ctx context.Context, userId uuid.UUID) ([]dtoChats.ChatViewInformationDTO, error) {
	chats, err := s.chatsRepo.GetChats(ctx, userId)
	if err != nil {
		return nil, err
	}

	lastMessages, err := s.chatsRepo.GetLastMessagesOfChats(ctx, userId)
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
			chatDTO.LastMessage = dtoMessage.MessageDTO{
				SenderName:   lastMsg.UserName,
				Text:         lastMsg.Text,
				CreatedAt:    lastMsg.CreatedAt,
				SenderAvatar: lastMsg.UserAvatar,
				ChatId:       lastMsg.ChatID,
			}
		}

		result = append(result, chatDTO)
	}

	return result, nil
}

func (s *ChatsService) GetInformationAboutChat(ctx context.Context, userId, chatId uuid.UUID) (*dtoChats.ChatDetailedInformationDTO, error) {
	chat, err := s.chatsRepo.GetChat(ctx, userId, chatId)
	if err != nil {
		return nil, err
	}

	messages, err := s.chatsRepo.GetMessagesOfChat(ctx, chatId, 0, 20)
	if err != nil {
		return nil, err
	}

	users, err := s.chatsRepo.GetUsersOfChat(ctx, chatId)
	if err != nil {
		return nil, err
	}

	userInfo, err := s.chatsRepo.GetUserInfo(ctx, userId, chatId)
	if err != nil {
		return nil, err
	}

	messagesDTO := make([]dtoMessage.MessageDTO, len(messages))
	for i, message := range messages {
		messagesDTO[i] = dtoMessage.MessageDTO{
			SenderName:   message.UserName,
			Text:         message.Text,
			CreatedAt:    message.CreatedAt,
			SenderAvatar: message.UserAvatar,
			ChatId:       message.ChatID,
		}
	}

	usersDTO := make([]dtoChats.UserInfoChatDTO, len(users))
	for i, user := range users {
		usersDTO[i] = dtoChats.UserInfoChatDTO{
			UserId:     user.UserID,
			UserName:   user.UserName,
			UserAvatar: user.UserAvatar,
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

func (s *ChatsService) CreateChat(ctx context.Context, chatDTO dtoChats.ChatCreateInformationDTO) (uuid.UUID, error) {
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
		return uuid.Nil, fmt.Errorf("can't create chat: %w", err)
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
		return uuid.Nil, fmt.Errorf("can't create chat: %w", err)
	}
	return chat.ID, nil
}
