package messages

import (
	"context"
	"fmt"
	"testing"
	"time"

	modelsMessage "github.com/go-park-mail-ru/2025_2_Undefined/internal/models/message"
	"github.com/google/uuid"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
)

func TestMessageRepository_InsertMessage_Success(t *testing.T) {
	mock, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("failed to open sqlmock database: %v", err)
	}
	defer mock.Close()

	repo := NewMessageRepository(mock)

	userID := uuid.New()

	msg := modelsMessage.CreateMessage{
		ChatID:    uuid.New(),
		UserID:    &userID,
		Text:      "hello",
		CreatedAt: time.Now(),
		Type:      "text",
	}

	expectedID := uuid.New()

	rows := pgxmock.NewRows([]string{"id"}).AddRow(expectedID)

	mock.ExpectQuery(`INSERT INTO message (chat_id, user_id, text, created_at, message_type) VALUES
						($1, $2, $3, $4, $5::message_type_enum)
						RETURNING id`).
		WithArgs(msg.ChatID, msg.UserID, msg.Text, msg.CreatedAt, msg.Type).
		WillReturnRows(rows)

	ctx := context.Background()
	gotID, err := repo.InsertMessage(ctx, msg)
	assert.NoError(t, err)
	assert.Equal(t, expectedID, gotID)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestMessageRepository_InsertMessage_DBError(t *testing.T) {
	mock, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("failed to open sqlmock database: %v", err)
	}
	defer mock.Close()

	repo := NewMessageRepository(mock)

	userID := uuid.New()

	msg := modelsMessage.CreateMessage{
		ChatID:    uuid.New(),
		UserID:    &userID,
		Text:      "hello",
		CreatedAt: time.Now(),
		Type:      "text",
	}

	mock.ExpectQuery(`INSERT INTO message (chat_id, user_id, text, created_at, message_type) VALUES
						($1, $2, $3, $4, $5::message_type_enum)
						RETURNING id`).
		WithArgs(msg.ChatID, msg.UserID, msg.Text, msg.CreatedAt, msg.Type).
		WillReturnError(fmt.Errorf("db error"))

	ctx := context.Background()
	gotID, err := repo.InsertMessage(ctx, msg)
	assert.Error(t, err)
	assert.Equal(t, uuid.Nil, gotID)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestMessageRepository_GetLastMessagesOfChats_Success(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer mock.Close()

	repo := NewMessageRepository(mock)
	ctx := context.Background()
	userID := uuid.New()
	msgUserID1 := uuid.New()
	msgUserID2 := uuid.New()
	userName1 := "User1"
	userName2 := "User2"
	now := time.Now()

	rows := pgxmock.NewRows([]string{"id", "chat_id", "user_id", "name", "text", "created_at", "updated_at", "message_type", "attachment_id", "attachment_type", "file_name", "file_size", "content_disposition", "duration"}).
		AddRow(uuid.New(), uuid.New(), &msgUserID1, &userName1, "Hello", now, now, "text", nil, nil, nil, nil, nil, nil).
		AddRow(uuid.New(), uuid.New(), &msgUserID2, &userName2, "Hi", now, now, "text", nil, nil, nil, nil, nil, nil)

	mock.ExpectQuery(`SELECT DISTINCT ON \(msg.chat_id\)`).
		WithArgs(userID).
		WillReturnRows(rows)

	messages, err := repo.GetLastMessagesOfChats(ctx, userID)

	assert.NoError(t, err)
	assert.Len(t, messages, 2)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestMessageRepository_GetMessagesOfChat_Success(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer mock.Close()

	repo := NewMessageRepository(mock)
	ctx := context.Background()
	chatID := uuid.New()
	msgUserID1 := uuid.New()
	msgUserID2 := uuid.New()
	userName1 := "User1"
	userName2 := "User2"
	now := time.Now()

	rows := pgxmock.NewRows([]string{"id", "chat_id", "user_id", "name", "text", "created_at", "updated_at", "message_type", "attachment_id", "attachment_type", "file_name", "file_size", "content_disposition", "duration"}).
		AddRow(uuid.New(), chatID, &msgUserID1, &userName1, "Message 1", now, now, "text", nil, nil, nil, nil, nil, nil).
		AddRow(uuid.New(), chatID, &msgUserID2, &userName2, "Message 2", now, now, "text", nil, nil, nil, nil, nil, nil)

	mock.ExpectQuery(`SELECT\s+msg\.id`).
		WithArgs(chatID, 0, 10).
		WillReturnRows(rows)

	messages, err := repo.GetMessagesOfChat(ctx, chatID, 0, 10)

	assert.NoError(t, err)
	assert.Len(t, messages, 2)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestMessageRepository_GetMessageByID_Success(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer mock.Close()

	repo := NewMessageRepository(mock)
	ctx := context.Background()
	messageID := uuid.New()
	chatID := uuid.New()
	userID := uuid.New()
	now := time.Now()

	row := pgxmock.NewRows([]string{"id", "chat_id", "user_id", "text", "created_at", "updated_at", "message_type"}).
		AddRow(messageID, chatID, &userID, "Test message", now, now, "text")

	mock.ExpectQuery(`SELECT id, chat_id, user_id, text, created_at, updated_at, message_type::text`).
		WithArgs(messageID).
		WillReturnRows(row)

	message, err := repo.GetMessageByID(ctx, messageID)

	assert.NoError(t, err)
	assert.Equal(t, messageID, message.ID)
	assert.Equal(t, chatID, message.ChatID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestMessageRepository_UpdateMessage_Success(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer mock.Close()

	repo := NewMessageRepository(mock)
	ctx := context.Background()
	messageID := uuid.New()
	newText := "Updated text"

	mock.ExpectExec(`UPDATE message SET text = \$1, updated_at = NOW\(\) WHERE id = \$2`).
		WithArgs(newText, messageID).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))

	err = repo.UpdateMessage(ctx, messageID, newText)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestMessageRepository_DeleteMessage_Success(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer mock.Close()

	repo := NewMessageRepository(mock)
	ctx := context.Background()
	messageID := uuid.New()

	mock.ExpectExec(`DELETE FROM message WHERE id = \$1`).
		WithArgs(messageID).
		WillReturnResult(pgxmock.NewResult("DELETE", 1))

	err = repo.DeleteMessage(ctx, messageID)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestMessageRepository_GetLastMessagesOfChatsByIDs_Success(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer mock.Close()

	repo := NewMessageRepository(mock)
	ctx := context.Background()
	chatIDs := []uuid.UUID{uuid.New(), uuid.New()}
	msgUserID1 := uuid.New()
	msgUserID2 := uuid.New()
	userName1 := "User1"
	userName2 := "User2"
	now := time.Now()

	rows := pgxmock.NewRows([]string{"id", "chat_id", "user_id", "name", "text", "created_at", "updated_at", "message_type", "attachment_id", "attachment_type", "file_name", "file_size", "content_disposition", "duration"}).
		AddRow(uuid.New(), chatIDs[0], &msgUserID1, &userName1, "Last message 1", now, now, "text", nil, nil, nil, nil, nil, nil).
		AddRow(uuid.New(), chatIDs[1], &msgUserID2, &userName2, "Last message 2", now, now, "text", nil, nil, nil, nil, nil, nil)

	mock.ExpectQuery(`SELECT DISTINCT ON \(msg.chat_id\)`).
		WithArgs(chatIDs).
		WillReturnRows(rows)

	messages, err := repo.GetLastMessagesOfChatsByIDs(ctx, chatIDs)

	assert.NoError(t, err)
	assert.Len(t, messages, 2)
	assert.NotNil(t, messages[chatIDs[0]])
	assert.NotNil(t, messages[chatIDs[1]])
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestMessageRepository_SearchMessagesInChat_Success(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer mock.Close()

	repo := NewMessageRepository(mock)
	ctx := context.Background()
	userID := uuid.New()
	chatID := uuid.New()
	searchText := "hello"
	msgUserID := uuid.New()
	userName := "User1"
	now := time.Now()

	rows := pgxmock.NewRows([]string{"id", "chat_id", "user_id", "name", "text", "created_at", "updated_at", "message_type", "attachment_id", "attachment_type", "file_name", "file_size", "content_disposition", "duration"}).
		AddRow(uuid.New(), chatID, &msgUserID, &userName, "hello world", now, now, "text", nil, nil, nil, nil, nil, nil).
		AddRow(uuid.New(), chatID, &msgUserID, &userName, "say hello", now, now, "text", nil, nil, nil, nil, nil, nil)

	mock.ExpectQuery(`SELECT\s+msg\.id`).
		WithArgs(userID, chatID, searchText).
		WillReturnRows(rows)

	messages, err := repo.SearchMessagesInChat(ctx, userID, chatID, searchText)

	assert.NoError(t, err)
	assert.Len(t, messages, 2)
	assert.NoError(t, mock.ExpectationsWereMet())
}
