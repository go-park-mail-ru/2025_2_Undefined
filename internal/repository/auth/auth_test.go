package repository

import (
	"context"
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/models/errs"
	models "github.com/go-park-mail-ru/2025_2_Undefined/internal/models/user"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
)

func TestAuthRepository_CreateUser_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := New(db)
	ctx := context.Background()

	name := "Test User"
	phone := "+79998887766"
	passwordHash := "hashed_password"
	userID := uuid.New()
	username := "user_123456789012345"

	mock.ExpectBegin()

	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM "user" WHERE username = \$1\)`).
		WithArgs(sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	mock.ExpectQuery(`INSERT INTO "user"`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), name, phone, passwordHash, models.UserAccount, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "phone_number", "user_type"}).
			AddRow(userID, username, phone, models.UserAccount))

	mock.ExpectCommit()

	user, err := repo.CreateUser(ctx, name, phone, passwordHash)

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, username, user.Username)
	assert.Equal(t, name, user.Name)
	assert.Equal(t, phone, user.PhoneNumber)
	assert.Equal(t, passwordHash, user.PasswordHash)
	assert.Equal(t, models.UserAccount, user.AccountType)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAuthRepository_CreateUser_DuplicateKey(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := New(db)
	ctx := context.Background()

	name := "Test User"
	phone := "+79998887766"
	passwordHash := "hashed_password"

	mock.ExpectBegin()

	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM "user" WHERE username = \$1\)`).
		WithArgs(sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	pqErr := &pq.Error{Code: "23505"}
	mock.ExpectQuery(`INSERT INTO "user"`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), name, phone, passwordHash, models.UserAccount, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnError(pqErr)

	mock.ExpectRollback()

	user, err := repo.CreateUser(ctx, name, phone, passwordHash)

	assert.Error(t, err)
	assert.Equal(t, errs.ErrIsDuplicateKey, err)
	assert.Nil(t, user)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAuthRepository_CreateUser_BeginTxError(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := New(db)
	ctx := context.Background()

	mock.ExpectBegin().WillReturnError(sql.ErrConnDone)

	user, err := repo.CreateUser(ctx, "Test User", "+79998887766", "hashed_password")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "begin transaction")
	assert.Nil(t, user)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAuthRepository_CreateUser_GenerateUsernameError(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := New(db)
	ctx := context.Background()

	mock.ExpectBegin()

	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM "user" WHERE username = \$1\)`).
		WithArgs(sqlmock.AnyArg()).
		WillReturnError(sql.ErrConnDone)

	mock.ExpectRollback()

	user, err := repo.CreateUser(ctx, "Test User", "+79998887766", "hashed_password")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "generate username")
	assert.Nil(t, user)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAuthRepository_CreateUser_CommitError(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := New(db)
	ctx := context.Background()

	name := "Test User"
	phone := "+79998887766"
	passwordHash := "hashed_password"
	userID := uuid.New()
	username := "user_123456789012345"

	mock.ExpectBegin()

	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM "user" WHERE username = \$1\)`).
		WithArgs(sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	mock.ExpectQuery(`INSERT INTO "user"`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), name, phone, passwordHash, models.UserAccount, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "phone_number", "user_type"}).
			AddRow(userID, username, phone, models.UserAccount))

	mock.ExpectCommit().WillReturnError(sql.ErrTxDone)

	user, err := repo.CreateUser(ctx, name, phone, passwordHash)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "commit transaction")
	assert.Nil(t, user)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAuthRepository_CreateUser_UsernameCollision(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := New(db)
	ctx := context.Background()

	name := "Test User"
	phone := "+79998887766"
	passwordHash := "hashed_password"
	userID := uuid.New()

	mock.ExpectBegin()

	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM "user" WHERE username = \$1\)`).
		WithArgs(sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM "user" WHERE username = \$1\)`).
		WithArgs(sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	mock.ExpectQuery(`INSERT INTO "user"`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), name, phone, passwordHash, models.UserAccount, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "phone_number", "user_type"}).
			AddRow(userID, "user_123456789012346", phone, models.UserAccount))

	mock.ExpectCommit()

	user, err := repo.CreateUser(ctx, name, phone, passwordHash)

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Contains(t, user.Username, "user_")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestNew(t *testing.T) {
	db, _, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := New(db)

	assert.NotNil(t, repo)
	assert.Equal(t, db, repo.db)
}
