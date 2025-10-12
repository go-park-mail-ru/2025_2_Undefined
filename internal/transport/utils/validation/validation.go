package validation

import (
	"regexp"
	"strings"
	"unicode"

	"github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/dto"
	AuthModels "github.com/go-park-mail-ru/2025_2_Undefined/internal/models/auth"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/models/errs"
)

// ValidateRegisterRequest проверяет все поля регистрации и возвращает все найденные ошибки
func ValidateRegisterRequest(req *AuthModels.RegisterRequest) []errs.ValidationError {
	var errors []errs.ValidationError

	// Проверка обязательных полей
	if req.PhoneNumber == "" {
		errors = append(errors, errs.ValidationError{Field: "phone_number", Message: "Номер телефона обязателен"})
	}
	if req.Email == "" {
		errors = append(errors, errs.ValidationError{Field: "email", Message: "Email обязателен"})
	}
	if req.Username == "" {
		errors = append(errors, errs.ValidationError{Field: "username", Message: "Имя пользователя обязательно"})
	}
	if req.Password == "" {
		errors = append(errors, errs.ValidationError{Field: "password", Message: "Пароль обязателен"})
	}
	if req.Name == "" {
		errors = append(errors, errs.ValidationError{Field: "name", Message: "Имя обязательно"})
	}

	// Валидация номера телефона
	if req.PhoneNumber != "" {
		normalizedPhone, isValid := ValidateAndNormalizePhone(req.PhoneNumber)
		if !isValid {
			errors = append(errors, errs.ValidationError{Field: "phone_number", Message: "Неверный формат номера телефона"})
		} else {
			req.PhoneNumber = normalizedPhone // Обновляем нормализованным значением
		}
	}

	// Валидация email
	if req.Email != "" && !ValidateEmail(req.Email) {
		errors = append(errors, errs.ValidationError{Field: "email", Message: "Неверный формат email"})
	}

	// Валидация пароля
	if req.Password != "" && !ValidatePassword(req.Password) {
		errors = append(errors, errs.ValidationError{Field: "password", Message: "Пароль должен содержать минимум 8 символов и только латинские буквы, цифры и специальные символы"})
	}

	// Валидация username
	if req.Username != "" && !ValidateUsername(req.Username) {
		errors = append(errors, errs.ValidationError{Field: "username", Message: "Неверный формат имени пользователя"})
	}

	// Валидация имени
	if req.Name != "" && !ValidateName(req.Name) {
		errors = append(errors, errs.ValidationError{Field: "name", Message: "Неверный формат имени"})
	}

	return errors
}

// ValidateLoginRequest проверяет все поля входа и возвращает все найденные ошибки
func ValidateLoginRequest(req *AuthModels.LoginRequest) []errs.ValidationError {
	var errors []errs.ValidationError

	// Проверка обязательных полей
	if req.PhoneNumber == "" {
		errors = append(errors, errs.ValidationError{Field: "phone_number", Message: "Номер телефона обязателен"})
	}
	if req.Password == "" {
		errors = append(errors, errs.ValidationError{Field: "password", Message: "Пароль обязателен"})
	}

	// Валидация номера телефона
	if req.PhoneNumber != "" {
		normalizedPhone, isValid := ValidateAndNormalizePhone(req.PhoneNumber)
		if !isValid {
			errors = append(errors, errs.ValidationError{Field: "phone_number", Message: "Неверный формат номера телефона"})
		} else {
			req.PhoneNumber = normalizedPhone // Обновляем нормализованным значением
		}
	}

	// Валидация пароля
	if req.Password != "" && !ValidatePassword(req.Password) {
		errors = append(errors, errs.ValidationError{Field: "password", Message: "В пароле разрешены только латинские буквы, цифры и специальные символы"})
	}

	return errors
}

// ConvertToValidationErrorsDTO конвертирует errs.ValidationError в DTO
func ConvertToValidationErrorsDTO(errors []errs.ValidationError) dto.ValidationErrorsDTO {
	var dtoErrors []dto.ValidationErrorDTO
	for _, err := range errors {
		dtoErrors = append(dtoErrors, dto.ValidationErrorDTO{
			Field:   err.Field,
			Message: err.Message,
		})
	}

	return dto.ValidationErrorsDTO{
		Message: "Ошибка валидации",
		Errors:  dtoErrors,
	}
}

func ValidateAndNormalizePhone(phone string) (string, bool) {
	phoneWithoutSpace := strings.ReplaceAll(phone, " ", "")
	if phone == "" {
		return phone, false
	}
	// Должен начинаться с +7 или 8
	if !strings.HasPrefix(phoneWithoutSpace, "+7") && !strings.HasPrefix(phoneWithoutSpace, "8") {
		return phone, false
	}
	cleanPhone := phoneWithoutSpace
	if strings.HasPrefix(phoneWithoutSpace, "+7") {
		cleanPhone = phoneWithoutSpace[2:]
	} else if strings.HasPrefix(phoneWithoutSpace, "8") {
		cleanPhone = phoneWithoutSpace[1:]
	}

	// Должно быть 10 цифр
	if len(cleanPhone) != 10 {
		return phone, false
	}
	for _, char := range cleanPhone {
		if !unicode.IsDigit(char) {
			return phone, false
		}
	}
	return "+7" + cleanPhone, true
}

func ValidateEmail(email string) bool {
	reg, err := regexp.Compile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if err != nil {
		return false
	}
	if email == "" {
		return false
	}
	if !reg.MatchString(email) {
		return false
	}
	return true
}

func ValidatePassword(password string) bool {
	if len(password) < 8 {
		return false
	}
	passwordRegex, err := regexp.Compile(`^[a-zA-Z0-9!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?]+$`)
	if err != nil {
		return false
	}
	if !passwordRegex.MatchString(password) {
		return false
	}
	return true
}

func ValidateUsername(username string) bool {
	if len(username) < 3 || len(username) > 20 {
		return false
	}
	usernameRegex, err := regexp.Compile(`^[a-zA-Z0-9_]+$`)
	if err != nil {
		return false
	}
	if !usernameRegex.MatchString(username) {
		return false
	}
	return true
}

func ValidateName(name string) bool {
	if len(name) < 1 || len(name) > 20 {
		return false
	}
	return true
}
