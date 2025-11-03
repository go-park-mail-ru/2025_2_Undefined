package response

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-park-mail-ru/2025_2_Undefined/internal/models/domains"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/models/errs"
	dto "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/dto/utils"
	"github.com/sirupsen/logrus"
)

func SendError(ctx context.Context, op string, w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	logger := domains.GetLogger(ctx).WithFields(logrus.Fields{
		"operation": op,
		"status":    status,
	})

	resp, err := json.Marshal(dto.ErrorDTO{Message: message})
	if err != nil {
		logger.Errorf("failed to marshal response: %s", err.Error())
		return
	}

	if _, err := w.Write(resp); err != nil {
		logger.Errorf("failed to write response: %s", err.Error())
		return
	}

	logger.WithField("error_message", message).Error("request failed")
}

// SendErrorWithAutoStatus автоматически определяет HTTP статус код на основе типа ошибки.
//
// Функция использует errors.Is() для проверки типа ошибки, что позволяет корректно
// обрабатывать wrapped ошибки (созданные через fmt.Errorf("context: %w", err)).
// Автоматически выбирает соответствующий HTTP статус код:
//
// - ErrServiceIsOverloaded -> 503 Service Unavailable
// - ErrNotFound, ErrUserNotFound -> 404 Not Found
// - ErrInvalidToken, ErrInvalidCredentials, ErrJWTIsRequired -> 401 Unauthorized
// - ErrIsDuplicateKey -> 409 Conflict
// - ErrRequiredFieldsMissing -> 422 Unprocessable Entity
// - Все остальные ошибки -> 400 Bad Request (по умолчанию)
//
// Преимущества errors.Is() над строковым сравнением:
// - Работает с wrapped ошибками
// - Не зависит от изменения текста ошибки
// - Более производительно и типобезопасно
func SendErrorWithAutoStatus(ctx context.Context, op string, w http.ResponseWriter, err error) {
	// 503 Service Unavailable - сервис недоступен или перегружен
	if errors.Is(err, errs.ErrServiceIsOverloaded) {
		SendError(ctx, op, w, http.StatusServiceUnavailable, err.Error())
		return
	}

	// 404 Not Found - ресурс не найден
	if errors.Is(err, errs.ErrNotFound) || errors.Is(err, errs.ErrUserNotFound) || errors.Is(err, sql.ErrNoRows) {
		SendError(ctx, op, w, http.StatusNotFound, err.Error())
		return
	}

	// 401 Unauthorized - неавторизованный доступ
	if errors.Is(err, errs.ErrInvalidToken) ||
		errors.Is(err, errs.ErrInvalidCredentials) ||
		errors.Is(err, errs.ErrJWTIsRequired) {
		SendError(ctx, op, w, http.StatusUnauthorized, err.Error())
		return
	}

	// 409 Conflict - конфликт данных (дубликат)
	if errors.Is(err, errs.ErrIsDuplicateKey) {
		SendError(ctx, op, w, http.StatusConflict, err.Error())
		return
	}

	// 422 Unprocessable Entity - отсутствуют обязательные поля
	if errors.Is(err, errs.ErrRequiredFieldsMissing) {
		SendError(ctx, op, w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	// По умолчанию - 400 Bad Request
	SendError(ctx, op, w, http.StatusBadRequest, err.Error())
}

func SendValidationErrors(ctx context.Context, op string, w http.ResponseWriter, status int, validationErrors dto.ValidationErrorsDTO) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	logger := domains.GetLogger(ctx).WithFields(logrus.Fields{
		"operation": op,
		"status":    status,
	})

	resp, err := json.Marshal(validationErrors)
	if err != nil {
		logger.Errorf("failed to marshal validation errors response: %s", err.Error())
		return
	}

	if _, err := w.Write(resp); err != nil {
		logger.Errorf("failed to write validation errors response: %s", err.Error())
		return
	}

	logger.WithField("validation_errors", validationErrors).Error("validation failed")
}

func SendJSONResponse(ctx context.Context, op string, w http.ResponseWriter, status int, body any) {
	if body == nil {
		w.WriteHeader(status)
		return
	}

	logger := domains.GetLogger(ctx).WithFields(logrus.Fields{
		"operation": op,
		"status":    status,
	})

	w.Header().Set("Content-Type", "application/json")
	resp, err := json.Marshal(body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logger.Errorf("failed to marshal response %s", err.Error())
		return
	}

	w.WriteHeader(status)
	if _, err := w.Write(resp); err != nil {
		logger.Errorf("failed to write response %s", err.Error())
	}
}
