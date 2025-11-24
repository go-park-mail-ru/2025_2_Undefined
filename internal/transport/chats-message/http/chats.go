package chats

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/go-park-mail-ru/2025_2_Undefined/internal/models/domains"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/models/errs"
	mappers "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/chats-message/mappers"
	dtoChats "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/dto/chats"
	dtoUtils "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/dto/utils"
	gen "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/generated/chats"
	contextUtils "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/utils/context"
	utils "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/utils/response"
	"github.com/google/uuid"
)

type ChatsGRPCProxyHandler struct {
	chatsClient   gen.ChatServiceClient
	messageClient gen.MessageServiceClient
}

func NewChatsGRPCProxyHandler(chatsClient gen.ChatServiceClient, messageClient gen.MessageServiceClient) *ChatsGRPCProxyHandler {
	return &ChatsGRPCProxyHandler{
		chatsClient:   chatsClient,
		messageClient: messageClient,
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
func (h *ChatsGRPCProxyHandler) GetChats(w http.ResponseWriter, r *http.Request) {
	const op = "ChatsGRPCProxyHandler.GetChats"

	userUUID, err := contextUtils.GetUserIDFromContext(r)
	if err != nil {
		utils.SendError(r.Context(), op, w, http.StatusUnauthorized, err.Error())
		return
	}

	request := &gen.GetChatsReq{UserId: userUUID.String()}

	chats, err := h.chatsClient.GetChats(r.Context(), request)
	if err != nil {
		utils.HandleGRPCError(r.Context(), w, err, op)
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
// @Param X-CSRF-Token header string true "CSRF Token"
// @Security     ApiKeyAuth
// @Param        chat  body      dto.ChatCreateInformationDTO  true  "Данные для создания чата"
// @Success      201   {object}  dto.IdDTO                     "ID созданного чата"
// @Failure      400   {object}  dto.ErrorDTO                  "Некорректный запрос"
// @Failure      401   {object}  dto.ErrorDTO                  "Неавторизованный доступ"
// @Router       /chats [post]
func (h *ChatsGRPCProxyHandler) PostChats(w http.ResponseWriter, r *http.Request) {
	const op = "ChatsGRPCProxyHandler.PostChats"
	chatDTO := &dtoChats.ChatCreateInformationDTO{}

	if err := json.NewDecoder(r.Body).Decode(chatDTO); err != nil {
		utils.SendError(r.Context(), op, w, http.StatusBadRequest, err.Error())
		return
	}

	request := &gen.CreateChatReq{
		Name:    chatDTO.Name,
		Type:    chatDTO.Type,
		Members: mappers.DTOAddChatMembersToProto(chatDTO.Members),
	}

	response, err := h.chatsClient.CreateChat(r.Context(), request)
	if err != nil {
		utils.HandleGRPCError(r.Context(), w, err, op)
		return
	}

	idOfCreatedChat, err := uuid.Parse(response.GetId())
	if err != nil {
		utils.SendError(r.Context(), op, w, http.StatusInternalServerError, "invalid chat id from service")
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
func (h *ChatsGRPCProxyHandler) GetInformationAboutChat(w http.ResponseWriter, r *http.Request) {
	const op = "ChatsGRPCProxyHandler.GetInformationAboutChat"
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

	userID, err := contextUtils.GetUserIDFromContext(r)
	if err != nil {
		utils.SendError(r.Context(), op, w, http.StatusUnauthorized, err.Error())
		return
	}

	request := &gen.GetChatReq{
		ChatId: chatUUID.String(),
		UserId: userID.String(),
	}

	informationDTO, err := h.chatsClient.GetChat(r.Context(), request)
	if err != nil {
		utils.HandleGRPCError(r.Context(), w, err, op)
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
func (h *ChatsGRPCProxyHandler) GetUsersDialog(w http.ResponseWriter, r *http.Request) {
	const op = "ChatsGRPCProxyHandler.GetUsersDialog"

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

	userID, err := contextUtils.GetUserIDFromContext(r)
	if err != nil {
		utils.SendError(r.Context(), op, w, http.StatusUnauthorized, err.Error())
		return
	}

	request := &gen.GetUsersDialogReq{
		User1Id: userID.String(),
		User2Id: otherUserID.String(),
	}

	response, err := h.chatsClient.GetUsersDialog(r.Context(), request)
	if err != nil {
		utils.HandleGRPCError(r.Context(), w, err, op)
		return
	}

	dialogID, err := uuid.Parse(response.GetId())
	if err != nil {
		utils.SendError(r.Context(), op, w, http.StatusInternalServerError, "invalid dialog id from service")
		return
	}

	utils.SendJSONResponse(r.Context(), op, w, http.StatusOK, dtoUtils.IdDTO{ID: dialogID})
}

// AddUsersToChat добавляет пользователей в чат
// @Summary      Добавить пользователей в чат
// @Description  Добавляет указанных пользователей в существующий чат
// @Tags         chats
// @Accept       json
// @Produce      json
// @Param X-CSRF-Token header string true "CSRF Token"
// @Security     ApiKeyAuth
// @Param        chatId  path      string                     true  "ID чата"  format(uuid)
// @Param        users   body      dto.AddUsersToChatDTO      true  "Список пользователей для добавления"
// @Success      200
// @Failure      400     {object}  dto.ErrorDTO               "Некорректный запрос"
// @Failure      401     {object}  dto.ErrorDTO               "Неавторизованный доступ"
// @Router       /chats/{chatId}/members [patch]
func (h *ChatsGRPCProxyHandler) AddUsersToChat(w http.ResponseWriter, r *http.Request) {
	const op = "ChatsGRPCProxyHandler.AddUsersToChat"

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

	userID, err := contextUtils.GetUserIDFromContext(r)
	if err != nil {
		utils.SendError(r.Context(), op, w, http.StatusUnauthorized, err.Error())
		return
	}

	addUsersDTO := &dtoChats.AddUsersToChatDTO{}
	if err := json.NewDecoder(r.Body).Decode(addUsersDTO); err != nil {
		utils.SendErrorWithAutoStatus(r.Context(), op, w, err)
		return
	}

	request := &gen.AddUserToChatReq{
		ChatId:  chatUUID.String(),
		Members: mappers.DTOAddChatMembersToProto(addUsersDTO.Users),
		UserId:  userID.String(),
	}

	_, err = h.chatsClient.AddUserToChat(r.Context(), request)
	if err != nil {
		utils.HandleGRPCError(r.Context(), w, err, op)
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
// @Param X-CSRF-Token header string true "CSRF Token"
// @Security     ApiKeyAuth
// @Param        chatId  path      string  true  "ID чата"  format(uuid)
// @Success      200
// @Failure      400     {object}  dto.ErrorDTO  "Некорректный запрос"
// @Failure      401     {object}  dto.ErrorDTO  "Неавторизованный доступ"
// @Failure      403     {object}  dto.ErrorDTO  "Нет прав для удаления чата"
// @Failure      404     {object}  dto.ErrorDTO  "Чат не найден"
// @Router       /chats/{chatId} [delete]
func (h *ChatsGRPCProxyHandler) DeleteChat(w http.ResponseWriter, r *http.Request) {
	const op = "ChatsGRPCProxyHandler.DeleteChat"

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

	userID, err := contextUtils.GetUserIDFromContext(r)
	if err != nil {
		utils.SendError(r.Context(), op, w, http.StatusUnauthorized, err.Error())
		return
	}

	request := &gen.GetChatReq{
		ChatId: chatUUID.String(),
		UserId: userID.String(),
	}

	_, err = h.chatsClient.DeleteChat(r.Context(), request)
	if err != nil {
		utils.HandleGRPCError(r.Context(), w, err, op)
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
// @Param X-CSRF-Token header string true "CSRF Token"
// @Security     ApiKeyAuth
// @Param        chatId  path      string          true  "ID чата"  format(uuid)
// @Param        chat    body      dto.ChatUpdateDTO true "Поля для обновления чата"
// @Success      200
// @Failure      400     {object}  dto.ErrorDTO  "Некорректный запрос"
// @Failure      401     {object}  dto.ErrorDTO  "Неавторизованный доступ"
// @Failure      403     {object}  dto.ErrorDTO  "Нет прав для изменения чата"
// @Failure      404     {object}  dto.ErrorDTO  "Чат не найден"
// @Router       /chats/{chatId} [patch]
func (h *ChatsGRPCProxyHandler) UpdateChat(w http.ResponseWriter, r *http.Request) {
	const op = "ChatsGRPCProxyHandler.UpdateChat"

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

	userID, err := contextUtils.GetUserIDFromContext(r)
	if err != nil {
		utils.SendError(r.Context(), op, w, http.StatusUnauthorized, err.Error())
		return
	}

	updateDTO := &dtoChats.ChatUpdateDTO{}
	if err := json.NewDecoder(r.Body).Decode(updateDTO); err != nil {
		utils.SendError(r.Context(), op, w, http.StatusBadRequest, err.Error())
		return
	}

	var namePtr *string
	if updateDTO.Name != "" {
		namePtr = &updateDTO.Name
	}

	var descriptionPtr *string
	if updateDTO.Description != "" {
		descriptionPtr = &updateDTO.Description
	}

	request := &gen.UpdateChatReq{
		ChatId:      chatUUID.String(),
		Name:        namePtr,
		Description: descriptionPtr,
		UserId:      userID.String(),
	}

	_, err = h.chatsClient.UpdateChat(r.Context(), request)
	if err != nil {
		utils.HandleGRPCError(r.Context(), w, err, op)
		return
	}

	utils.SendJSONResponse(r.Context(), op, w, http.StatusOK, nil)
}

// GetChatAvatars получает аватарки нескольких чатов
// @Summary      Получить аватарки чатов
// @Description  Возвращает аватарки для списка чатов по их ID
// @Tags         chats
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        request  body      dto.GetAvatarsRequest   true  "Список ID чатов"
// @Success      200      {object}  dto.GetAvatarsResponse  "Аватарки чатов"
// @Failure      400      {object}  dto.ErrorDTO            "Некорректный запрос"
// @Failure      401      {object}  dto.ErrorDTO            "Неавторизованный доступ"
// @Router       /chats/avatars/query [post]
func (h *ChatsGRPCProxyHandler) GetChatAvatars(w http.ResponseWriter, r *http.Request) {
	const op = "ChatsGRPCProxyHandler.GetChatAvatars"

	var req dtoUtils.GetAvatarsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.SendError(r.Context(), op, w, http.StatusBadRequest, err.Error())
		return
	}

	if len(req.IDs) == 0 {
		utils.SendJSONResponse(r.Context(), op, w, http.StatusOK, dtoUtils.GetAvatarsResponse{Avatars: make(map[string]*string)})
		return
	}

	// Валидация UUID
	for _, idStr := range req.IDs {
		if _, err := uuid.Parse(idStr); err != nil {
			utils.SendError(r.Context(), op, w, http.StatusBadRequest, "invalid chat id format: "+idStr)
			return
		}
	}

	request := &gen.GetChatAvatarsReq{ChatIds: req.IDs}

	response, err := h.chatsClient.GetChatAvatars(r.Context(), request)
	if err != nil {
		utils.HandleGRPCError(r.Context(), w, err, op)
		return
	}

	utils.SendJSONResponse(r.Context(), op, w, http.StatusOK, dtoUtils.GetAvatarsResponse{Avatars: dtoUtils.StringMapToPointerMap(response.Avatars)})
}

// UploadChatAvatar загружает аватарку чата
// @Summary      Загрузить аватарку чата
// @Description  Загружает новый аватар для чата (только админ)
// @Tags         chats
// @Accept       multipart/form-data
// @Produce      json
// @Param X-CSRF-Token header string true "CSRF Token"
// @Security     ApiKeyAuth
// @Param        chatId path string true "ID чата" format(uuid)
// @Param        avatar formData file true "Файл аватара"
// @Success      200  {object}  map[string]string  "URL загруженного аватара"
// @Failure      400  {object}  dto.ErrorDTO      "Неверный формат запроса"
// @Failure      401  {object}  dto.ErrorDTO      "Неавторизованный доступ"
// @Failure      403  {object}  dto.ErrorDTO      "Нет прав для изменения чата"
// @Failure      404  {object}  dto.ErrorDTO      "Чат не найден"
// @Router       /chats/{chatId}/avatar [post]
func (h *ChatsGRPCProxyHandler) UploadChatAvatar(w http.ResponseWriter, r *http.Request) {
	const op = "ChatsGRPCProxyHandler.UploadChatAvatar"
	logger := domains.GetLogger(r.Context()).WithField("op", op)

	// Получить chatId из path
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 3 {
		utils.SendError(r.Context(), op, w, http.StatusBadRequest, "invalid chatId in path")
		return
	}
	chatIdStr := parts[len(parts)-2]
	chatId, err := uuid.Parse(chatIdStr)
	if err != nil {
		utils.SendError(r.Context(), op, w, http.StatusBadRequest, "invalid chatId")
		return
	}

	userID, err := contextUtils.GetUserIDFromContext(r)
	if err != nil {
		utils.SendError(r.Context(), op, w, http.StatusUnauthorized, err.Error())
		return
	}

	// Проверка, что пользователь — админ чата (через gRPC GetChat)
	chatInfo, err := h.chatsClient.GetChat(r.Context(), &gen.GetChatReq{
		ChatId: chatId.String(),
		UserId: userID.String(),
	})
	if err != nil {
		utils.HandleGRPCError(r.Context(), w, err, op)
		return
	}
	if !chatInfo.IsAdmin {
		utils.SendError(r.Context(), op, w, http.StatusForbidden, "only admin can upload chat avatar")
		return
	}

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		logger.WithError(err).Error("failed to parse multipart form")
		utils.SendError(r.Context(), op, w, http.StatusBadRequest, "failed to parse form")
		return
	}

	file, header, err := r.FormFile("avatar")
	if err != nil {
		logger.WithError(err).Error("failed to get avatar file")
		utils.SendError(r.Context(), op, w, http.StatusBadRequest, "avatar file is required")
		return
	}
	defer file.Close()

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, file); err != nil {
		logger.WithError(err).Error("failed to read file")
		utils.SendError(r.Context(), op, w, http.StatusInternalServerError, "failed to read file")
		return
	}

	res, err := h.chatsClient.UploadChatAvatar(r.Context(), &gen.UploadChatAvatarReq{
		UserId:      userID.String(),
		ChatId:      chatId.String(),
		Data:        buf.Bytes(),
		Filename:    header.Filename,
		ContentType: header.Header.Get("Content-Type"),
	})
	if err != nil {
		logger.WithError(err).Error("failed to upload chat avatar")
		utils.HandleGRPCError(r.Context(), w, err, op)
		return
	}

	utils.SendJSONResponse(r.Context(), op, w, http.StatusOK, map[string]string{"avatar_url": res.AvatarUrl})
}

// SearchChats ищет чаты по имени
// @Summary      Поиск чатов по имени
// @Description  Позволяет искать чаты по части или полному имени
// @Tags         chats
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        name  query     string  true  "Имя или часть имени чата для поиска"
// @Success      200   {array}   dto.ChatViewInformationDTO  "Список найденных чатов"
// @Failure      400   {object}  dto.ErrorDTO                "Некорректный запрос"
// @Failure      401   {object}  dto.ErrorDTO                "Неавторизованный доступ"
// @Router       /chats/search [get]
func (h *ChatsGRPCProxyHandler) SearchChats(w http.ResponseWriter, r *http.Request) {
	const op = "ChatsGRPCProxyHandler.SearchChats"

	queryValues := r.URL.Query()
	searchQuery := queryValues.Get("name")
	if searchQuery == "" {
		utils.SendError(r.Context(), op, w, http.StatusBadRequest, "name in query is required")
		return
	}

	userID, err := contextUtils.GetUserIDFromContext(r)
	if err != nil {
		utils.SendError(r.Context(), op, w, http.StatusUnauthorized, err.Error())
		return
	}

	request := &gen.SearchChatsReq{
		UserId: userID.String(),
		Name:   searchQuery,
	}

	response, err := h.chatsClient.SearchChats(r.Context(), request)
	if err != nil {
		utils.HandleGRPCError(r.Context(), w, err, op)
		return
	}

	dtoChats := mappers.ProtoSearchChatsResToDTO(response)
	utils.SendJSONResponse(r.Context(), op, w, http.StatusOK, dtoChats)
}
