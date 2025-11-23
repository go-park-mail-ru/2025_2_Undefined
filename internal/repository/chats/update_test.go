package repository

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
)

func TestChatsRepository_UpdateChat_Success(t *testing.T) {
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

	// Обновление чата
	mock.ExpectExec(updateChatQuery).
		WithArgs("New Name", "New Description", chatID).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))

	ctx := context.Background()
	err = repo.UpdateChat(ctx, userID, chatID, "New Name", "New Description")

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
