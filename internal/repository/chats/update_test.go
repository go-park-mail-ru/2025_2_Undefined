package repository

import (
	"context"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

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
	mock.ExpectQuery(regexp.QuoteMeta(checkUserRoleQuery)).
		WithArgs(userID, chatID, "admin").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	// Обновление чата
	mock.ExpectExec(regexp.QuoteMeta(updateChatQuery)).
		WithArgs("New Name", "New Description", chatID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	ctx := context.Background()
	err = repo.UpdateChat(ctx, userID, chatID, "New Name", "New Description")

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
