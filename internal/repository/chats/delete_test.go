package repository

import (
	"context"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

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
	mock.ExpectQuery(regexp.QuoteMeta(checkUserRoleQuery)).
		WithArgs(userID, chatID, "admin").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	// Удаление чата
	mock.ExpectExec(regexp.QuoteMeta(deleteChatQuery)).
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

	mock.ExpectQuery(regexp.QuoteMeta(checkUserRoleQuery)).
		WithArgs(userID, chatID, "admin").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	ctx := context.Background()
	err = repo.DeleteChat(ctx, userID, chatID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user is not admin")
	assert.NoError(t, mock.ExpectationsWereMet())
}
