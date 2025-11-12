package repository

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
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

	mock.ExpectQuery(regexp.QuoteMeta(getChatsQuery)).
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

	mock.ExpectQuery(regexp.QuoteMeta(getChatsQuery)).
		WithArgs(userID).
		WillReturnError(fmt.Errorf("database error"))

	ctx := context.Background()
	chats, err := repo.GetChats(ctx, userID)

	assert.Error(t, err)
	assert.Nil(t, chats)
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

	mock.ExpectQuery(regexp.QuoteMeta(getChatQuery)).
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

	mock.ExpectQuery(regexp.QuoteMeta(getChatQuery)).
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

	mock.ExpectQuery(regexp.QuoteMeta(getUsersOfChat)).
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

	mock.ExpectQuery(regexp.QuoteMeta(getUserInfo)).
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

func TestChatsRepository_GetUsersDialog_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock database: %v", err)
	}
	defer db.Close()

	repo := NewChatsRepository(db)

	user1ID := uuid.New()
	user2ID := uuid.New()
	expectedChatID := uuid.New()

	mock.ExpectQuery(regexp.QuoteMeta(getUsersDialogQuery)).
		WithArgs(user1ID, user2ID).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(expectedChatID.String()))

	ctx := context.Background()
	chatID, err := repo.GetUsersDialog(ctx, user1ID, user2ID)

	assert.NoError(t, err)
	assert.Equal(t, expectedChatID, chatID)
	assert.NoError(t, mock.ExpectationsWereMet())
}
