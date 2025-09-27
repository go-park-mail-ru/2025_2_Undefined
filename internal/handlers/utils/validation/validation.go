package validation

import (
	"regexp"
	"strings"
	"unicode"
)

func ValidatePhone(phone string) (string, bool) {
	phoneWithoutSpace := strings.ReplaceAll(phone, " ", "")
	phoneWithoutSpace = strings.ReplaceAll(phoneWithoutSpace, "(", "")
	phoneWithoutSpace = strings.ReplaceAll(phoneWithoutSpace, ")", "")
	phoneWithoutSpace = strings.ReplaceAll(phoneWithoutSpace, "-", "")

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

	return "+7" + cleanPhone, true
}

func ValidateEmail(email string) bool {
	reg := regexp.Compile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if email == "" {
		return false
	}
	if !reg.MatchString(email) {
		return false
	}

}
