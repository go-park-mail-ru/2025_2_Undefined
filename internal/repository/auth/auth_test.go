package repository
/*
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

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM "user"`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))

	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM "user" WHERE username = \$1\)`).
		WithArgs("user_5").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	mock.ExpectQuery(`INSERT INTO "user"`).
		WithArgs(sqlmock.AnyArg(), "user_5", name, phone, passwordHash, models.UserAccount, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "phone_number", "user_type"}).
			AddRow(userID, "user_5", phone, models.UserAccount))

	user, err := repo.CreateUser(ctx, name, phone, passwordHash)

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, "user_5", user.Username)
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

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM "user"`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))

	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM "user" WHERE username = \$1\)`).
		WithArgs("user_5").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	pqErr := &pq.Error{Code: "23505"}
	mock.ExpectQuery(`INSERT INTO "user"`).
		WithArgs(sqlmock.AnyArg(), "user_5", name, phone, passwordHash, models.UserAccount, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnError(pqErr)

	user, err := repo.CreateUser(ctx, name, phone, passwordHash)

	assert.Error(t, err)
	assert.Equal(t, errs.ErrIsDuplicateKey, err)
	assert.Nil(t, user)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAuthRepository_CreateUser_QueryError(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := New(db)
	ctx := context.Background()

	name := "Test User"
	phone := "+79998887766"
	passwordHash := "hashed_password"

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM "user"`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))

	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM "user" WHERE username = \$1\)`).
		WithArgs("user_5").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	mock.ExpectQuery(`INSERT INTO "user"`).
		WithArgs(sqlmock.AnyArg(), "user_5", name, phone, passwordHash, models.UserAccount, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnError(sql.ErrConnDone)

	user, err := repo.CreateUser(ctx, name, phone, passwordHash)

	assert.Error(t, err)
	assert.Equal(t, sql.ErrConnDone, err)
	assert.Nil(t, user)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAuthRepository_CreateUser_UsernameExists(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := New(db)
	ctx := context.Background()

	name := "Test User"
	phone := "+79998887766"
	passwordHash := "hashed_password"
	userID := uuid.New()

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM "user"`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))

	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM "user" WHERE username = \$1\)`).
		WithArgs("user_5").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	mock.ExpectQuery(`INSERT INTO "user"`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), name, phone, passwordHash, models.UserAccount, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "phone_number", "user_type"}).
			AddRow(userID, "user_"+userID.String()[:8], phone, models.UserAccount))

	user, err := repo.CreateUser(ctx, name, phone, passwordHash)

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Contains(t, user.Username, "user_")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAuthRepository_CreateUser_GenerateUsernameError(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := New(db)
	ctx := context.Background()

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM "user"`).
		WillReturnError(sql.ErrConnDone)

	user, err := repo.CreateUser(ctx, "Test User", "+79998887766", "hashed_password")

	assert.Error(t, err)
	assert.Equal(t, sql.ErrConnDone, err)
	assert.Nil(t, user)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAuthRepository_generateUsername_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := New(db)
	ctx := context.Background()

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM "user"`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(10))

	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM "user" WHERE username = \$1\)`).
		WithArgs("user_10").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	username, err := repo.generateUsername(ctx)

	assert.NoError(t, err)
	assert.Equal(t, "user_10", username)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAuthRepository_generateUsername_CountError(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := New(db)
	ctx := context.Background()

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM "user"`).
		WillReturnError(sql.ErrConnDone)

	username, err := repo.generateUsername(ctx)

	assert.Error(t, err)
	assert.Equal(t, sql.ErrConnDone, err)
	assert.Empty(t, username)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAuthRepository_generateUsername_CheckUsernameError(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := New(db)
	ctx := context.Background()

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM "user"`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(10))

	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM "user" WHERE username = \$1\)`).
		WithArgs("user_10").
		WillReturnError(sql.ErrConnDone)

	username, err := repo.generateUsername(ctx)

	assert.Error(t, err)
	assert.Equal(t, sql.ErrConnDone, err)
	assert.Empty(t, username)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAuthRepository_generateUsername_FallbackToUUID(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := New(db)
	ctx := context.Background()

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM "user"`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(7))

	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM "user" WHERE username = \$1\)`).
		WithArgs("user_7").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	username, err := repo.generateUsername(ctx)

	assert.NoError(t, err)
	assert.Contains(t, username, "user_")
	assert.True(t, len(username) > len("user_7"))
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAuthRepository_checkUsernameExists_True(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := New(db)
	ctx := context.Background()

	username := "test_user"

	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM "user" WHERE username = \$1\)`).
		WithArgs(username).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	exists, err := repo.checkUsernameExists(ctx, username)

	assert.NoError(t, err)
	assert.True(t, exists)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAuthRepository_checkUsernameExists_False(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := New(db)
	ctx := context.Background()

	username := "test_user"

	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM "user" WHERE username = \$1\)`).
		WithArgs(username).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	exists, err := repo.checkUsernameExists(ctx, username)

	assert.NoError(t, err)
	assert.False(t, exists)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAuthRepository_checkUsernameExists_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := New(db)
	ctx := context.Background()

	username := "test_user"

	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM "user" WHERE username = \$1\)`).
		WithArgs(username).
		WillReturnError(sql.ErrConnDone)

	exists, err := repo.checkUsernameExists(ctx, username)

	assert.Error(t, err)
	assert.Equal(t, sql.ErrConnDone, err)
	assert.False(t, exists)
	assert.NoError(t, mock.ExpectationsWereMet())
}
*/