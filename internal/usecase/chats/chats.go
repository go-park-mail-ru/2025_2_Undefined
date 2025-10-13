package usecase

import (
	"fmt"

	models "github.com/go-park-mail-ru/2025_2_Undefined/internal/models/chats"
	dto "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/dto/chats"
	"github.com/google/uuid"
)

type ChatsRepository interface {
	GetChats(userId uuid.UUID) ([]models.Chat, error)
	GetLastMessagesOfChats(userId uuid.UUID) ([]models.Message, error)
	GetChat(userId, chatId uuid.UUID) (*models.Chat, error)
	GetUsersOfChat(chatId uuid.UUID) ([]models.UserInfo, error)
	GetMessagesOfChat(chatId uuid.UUID, offset, limit int) ([]models.Message, error)
	CreateChat(chat models.Chat, usersInfo []models.UserInfo, usersNames []string) error
	GetUserInfo(userId, chatId uuid.UUID) (*models.UserInfo, error)
}

type UserRepository interface {
	GetUsersNames(usersIds []uuid.UUID) ([]string, error)
}

type ChatsService struct {
	chatsRepo ChatsRepository
	usersRepo UserRepository
}

func NewChatsService(chatsRepo ChatsRepository, usersRepo UserRepository) *ChatsService {
	return &ChatsService{
		chatsRepo: chatsRepo,
		usersRepo: usersRepo,
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

	// Создаем мапу для быстрого поиска последних сообщений по chat_id
	messageMap := make(map[uuid.UUID]models.Message)
	for _, msg := range lastMessages {
		messageMap[msg.ChatID] = msg
	}

	result := make([]dto.ChatViewInformationDTO, 0, len(chats))
	for _, chat := range chats {
		chatDTO := dto.ChatViewInformationDTO{
			ID:   chat.ID,
			Name: chat.Name,
		}

		if lastMsg, exists := messageMap[chat.ID]; exists {
			chatDTO.LastMessage = dto.MessageDTO{
				Sender:    lastMsg.UserID,
				Text:      lastMsg.Text,
				CreatedAt: lastMsg.CreatedAt,
			}
		}

		result = append(result, chatDTO)
	}

	return result, nil
}

func (s *ChatsService) GetInformationAboutChat(userId, chatId uuid.UUID) (*dto.ChatDetailedInformationDTO, error) {
	chat, err := s.chatsRepo.GetChat(userId, chatId)
	if err != nil {
		return nil, err
	}

	messages, err := s.chatsRepo.GetMessagesOfChat(chatId, 0, 20)
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

	if chat.Type == models.ChatTypeDialog {
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

	usersIds := make([]uuid.UUID, len(chatDTO.Members))
	for i, memberDTO := range chatDTO.Members {
		usersIds[i] = memberDTO.UserId
	}

	usersNames, err := s.usersRepo.GetUsersNames(usersIds)
	if err != nil {
		return uuid.Nil, fmt.Errorf("can't create chat: %w", err)
	}

	usersInfo := make([]models.UserInfo, len(chatDTO.Members))
	for i, memberDTO := range chatDTO.Members {
		usersInfo[i] = models.UserInfo{
			UserID: memberDTO.UserId,
			ChatID: chat.ID,
			Role:   memberDTO.Role,
		}
	}

	err = s.chatsRepo.CreateChat(chat, usersInfo, usersNames)
	if err != nil {
		return uuid.Nil, fmt.Errorf("can't create chat: %w", err)
	}
	return chat.ID, nil
}
