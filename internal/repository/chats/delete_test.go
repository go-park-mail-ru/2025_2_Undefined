package repository

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
)

func TestChatsRepository_DeleteChat_Success(t *testing.T) {
	mock, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("failed to create pgxmock pool: %v", err)
	}
	defer mock.Close()

	repo := NewChatsRepository(mock)
	userID := uuid.New()
	chatID := uuid.New()

	// Проверка роли админа
	mock.ExpectQuery(checkUserRoleQuery).
		WithArgs(userID, chatID, "admin").
		WillReturnRows(pgxmock.NewRows([]string{"exists"}).AddRow(true))

	// Удаление чата
	mock.ExpectExec(deleteChatQuery).
		WithArgs(chatID).
		WillReturnResult(pgxmock.NewResult("DELETE", 1))

	ctx := context.Background()
	err = repo.DeleteChat(ctx, userID, chatID)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestChatsRepository_DeleteChat_NotAdmin(t *testing.T) {
	mock, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("failed to create pgxmock pool: %v", err)
	}
	defer mock.Close()

	repo := NewChatsRepository(mock)
	userID := uuid.New()
	chatID := uuid.New()

	mock.ExpectQuery(checkUserRoleQuery).
		WithArgs(userID, chatID, "admin").
		WillReturnRows(pgxmock.NewRows([]string{"exists"}).AddRow(false))

	ctx := context.Background()
	err = repo.DeleteChat(ctx, userID, chatID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user is not admin")
	assert.NoError(t, mock.ExpectationsWereMet())
}
