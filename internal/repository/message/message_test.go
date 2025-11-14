package messages

import (
	"context"
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	modelsMessage "github.com/go-park-mail-ru/2025_2_Undefined/internal/models/message"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestMessageRepository_InsertMessage_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock database: %v", err)
	}
	defer db.Close()

	repo := NewMessageRepository(db)

	userID := uuid.New()

	msg := modelsMessage.CreateMessage{
		ChatID:    uuid.New(),
		UserID:    &userID,
		Text:      "hello",
		CreatedAt: time.Now(),
		Type:      "text",
	}

	expectedID := uuid.New()

	rows := sqlmock.NewRows([]string{"id"}).AddRow(expectedID.String())

	mock.ExpectQuery(regexp.QuoteMeta(insertMessageQuery)).
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
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock database: %v", err)
	}
	defer db.Close()

	repo := NewMessageRepository(db)

	userID := uuid.New()

	msg := modelsMessage.CreateMessage{
		ChatID:    uuid.New(),
		UserID:    &userID,
		Text:      "hello",
		CreatedAt: time.Now(),
		Type:      "text",
	}

	mock.ExpectQuery(regexp.QuoteMeta(insertMessageQuery)).
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
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock database: %v", err)
	}
	defer db.Close()

	repo := NewMessageRepository(db)

	userID := uuid.New()
	messageID := uuid.New()
	chatID := uuid.New()
	senderID := uuid.New()
	createdAt := time.Now()

	rows := sqlmock.NewRows([]string{"id", "chat_id", "user_id", "name", "attachment_id", "text", "created_at", "message_type"}).
		AddRow(messageID.String(), chatID.String(), senderID.String(), "John Doe", nil, "Hello world", createdAt, "text")

	mock.ExpectQuery(regexp.QuoteMeta(getLastMessagesOfChatsQuery)).
		WithArgs(userID).
		WillReturnRows(rows)

	ctx := context.Background()
	messages, err := repo.GetLastMessagesOfChats(ctx, userID)

	assert.NoError(t, err)
	assert.Len(t, messages, 1)
	assert.Equal(t, messageID, messages[0].ID)
	assert.Equal(t, chatID, messages[0].ChatID)
	assert.Equal(t, senderID, *messages[0].UserID)
	assert.Equal(t, "John Doe", messages[0].UserName)
	assert.Equal(t, "Hello world", messages[0].Text)
	assert.Equal(t, "text", messages[0].Type)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestMessageRepository_GetMessagesOfChat_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock database: %v", err)
	}
	defer db.Close()

	repo := NewMessageRepository(db)

	chatID := uuid.New()
	messageID := uuid.New()
	userID := uuid.New()
	createdAt := time.Now()
	offset := 0
	limit := 10

	rows := sqlmock.NewRows([]string{"id", "chat_id", "user_id", "name", "attachment_id", "text", "created_at", "message_type"}).
		AddRow(messageID.String(), chatID.String(), userID.String(), "John", nil, "Test message", createdAt, "text")

	mock.ExpectQuery(regexp.QuoteMeta(getMessagesOfChatQuery)).
		WithArgs(chatID, offset, limit).
		WillReturnRows(rows)

	ctx := context.Background()
	messages, err := repo.GetMessagesOfChat(ctx, chatID, offset, limit)

	assert.NoError(t, err)
	assert.Len(t, messages, 1)
	assert.Equal(t, messageID, messages[0].ID)
	assert.Equal(t, chatID, messages[0].ChatID)
	assert.Equal(t, userID, *messages[0].UserID)
	assert.Equal(t, "John", messages[0].UserName)
	assert.Equal(t, "Test message", messages[0].Text)
	assert.Equal(t, "text", messages[0].Type)
	assert.NoError(t, mock.ExpectationsWereMet())
}
