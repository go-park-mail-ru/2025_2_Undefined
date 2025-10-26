package transport

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-park-mail-ru/2025_2_Undefined/internal/models/errs"
	ContactDTO "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/dto/contact"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/utils/response"
	"github.com/google/uuid"
)

type SessionUtilsI interface {
	GetUserIDFromSession(r *http.Request) (uuid.UUID, error)
}

type ContactUsecase interface {
	CreateContact(req *ContactDTO.PostContactDTO, userID uuid.UUID) error
	GetContacts(userID uuid.UUID) ([]*ContactDTO.GetContactsDTO, error)
}

type ContactHandler struct {
	uc           ContactUsecase
	sessionUtils SessionUtilsI
}

func New(uc ContactUsecase, sessionUtils SessionUtilsI) *ContactHandler {
	return &ContactHandler{
		uc:           uc,
		sessionUtils: sessionUtils,
	}
}

// CreateContact создает новый контакт
// @Summary      Добавление контакта
// @Description  Добавляет нового пользователя в список контактов текущего пользователя
// @Tags         contacts
// @Accept       json
// @Produce      json
// @Param        contact  body  dto.PostContactDTO  true  "Данные контакта для добавления"
// @Success      201   "Контакт успешно добавлен"
// @Failure      400   {object}  dto.ErrorDTO  			  "Неверные данные запроса"
// @Failure      401   {object}  dto.ErrorDTO 			  "Неавторизованный доступ"
// @Failure      500   {object}  dto.ErrorDTO          	  "Внутренняя ошибка сервера"
// @Security     ApiKeyAuth
// @Router       /contacts [post]
func (h *ContactHandler) CreateContact(w http.ResponseWriter, r *http.Request) {
	userID, err := h.sessionUtils.GetUserIDFromSession(r)
	if err != nil {
		response.SendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req ContactDTO.PostContactDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.SendError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	if userID == req.ContactUserID {
		response.SendError(w, http.StatusBadRequest, "Cannot add yourself as contact")
		return
	}

	if err := h.uc.CreateContact(&req, userID); err != nil {
		if errors.Is(err, errs.ErrIsDuplicateKey) {
			response.SendError(w, http.StatusConflict, "contact already exists")
			return
		}
		response.SendError(w, http.StatusInternalServerError, "failed to create contact")
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// GetContacts получает список контактов пользователя
// @Summary      Получение списка контактов
// @Description  Возвращает список всех контактов текущего пользователя с полной информацией о пользователях
// @Tags         contacts
// @Accept       json
// @Produce      json
// @Success      200   {array}   dto.GetContactsDTO      "Список контактов успешно получен"
// @Failure      401   {object}  dto.ErrorDTO   		 "Неавторизованный доступ"
// @Failure      500   {object}  dto.ErrorDTO 			 "Внутренняя ошибка сервера"
// @Security     ApiKeyAuth
// @Router       /contacts [get]
func (h *ContactHandler) GetContacts(w http.ResponseWriter, r *http.Request) {
	userID, err := h.sessionUtils.GetUserIDFromSession(r)
	if err != nil {
		response.SendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	contacts, err := h.uc.GetContacts(userID)
	if err != nil {
		response.SendError(w, http.StatusInternalServerError, "Failed to get contacts")
		return
	}

	response.SendJSONResponse(w, http.StatusOK, contacts)
}
