package inmemory

import (
	"sync"
	"testing"
	"time"

	models "github.com/go-park-mail-ru/2025_2_Undefined/internal/models/chats"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/models/errs"
	userModels "github.com/go-park-mail-ru/2025_2_Undefined/internal/models/user"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

type MockUserRepo struct {
	users map[uuid.UUID]*userModels.User
}

func NewMockUserRepo() *MockUserRepo {
	return &MockUserRepo{
		users: make(map[uuid.UUID]*userModels.User),
	}
}

func (m *MockUserRepo) GetByID(id uuid.UUID) (*userModels.User, error) {
	user, exists := m.users[id]
	if !exists {
		return nil, errs.ErrNotFound
	}
	return user, nil
}

func (m *MockUserRepo) GetByUsername(username string) (*userModels.User, error) {
	for _, user := range m.users {
		if user.Username == username {
			return user, nil
		}
	}
	return nil, errs.ErrNotFound
}

func (m *MockUserRepo) GetByEmail(email string) (*userModels.User, error) {
	for _, user := range m.users {
		if user.Email == email {
			return user, nil
		}
	}
	return nil, errs.ErrNotFound
}

func (m *MockUserRepo) GetByPhone(phone string) (*userModels.User, error) {
	for _, user := range m.users {
		if user.PhoneNumber == phone {
			return user, nil
		}
	}
	return nil, errs.ErrNotFound
}

func (m *MockUserRepo) Create(user *userModels.User) error {
	if user.ID == uuid.Nil {
		user.ID = uuid.New()
	}
	m.users[user.ID] = user
	return nil
}

func (m *MockUserRepo) Update(user *userModels.User) error {
	if _, exists := m.users[user.ID]; !exists {
		return errs.ErrNotFound
	}
	m.users[user.ID] = user
	return nil
}

func (m *MockUserRepo) Delete(id uuid.UUID) error {
	if _, exists := m.users[id]; !exists {
		return errs.ErrNotFound
	}
	delete(m.users, id)
	return nil
}

func TestGetChat_Success(t *testing.T) {
	repoUser := NewMockUserRepo()
	repo := NewChatsRepo(repoUser)
	chatID := uuid.New()
	userID := uuid.New()
	chat := models.Chat{ID: chatID, Name: "Test Chat", Type: models.ChatDialog}
	users := []models.UserInfo{{UserID: userID, ChatID: chatID, Role: models.RoleAdmin}}
	err := repo.CreateChat(chat, users)
	assert.NoError(t, err)
	foundChat, err := repo.GetChat(userID, chatID)
	assert.NoError(t, err)
	assert.Equal(t, chat, foundChat)
}

func TestGetChat_Error(t *testing.T) {
	repoUser := NewMockUserRepo()
	repo := NewChatsRepo(repoUser)
	userID := uuid.New()
	chatID := uuid.New()
	chat, err := repo.GetChat(userID, chatID)
	assert.Error(t, err)
	assert.Equal(t, errs.ErrNotFound, err)
	assert.Equal(t, models.Chat{}, chat)
}

func TestGetChats(t *testing.T) {
	repoUser := NewMockUserRepo()
	repo := NewChatsRepo(repoUser)
	userID := uuid.New()
	chatID := uuid.New()
	chat := models.Chat{ID: chatID, Name: "Chat", Type: models.ChatDialog}
	users := []models.UserInfo{{UserID: userID, ChatID: chatID, Role: models.RoleAdmin}}
	err := repo.CreateChat(chat, users)
	assert.NoError(t, err)
	chats, err := repo.GetChats(userID)
	assert.NoError(t, err)
	assert.Len(t, chats, 1)
}

func TestGetLastMessagesOfChats(t *testing.T) {
	repoUser := NewMockUserRepo()
	repo := NewChatsRepo(repoUser)
	userID := uuid.New()
	chatID := uuid.New()
	chat := models.Chat{ID: chatID, Name: "Chat", Type: models.ChatDialog}
	users := []models.UserInfo{{UserID: userID, ChatID: chatID, Role: models.RoleAdmin}}
	err := repo.CreateChat(chat, users)
	assert.NoError(t, err)

	message := models.Message{
		ID:        uuid.New(),
		ChatID:    chatID,
		UserID:    userID,
		Text:      "Test message",
		CreatedAt: time.Now(),
		Type:      models.UserMessage,
	}
	repo.mutex.Lock()
	repo.chatMessages[chatID] = append(repo.chatMessages[chatID], message)
	repo.mutex.Unlock()

	messages, err := repo.GetLastMessagesOfChats(userID)
	assert.NoError(t, err)
	assert.Len(t, messages, 1)
	assert.Equal(t, message, messages[0])
}

func TestGetUsersOfChat(t *testing.T) {
	repoUser := NewMockUserRepo()
	repo := NewChatsRepo(repoUser)
	user := models.UserInfo{
		UserID: uuid.New(),
		ChatID: uuid.New(),
		Role:   models.RoleAdmin,
	}
	chat := models.Chat{ID: user.ChatID, Name: "Chat", Type: models.ChatDialog}
	users := []models.UserInfo{user}
	_ = repo.CreateChat(chat, users)
	chatUsers, err := repo.GetUsersOfChat(user.ChatID)
	assert.NoError(t, err)
	assert.Len(t, chatUsers, 1)
	assert.Equal(t, user, chatUsers[0])
}

func TestGetMessagesOfChat(t *testing.T) {
	repoUser := NewMockUserRepo()
	repo := NewChatsRepo(repoUser)
	chatID := uuid.New()
	userID := uuid.New()
	chat := models.Chat{ID: chatID, Name: "Chat", Type: models.ChatDialog}
	users := []models.UserInfo{{UserID: userID, ChatID: chatID, Role: models.RoleAdmin}}
	err := repo.CreateChat(chat, users)
	assert.NoError(t, err)

	// Добавляем 5 пользовательских сообщений
	testMessages := make([]models.Message, 5)
	repo.mutex.Lock()
	for i := range testMessages {
		message := models.Message{
			ID:        uuid.New(),
			ChatID:    chatID,
			UserID:    userID,
			Text:      "Test message",
			CreatedAt: time.Now().AddDate(0, 0, i),
			Type:      models.UserMessage,
		}
		testMessages[i] = message
		repo.chatMessages[chatID] = append(repo.chatMessages[chatID], message)
	}
	repo.mutex.Unlock()

	// Получаем 3 последних сообщения
	messages, err := repo.GetMessagesOfChat(chatID, 3, 0)
	assert.NoError(t, err)
	assert.Len(t, messages, 3)

	// Ожидаем последние 3 сообщения в обратном порядке
	expected := make([]models.Message, 3)
	for i := range expected {
		expected[i] = testMessages[4-i]
	}
	assert.Equal(t, expected, messages)
}

func TestGetUserInfo_Success(t *testing.T) {
	repoUser := NewMockUserRepo()
	repo := NewChatsRepo(repoUser)
	chatID := uuid.New()
	userID := uuid.New()
	chat := models.Chat{ID: chatID, Name: "Chat", Type: models.ChatDialog}
	users := []models.UserInfo{{UserID: userID, ChatID: chatID, Role: models.RoleAdmin}}
	err := repo.CreateChat(chat, users)
	assert.NoError(t, err)
	userInfo, err := repo.GetUserInfo(userID, chatID)
	assert.NoError(t, err)
	assert.Equal(t, userID, userInfo.UserID)
	assert.Equal(t, chatID, userInfo.ChatID)
	assert.Equal(t, models.RoleAdmin, userInfo.Role)
}

func TestGetUserInfo_Error(t *testing.T) {
	repoUser := NewMockUserRepo()
	repo := NewChatsRepo(repoUser)
	userID := uuid.New()
	chatID := uuid.New()
	userInfo, err := repo.GetUserInfo(userID, chatID)
	assert.Error(t, err)
	assert.Equal(t, errs.ErrNotFound, err)
	assert.Equal(t, models.UserInfo{}, userInfo)
}

func TestConcurrentCreateChat(t *testing.T) {
	repoUser := NewMockUserRepo()
	repo := NewChatsRepo(repoUser)
	const numGoroutines = 10
	var wg sync.WaitGroup
	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			chatID := uuid.New()
			userID := uuid.New()
			chat := models.Chat{ID: chatID, Name: "Concurrent Chat", Type: models.ChatDialog}
			users := []models.UserInfo{{UserID: userID, ChatID: chatID, Role: models.RoleAdmin}}
			err := repo.CreateChat(chat, users)
			assert.NoError(t, err)
		}()
	}
	wg.Wait()
	assert.Equal(t, numGoroutines, len(repo.chats))
}
