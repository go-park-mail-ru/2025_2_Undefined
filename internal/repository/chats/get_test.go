package repository

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
)

func TestChatsRepository_GetChats_Success(t *testing.T) {
	mock, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("failed to create pgxmock pool: %v", err)
	}
	defer mock.Close()

	repo := NewChatsRepository(mock)

	userID := uuid.New()
	chatID1 := uuid.New()
	chatID2 := uuid.New()

	rows := pgxmock.NewRows([]string{"id", "chat_type", "name", "description"}).
		AddRow(chatID1, "dialog", "Chat 1", "Description 1").
		AddRow(chatID2, "group", "Chat 2", "Description 2")

	mock.ExpectQuery(getChatsQuery).
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
	mock, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("failed to create pgxmock pool: %v", err)
	}
	defer mock.Close()

	repo := NewChatsRepository(mock)
	userID := uuid.New()

	mock.ExpectQuery(getChatsQuery).
		WithArgs(userID).
		WillReturnError(fmt.Errorf("database error"))

	ctx := context.Background()
	chats, err := repo.GetChats(ctx, userID)

	assert.Error(t, err)
	assert.Nil(t, chats)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestChatsRepository_GetChat_Success(t *testing.T) {
	mock, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("failed to create pgxmock pool: %v", err)
	}
	defer mock.Close()

	repo := NewChatsRepository(mock)

	chatID := uuid.New()

	rows := pgxmock.NewRows([]string{"id", "chat_type", "name", "description"}).
		AddRow(chatID, "group", "Test Chat", "Test Description")

	mock.ExpectQuery(getChatQuery).
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
	mock, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("failed to create pgxmock pool: %v", err)
	}
	defer mock.Close()

	repo := NewChatsRepository(mock)

	chatID := uuid.New()

	mock.ExpectQuery(getChatQuery).
		WithArgs(chatID).
		WillReturnError(pgx.ErrNoRows)

	ctx := context.Background()
	chat, err := repo.GetChat(ctx, chatID)

	assert.Error(t, err)
	assert.Nil(t, chat)
	assert.Equal(t, pgx.ErrNoRows, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestChatsRepository_GetUsersOfChat_Success(t *testing.T) {
	mock, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("failed to create pgxmock pool: %v", err)
	}
	defer mock.Close()

	repo := NewChatsRepository(mock)

	chatID := uuid.New()
	userID1 := uuid.New()
	userID2 := uuid.New()

	rows := pgxmock.NewRows([]string{"user_id", "chat_id", "name", "chat_member_role"}).
		AddRow(userID1, chatID, "User 1", "admin").
		AddRow(userID2, chatID, "User 2", "member")

	mock.ExpectQuery(getUsersOfChat).
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
	mock, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("failed to create pgxmock pool: %v", err)
	}
	defer mock.Close()

	repo := NewChatsRepository(mock)

	userID := uuid.New()
	chatID := uuid.New()

	rows := pgxmock.NewRows([]string{"user_id", "chat_id", "name", "chat_member_role"}).
		AddRow(userID, chatID, "John Doe", "admin")

	mock.ExpectQuery(getUserInfo).
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
	mock, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("failed to create pgxmock pool: %v", err)
	}
	defer mock.Close()

	repo := NewChatsRepository(mock)

	user1ID := uuid.New()
	user2ID := uuid.New()
	expectedChatID := uuid.New()

	mock.ExpectQuery(getUsersDialogQuery).
		WithArgs(user1ID, user2ID).
		WillReturnRows(pgxmock.NewRows([]string{"id"}).AddRow(expectedChatID))

	ctx := context.Background()
	chatID, err := repo.GetUsersDialog(ctx, user1ID, user2ID)

	assert.NoError(t, err)
	assert.Equal(t, expectedChatID, chatID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestChatsRepository_GetChatAvatars_Success(t *testing.T) {
	mock, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("failed to create pgxmock pool: %v", err)
	}
	defer mock.Close()

	repo := NewChatsRepository(mock)

	chatID1 := uuid.New()
	chatID2 := uuid.New()
	avatarID1 := uuid.New()
	avatarID2 := uuid.New()
	chatIDs := []uuid.UUID{chatID1, chatID2}

	rows := pgxmock.NewRows([]string{"chat_id", "attachment_id"}).
		AddRow(chatID1, avatarID1).
		AddRow(chatID2, avatarID2)

	mock.ExpectQuery(getChatAvatarsQuery).
		WithArgs(chatIDs).
		WillReturnRows(rows)

	ctx := context.Background()
	avatars, err := repo.GetChatAvatars(ctx, chatIDs)

	assert.NoError(t, err)
	assert.Len(t, avatars, 2)
	assert.Equal(t, avatarID1, avatars[chatID1.String()])
	assert.Equal(t, avatarID2, avatars[chatID2.String()])
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestChatsRepository_GetChatAvatars_EmptyInput(t *testing.T) {
	mock, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("failed to create pgxmock pool: %v", err)
	}
	defer mock.Close()

	repo := NewChatsRepository(mock)

	ctx := context.Background()
	avatars, err := repo.GetChatAvatars(ctx, []uuid.UUID{})

	assert.NoError(t, err)
	assert.Empty(t, avatars)
	assert.NoError(t, mock.ExpectationsWereMet())
}
