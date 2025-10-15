package dto

// ValidationErrorDTO представляет ошибку валидации для конкретного поля
type ValidationErrorDTO struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ValidationErrorsDTO представляет множественные ошибки валидации
type ValidationErrorsDTO struct {
	Message string               `json:"message"`
	Errors  []ValidationErrorDTO `json:"errors"`
}

type ErrorDTO struct {
	Message string `json:"message"`
}
