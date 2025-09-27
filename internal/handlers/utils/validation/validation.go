package validation

import (
	"log"
	"regexp"
	"strings"
	"unicode"
)

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
	if strings.HasPrefix(phone, "+7") {
		cleanPhone = phone[2:]
	} else if strings.HasPrefix(phone, "8") {
		cleanPhone = phone[1:]
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
	log.Print("+7" + cleanPhone)
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
