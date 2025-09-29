package inmemory

import (
	"fmt"
	"sync"
	"time"

	models "github.com/go-park-mail-ru/2025_2_Undefined/internal/models/chats"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/models/errs"
	"github.com/google/uuid"
)

type ChatsRepo struct {
	chats        map[uuid.UUID]models.Chat       // информация по чату по id чата
	usersInfo    map[uuid.UUID][]models.UserInfo // информация по чатам пользователя по его id
	chatMessages map[uuid.UUID][]models.Message  // сообщения заданного чата

	mutex sync.RWMutex
}

func NewChatsRepo() *ChatsRepo {
	return &ChatsRepo{
		chats:        make(map[uuid.UUID]models.Chat),
		usersInfo:    make(map[uuid.UUID][]models.UserInfo),
		chatMessages: make(map[uuid.UUID][]models.Message),
	}
}

func (r *ChatsRepo) GetChats(userId uuid.UUID) ([]models.Chat, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	userInfo, ok := r.usersInfo[userId]
	if !ok {
		return make([]models.Chat, 0), nil
	}

	res := make([]models.Chat, 0, len(userInfo))
	for i := range userInfo {
		res = append(res, r.chats[userInfo[i].ChatID])
	}
	return res, nil
}

func (r *ChatsRepo) GetLastMessagesOfChats(userId uuid.UUID) ([]models.Message, error) {
	chats, err := r.GetChats(userId)
	if err != nil {
		return nil, err
	}
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	last_messages := make([]models.Message, 0, len(chats))
	for _, chat := range chats {
		chatMessages := r.chatMessages[chat.ID]
		if len(chatMessages) == 0 {
			last_messages = append(last_messages, models.Message{})
			continue
		}
		last_messages = append(last_messages, chatMessages[len(chatMessages)-1])
	}
	return last_messages, nil
}

func (r *ChatsRepo) GetChat(userId, chatId uuid.UUID) (models.Chat, error) {
	userChats, err := r.GetChats(userId)
	if err != nil {
		return models.Chat{}, err
	}

	for _, chat := range userChats {
		if chat.ID == chatId {
			return chat, nil
		}
	}
	return models.Chat{}, errs.ErrNotFound
}

func (r *ChatsRepo) GetUsersOfChat(chatId uuid.UUID) ([]models.UserInfo, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	result := make([]models.UserInfo, 0)
	for _, userInfo := range r.usersInfo {
		for _, info := range userInfo {
			if info.ChatID == chatId {
				result = append(result, info)
			}
		}
	}
	return result, nil
}

func (r *ChatsRepo) GetMessagesOfChat(chatId uuid.UUID, limit, offset int) ([]models.Message, error) {
	r.mutex.RLock()
	messages := r.chatMessages[chatId]
	r.mutex.RUnlock()
	result := make([]models.Message, 0, limit)
	total := len(messages)
	for i := offset; i < offset+limit && (total-1-i) >= 0; i++ {
		result = append(result, messages[total-1-i])
	}
	return result, nil
}

func (r *ChatsRepo) CreateChat(chat models.Chat, usersInfo []models.UserInfo) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.chats[chat.ID] = chat

	for _, userInfo := range usersInfo {
		r.usersInfo[userInfo.UserID] = append(r.usersInfo[userInfo.UserID], userInfo)
		// Имя пользователя берется из бд
		// ! Пофиксить это, когда будет единая структура Repository !
		id := uuid.New()
		var introductionMessage = models.Message{
			ID:        id,
			ChatID:    userInfo.ChatID,
			UserID:    userInfo.UserID,
			Text:      fmt.Sprintf("Пользователь %s вступил в чат", userInfo.UserID.String()), // fix
			CreatedAt: time.Now(),
			Type:      models.SystemMessage,
		}
		r.chatMessages[userInfo.ChatID] = append(r.chatMessages[userInfo.ChatID], introductionMessage)
	}

	return nil
}

func (r *ChatsRepo) GetUserInfo(userId, chatId uuid.UUID) (models.UserInfo, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	userInfos, ok := r.usersInfo[userId]
	if !ok {
		return models.UserInfo{}, errs.ErrNotFound
	}
	for _, userInfo := range userInfos {
		if userInfo.ChatID == chatId {
			return userInfo, nil
		}
	}
	return models.UserInfo{}, errs.ErrNotFound
}
