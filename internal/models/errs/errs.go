package errs

import "errors"

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
)

// ValidationError представляет ошибку валидации поля
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}
