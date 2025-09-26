package errs

import "errors"

var (
	ErrInvalidToken          = errors.New("invalid token")
	ErrNotFound              = errors.New("not found")
	ErrBadRequest            = errors.New("bad request")
	ErrInvalidCredentials    = errors.New("invalid credentials")
	ErrRequiredFieldsMissing = errors.New("required fields are missing")
	ErrUserNotFound          = errors.New("user not found")
)
