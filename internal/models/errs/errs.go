package errs

import "errors"

// PostgreSQL коды ошибок
const (
	PostgresErrorUniqueViolationCode     = "23505"
	PostgresErrorForeignKeyViolationCode = "23503"
)

var (
	ErrInvalidToken          = errors.New("invalid token")
	ErrNotFound              = errors.New("not found")
	ErrBadRequest            = errors.New("bad request")
	ErrInvalidCredentials    = errors.New("invalid credentials")
	ErrRequiredFieldsMissing = errors.New("required fields missing")
	ErrUserNotFound          = errors.New("user not found")
	ErrJWTIsRequired         = errors.New("JWT token required")
	ErrIsDuplicateKey        = errors.New("duplicate key")
	ErrServiceIsOverloaded   = errors.New("service is overloaded, try again later")
	ErrNoRights              = errors.New("no rights to perform this action")
	ErrSessionNotFound       = errors.New("session not found")
	ErrInternalServerError   = errors.New("internal server error")
	ErrContactAlreadyExists  = errors.New("contact already exists")
	ErrContactNotFound       = errors.New("contact not found")
)

var (
	ValidateUserAlreadyExists = "a user with such a phone already exists"
)

// ValidationError представляет ошибку валидации поля
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}
