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
