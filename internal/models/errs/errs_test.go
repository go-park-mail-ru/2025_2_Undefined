package errs

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrorConstants(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{"ErrInvalidToken", ErrInvalidToken, "invalid token"},
		{"ErrNotFound", ErrNotFound, "not found"},
		{"ErrBadRequest", ErrBadRequest, "bad request"},
		{"ErrInvalidCredentials", ErrInvalidCredentials, "invalid credentials"},
		{"ErrRequiredFieldsMissing", ErrRequiredFieldsMissing, "required fields missing"},
		{"ErrUserNotFound", ErrUserNotFound, "user not found"},
		{"ErrJWTIsRequired", ErrJWTIsRequired, "JWT token required"},
		{"ErrIsDuplicateKey", ErrIsDuplicateKey, "duplicate key"},
		{"ErrServiceIsOverloaded", ErrServiceIsOverloaded, "service is overloaded, try again later"},
		{"ErrNoRights", ErrNoRights, "no rights to perform this action"},
		{"ErrSessionNotFound", ErrSessionNotFound, "session not found"},
		{"ErrInternalServerError", ErrInternalServerError, "internal server error"},
		{"ErrContactAlreadyExists", ErrContactAlreadyExists, "contact already exists"},
		{"ErrContactNotFound", ErrContactNotFound, "contact not found"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.err.Error())
		})
	}
}

func TestErrorComparison(t *testing.T) {
	assert.True(t, errors.Is(ErrUserNotFound, ErrUserNotFound))
	assert.False(t, errors.Is(ErrUserNotFound, ErrNotFound))
}

func TestPostgresErrorCodes(t *testing.T) {
	assert.Equal(t, "23505", PostgresErrorUniqueViolationCode)
	assert.Equal(t, "23503", PostgresErrorForeignKeyViolationCode)
}

func TestValidateUserAlreadyExists(t *testing.T) {
	assert.Equal(t, "a user with such a phone already exists", ValidateUserAlreadyExists)
}
