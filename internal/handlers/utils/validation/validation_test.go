package validation

import (
	"testing"

	AuthModels "github.com/go-park-mail-ru/2025_2_Undefined/internal/models/auth"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/models/errs"
)

func TestValidateRegisterRequest(t *testing.T) {
	tests := []struct {
		name     string
		req      *AuthModels.RegisterRequest
		wantErrs int
	}{
		{
			name: "valid request",
			req: &AuthModels.RegisterRequest{
				PhoneNumber: "+79123456789",
				Email:       "test@example.com",
				Username:    "testuser",
				Password:    "password123",
				Name:        "Test User",
			},
			wantErrs: 0,
		},
		{
			name:     "empty request",
			req:      &AuthModels.RegisterRequest{},
			wantErrs: 5,
		},
		{
			name: "invalid phone",
			req: &AuthModels.RegisterRequest{
				PhoneNumber: "invalid",
				Email:       "test@example.com",
				Username:    "testuser",
				Password:    "password123",
				Name:        "Test User",
			},
			wantErrs: 1,
		},
		{
			name: "invalid email",
			req: &AuthModels.RegisterRequest{
				PhoneNumber: "+79123456789",
				Email:       "invalid-email",
				Username:    "testuser",
				Password:    "password123",
				Name:        "Test User",
			},
			wantErrs: 1,
		},
		{
			name: "invalid password",
			req: &AuthModels.RegisterRequest{
				PhoneNumber: "+79123456789",
				Email:       "test@example.com",
				Username:    "testuser",
				Password:    "123",
				Name:        "Test User",
			},
			wantErrs: 1,
		},
		{
			name: "invalid username",
			req: &AuthModels.RegisterRequest{
				PhoneNumber: "+79123456789",
				Email:       "test@example.com",
				Username:    "ab",
				Password:    "password123",
				Name:        "Test User",
			},
			wantErrs: 1,
		},
		{
			name: "invalid name",
			req: &AuthModels.RegisterRequest{
				PhoneNumber: "+79123456789",
				Email:       "test@example.com",
				Username:    "testuser",
				Password:    "password123",
				Name:        "",
			},
			wantErrs: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := ValidateRegisterRequest(tt.req)
			if len(errors) != tt.wantErrs {
				t.Errorf("ValidateRegisterRequest() got %d errors, want %d", len(errors), tt.wantErrs)
			}
		})
	}
}

func TestValidateLoginRequest(t *testing.T) {
	tests := []struct {
		name     string
		req      *AuthModels.LoginRequest
		wantErrs int
	}{
		{
			name: "valid request",
			req: &AuthModels.LoginRequest{
				PhoneNumber: "+79123456789",
				Password:    "password123",
			},
			wantErrs: 0,
		},
		{
			name:     "empty request",
			req:      &AuthModels.LoginRequest{},
			wantErrs: 2,
		},
		{
			name: "invalid phone",
			req: &AuthModels.LoginRequest{
				PhoneNumber: "invalid",
				Password:    "password123",
			},
			wantErrs: 1,
		},
		{
			name: "invalid password",
			req: &AuthModels.LoginRequest{
				PhoneNumber: "+79123456789",
				Password:    "пароль",
			},
			wantErrs: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := ValidateLoginRequest(tt.req)
			if len(errors) != tt.wantErrs {
				t.Errorf("ValidateLoginRequest() got %d errors, want %d", len(errors), tt.wantErrs)
			}
		})
	}
}

func TestConvertToValidationErrorsDTO(t *testing.T) {
	errors := []errs.ValidationError{
		{Field: "email", Message: "Email is required"},
		{Field: "password", Message: "Password is required"},
	}

	result := ConvertToValidationErrorsDTO(errors)

	if result.Message != "Ошибка валидации" {
		t.Errorf("Expected message 'Ошибка валидации', got '%s'", result.Message)
	}

	if len(result.Errors) != 2 {
		t.Errorf("Expected 2 errors, got %d", len(result.Errors))
	}

	if result.Errors[0].Field != "email" || result.Errors[0].Message != "Email is required" {
		t.Errorf("First error mismatch")
	}
}

func TestValidateAndNormalizePhone(t *testing.T) {
	tests := []struct {
		name      string
		phone     string
		wantPhone string
		wantValid bool
	}{
		{"valid +7", "+79123456789", "+79123456789", true},
		{"valid 8", "89123456789", "+79123456789", true},
		{"with spaces", "+7 912 345 67 89", "+79123456789", true},
		{"empty", "", "", false},
		{"invalid prefix", "79123456789", "79123456789", false},
		{"too short", "+791234567", "+791234567", false},
		{"too long", "+791234567890", "+791234567890", false},
		{"with letters", "+7912345678a", "+7912345678a", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPhone, gotValid := ValidateAndNormalizePhone(tt.phone)
			if gotPhone != tt.wantPhone {
				t.Errorf("ValidateAndNormalizePhone() phone = %v, want %v", gotPhone, tt.wantPhone)
			}
			if gotValid != tt.wantValid {
				t.Errorf("ValidateAndNormalizePhone() valid = %v, want %v", gotValid, tt.wantValid)
			}
		})
	}
}

func TestValidateEmail(t *testing.T) {
	tests := []struct {
		name  string
		email string
		want  bool
	}{
		{"valid email", "test@example.com", true},
		{"valid with dots", "test.user@example.com", true},
		{"valid with plus", "test+tag@example.com", true},
		{"empty", "", false},
		{"no @", "testexample.com", false},
		{"no domain", "test@", false},
		{"no local", "@example.com", false},
		{"invalid domain", "test@invalid", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ValidateEmail(tt.email); got != tt.want {
				t.Errorf("ValidateEmail() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		want     bool
	}{
		{"valid password", "password123", true},
		{"valid with special", "Pass123!@#", true},
		{"too short", "pass", false},
		{"empty", "", false},
		{"with cyrillic", "пароль123", false},
		{"only numbers", "12345678", true},
		{"only letters", "password", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ValidatePassword(tt.password); got != tt.want {
				t.Errorf("ValidatePassword() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidateUsername(t *testing.T) {
	tests := []struct {
		name     string
		username string
		want     bool
	}{
		{"valid username", "testuser", true},
		{"valid with numbers", "test123", true},
		{"valid with underscore", "test_user", true},
		{"too short", "ab", false},
		{"too long", "verylongusernamethatistoolong", false},
		{"with special chars", "test-user", false},
		{"with spaces", "test user", false},
		{"empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ValidateUsername(tt.username); got != tt.want {
				t.Errorf("ValidateUsername() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidateName(t *testing.T) {
	tests := []struct {
		name     string
		namePart string
		want     bool
	}{
		{"valid name", "John", true},
		{"valid long name", "VeryLongNameButValid", true},
		{"empty", "", false},
		{"too long", "VeryLongNameThatExceedsLimit", false},
		{"single char", "A", true},
		{"20 chars", "12345678901234567890", true},
		{"21 chars", "123456789012345678901", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ValidateName(tt.namePart); got != tt.want {
				t.Errorf("ValidateName() = %v, want %v", got, tt.want)
			}
		})
	}
}
