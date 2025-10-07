package usecase

import (
	"fmt"

	"github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/dto"
	models "github.com/go-park-mail-ru/2025_2_Undefined/internal/models/chats"
	repositoryInterface "github.com/go-park-mail-ru/2025_2_Undefined/internal/repository/chats"
	"github.com/google/uuid"
)

type ChatsService struct {
	chatsRepo repositoryInterface.ChatsRepository
}

func NewChatsService(chatsRepo repositoryInterface.ChatsRepository) *ChatsService {
	return &ChatsService{
		chatsRepo: chatsRepo,
	}
}

func (s *ChatsService) GetChats(userId uuid.UUID) ([]dto.ChatViewInformationDTO, error) {
	chats, err := s.chatsRepo.GetChats(userId)
	if err != nil {
		return nil, err
	}
	lastMessages, err := s.chatsRepo.GetLastMessagesOfChats(userId)
	if err != nil {
		return nil, err
	}

	result := make([]dto.ChatViewInformationDTO, 0, len(chats))
	for i := range chats {
		message := dto.MessageDTO{
			Sender:    lastMessages[i].UserID,
			Text:      lastMessages[i].Text,
			CreatedAt: lastMessages[i].CreatedAt,
		}
		result = append(result, dto.ChatViewInformationDTO{
			ID:          chats[i].ID,
			Name:        chats[i].Name,
			LastMessage: message,
		})
	}

	return result, nil
}

func (s *ChatsService) GetInformationAboutChat(userId, chatId uuid.UUID) (*dto.ChatDetailedInformationDTO, error) {
	chat, err := s.chatsRepo.GetChat(userId, chatId)
	if err != nil {
		return nil, err
	}
	messages, err := s.chatsRepo.GetMessagesOfChat(chatId, 20, 0)
	if err != nil {
		return nil, err
	}
	users, err := s.chatsRepo.GetUsersOfChat(chatId)
	if err != nil {
		return nil, err
	}
	userInfo, err := s.chatsRepo.GetUserInfo(userId, chatId)
	if err != nil {
		return nil, err
	}

	messagesDTO := make([]dto.MessageDTO, len(messages))
	for i, message := range messages {
		messagesDTO[i] = dto.MessageDTO{
			Sender:    message.UserID,
			Text:      message.Text,
			CreatedAt: message.CreatedAt,
		}
	}
	usersDTO := make([]dto.UserInfoChatDTO, len(users))
	for i, user := range users {
		usersDTO[i] = dto.UserInfoChatDTO{
			UserId: user.UserID,
			Role:   user.Role,
		}
	}

	var isAdmin, canChat, isMember, isPrivate bool = false, false, false, false
	switch userInfo.Role {
	case models.RoleAdmin:
		isAdmin = true
		fallthrough
	case models.RoleMember:
		isMember = true
		canChat = true
	}

	if chat.Type == models.ChatDialog {
		isPrivate = true
	}

	result := &dto.ChatDetailedInformationDTO{
		ID:        chat.ID,
		Name:      chat.Name,
		IsAdmin:   isAdmin,
		CanChat:   canChat,
		IsMember:  isMember,
		IsPrivate: isPrivate,
		Members:   usersDTO,
		Messages:  messagesDTO,
	}

	return result, nil
}

func (s *ChatsService) CreateChat(chatDTO dto.ChatCreateInformationDTO) (uuid.UUID, error) {
	chat := models.Chat{
		ID:          uuid.New(),
		Name:        chatDTO.Name,
		Type:        chatDTO.Type,
		Description: "",
	}
	users := make([]models.UserInfo, len(chatDTO.Members))
	for i, memberDTO := range chatDTO.Members {
		users[i] = models.UserInfo{
			UserID: memberDTO.UserId,
			ChatID: chat.ID,
			Role:   memberDTO.Role,
		}
	}

	err := s.chatsRepo.CreateChat(chat, users)
	if err != nil {
		return uuid.Nil, fmt.Errorf("не удалось создать чат: %w", err)
	}
	return chat.ID, nil
}
