package errs

import "errors"

var (
	ErrInvalidToken          = errors.New("неверный токен")
	ErrNotFound              = errors.New("не найдено")
	ErrBadRequest            = errors.New("неверный запрос")
	ErrInvalidCredentials    = errors.New("неверные учетные данные")
	ErrRequiredFieldsMissing = errors.New("отсутствуют обязательные поля")
	ErrUserNotFound          = errors.New("пользователь не найден")
	ErrJWTIsRequired         = errors.New("требуется JWT токен")
)

// ValidationError представляет ошибку валидации поля
type ValidationError struct {
	Field   string
	Message string
}
