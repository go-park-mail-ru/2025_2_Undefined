package repository

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	modelsChats "github.com/go-park-mail-ru/2025_2_Undefined/internal/models/chats"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestChatsRepository_GetChats_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock database: %v", err)
	}
	defer db.Close()

	repo := NewChatsRepository(db)

	userID := uuid.New()
	chatID1 := uuid.New()
	chatID2 := uuid.New()

	rows := sqlmock.NewRows([]string{"id", "chat_type", "name", "description"}).
		AddRow(chatID1.String(), "dialog", "Chat 1", "Description 1").
		AddRow(chatID2.String(), "group", "Chat 2", "Description 2")

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT c.id, c.chat_type::text, c.name, c.description 
		FROM chat c
		JOIN chat_member cm ON cm.chat_id = c.id
		WHERE cm.user_id = $1`)).
		WithArgs(userID).
		WillReturnRows(rows)

	ctx := context.Background()
	chats, err := repo.GetChats(ctx, userID)

	assert.NoError(t, err)
	assert.Len(t, chats, 2)
	assert.Equal(t, chatID1, chats[0].ID)
	assert.Equal(t, "dialog", chats[0].Type)
	assert.Equal(t, "Chat 1", chats[0].Name)
	assert.Equal(t, "Description 1", chats[0].Description)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestChatsRepository_GetChats_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock database: %v", err)
	}
	defer db.Close()

	repo := NewChatsRepository(db)
	userID := uuid.New()

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT c.id, c.chat_type::text, c.name, c.description 
		FROM chat c
		JOIN chat_member cm ON cm.chat_id = c.id
		WHERE cm.user_id = $1`)).
		WithArgs(userID).
		WillReturnError(fmt.Errorf("database error"))

	ctx := context.Background()
	chats, err := repo.GetChats(ctx, userID)

	assert.Error(t, err)
	assert.Nil(t, chats)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestChatsRepository_GetLastMessagesOfChats_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock database: %v", err)
	}
	defer db.Close()

	repo := NewChatsRepository(db)

	userID := uuid.New()
	messageID := uuid.New()
	chatID := uuid.New()
	senderID := uuid.New()
	createdAt := time.Now()

	rows := sqlmock.NewRows([]string{"id", "chat_id", "user_id", "name", "attachment_id", "text", "created_at", "message_type"}).
		AddRow(messageID.String(), chatID.String(), senderID.String(), "John Doe", nil, "Hello world", createdAt, "text")

	mock.ExpectQuery(regexp.QuoteMeta(`
		WITH latest_avatars AS (
			SELECT DISTINCT ON (user_id) user_id, attachment_id
			FROM avatar_user 
			ORDER BY user_id, created_at DESC
		)
		SELECT DISTINCT ON (msg.chat_id)
			msg.id, msg.chat_id, msg.user_id, usr.name, 
			la.attachment_id,
			msg.text, msg.created_at, msg.message_type::text
		FROM message msg
		JOIN chat_member cm ON cm.chat_id = msg.chat_id
		JOIN "user" usr ON usr.id = msg.user_id
		LEFT JOIN latest_avatars la ON la.user_id = msg.user_id
		WHERE cm.user_id = $1
		ORDER BY msg.chat_id, msg.created_at DESC`)).
		WithArgs(userID).
		WillReturnRows(rows)

	ctx := context.Background()
	messages, err := repo.GetLastMessagesOfChats(ctx, userID)

	assert.NoError(t, err)
	assert.Len(t, messages, 1)
	assert.Equal(t, messageID, messages[0].ID)
	assert.Equal(t, chatID, messages[0].ChatID)
	assert.Equal(t, senderID, messages[0].UserID)
	assert.Equal(t, "John Doe", messages[0].UserName)
	assert.Equal(t, "Hello world", messages[0].Text)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestChatsRepository_GetChat_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock database: %v", err)
	}
	defer db.Close()

	repo := NewChatsRepository(db)

	chatID := uuid.New()

	rows := sqlmock.NewRows([]string{"id", "chat_type", "name", "description"}).
		AddRow(chatID.String(), "group", "Test Chat", "Test Description")

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT c.id, c.chat_type::text, c.name, c.description 
		FROM chat c
		WHERE c.id = $1`)).
		WithArgs(chatID).
		WillReturnRows(rows)

	ctx := context.Background()
	chat, err := repo.GetChat(ctx, chatID)

	assert.NoError(t, err)
	assert.NotNil(t, chat)
	assert.Equal(t, chatID, chat.ID)
	assert.Equal(t, "group", chat.Type)
	assert.Equal(t, "Test Chat", chat.Name)
	assert.Equal(t, "Test Description", chat.Description)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestChatsRepository_GetChat_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock database: %v", err)
	}
	defer db.Close()

	repo := NewChatsRepository(db)

	chatID := uuid.New()

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT c.id, c.chat_type::text, c.name, c.description 
		FROM chat c
		WHERE c.id = $1`)).
		WithArgs(chatID).
		WillReturnError(sql.ErrNoRows)

	ctx := context.Background()
	chat, err := repo.GetChat(ctx, chatID)

	assert.Error(t, err)
	assert.Nil(t, chat)
	assert.Equal(t, sql.ErrNoRows, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestChatsRepository_GetUsersOfChat_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock database: %v", err)
	}
	defer db.Close()

	repo := NewChatsRepository(db)

	chatID := uuid.New()
	userID1 := uuid.New()
	userID2 := uuid.New()

	rows := sqlmock.NewRows([]string{"user_id", "chat_id", "name", "attachment_id", "chat_member_role"}).
		AddRow(userID1.String(), chatID.String(), "User 1", nil, "admin").
		AddRow(userID2.String(), chatID.String(), "User 2", nil, "member")

	mock.ExpectQuery(regexp.QuoteMeta(`
		WITH latest_avatars AS (
			SELECT DISTINCT ON (user_id) user_id, attachment_id
			FROM avatar_user 
			ORDER BY user_id, created_at DESC
		)
		SELECT 
			cm.user_id, cm.chat_id, usr.name, 
			la.attachment_id,
			cm.chat_member_role::text
		FROM chat_member cm
		JOIN "user" usr ON usr.id = cm.user_id
		LEFT JOIN latest_avatars la ON la.user_id = cm.user_id
		WHERE cm.chat_id = $1`)).
		WithArgs(chatID).
		WillReturnRows(rows)

	ctx := context.Background()
	users, err := repo.GetUsersOfChat(ctx, chatID)

	assert.NoError(t, err)
	assert.Len(t, users, 2)
	assert.Equal(t, userID1, users[0].UserID)
	assert.Equal(t, "User 1", users[0].UserName)
	assert.Equal(t, "admin", users[0].Role)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestChatsRepository_GetMessagesOfChat_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock database: %v", err)
	}
	defer db.Close()

	repo := NewChatsRepository(db)

	chatID := uuid.New()
	messageID := uuid.New()
	userID := uuid.New()
	createdAt := time.Now()
	offset := 0
	limit := 10

	rows := sqlmock.NewRows([]string{"id", "chat_id", "user_id", "name", "attachment_id", "text", "created_at", "message_type"}).
		AddRow(messageID.String(), chatID.String(), userID.String(), "John", nil, "Test message", createdAt, "text")

	mock.ExpectQuery(regexp.QuoteMeta(`
		WITH latest_avatars AS (
			SELECT DISTINCT ON (user_id) user_id, attachment_id
			FROM avatar_user 
			ORDER BY user_id, created_at DESC
		)
		SELECT 
			msg.id, msg.chat_id, msg.user_id, usr.name, 
			la.attachment_id,
			msg.text, msg.created_at, msg.message_type::text
		FROM message msg
		JOIN "user" usr ON usr.id = msg.user_id
		LEFT JOIN latest_avatars la ON la.user_id = msg.user_id
		WHERE chat_id = $1
		ORDER BY msg.created_at DESC
		LIMIT $3 OFFSET $2`)).
		WithArgs(chatID, offset, limit).
		WillReturnRows(rows)

	ctx := context.Background()
	messages, err := repo.GetMessagesOfChat(ctx, chatID, offset, limit)

	assert.NoError(t, err)
	assert.Len(t, messages, 1)
	assert.Equal(t, messageID, messages[0].ID)
	assert.Equal(t, chatID, messages[0].ChatID)
	assert.Equal(t, "Test message", messages[0].Text)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestChatsRepository_CreateChat_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock database: %v", err)
	}
	defer db.Close()

	repo := NewChatsRepository(db)

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
	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO chat (id, chat_type, name, description) 
        VALUES ($1, $2::chat_type_enum, $3, $4)`)).
		WithArgs(chatID, "group", "Test Group", "Test Description").
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Вставка участников чата
	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO chat_member (user_id, chat_id, chat_member_role) VALUES ($1, $2, $3::chat_member_role_enum), ($4, $5, $6::chat_member_role_enum)`)).
		WithArgs(userID1, chatID, "admin", userID2, chatID, "member").
		WillReturnResult(sqlmock.NewResult(2, 2))

	// Вставка системных сообщений
	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO message (chat_id, user_id, text, message_type) VALUES ($1, $2, $3, $4::message_type_enum), ($5, $6, $7, $8::message_type_enum)`)).
		WithArgs(chatID, userID1, "Пользователь User1 вступил в чат", "system", chatID, userID2, "Пользователь User2 вступил в чат", "system").
		WillReturnResult(sqlmock.NewResult(2, 2))

	// Коммит транзакции
	mock.ExpectCommit()

	ctx := context.Background()
	err = repo.CreateChat(ctx, chat, usersInfo, usersNames)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestChatsRepository_CreateChat_InvalidInput(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock database: %v", err)
	}
	defer db.Close()

	repo := NewChatsRepository(db)

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

func TestChatsRepository_GetUserInfo_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock database: %v", err)
	}
	defer db.Close()

	repo := NewChatsRepository(db)

	userID := uuid.New()
	chatID := uuid.New()

	rows := sqlmock.NewRows([]string{"user_id", "chat_id", "name", "attachment_id", "chat_member_role"}).
		AddRow(userID.String(), chatID.String(), "John Doe", nil, "admin")

	mock.ExpectQuery(regexp.QuoteMeta(`
		WITH latest_avatars AS (
			SELECT DISTINCT ON (user_id) user_id, attachment_id
			FROM avatar_user 
			ORDER BY user_id, created_at DESC
		)
		SELECT 
			cm.user_id, cm.chat_id, usr.name, 
			la.attachment_id,
			cm.chat_member_role::text
		FROM chat_member cm
		JOIN "user" usr ON usr.id = cm.user_id
		LEFT JOIN latest_avatars la ON la.user_id = cm.user_id
		WHERE cm.user_id = $1 AND cm.chat_id = $2`)).
		WithArgs(userID, chatID).
		WillReturnRows(rows)

	ctx := context.Background()
	userInfo, err := repo.GetUserInfo(ctx, userID, chatID)

	assert.NoError(t, err)
	assert.NotNil(t, userInfo)
	assert.Equal(t, userID, userInfo.UserID)
	assert.Equal(t, chatID, userInfo.ChatID)
	assert.Equal(t, "John Doe", userInfo.UserName)
	assert.Equal(t, "admin", userInfo.Role)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestNewChatsRepository(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock database: %v", err)
	}
	defer db.Close()

	repo := NewChatsRepository(db)

	assert.NotNil(t, repo)
	assert.Equal(t, db, repo.db)
}

func TestChatsRepository_InsertUsersToChat_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock database: %v", err)
	}
	defer db.Close()

	repo := NewChatsRepository(db)

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
	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO chat_member (user_id, chat_id, chat_member_role) VALUES ($1, $2, $3::chat_member_role_enum), ($4, $5, $6::chat_member_role_enum)`)).
		WithArgs(userID1, chatID, "admin", userID2, chatID, "member").
		WillReturnResult(sqlmock.NewResult(2, 2))

	// Ожидаем коммит транзакции
	mock.ExpectCommit()

	ctx := context.Background()
	err = repo.InsertUsersToChat(ctx, chatID, usersInfo)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestChatsRepository_CheckUserHasRole_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock database: %v", err)
	}
	defer db.Close()

	repo := NewChatsRepository(db)
	userID := uuid.New()
	chatID := uuid.New()

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT EXISTS(
			SELECT 1 FROM chat_member 
			WHERE user_id = $1 AND chat_id = $2 AND chat_member_role = $3::chat_member_role_enum
		)`)).
		WithArgs(userID, chatID, "admin").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	ctx := context.Background()
	hasRole, err := repo.CheckUserHasRole(ctx, userID, chatID, "admin")

	assert.NoError(t, err)
	assert.True(t, hasRole)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestChatsRepository_DeleteChat_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock database: %v", err)
	}
	defer db.Close()

	repo := NewChatsRepository(db)
	userID := uuid.New()
	chatID := uuid.New()

	// Проверка роли админа
	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT EXISTS(
			SELECT 1 FROM chat_member 
			WHERE user_id = $1 AND chat_id = $2 AND chat_member_role = $3::chat_member_role_enum
		)`)).
		WithArgs(userID, chatID, "admin").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	// Удаление чата
	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM chat WHERE id = $1`)).
		WithArgs(chatID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	ctx := context.Background()
	err = repo.DeleteChat(ctx, userID, chatID)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestChatsRepository_DeleteChat_NotAdmin(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock database: %v", err)
	}
	defer db.Close()

	repo := NewChatsRepository(db)
	userID := uuid.New()
	chatID := uuid.New()

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT EXISTS(
			SELECT 1 FROM chat_member 
			WHERE user_id = $1 AND chat_id = $2 AND chat_member_role = $3::chat_member_role_enum
		)`)).
		WithArgs(userID, chatID, "admin").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	ctx := context.Background()
	err = repo.DeleteChat(ctx, userID, chatID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user is not admin")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestChatsRepository_UpdateChat_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock database: %v", err)
	}
	defer db.Close()

	repo := NewChatsRepository(db)
	userID := uuid.New()
	chatID := uuid.New()

	// Проверка роли админа
	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT EXISTS(
			SELECT 1 FROM chat_member 
			WHERE user_id = $1 AND chat_id = $2 AND chat_member_role = $3::chat_member_role_enum
		)`)).
		WithArgs(userID, chatID, "admin").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	// Обновление чата
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE chat SET name = $1, description = $2 WHERE id = $3`)).
		WithArgs("New Name", "New Description", chatID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	ctx := context.Background()
	err = repo.UpdateChat(ctx, userID, chatID, "New Name", "New Description")

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
