package response

import (
	"context"
	"encoding/json"
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

// Функция отправки ошибки с проставлением статуса ответа в зависимости от сообщения.
// Сообщение сравнивается с ошибками из models/errs. Если нету совпадению, то возвращается
// http.StatusBadRequest
func SendErrorWithAutoStatus(ctx context.Context, op string, w http.ResponseWriter, message string) {
	if message == errs.ErrServiceIsOverloaded.Error() {
		SendError(ctx, op, w, http.StatusServiceUnavailable, message)
		return
	}

	SendError(ctx, op, w, http.StatusBadRequest, message)
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
