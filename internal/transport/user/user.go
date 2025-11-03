package transport

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-park-mail-ru/2025_2_Undefined/internal/models/errs"
	UserDto "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/dto/user"
	interfaceSession "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/interface/session"
	interfaceUser "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/interface/user"
	utils "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/utils/response"

	"github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/utils/validation"
)

type UserHandler struct {
	uc             interfaceUser.UserUsecase
	sessionUsecase interfaceSession.SessionUsecase
}

func New(uc interfaceUser.UserUsecase, sessionUsecase interfaceSession.SessionUsecase) *UserHandler {
	return &UserHandler{
		uc:             uc,
		sessionUsecase: sessionUsecase,
	}
}

// GetCurrentUser получает информацию о текущем пользователе
// @Summary      Получить информацию о текущем пользователе
// @Description  Возвращает полные данные о текущем авторизованном пользователе
// @Tags         user
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Success      200  {object}  dto.User   "Информация о пользователе"
// @Failure      401  {object}  dto.ErrorDTO      "Неавторизованный доступ"
// @Router       /me [get]
func (h *UserHandler) GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	const op = "UserHandler.GetCurrentUser"

	userID, err := h.sessionUsecase.GetUserIDFromSession(r)
	if err != nil {
		utils.SendError(r.Context(), op, w, http.StatusUnauthorized, err.Error())
		return
	}

	user, err := h.uc.GetUserById(r.Context(), userID)
	if err != nil {
		utils.SendError(r.Context(), op, w, http.StatusUnauthorized, errs.ErrUserNotFound.Error())
		return
	}
	utils.SendJSONResponse(r.Context(), op, w, http.StatusOK, user)
}

// GetSessionsByUser получает все сессии текущего пользователя
// @Summary      Получить список сессий пользователя
// @Description  Возвращает все активные сессии текущего авторизованного пользователя
// @Tags         user
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Success      200  {array}   dto.Session  "Список сессий пользователя"
// @Failure      401  {object}  dto.ErrorDTO     "Неавторизованный доступ"
// @Failure      500  {object}  dto.ErrorDTO     "Внутренняя ошибка сервера"
// @Router       /sessions [get]
func (h *UserHandler) GetSessionsByUser(w http.ResponseWriter, r *http.Request) {
	const op = "UserHandler.GetSessionsByUser"

	userID, err := h.sessionUsecase.GetUserIDFromSession(r)
	if err != nil {
		utils.SendError(r.Context(), op, w, http.StatusUnauthorized, err.Error())
		return
	}

	sessions, err := h.sessionUsecase.GetSessionsByUserID(userID)
	if err != nil {
		utils.SendError(r.Context(), op, w, http.StatusUnauthorized, err.Error())
		return
	}
	utils.SendJSONResponse(r.Context(), op, w, http.StatusOK, sessions)
}

// GetUserByPhone получает информацию о пользователе по номеру телефона
// @Summary      Получить информацию о пользователе по номеру телефона
// @Description  Возвращает полные данные о пользователе по указанному номеру телефона
// @Tags         user
// @Accept       json
// @Produce      json
// @Param        request body dto.GetUserByPhone true "Запрос с номером телефона"
// @Success      200    {object}  dto.User   "Информация о пользователе"
// @Failure      400    {object}  dto.ErrorDTO   "Неверный формат номера телефона"
// @Failure      404    {object}  dto.ErrorDTO   "Пользователь не найден"
// @Failure      500    {object}  dto.ErrorDTO   "Внутренняя ошибка сервера"
// @Router       /user/by-phone [post]
func (h *UserHandler) GetUserByPhone(w http.ResponseWriter, r *http.Request) {
	const op = "UserHandler.GetUserByPhone"

	var req UserDto.GetUserByPhone
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.SendError(r.Context(), op, w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.PhoneNumber == "" {
		utils.SendError(r.Context(), op, w, http.StatusBadRequest, "Phone number is required")
		return
	}

	user, err := h.uc.GetUserByPhone(r.Context(), req.PhoneNumber)
	if err != nil {
		if errors.Is(err, errs.ErrUserNotFound) {
			utils.SendError(r.Context(), op, w, http.StatusNotFound, err.Error())
			return
		}
		utils.SendError(r.Context(), op, w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SendJSONResponse(r.Context(), op, w, http.StatusOK, user)
}

// GetUserByUsername получает информацию о пользователе по имени пользователя
// @Summary      Получить информацию о пользователе по имени пользователя
// @Description  Возвращает полные данные о пользователе по указанному имени пользователя
// @Tags         user
// @Accept       json
// @Produce      json
// @Param        request body dto.GetUserByUsername true "Запрос с именем пользователя"
// @Success      200       {object}  dto.User   "Информация о пользователе"
// @Failure      400       {object}  dto.ErrorDTO   "Неверный формат имени пользователя"
// @Failure      404       {object}  dto.ErrorDTO   "Пользователь не найден"
// @Failure      500       {object}  dto.ErrorDTO   "Внутренняя ошибка сервера"
// @Router       /user/by-username [post]
func (h *UserHandler) GetUserByUsername(w http.ResponseWriter, r *http.Request) {
	const op = "UserHandler.GetUserByUsername"

	var req UserDto.GetUserByUsername
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.SendError(r.Context(), op, w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Username == "" {
		utils.SendError(r.Context(), op, w, http.StatusBadRequest, "Username is required")
		return
	}

	user, err := h.uc.GetUserByUsername(r.Context(), req.Username)
	if err != nil {
		if errors.Is(err, errs.ErrUserNotFound) {
			utils.SendError(r.Context(), op, w, http.StatusNotFound, err.Error())
			return
		}
		utils.SendError(r.Context(), op, w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SendJSONResponse(r.Context(), op, w, http.StatusOK, user)
}

// UploadUserAvatar загружает аватар пользователя
// @Summary      Загрузить аватар пользователя
// @Description  Позволяет текущему авторизованному пользователю загрузить или обновить свой аватар
// @Tags         user
// @Accept       multipart/form-data
// @Produce      json
// @Security     ApiKeyAuth
// @Param        avatar  formData  file  true  "Файл аватара"
// @Success      200     {object}  map[string]string  "URL загруженного аватара"
// @Failure      400     {object}  dto.ErrorDTO      "Ошибка загрузки файла"
// @Failure      401     {object}  dto.ErrorDTO      "Неавторизованный доступ"
// @Failure      500     {object}  dto.ErrorDTO      "Внутренняя ошибка сервера"
// @Router       /user/avatar [post]
func (h *UserHandler) UploadUserAvatar(w http.ResponseWriter, r *http.Request) {
	const op = "UserHandler.UploadUserAvatar"

	userID, err := h.sessionUsecase.GetUserIDFromSession(r)
	if err != nil {
		utils.SendError(r.Context(), op, w, http.StatusUnauthorized, err.Error())
		return
	}

	file, header, err := r.FormFile("avatar")
	if err != nil {
		utils.SendError(r.Context(), op, w, http.StatusBadRequest, "Failed to read avatar file")
		return
	}

	defer file.Close()

	// Валидация типа файла на основе Content-Type
	contentType := header.Header.Get("Content-Type")
	if !validation.ValidImageType(contentType) {
		utils.SendError(r.Context(), op, w, http.StatusBadRequest, "Invalid file type. Only images (jpeg, png) are allowed")
		return
	}

	fileData := make([]byte, r.ContentLength)

	_, err = file.Read(fileData)
	if err != nil {
		utils.SendError(r.Context(), op, w, http.StatusInternalServerError, "Failed to read avatar file data")
		return
	}

	filename := header.Filename
	if filename == "" {
		filename = "avatar" + validation.GetFileExtensionFromContentType(contentType)
	}

	avatarURL, err := h.uc.UploadUserAvatar(r.Context(), userID, fileData, filename, contentType)
	if err != nil {
		utils.SendError(r.Context(), op, w, http.StatusBadRequest, "Failed to upload avatar")
		return
	}

	utils.SendJSONResponse(r.Context(), op, w, http.StatusOK, map[string]string{"avatar_url": avatarURL})
}
