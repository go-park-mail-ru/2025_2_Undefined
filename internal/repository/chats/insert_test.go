package repository

import (
	"context"
	"testing"

	modelsChats "github.com/go-park-mail-ru/2025_2_Undefined/internal/models/chats"
	"github.com/google/uuid"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
)

func TestChatsRepository_CreateChat_Success(t *testing.T) {
	mock, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("failed to create pgxmock pool: %v", err)
	}
	defer mock.Close()

	repo := NewChatsRepository(mock)

	chatID := uuid.New()
	userID1 := uuid.New()
	userID2 := uuid.New()

	chat := modelsChats.Chat{
		ID:          chatID,
		Type:        "group",
		Name:        "Test Group",
		Description: "Test Description",
	}

	usersInfo := []modelsChats.UserInfo{
		{UserID: userID1, ChatID: chatID, Role: "admin"},
		{UserID: userID2, ChatID: chatID, Role: "member"},
	}

	usersNames := []string{"User1", "User2"}

	// Начинаем транзакцию
	mock.ExpectBegin()

	// Вставка чата
	mock.ExpectExec(`INSERT INTO chat (id, chat_type, name, description) 
        VALUES ($1, $2::chat_type_enum, $3, $4)`).
		WithArgs(chatID, "group", "Test Group", "Test Description").
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	// Вставка участников чата
	mock.ExpectExec(`INSERT INTO chat_member (user_id, chat_id, chat_member_role) VALUES ($1, $2, $3::chat_member_role_enum), ($4, $5, $6::chat_member_role_enum)`).
		WithArgs(userID1, chatID, "admin", userID2, chatID, "member").
		WillReturnResult(pgxmock.NewResult("INSERT", 2))

	// Вставка системных сообщений
	mock.ExpectExec(`INSERT INTO message (chat_id, user_id, text, message_type) VALUES ($1, $2, $3, $4::message_type_enum), ($5, $6, $7, $8::message_type_enum), ($9, $10, $11, $12::message_type_enum)`).
		WithArgs(chatID, nil, "Чат создан", "system", chatID, userID1, "Пользователь User1 вступил в чат", "system", chatID, userID2, "Пользователь User2 вступил в чат", "system").
		WillReturnResult(pgxmock.NewResult("INSERT", 3))

	// Коммит транзакции
	mock.ExpectCommit()

	ctx := context.Background()
	err = repo.CreateChat(ctx, chat, usersInfo, usersNames)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestChatsRepository_CreateChat_InvalidInput(t *testing.T) {
	mock, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("failed to create pgxmock pool: %v", err)
	}
	defer mock.Close()

	repo := NewChatsRepository(mock)

	chat := modelsChats.Chat{
		ID:   uuid.New(),
		Type: "group",
		Name: "Test Group",
	}

	// Разное количество пользователей и имен
	usersInfo := []modelsChats.UserInfo{{UserID: uuid.New(), Role: "admin"}}
	usersNames := []string{"User1", "User2"}

	ctx := context.Background()
	err = repo.CreateChat(ctx, chat, usersInfo, usersNames)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid input")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestChatsRepository_InsertUsersToChat_Success(t *testing.T) {
	mock, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("failed to create pgxmock pool: %v", err)
	}
	defer mock.Close()

	repo := NewChatsRepository(mock)

	chatID := uuid.New()
	userID1 := uuid.New()
	userID2 := uuid.New()

	usersInfo := []modelsChats.UserInfo{
		{UserID: userID1, ChatID: chatID, Role: "admin"},
		{UserID: userID2, ChatID: chatID, Role: "member"},
	}

	// Ожидаем начало транзакции
	mock.ExpectBegin()

	// Ожидаем вставку участников чата
	mock.ExpectExec(`INSERT INTO chat_member (user_id, chat_id, chat_member_role) VALUES ($1, $2, $3::chat_member_role_enum), ($4, $5, $6::chat_member_role_enum)`).
		WithArgs(userID1, chatID, "admin", userID2, chatID, "member").
		WillReturnResult(pgxmock.NewResult("INSERT", 2))

	// Ожидаем коммит транзакции
	mock.ExpectCommit()

	ctx := context.Background()
	err = repo.InsertUsersToChat(ctx, chatID, usersInfo)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
