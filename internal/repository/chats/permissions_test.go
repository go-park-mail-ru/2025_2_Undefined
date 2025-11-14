package repository

import (
	"context"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestChatsRepository_CheckUserHasRole_Success(t *testing.T) {
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
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	ctx := context.Background()
	hasRole, err := repo.CheckUserHasRole(ctx, userID, chatID, "admin")

	assert.NoError(t, err)
	assert.True(t, hasRole)
	assert.NoError(t, mock.ExpectationsWereMet())
}
