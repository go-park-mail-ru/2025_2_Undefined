package transport

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/go-park-mail-ru/2025_2_Undefined/config"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/models/errs"
	SessionDto "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/dto/session"
	UserDto "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/dto/user"
	gen "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/generated/auth"
	interfaceUser "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/interface/user"
	contextUtils "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/utils/context"
	utils "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/utils/response"
	"github.com/google/uuid"

	"github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/utils/validation"
)

type UserHandler struct {
	uc            interfaceUser.UserUsecase
	authClient    gen.AuthServiceClient
	sessionConfig *config.SessionConfig
}

func New(uc interfaceUser.UserUsecase, authClient gen.AuthServiceClient, sessionConfig *config.SessionConfig) *UserHandler {
	return &UserHandler{
		uc:            uc,
		authClient:    authClient,
		sessionConfig: sessionConfig,
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

	userID, err := contextUtils.GetUserIDFromContext(r)
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

	userID, err := contextUtils.GetUserIDFromContext(r)
	if err != nil {
		utils.SendError(r.Context(), op, w, http.StatusUnauthorized, err.Error())
		return
	}

	// Вызываем gRPC для получения сессий
	res, err := h.authClient.GetSessionsByUserID(r.Context(), &gen.GetSessionsByUserIDReq{
		UserId: userID.String(),
	})
	if err != nil {
		utils.SendError(r.Context(), op, w, http.StatusInternalServerError, err.Error())
		return
	}

	// Преобразуем в DTO
	var sessions []*SessionDto.Session
	for _, s := range res.Sessions {
		sessionID, _ := uuid.Parse(s.Id)
		userID, _ := uuid.Parse(s.UserId)
		createdAt, _ := time.Parse("2006-01-02 15:04:05", s.CreatedAt)
		lastSeen, _ := time.Parse("2006-01-02 15:04:05", s.LastSeen)

		sessions = append(sessions, &SessionDto.Session{
			ID:         sessionID,
			UserID:     userID,
			Device:     s.Device,
			Created_at: createdAt,
			Last_seen:  lastSeen,
		})
	}

	utils.SendJSONResponse(r.Context(), op, w, http.StatusOK, sessions)
}

// GetSessionsByUser удалить сессию пользователя
// @Summary      удалить сессию пользователя
// @Description  удалить конкретную сессию пользователя
// @Tags         user
// @Accept       json
// @Produce      json
// @Param X-CSRF-Token header string true "CSRF Token"
// @Security     ApiKeyAuth
// @Param        session body dto.DeleteSession  true "Сессию которую надо удалить"
// @Success      200  "сессия удалена"
// @Failure      401  {object}  dto.ErrorDTO     "Неавторизованный доступ"
// @Failure      404  {object}  dto.ErrorDTO     "Не удалось удалить"
// @Router       /session [delete]
func (h *UserHandler) DeleteSession(w http.ResponseWriter, r *http.Request) {
	const op = "UserHandler.DeleteSession"

	var req SessionDto.DeleteSession
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.SendError(r.Context(), op, w, http.StatusBadRequest, "Invalid request body")
		return
	}

	userID, err := contextUtils.GetUserIDFromContext(r)
	if err != nil {
		utils.SendError(r.Context(), op, w, http.StatusUnauthorized, err.Error())
		return
	}

	// Вызываем gRPC для удаления сессии
	_, err = h.authClient.DeleteSession(r.Context(), &gen.DeleteSessionReq{
		UserId:    userID.String(),
		SessionId: req.ID.String(),
	})
	if err != nil {
		utils.SendError(r.Context(), op, w, http.StatusNotFound, err.Error())
		return
	}

	utils.SendJSONResponse(r.Context(), op, w, http.StatusOK, nil)
}

// GetSessionsByUser удалить сессии пользователя, кроме текущей
// @Summary      удалить сессии пользователя, кроме текущей
// @Description  удалить сессии пользователя, кроме текущей
// @Tags         user
// @Accept       json
// @Produce      json
// @Param X-CSRF-Token header string true "CSRF Token"
// @Security     ApiKeyAuth
// @Success      200  "сессии удалены"
// @Failure      401  {object}  dto.ErrorDTO     "Неавторизованный доступ"
// @Failure      404  {object}  dto.ErrorDTO     "Не удалось удалить сессии"
// @Router       /sessions [delete]
func (h *UserHandler) DeleteAllSessionWithoutCurrent(w http.ResponseWriter, r *http.Request) {
	const op = "UserHandler.DeleteAllSessionWithoutCurrent"

	userID, err := contextUtils.GetUserIDFromContext(r)
	if err != nil {
		utils.SendError(r.Context(), op, w, http.StatusUnauthorized, err.Error())
		return
	}

	sessionID, err := contextUtils.GetSessionIDFromCookie(r, h.sessionConfig.Signature)
	if err != nil {
		utils.SendError(r.Context(), op, w, http.StatusUnauthorized, err.Error())
		return
	}

	// Вызываем gRPC для удаления всех сессий кроме текущей
	_, err = h.authClient.DeleteAllSessionsExceptCurrent(r.Context(), &gen.DeleteAllSessionsExceptCurrentReq{
		UserId:           userID.String(),
		CurrentSessionId: sessionID.String(),
	})
	if err != nil {
		utils.SendError(r.Context(), op, w, http.StatusNotFound, err.Error())
		return
	}

	utils.SendJSONResponse(r.Context(), op, w, http.StatusOK, nil)
}

// GetUserByPhone получает информацию о пользователе по номеру телефона
// @Summary      Получить информацию о пользователе по номеру телефона
// @Description  Возвращает полные данные о пользователе по указанному номеру телефона
// @Tags         user
// @Accept       json
// @Produce      json
// @Param X-CSRF-Token header string true "CSRF Token"
// @Security     ApiKeyAuth
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
// @Param X-CSRF-Token header string true "CSRF Token"
// @Security     ApiKeyAuth
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
// @Param X-CSRF-Token header string true "CSRF Token"
// @Security     ApiKeyAuth
// @Param        avatar  formData  file  true  "Файл аватара"
// @Success      200     {object}  map[string]string  "URL загруженного аватара"
// @Failure      400     {object}  dto.ErrorDTO      "Ошибка загрузки файла"
// @Failure      401     {object}  dto.ErrorDTO      "Неавторизованный доступ"
// @Failure      500     {object}  dto.ErrorDTO      "Внутренняя ошибка сервера"
// @Router       /user/avatar [post]
func (h *UserHandler) UploadUserAvatar(w http.ResponseWriter, r *http.Request) {
	const op = "UserHandler.UploadUserAvatar"

	userID, err := contextUtils.GetUserIDFromContext(r)
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

// UpdateUserInfo обновляет информацию о пользователе
// @Summary      Обновить информацию пользователя
// @Description  Частичное обновление информации пользователя. Можно обновить только нужные поля.
// @Tags         user
// @Accept       json
// @Produce      json
// @Param X-CSRF-Token header string true "CSRF Token"
// @Param        request body dto.UpdateUserInfo true "Данные для обновления"
// @Success      200  "Информация пользователя обновилось"
// @Failure      400  {object}  dto.ErrorDTO
// @Failure      401  {object}  dto.ErrorDTO
// @Failure      403  {object}  dto.ErrorDTO
// @Failure      500  {object}  dto.ErrorDTO
// @Router       /me [patch]
func (h *UserHandler) UpdateUserInfo(w http.ResponseWriter, r *http.Request) {
	const op = "UserHandler.UpdateUserInfo"

	var req UserDto.UpdateUserInfo
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.SendError(r.Context(), op, w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Name == nil && req.Username == nil && req.Bio == nil {
		utils.SendError(r.Context(), op, w, http.StatusBadRequest, "At least one field must be provided for update")
		return
	}

	userID, err := contextUtils.GetUserIDFromContext(r)
	if err != nil {
		utils.SendError(r.Context(), op, w, http.StatusUnauthorized, err.Error())
		return
	}

	if req.Username != nil {
		if !validation.ValidateUsername(*req.Username) {
			utils.SendError(r.Context(), op, w, http.StatusBadRequest, "Invalid username format")
			return
		}
	}

	if req.Name != nil {
		if !validation.ValidateName(*req.Name) {
			utils.SendError(r.Context(), op, w, http.StatusBadRequest, "Invalid name format")
			return
		}
	}

	err = h.uc.UpdateUserInfo(r.Context(), userID, req.Name, req.Username, req.Bio)
	if err != nil {
		utils.SendError(r.Context(), op, w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SendJSONResponse(r.Context(), op, w, http.StatusOK, nil)
}
