package transport

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/go-park-mail-ru/2025_2_Undefined/internal/models/errs"
	dtoChats "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/dto/chats"
	dtoUtils "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/dto/utils"
	chatsInterface "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/interface/chats"
	sessionInterface "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/interface/session"
	utils "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/utils/response"
	"github.com/google/uuid"
)

type ChatsHandler struct {
	chatUsecase    chatsInterface.ChatsUsecase
	sessionUsecase sessionInterface.SessionUsecase
}

func NewChatsHandler(chatUsecase chatsInterface.ChatsUsecase, sessionUsecase sessionInterface.SessionUsecase) *ChatsHandler {
	return &ChatsHandler{
		chatUsecase:    chatUsecase,
		sessionUsecase: sessionUsecase,
	}
}

// GetChats получает список всех чатов пользователя
// @Summary      Получить список чатов
// @Description  Возвращает список всех чатов текущего пользователя с информацией о последнем сообщении
// @Tags         chats
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Success      200  {array}   dto.ChatViewInformationDTO  "Список чатов"
// @Failure      400  {object}  dto.ErrorDTO                "Некорректный запрос"
// @Failure      401  {object}  dto.ErrorDTO                "Неавторизованный доступ"
// @Router       /chats [get]
func (h *ChatsHandler) GetChats(w http.ResponseWriter, r *http.Request) {
	const op = "ChatsHandler.GetChats"
	// Получаем id пользователя из сессии
	userUUID, err := h.sessionUsecase.GetUserIDFromSession(r)
	if err != nil {
		utils.SendError(r.Context(), op, w, http.StatusUnauthorized, err.Error())
		return
	}

	chats, err := h.chatUsecase.GetChats(r.Context(), userUUID)
	if err != nil {
		utils.SendError(r.Context(), op, w, http.StatusBadRequest, err.Error())
		return
	}
	utils.SendJSONResponse(r.Context(), op, w, http.StatusOK, chats)
}

// PostChats создает новый чат
// @Summary      Создать новый чат
// @Description  Создает новый чат с указанными участниками и настройками
// @Tags         chats
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        chat  body      dto.ChatCreateInformationDTO  true  "Данные для создания чата"
// @Success      201   {object}  dto.IdDTO                     "ID созданного чата"
// @Failure      400   {object}  dto.ErrorDTO                  "Некорректный запрос"
// @Failure      401   {object}  dto.ErrorDTO                  "Неавторизованный доступ"
// @Router       /chats [post]
func (h *ChatsHandler) PostChats(w http.ResponseWriter, r *http.Request) {
	const op = "ChatsHandler.PostChats"
	chatDTO := &dtoChats.ChatCreateInformationDTO{}

	if err := json.NewDecoder(r.Body).Decode(chatDTO); err != nil {
		utils.SendError(r.Context(), op, w, http.StatusBadRequest, err.Error())
		return
	}

	if strings.TrimSpace(chatDTO.Name) == "" {
		utils.SendError(r.Context(), op, w, http.StatusBadRequest, "name is required and cannot be empty")
		return
	}
	if strings.TrimSpace(chatDTO.Type) == "" {
		utils.SendError(r.Context(), op, w, http.StatusBadRequest, "type is required and cannot be empty")
		return
	}
	if len(chatDTO.Members) == 0 {
		utils.SendError(r.Context(), op, w, http.StatusBadRequest, "members field is required and cannot be empty")
		return
	}

	for i, member := range chatDTO.Members {
		if member.UserId == uuid.Nil {
			utils.SendError(r.Context(), op, w, http.StatusBadRequest, "user_id is required for all members")
			return
		}

		if strings.TrimSpace(member.Role) == "" {
			utils.SendError(r.Context(), op, w, http.StatusBadRequest, "role is required for all members")
			return
		}

		if member.Role != "admin" && member.Role != "writer" && member.Role != "viewer" {
			utils.SendError(r.Context(), op, w, http.StatusBadRequest, "role must be one of: admin, writer, viewer")
			return
		}
		chatDTO.Members[i].Role = strings.TrimSpace(member.Role)
	}

	// Проверка на дубликаты пользователей в создаваемом чате
	memberIds := make(map[uuid.UUID]bool)
	for _, member := range chatDTO.Members {
		if memberIds[member.UserId] {
			utils.SendError(r.Context(), op, w, http.StatusBadRequest, "duplicate user_id found in members")
			return
		}
		memberIds[member.UserId] = true
	}

	// Обрезаем пробелы в основных полях
	chatDTO.Type = strings.TrimSpace(chatDTO.Type)

	idOfCreatedChat, err := h.chatUsecase.CreateChat(r.Context(), *chatDTO)
	if err != nil {
		utils.SendError(r.Context(), op, w, http.StatusBadRequest, err.Error())
		return
	}

	utils.SendJSONResponse(r.Context(), op, w, http.StatusCreated, dtoUtils.IdDTO{ID: idOfCreatedChat})
}

// GetInformationAboutChat получает детальную информацию о чате
// @Summary      Получить информацию о чате
// @Description  Возвращает детальную информацию о конкретном чате, включая сообщения и участников
// @Tags         chats
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        chatId  path      string  true  "ID чата"  format(uuid)
// @Success      200     {object}  dto.ChatDetailedInformationDTO  "Детальная информация о чате"
// @Failure      400     {object}  dto.ErrorDTO                    "Некорректный запрос"
// @Failure      401     {object}  dto.ErrorDTO                    "Неавторизованный доступ"
// @Failure      404     {object}  dto.ErrorDTO                    "Не существует такого чата"
// @Router       /chats/{chatId} [get]
func (h *ChatsHandler) GetInformationAboutChat(w http.ResponseWriter, r *http.Request) {
	const op = "ChatsHandler.GetInformationAboutChat"
	// Получаем id чата из пути
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 2 {
		utils.SendError(r.Context(), op, w, http.StatusBadRequest, errs.ErrBadRequest.Error())
		return
	}

	idStr := parts[len(parts)-1]
	chatUUID, err := uuid.Parse(idStr)
	if err != nil {
		utils.SendError(r.Context(), op, w, http.StatusBadRequest, errs.ErrBadRequest.Error())
		return
	}

	// Получаем id пользователя из сессии
	userUUID, err := h.sessionUsecase.GetUserIDFromSession(r)
	if err != nil {
		utils.SendError(r.Context(), op, w, http.StatusUnauthorized, err.Error())
		return
	}

	informationDTO, err := h.chatUsecase.GetInformationAboutChat(r.Context(), userUUID, chatUUID)
	if err != nil {
		utils.SendErrorWithAutoStatus(r.Context(), op, w, err)
		return
	}

	utils.SendJSONResponse(r.Context(), op, w, http.StatusOK, informationDTO)
}

// GetUsersDialog получает личный чат (диалог) между текущим пользователем и другим пользователем
// @Summary      Получить личный диалог с пользователем
// @Description  Возвращает информацию о личном чате между авторизованным пользователем и указанным пользователем. Если чата нет — возвращается ошибку.
// @Tags         chats
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        otherUserId  path      string  true  "ID другого пользователя"  format(uuid)
// @Success      200  {object}  dto.IdDTO "Информация о личном чате"
// @Failure      400  {object}  dto.ErrorDTO                "Некорректный ID пользователя"
// @Failure      401  {object}  dto.ErrorDTO                "Неавторизованный доступ"
// @Failure      404  {object}  dto.ErrorDTO                "Пользователь не найден"
// @Router       /chats/dialog/{otherUserId} [get]
func (h *ChatsHandler) GetUsersDialog(w http.ResponseWriter, r *http.Request) {
	const op = "ChatsHandler.GetUsersDialog"

	// Получаем id пользователя из пути
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 3 {
		utils.SendError(r.Context(), op, w, http.StatusBadRequest, errs.ErrBadRequest.Error())
		return
	}

	IDStr := parts[len(parts)-1]
	otherUserID, err := uuid.Parse(IDStr)
	if err != nil {
		utils.SendError(r.Context(), op, w, http.StatusBadRequest, errs.ErrBadRequest.Error())
		return
	}

	// Получаем id пользователя из сессии
	userID, err := h.sessionUsecase.GetUserIDFromSession(r)
	if err != nil {
		utils.SendError(r.Context(), op, w, http.StatusUnauthorized, err.Error())
		return
	}

	DTO, err := h.chatUsecase.GetUsersDialog(r.Context(), userID, otherUserID)
	if err != nil {
		utils.SendErrorWithAutoStatus(r.Context(), op, w, err)
		return
	}

	utils.SendJSONResponse(r.Context(), op, w, http.StatusOK, DTO)
}

// AddUsersToChat добавляет пользователей в чат
// @Summary      Добавить пользователей в чат
// @Description  Добавляет указанных пользователей в существующий чат
// @Tags         chats
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        chatId  path      string                     true  "ID чата"  format(uuid)
// @Param        users   body      dto.AddUsersToChatDTO      true  "Список пользователей для добавления"
// @Success      200
// @Failure      400     {object}  dto.ErrorDTO               "Некорректный запрос"
// @Failure      401     {object}  dto.ErrorDTO               "Неавторизованный доступ"
// @Router       /chats/{chatId}/members [patch]
func (h *ChatsHandler) AddUsersToChat(w http.ResponseWriter, r *http.Request) {
	const op = "ChatsHandler.AddUsersToChat"

	// Получаем id чата из пути
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 3 {
		utils.SendErrorWithAutoStatus(r.Context(), op, w, errs.ErrBadRequest)
		return
	}

	idStr := parts[len(parts)-2]
	chatUUID, err := uuid.Parse(idStr)
	if err != nil {
		utils.SendErrorWithAutoStatus(r.Context(), op, w, err)
		return
	}

	// Получаем id пользователя из сессии
	userUUID, err := h.sessionUsecase.GetUserIDFromSession(r)
	if err != nil {
		utils.SendError(r.Context(), op, w, http.StatusUnauthorized, err.Error())
		return
	}

	addUsersDTO := &dtoChats.AddUsersToChatDTO{}
	if err := json.NewDecoder(r.Body).Decode(addUsersDTO); err != nil {
		utils.SendErrorWithAutoStatus(r.Context(), op, w, err)
		return
	}

	if len(addUsersDTO.Users) == 0 {
		utils.SendError(r.Context(), op, w, http.StatusBadRequest, "users field is required and cannot be empty")
		return
	}

	for i, user := range addUsersDTO.Users {
		if user.UserId == uuid.Nil {
			utils.SendError(r.Context(), op, w, http.StatusBadRequest, "user_id is required for all users")
			return
		}

		if strings.TrimSpace(user.Role) == "" {
			utils.SendError(r.Context(), op, w, http.StatusBadRequest, "role is required for all users")
			return
		}

		if user.Role != "admin" && user.Role != "writer" && user.Role != "viewer" {
			utils.SendError(r.Context(), op, w, http.StatusBadRequest, "role must be one of: admin, writer, viewer")
			return
		}
		addUsersDTO.Users[i].Role = strings.TrimSpace(user.Role)
	}

	// Проверка на дубликаты пользователей
	userIds := make(map[uuid.UUID]bool)
	for _, user := range addUsersDTO.Users {
		if userIds[user.UserId] {
			utils.SendError(r.Context(), op, w, http.StatusBadRequest, "duplicate user_id found in request")
			return
		}
		userIds[user.UserId] = true
	}

	err = h.chatUsecase.AddUsersToChat(r.Context(), chatUUID, userUUID, addUsersDTO.Users)
	if err != nil {
		utils.SendErrorWithAutoStatus(r.Context(), op, w, err)
		return
	}

	utils.SendJSONResponse(r.Context(), op, w, http.StatusOK, nil)
}

// DeleteChat удаляет чат
// @Summary      Удалить чат
// @Description  Удаляет существующий чат. Только администратор чата может удалить чат.
// @Tags         chats
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        chatId  path      string  true  "ID чата"  format(uuid)
// @Success      200
// @Failure      400     {object}  dto.ErrorDTO  "Некорректный запрос"
// @Failure      401     {object}  dto.ErrorDTO  "Неавторизованный доступ"
// @Failure      403     {object}  dto.ErrorDTO  "Нет прав для удаления чата"
// @Failure      404     {object}  dto.ErrorDTO  "Чат не найден"
// @Router       /chats/{chatId} [delete]
func (h *ChatsHandler) DeleteChat(w http.ResponseWriter, r *http.Request) {
	const op = "ChatsHandler.DeleteChat"

	// Получаем id чата из пути
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 2 {
		utils.SendError(r.Context(), op, w, http.StatusBadRequest, errs.ErrBadRequest.Error())
		return
	}

	idStr := parts[len(parts)-1]
	chatUUID, err := uuid.Parse(idStr)
	if err != nil {
		utils.SendError(r.Context(), op, w, http.StatusBadRequest, errs.ErrBadRequest.Error())
		return
	}

	userUUID, err := h.sessionUsecase.GetUserIDFromSession(r)
	if err != nil {
		utils.SendError(r.Context(), op, w, http.StatusUnauthorized, err.Error())
		return
	}

	err = h.chatUsecase.DeleteChat(r.Context(), userUUID, chatUUID)
	if err != nil {
		if strings.Contains(err.Error(), "user is not admin") {
			utils.SendError(r.Context(), op, w, http.StatusForbidden, "user must have role admin to delete chat")
			return
		}

		if errors.Is(err, sql.ErrNoRows) {
			utils.SendError(r.Context(), op, w, http.StatusNotFound, "chat not found")
			return
		}

		utils.SendError(r.Context(), op, w, http.StatusBadRequest, err.Error())
		return
	}

	utils.SendJSONResponse(r.Context(), op, w, http.StatusOK, nil)
}

// UpdateChat изменяет название и описание чата (тип менять нельзя)
// @Summary      Обновить чат
// @Description  Позволяет администратору изменить название и описание чата
// @Tags         chats
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        chatId  path      string          true  "ID чата"  format(uuid)
// @Param        chat    body      dto.ChatUpdateDTO true "Поля для обновления чата"
// @Success      200
// @Failure      400     {object}  dto.ErrorDTO  "Некорректный запрос"
// @Failure      401     {object}  dto.ErrorDTO  "Неавторизованный доступ"
// @Failure      403     {object}  dto.ErrorDTO  "Нет прав для изменения чата"
// @Failure      404     {object}  dto.ErrorDTO  "Чат не найден"
// @Router       /chats/{chatId} [patch]
func (h *ChatsHandler) UpdateChat(w http.ResponseWriter, r *http.Request) {
	const op = "ChatsHandler.UpdateChat"

	// Получаем id чата из пути
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 2 {
		utils.SendError(r.Context(), op, w, http.StatusBadRequest, errs.ErrBadRequest.Error())
		return
	}

	idStr := parts[len(parts)-1]
	chatUUID, err := uuid.Parse(idStr)
	if err != nil {
		utils.SendError(r.Context(), op, w, http.StatusBadRequest, errs.ErrBadRequest.Error())
		return
	}

	// Получаем id пользователя из сессии
	userUUID, err := h.sessionUsecase.GetUserIDFromSession(r)
	if err != nil {
		utils.SendError(r.Context(), op, w, http.StatusUnauthorized, err.Error())
		return
	}

	updateDTO := &dtoChats.ChatUpdateDTO{}
	if err := json.NewDecoder(r.Body).Decode(updateDTO); err != nil {
		utils.SendError(r.Context(), op, w, http.StatusBadRequest, err.Error())
		return
	}

	if strings.TrimSpace(updateDTO.Name) == "" {
		utils.SendError(r.Context(), op, w, http.StatusBadRequest, "name is required and cannot be empty")
		return
	}

	err = h.chatUsecase.UpdateChat(r.Context(), userUUID, chatUUID, updateDTO.Name, updateDTO.Description)
	if err != nil {
		if strings.Contains(err.Error(), "user is not admin") {
			utils.SendError(r.Context(), op, w, http.StatusForbidden, "user must have role admin to update chat")
			return
		}

		if errors.Is(err, sql.ErrNoRows) {
			utils.SendError(r.Context(), op, w, http.StatusNotFound, "chat not found")
			return
		}

		utils.SendError(r.Context(), op, w, http.StatusBadRequest, err.Error())
		return
	}

	utils.SendJSONResponse(r.Context(), op, w, http.StatusOK, nil)
}
