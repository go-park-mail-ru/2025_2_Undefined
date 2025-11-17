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
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
// - ErrNoRights -> 403 Forbidden
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
		SendError(ctx, op, w, http.StatusNotFound, errs.ErrNotFound.Error())
		return
	}

	// 401 Unauthorized - неавторизованный доступ
	if errors.Is(err, errs.ErrInvalidToken) ||
		errors.Is(err, errs.ErrInvalidCredentials) ||
		errors.Is(err, errs.ErrJWTIsRequired) {
		SendError(ctx, op, w, http.StatusUnauthorized, err.Error())
		return
	}

	// 403 Forbidden - недостаточно прав
	if errors.Is(err, errs.ErrNoRights) {
		SendError(ctx, op, w, http.StatusForbidden, err.Error())
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

func HandleGRPCError(ctx context.Context, w http.ResponseWriter, err error, op string) {
	logger := domains.GetLogger(ctx)
	st, ok := status.FromError(err)
	if !ok {
		logger.WithError(err).Error(op + ": unexpected error type")
		SendJSONError(ctx, w, http.StatusInternalServerError, "internal server error")
		return
	}

	switch st.Code() {
	case codes.Unauthenticated:
		SendJSONError(ctx, w, http.StatusUnauthorized, st.Message())
	case codes.AlreadyExists:
		SendJSONError(ctx, w, http.StatusConflict, st.Message())
	case codes.NotFound:
		SendJSONError(ctx, w, http.StatusNotFound, st.Message())
	case codes.InvalidArgument:
		SendJSONError(ctx, w, http.StatusBadRequest, st.Message())
	default:
		logger.WithError(err).Error(op + ": unexpected gRPC status code")
		SendJSONError(ctx, w, http.StatusInternalServerError, "internal server error")
	}
}

func SendJSONError(ctx context.Context, w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	resp, err := json.Marshal(dto.ErrorDTO{Message: message})
	if err != nil {
		domains.GetLogger(ctx).Error("failed to marshal response: ", err.Error())
		return
	}

	if _, err := w.Write(resp); err != nil {
		domains.GetLogger(ctx).Error("failed to write response: ", err.Error())
	}
}
