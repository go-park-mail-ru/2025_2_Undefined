package repository

import (
	"context"
	"fmt"
	"testing"
	"time"

	ContactModels "github.com/go-park-mail-ru/2025_2_Undefined/internal/models/contact"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/models/errs"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
)

func TestContactRepository_CreateContact_Success(t *testing.T) {
	mock, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
	assert.NoError(t, err)
	defer mock.Close()

	repo := New(mock)
	ctx := context.Background()

	userID := uuid.New()
	contactUserID := uuid.New()

	mock.ExpectQuery(`
		INSERT INTO contact (user_id, contact_user_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4)
		 RETURNING user_id`).
		WithArgs(userID, contactUserID, pgxmock.AnyArg(), pgxmock.AnyArg()).
		WillReturnRows(pgxmock.NewRows([]string{"user_id"}).AddRow(userID))

	err = repo.CreateContact(ctx, userID, contactUserID)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestContactRepository_CreateContact_DuplicateKey(t *testing.T) {
	mock, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
	assert.NoError(t, err)
	defer mock.Close()

	repo := New(mock)
	ctx := context.Background()

	userID := uuid.New()
	contactUserID := uuid.New()

	pgErr := &pgconn.PgError{Code: "23505"}
	mock.ExpectQuery(`
		INSERT INTO contact (user_id, contact_user_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4)
		 RETURNING user_id`).
		WithArgs(userID, contactUserID, pgxmock.AnyArg(), pgxmock.AnyArg()).
		WillReturnError(pgErr)

	err = repo.CreateContact(ctx, userID, contactUserID)

	assert.Error(t, err)
	assert.Equal(t, errs.ErrIsDuplicateKey, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestContactRepository_CreateContact_UserNotFound(t *testing.T) {
	mock, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
	assert.NoError(t, err)
	defer mock.Close()

	repo := New(mock)
	ctx := context.Background()

	userID := uuid.New()
	contactUserID := uuid.New()

	pgErr := &pgconn.PgError{Code: "23503"}
	mock.ExpectQuery(`
		INSERT INTO contact (user_id, contact_user_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4)
		 RETURNING user_id`).
		WithArgs(userID, contactUserID, pgxmock.AnyArg(), pgxmock.AnyArg()).
		WillReturnError(pgErr)

	err = repo.CreateContact(ctx, userID, contactUserID)

	assert.Error(t, err)
	assert.Equal(t, errs.ErrUserNotFound, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestContactRepository_CreateContact_QueryError(t *testing.T) {
	mock, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
	assert.NoError(t, err)
	defer mock.Close()

	repo := New(mock)
	ctx := context.Background()

	userID := uuid.New()
	contactUserID := uuid.New()

	mock.ExpectQuery(`
		INSERT INTO contact (user_id, contact_user_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4)
		 RETURNING user_id`).
		WithArgs(userID, contactUserID, pgxmock.AnyArg(), pgxmock.AnyArg()).
		WillReturnError(fmt.Errorf("connection error"))

	err = repo.CreateContact(ctx, userID, contactUserID)

	assert.Error(t, err)
	assert.Equal(t, fmt.Errorf("connection error"), err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestContactRepository_GetContactsByUserID_Success(t *testing.T) {
	mock, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
	assert.NoError(t, err)
	defer mock.Close()

	repo := New(mock)
	ctx := context.Background()

	userID := uuid.New()
	contactUserID := uuid.New()
	createdAt := time.Now()
	updatedAt := time.Now()

	rows := pgxmock.NewRows([]string{"user_id", "contact_user_id", "created_at", "updated_at"}).
		AddRow(userID, contactUserID, createdAt, updatedAt)

	mock.ExpectQuery(`SELECT user_id, contact_user_id, created_at, updated_at
		FROM contact
		WHERE user_id = $1`).
		WithArgs(userID).
		WillReturnRows(rows)

	contacts, err := repo.GetContactsByUserID(ctx, userID)

	assert.NoError(t, err)
	assert.Len(t, contacts, 1)
	assert.IsType(t, []*ContactModels.Contact{}, contacts)
	assert.Equal(t, userID, contacts[0].UserID)
	assert.Equal(t, contactUserID, contacts[0].ContactUserID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestContactRepository_GetContactsByUserID_NoContacts(t *testing.T) {
	mock, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
	assert.NoError(t, err)
	defer mock.Close()

	repo := New(mock)
	ctx := context.Background()

	userID := uuid.New()

	rows := pgxmock.NewRows([]string{"user_id", "contact_user_id", "created_at", "updated_at"})

	mock.ExpectQuery(`SELECT user_id, contact_user_id, created_at, updated_at
		FROM contact
		WHERE user_id = $1`).
		WithArgs(userID).
		WillReturnRows(rows)

	contacts, err := repo.GetContactsByUserID(ctx, userID)

	assert.NoError(t, err)
	assert.Empty(t, contacts)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestContactRepository_GetContactsByUserID_QueryError(t *testing.T) {
	mock, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
	assert.NoError(t, err)
	defer mock.Close()

	repo := New(mock)
	ctx := context.Background()

	userID := uuid.New()

	mock.ExpectQuery(`SELECT user_id, contact_user_id, created_at, updated_at
		FROM contact
		WHERE user_id = $1`).
		WithArgs(userID).
		WillReturnError(fmt.Errorf("connection error"))

	contacts, err := repo.GetContactsByUserID(ctx, userID)

	assert.Error(t, err)
	assert.Nil(t, contacts)
	assert.Equal(t, fmt.Errorf("connection error"), err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestContactRepository_GetContactsByUserID_ScanError(t *testing.T) {
	mock, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
	assert.NoError(t, err)
	defer mock.Close()

	repo := New(mock)
	ctx := context.Background()

	userID := uuid.New()

	rows := pgxmock.NewRows([]string{"user_id", "contact_user_id", "created_at", "updated_at"}).
		AddRow("invalid-uuid", uuid.New(), time.Now(), time.Now())

	mock.ExpectQuery(`SELECT user_id, contact_user_id, created_at, updated_at
		FROM contact
		WHERE user_id = $1`).
		WithArgs(userID).
		WillReturnRows(rows)

	contacts, err := repo.GetContactsByUserID(ctx, userID)

	assert.Error(t, err)
	assert.Nil(t, contacts)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestContactRepository_GetContactsByUserID_RowsError(t *testing.T) {
	mock, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
	assert.NoError(t, err)
	defer mock.Close()

	repo := New(mock)
	ctx := context.Background()

	userID := uuid.New()
	contactUserID := uuid.New()

	rows := pgxmock.NewRows([]string{"user_id", "contact_user_id", "created_at", "updated_at"}).
		AddRow(userID, contactUserID, time.Now(), time.Now()).
		RowError(0, fmt.Errorf("connection error"))

	mock.ExpectQuery(`SELECT user_id, contact_user_id, created_at, updated_at
		FROM contact
		WHERE user_id = $1`).
		WithArgs(userID).
		WillReturnRows(rows)

	contacts, err := repo.GetContactsByUserID(ctx, userID)

	assert.Error(t, err)
	assert.Nil(t, contacts)
	assert.Equal(t, fmt.Errorf("connection error"), err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestContactRepository_GetAllContacts_Success(t *testing.T) {
	mock, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
	assert.NoError(t, err)
	defer mock.Close()

	repo := New(mock)
	ctx := context.Background()

	userID1 := uuid.New()
	contactUserID1 := uuid.New()
	userID2 := uuid.New()
	contactUserID2 := uuid.New()
	createdAt := time.Now()
	updatedAt := time.Now()

	rows := pgxmock.NewRows([]string{"user_id", "contact_user_id", "created_at", "updated_at"}).
		AddRow(userID1, contactUserID1, createdAt, updatedAt).
		AddRow(userID2, contactUserID2, createdAt, updatedAt)

	mock.ExpectQuery(`SELECT user_id, contact_user_id, created_at, updated_at
		FROM contact`).
		WillReturnRows(rows)

	contacts, err := repo.GetAllContacts(ctx)

	assert.NoError(t, err)
	assert.Len(t, contacts, 2)
	assert.Equal(t, userID1, contacts[0].UserID)
	assert.Equal(t, contactUserID1, contacts[0].ContactUserID)
	assert.Equal(t, userID2, contacts[1].UserID)
	assert.Equal(t, contactUserID2, contacts[1].ContactUserID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestContactRepository_GetAllContacts_QueryError(t *testing.T) {
	mock, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
	assert.NoError(t, err)
	defer mock.Close()

	repo := New(mock)
	ctx := context.Background()

	mock.ExpectQuery(`SELECT user_id, contact_user_id, created_at, updated_at
		FROM contact`).
		WillReturnError(fmt.Errorf("database error"))

	contacts, err := repo.GetAllContacts(ctx)

	assert.Error(t, err)
	assert.Nil(t, contacts)
	assert.NoError(t, mock.ExpectationsWereMet())
}
