package response

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-park-mail-ru/2025_2_Undefined/internal/models/errs"
	"github.com/stretchr/testify/assert"
)

func TestSendErrorWithAutoStatus(t *testing.T) {
	tests := []struct {
		name           string
		err            error
		expectedStatus int
	}{
		{
			name:           "Service overloaded error",
			err:            errs.ErrServiceIsOverloaded,
			expectedStatus: http.StatusServiceUnavailable,
		},
		{
			name:           "Not found error",
			err:            errs.ErrNotFound,
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "User not found error",
			err:            errs.ErrUserNotFound,
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "Invalid token error",
			err:            errs.ErrInvalidToken,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Invalid credentials error",
			err:            errs.ErrInvalidCredentials,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "JWT required error",
			err:            errs.ErrJWTIsRequired,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Duplicate key error",
			err:            errs.ErrIsDuplicateKey,
			expectedStatus: http.StatusConflict,
		},
		{
			name:           "Required fields missing error",
			err:            errs.ErrRequiredFieldsMissing,
			expectedStatus: http.StatusUnprocessableEntity,
		},
		{
			name:           "Bad request error",
			err:            errs.ErrBadRequest,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Unknown error",
			err:            errors.New("unknown error message"),
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Wrapped service overloaded error",
			err:            fmt.Errorf("database connection failed: %w", errs.ErrServiceIsOverloaded),
			expectedStatus: http.StatusServiceUnavailable,
		},
		{
			name:           "Wrapped user not found error",
			err:            fmt.Errorf("validation failed: %w", errs.ErrUserNotFound),
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			ctx := context.Background()

			SendErrorWithAutoStatus(ctx, "TestOperation", w, tt.err)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
		})
	}
}
