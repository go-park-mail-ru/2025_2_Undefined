package appeal

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	dtoAppeal "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/dto/appeal"
	dtoUtils "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/dto/utils"
	appealInterface "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/interface/appeal"
	sessionInterface "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/interface/session"
	utils "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/utils/response"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type AppealHandler struct {
	appealUsecase appealInterface.AppealUsecase
	sessionUtils  sessionInterface.SessionUtils
}

func NewAppealHandler(appealUsecase appealInterface.AppealUsecase, sessionUtils sessionInterface.SessionUtils) *AppealHandler {
	return &AppealHandler{
		appealUsecase: appealUsecase,
		sessionUtils:  sessionUtils,
	}
}

// GetAppeals возвращает список обращений текущего пользователя
// @Summary      Получить список обращений текущего пользователя
// @Description  Возвращает обращения вместе с их сообщениями
// @Tags         appeals
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Success      200   {array}   dto.AppealDTO
// @Failure      401   {object}  dto.ErrorDTO  "Неавторизованный доступ"
// @Failure      500   {object}  dto.ErrorDTO  "Внутренняя ошибка сервера"
// @Router       /appeal/all [get]
func (h *AppealHandler) GetAppeals(w http.ResponseWriter, r *http.Request) {
	const op = "AppealHandler.GetAppeals"

	// Получаем id пользователя из сессии
	userID, err := h.sessionUtils.GetUserIDFromSession(r)
	if err != nil {
		utils.SendError(r.Context(), op, w, http.StatusUnauthorized, err.Error())
		return
	}

	appealsDTO, err := h.appealUsecase.GetAppeals(r.Context(), userID)
	if err != nil {
		utils.SendErrorWithAutoStatus(r.Context(), op, w, err)
		return
	}

	utils.SendJSONResponse(r.Context(), op, w, http.StatusOK, appealsDTO)
}

// GetAppealByID возвращает подробную информацию по одному обращению
// @Summary      Получить обращение по ID
// @Description  Возвращает обращение и его сообщения
// @Tags         appeals
// @Accept       json
// @Produce      json
// @Param        id   path     string  true  "Appeal ID"  Format(uuid)
// @Security     ApiKeyAuth
// @Success      200   {object}  dto.AppealDTO
// @Failure      400   {object}  dto.ErrorDTO  "Неверный id"
// @Failure      401   {object}  dto.ErrorDTO  "Неавторизованный доступ"
// @Failure      404   {object}  dto.ErrorDTO  "Обращение не найдено"
// @Router       /appeal/{id} [get]
func (h *AppealHandler) GetAppealByID(w http.ResponseWriter, r *http.Request) {
	const op = "AppealHandler.GetAppealByID"
	// Получаем id обращения из path
	vars := mux.Vars(r)
	idStr, ok := vars["id"]
	if !ok || idStr == "" {
		utils.SendError(r.Context(), op, w, http.StatusBadRequest, "id is required")
		return
	}

	appealID, err := uuid.Parse(idStr)
	if err != nil {
		utils.SendError(r.Context(), op, w, http.StatusBadRequest, "invalid id")
		return
	}

	// Получаем id пользователя из сессии
	userID, err := h.sessionUtils.GetUserIDFromSession(r)
	if err != nil {
		utils.SendError(r.Context(), op, w, http.StatusUnauthorized, err.Error())
		return
	}

	appealDTO, err := h.appealUsecase.GetAppealByID(r.Context(), userID, appealID)
	if err != nil {
		utils.SendErrorWithAutoStatus(r.Context(), op, w, err)
		return
	}

	utils.SendJSONResponse(r.Context(), op, w, http.StatusOK, appealDTO)
}

// PatchAppeal обновляет статус/категорию/заголовок обращения
// @Summary      Редактировать обращение
// @Description  Позволяет изменить статус, категорию или заголовок обращения
// @Tags         appeals
// @Accept       json
// @Produce      json
// @Param X-CSRF-Token header string true "CSRF Token"
// @Param        appeal  body  dto.EditAppealDTO  true  "Данные для редактирования"
// @Security     ApiKeyAuth
// @Success      200   {object}  dto.ErrorDTO  "OK"
// @Failure      400   {object}  dto.ErrorDTO  "Неверные данные"
// @Failure      401   {object}  dto.ErrorDTO  "Неавторизованный доступ"
// @Failure      403   {object}  dto.ErrorDTO  "Нет прав"
// @Router       /appeal [patch]
func (h *AppealHandler) PatchAppeal(w http.ResponseWriter, r *http.Request) {
	const op = "AppealHandler.PatchAppeal"

	editAppealDTO := &dtoAppeal.EditAppealDTO{}

	if err := json.NewDecoder(r.Body).Decode(editAppealDTO); err != nil {
		utils.SendError(r.Context(), op, w, http.StatusBadRequest, err.Error())
		return
	}

	if editAppealDTO.ID == uuid.Nil {
		utils.SendError(r.Context(), op, w, http.StatusBadRequest, "id is required")
		return
	}

	if editAppealDTO.Category == "" && editAppealDTO.Status == "" {
		utils.SendError(r.Context(), op, w, http.StatusBadRequest, "category or status is required")
		return
	}

	// Получаем id пользователя из сессии
	userID, err := h.sessionUtils.GetUserIDFromSession(r)
	if err != nil {
		utils.SendError(r.Context(), op, w, http.StatusUnauthorized, err.Error())
		return
	}

	err = h.appealUsecase.EditAppeal(r.Context(), userID, *editAppealDTO)
	if err != nil {
		utils.SendErrorWithAutoStatus(r.Context(), op, w, err)
		return
	}

	utils.SendJSONResponse(r.Context(), op, w, http.StatusOK, nil)
}

// PostAppeal создаёт новое обращение от авторизованного пользователя
// @Summary      Создать обращение
// @Description  Создаёт новое обращение для текущего пользователя
// @Tags         appeals
// @Accept       json
// @Produce      json
// @Param X-CSRF-Token header string true "CSRF Token"
// @Param        appeal  body  dto.CreateAppealDTO  true  "Данные обращения"
// @Security     ApiKeyAuth
// @Success      201   {object}  dto.IdDTO
// @Failure      400   {object}  dto.ErrorDTO
// @Failure      401   {object}  dto.ErrorDTO
// @Router       /appeal [post]
func (h *AppealHandler) PostAppeal(w http.ResponseWriter, r *http.Request) {
	const op = "AppealHandler.PostAppeal"

	createAppealDTO := &dtoAppeal.CreateAppealDTO{}
	if err := json.NewDecoder(r.Body).Decode(createAppealDTO); err != nil {
		utils.SendError(r.Context(), op, w, http.StatusBadRequest, err.Error())
		return
	}

	if createAppealDTO.Title == "" {
		utils.SendError(r.Context(), op, w, http.StatusBadRequest, "title is required")
		return
	}

	if createAppealDTO.Category == "" {
		utils.SendError(r.Context(), op, w, http.StatusBadRequest, "category is required")
		return
	}

	// Получаем id пользователя из сессии
	userID, err := h.sessionUtils.GetUserIDFromSession(r)
	if err != nil {
		utils.SendError(r.Context(), op, w, http.StatusUnauthorized, err.Error())
		return
	}

	createAppealDTO.UserID = userID

	appealID, err := h.appealUsecase.CreateAppeal(r.Context(), *createAppealDTO)
	if err != nil {
		utils.SendErrorWithAutoStatus(r.Context(), op, w, err)
		return
	}

	utils.SendJSONResponse(r.Context(), op, w, http.StatusCreated, dtoUtils.IdDTO{ID: appealID})
}

// PostAppealMessage добавляет сообщение в обращение от авторизованного пользователя
// @Summary      Добавить сообщение к обращению
// @Description  Добавляет новое сообщение в указанное обращение
// @Tags         appeals
// @Accept       json
// @Produce      json
// @Param X-CSRF-Token header string true "CSRF Token"
// @Param        message  body  dto.AppealCreateMessageDTO  true  "Данные сообщения"
// @Security     ApiKeyAuth
// @Success      201   {object}  dto.ErrorDTO
// @Failure      400   {object}  dto.ErrorDTO
// @Failure      401   {object}  dto.ErrorDTO
// @Failure      404   {object}  dto.ErrorDTO
// @Router       /appeal/message [post]
func (h *AppealHandler) PostAppealMessage(w http.ResponseWriter, r *http.Request) {
	const op = "AppealHandler.PostAppealMessage"
	messageDTO := &dtoAppeal.AppealCreateMessageDTO{}

	if err := json.NewDecoder(r.Body).Decode(messageDTO); err != nil {
		utils.SendError(r.Context(), op, w, http.StatusBadRequest, err.Error())
		return
	}

	if messageDTO.AppealID == uuid.Nil {
		utils.SendError(r.Context(), op, w, http.StatusBadRequest, "appeal_id is required")
		return
	}

	if messageDTO.Text == "" {
		utils.SendError(r.Context(), op, w, http.StatusBadRequest, "text is required")
		return
	}

	// Получаем id пользователя из сессии — отправитель сообщения
	userID, err := h.sessionUtils.GetUserIDFromSession(r)
	if err != nil {
		utils.SendError(r.Context(), op, w, http.StatusUnauthorized, err.Error())
		return
	}

	if err := h.appealUsecase.PostAppealMessage(r.Context(), userID, *messageDTO); err != nil {
		utils.SendErrorWithAutoStatus(r.Context(), op, w, err)
		return
	}

	utils.SendJSONResponse(r.Context(), op, w, http.StatusCreated, nil)
}

// GetAnonymousAppeals возвращает анонимные обращения по anonym_id (публичный роут)
// @Summary      Получить анонимные обращения
// @Description  Возвращает обращения анонимного пользователя по anonym_id
// @Tags         appeals
// @Accept       json
// @Produce      json
// @Param        anonym_id  query  string  true  "Anonym ID"  Format(uuid)
// @Success      200   {array}   dto.AnonymousAppealDTO
// @Failure      400   {object}  dto.ErrorDTO
// @Failure      500   {object}  dto.ErrorDTO
// @Router       /public/appeal [get]
// (public: нет Security)
func (h *AppealHandler) GetAnonymousAppeals(w http.ResponseWriter, r *http.Request) {
	const op = "AppealHandler.GetAnonymousAppeals"

	// Получаем anonym_id из query параметров
	anonymIDStr := r.URL.Query().Get("anonym_id")
	if anonymIDStr == "" {
		utils.SendError(r.Context(), op, w, http.StatusBadRequest, "anonym_id is required")
		return
	}

	anonymID, err := uuid.Parse(anonymIDStr)
	if err != nil {
		utils.SendError(r.Context(), op, w, http.StatusBadRequest, "invalid anonym_id")
		return
	}

	appealsDTO, err := h.appealUsecase.GetAnonymousAppeals(r.Context(), anonymID)
	if err != nil {
		utils.SendErrorWithAutoStatus(r.Context(), op, w, err)
		return
	}

	utils.SendJSONResponse(r.Context(), op, w, http.StatusOK, appealsDTO)
}

// GetAppealsForSupport возвращает обращения для техподдержки
// @Summary      Получить обращения для поддержки
// @Description  Возвращает обращения для роли поддержки текущего пользователя; админ видит все обращения
// @Tags         appeals
// @Accept       json
// @Produce      json
// @Param        limit   query    int  false  "limit"
// @Param        offset  query    int  false  "offset"
// @Security     ApiKeyAuth
// @Success      200   {array}   dto.AppealDTO
// @Failure      401   {object}  dto.ErrorDTO
// @Failure      500   {object}  dto.ErrorDTO
// @Router       /appeal/support [get]
func (h *AppealHandler) GetAppealsForSupport(w http.ResponseWriter, r *http.Request) {
	const op = "AppealHandler.GetAppealsForSupport"

	// Получаем id пользователя из сессии
	userID, err := h.sessionUtils.GetUserIDFromSession(r)
	if err != nil {
		utils.SendError(r.Context(), op, w, http.StatusUnauthorized, err.Error())
		return
	}

	// Параметры пагинации
	q := r.URL.Query()
	limit := 50
	offset := 0
	if l := q.Get("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 {
			limit = v
		}
	}
	if o := q.Get("offset"); o != "" {
		if v, err := strconv.Atoi(o); err == nil && v >= 0 {
			offset = v
		}
	}

	appealsDTO, err := h.appealUsecase.GetAppealsForSupport(r.Context(), userID, limit, offset)
	if err != nil {
		utils.SendErrorWithAutoStatus(r.Context(), op, w, err)
		return
	}

	utils.SendJSONResponse(r.Context(), op, w, http.StatusOK, appealsDTO)
}

// GetAppealsStats возвращает статистику обращений. Доступно только админу.
// Опционально можно передать query-параметры start_date и end_date в формате YYYY-MM-DD
// @Summary      Получить статистику обращений
// @Description  Возвращает агрегированную статистику по обращениям (по статусам, категориям, итоги). Доступно только админу.
// @Tags         appeals
// @Accept       json
// @Produce      json
// @Param        start_date  query  string  false  "Start date (YYYY-MM-DD)"
// @Param        end_date    query  string  false  "End date (YYYY-MM-DD)"
// @Security     ApiKeyAuth
// @Success      200   {object}  dto.AppealStatsDTO
// @Failure      400   {object}  dto.ErrorDTO  "Неверные параметры"
// @Failure      401   {object}  dto.ErrorDTO  "Неавторизованный доступ"
// @Failure      403   {object}  dto.ErrorDTO  "Нет прав"
// @Failure      500   {object}  dto.ErrorDTO  "Внутренняя ошибка сервера"
// @Router       /appeal/stats [get]
func (h *AppealHandler) GetAppealsStats(w http.ResponseWriter, r *http.Request) {
	const op = "AppealHandler.GetAppealsStats"

	// Получаем id пользователя из сессии
	adminID, err := h.sessionUtils.GetUserIDFromSession(r)
	if err != nil {
		utils.SendError(r.Context(), op, w, http.StatusUnauthorized, err.Error())
		return
	}

	q := r.URL.Query()
	var startDatePtr, endDatePtr *time.Time
	if s := q.Get("start_date"); s != "" {
		t, err := time.Parse("2006-01-02", s)
		if err != nil {
			utils.SendError(r.Context(), op, w, http.StatusBadRequest, "invalid start_date format, expected YYYY-MM-DD")
			return
		}
		startDatePtr = &t
	}
	if e := q.Get("end_date"); e != "" {
		t, err := time.Parse("2006-01-02", e)
		if err != nil {
			utils.SendError(r.Context(), op, w, http.StatusBadRequest, "invalid end_date format, expected YYYY-MM-DD")
			return
		}
		endDatePtr = &t
	}

	stats, err := h.appealUsecase.GetAppealsStats(r.Context(), adminID, startDatePtr, endDatePtr)
	if err != nil {
		utils.SendErrorWithAutoStatus(r.Context(), op, w, err)
		return
	}

	utils.SendJSONResponse(r.Context(), op, w, http.StatusOK, stats)
}

// ChangeUserRole изменяет роль указанного пользователя. Доступно только админу.
// @Summary      Изменить роль пользователя
// @Description  Позволяет администратору изменить роль указанного пользователя
// @Tags         appeals
// @Accept       json
// @Produce      json
// @Param        user_id  path  string  true  "User ID"  Format(uuid)
// @Param        role     body  dto.ChangeRoleDTO  true  "Новая роль"
// @Security     ApiKeyAuth
// @Success      200   {object}  dto.ErrorDTO  "OK"
// @Failure      400   {object}  dto.ErrorDTO  "Неверные данные"
// @Failure      401   {object}  dto.ErrorDTO  "Неавторизованный доступ"
// @Failure      403   {object}  dto.ErrorDTO  "Нет прав"
// @Failure      500   {object}  dto.ErrorDTO  "Внутренняя ошибка сервера"
// @Router       /appeal/role/{user_id} [patch]
func (h *AppealHandler) ChangeUserRole(w http.ResponseWriter, r *http.Request) {
	const op = "AppealHandler.ChangeUserRole"

	// Получаем id администратора из сессии
	adminID, err := h.sessionUtils.GetUserIDFromSession(r)
	if err != nil {
		utils.SendError(r.Context(), op, w, http.StatusUnauthorized, err.Error())
		return
	}

	// Получаем id целевого пользователя из path
	vars := mux.Vars(r)
	userIDStr, ok := vars["user_id"]
	if !ok || userIDStr == "" {
		utils.SendError(r.Context(), op, w, http.StatusBadRequest, "user_id is required")
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		utils.SendError(r.Context(), op, w, http.StatusBadRequest, "invalid user_id")
		return
	}

	dto := &dtoAppeal.ChangeRoleDTO{}
	if err := json.NewDecoder(r.Body).Decode(dto); err != nil {
		utils.SendError(r.Context(), op, w, http.StatusBadRequest, err.Error())
		return
	}

	if dto.Role == "" {
		utils.SendError(r.Context(), op, w, http.StatusBadRequest, "role is required")
		return
	}

	if err := h.appealUsecase.ChangeUserRole(r.Context(), adminID, userID, dto.Role); err != nil {
		utils.SendErrorWithAutoStatus(r.Context(), op, w, err)
		return
	}

	utils.SendJSONResponse(r.Context(), op, w, http.StatusOK, nil)
}

// DeleteUserRole удаляет роль у пользователя (возвращает к роли по умолчанию). Доступно только админу.
// @Summary      Удалить роль пользователя
// @Description  Позволяет администратору удалить роль пользователя (возврат к роли по умолчанию)
// @Tags         appeals
// @Accept       json
// @Produce      json
// @Param        user_id  path  string  true  "User ID"  Format(uuid)
// @Security     ApiKeyAuth
// @Success      200   {object}  dto.ErrorDTO  "OK"
// @Failure      400   {object}  dto.ErrorDTO  "Неверные данные"
// @Failure      401   {object}  dto.ErrorDTO  "Неавторизованный доступ"
// @Failure      403   {object}  dto.ErrorDTO  "Нет прав"
// @Failure      500   {object}  dto.ErrorDTO  "Внутренняя ошибка сервера"
// @Router       /appeal/role/{user_id} [delete]
func (h *AppealHandler) DeleteUserRole(w http.ResponseWriter, r *http.Request) {
	const op = "AppealHandler.DeleteUserRole"

	// Получаем id администратора из сессии
	adminID, err := h.sessionUtils.GetUserIDFromSession(r)
	if err != nil {
		utils.SendError(r.Context(), op, w, http.StatusUnauthorized, err.Error())
		return
	}

	// Получаем id целевого пользователя из path
	vars := mux.Vars(r)
	userIDStr, ok := vars["user_id"]
	if !ok || userIDStr == "" {
		utils.SendError(r.Context(), op, w, http.StatusBadRequest, "user_id is required")
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		utils.SendError(r.Context(), op, w, http.StatusBadRequest, "invalid user_id")
		return
	}

	if err := h.appealUsecase.DeleteUserRole(r.Context(), adminID, userID); err != nil {
		utils.SendErrorWithAutoStatus(r.Context(), op, w, err)
		return
	}

	utils.SendJSONResponse(r.Context(), op, w, http.StatusOK, nil)
}

// CreateAnonymousAppeal создает анонимное обращение
// @Summary      Создать анонимное обращение
// @Description  Создаёт новое анонимное обращение
// @Tags         appeals
// @Accept       json
// @Produce      json
// @Param        appeal  body  dto.CreateAnonymousAppealDTO  true  "Данные обращения"
// @Success      201   {object}  dto.IdDTO
// @Failure      400   {object}  dto.ErrorDTO
// @Failure      500   {object}  dto.ErrorDTO
// @Router       /public/appeal [post]
func (h *AppealHandler) CreateAnonymousAppeal(w http.ResponseWriter, r *http.Request) {
	const op = "AppealHandler.CreateAnonymousAppeal"

	appealDTO := &dtoAppeal.CreateAnonymousAppealDTO{}

	if err := json.NewDecoder(r.Body).Decode(appealDTO); err != nil {
		utils.SendError(r.Context(), op, w, http.StatusBadRequest, err.Error())
		return
	}

	if appealDTO.AnonymID == uuid.Nil {
		utils.SendError(r.Context(), op, w, http.StatusBadRequest, "anonym_id is required")
		return
	}

	if appealDTO.Title == "" {
		utils.SendError(r.Context(), op, w, http.StatusBadRequest, "title is required")
		return
	}

	if appealDTO.Category == "" {
		utils.SendError(r.Context(), op, w, http.StatusBadRequest, "category is required")
		return
	}

	appealID, err := h.appealUsecase.CreateAnonymousAppeal(r.Context(), *appealDTO)
	if err != nil {
		utils.SendErrorWithAutoStatus(r.Context(), op, w, err)
		return
	}

	response := dtoUtils.IdDTO{ID: appealID}
	utils.SendJSONResponse(r.Context(), op, w, http.StatusCreated, response)
}

// PostAnonymousMessage отправляет сообщение от анонимного пользователя
// @Summary      Добавить анонимное сообщение к обращению
// @Description  Добавляет новое сообщение от анонимного пользователя в указанное обращение
// @Tags         appeals
// @Accept       json
// @Produce      json
// @Param        message  body  dto.CreateAnonymousMessageDTO  true  "Данные сообщения"
// @Success      201   {object}  dto.ErrorDTO
// @Failure      400   {object}  dto.ErrorDTO
// @Failure      404   {object}  dto.ErrorDTO
// @Failure      500   {object}  dto.ErrorDTO
// @Router       /public/appeal/message [post]
func (h *AppealHandler) PostAnonymousMessage(w http.ResponseWriter, r *http.Request) {
	const op = "AppealHandler.PostAnonymousMessage"

	messageDTO := &dtoAppeal.CreateAnonymousMessageDTO{}

	if err := json.NewDecoder(r.Body).Decode(messageDTO); err != nil {
		utils.SendError(r.Context(), op, w, http.StatusBadRequest, err.Error())
		return
	}

	if messageDTO.AppealID == uuid.Nil {
		utils.SendError(r.Context(), op, w, http.StatusBadRequest, "appeal_id is required")
		return
	}

	if messageDTO.AnonymID == uuid.Nil {
		utils.SendError(r.Context(), op, w, http.StatusBadRequest, "anonym_id is required")
		return
	}

	if messageDTO.Text == "" {
		utils.SendError(r.Context(), op, w, http.StatusBadRequest, "text is required")
		return
	}

	if err := h.appealUsecase.PostAnonymousMessage(r.Context(), *messageDTO); err != nil {
		utils.SendErrorWithAutoStatus(r.Context(), op, w, err)
		return
	}

	utils.SendJSONResponse(r.Context(), op, w, http.StatusCreated, nil)
}
