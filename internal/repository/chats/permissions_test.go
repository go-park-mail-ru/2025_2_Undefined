package repository

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
)

func TestChatsRepository_CheckUserHasRole_Success(t *testing.T) {
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
		WillReturnRows(pgxmock.NewRows([]string{"exists"}).AddRow(true))

	ctx := context.Background()
	hasRole, err := repo.CheckUserHasRole(ctx, userID, chatID, "admin")

	assert.NoError(t, err)
	assert.True(t, hasRole)
	assert.NoError(t, mock.ExpectationsWereMet())
}
