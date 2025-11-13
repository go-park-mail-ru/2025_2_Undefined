package messages
/*
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

	msg := modelsMessage.CreateMessage{
		ChatID:    uuid.New(),
		UserID:    uuid.New(),
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

	msg := modelsMessage.CreateMessage{
		ChatID:    uuid.New(),
		UserID:    uuid.New(),
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
*/