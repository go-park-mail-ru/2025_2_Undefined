package transport

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/go-park-mail-ru/2025_2_Undefined/internal/models/domains"
	UserDTO "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/dto/user"
	dtoUtils "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/dto/utils"
	gen "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/generated/user"
	contextUtils "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/utils/context"
	grpcUtils "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/utils/grpc"
	utils "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/utils/response"
	"github.com/google/uuid"
)

type UserGRPCProxyHandler struct {
	userClient gen.UserServiceClient
}

func NewUserGRPCProxyHandler(userClient gen.UserServiceClient) *UserGRPCProxyHandler {
	return &UserGRPCProxyHandler{
		userClient: userClient,
	}
}

// GetCurrentUser получает информацию о текущем пользователе через gRPC
// @Summary      Получить информацию о текущем пользователе
// @Description  Возвращает полные данные о текущем авторизованном пользователе
// @Tags         user
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Success      200  {object}  dto.User   "Информация о пользователе"
// @Failure      401  {object}  dto.ErrorDTO      "Неавторизованный доступ"
// @Router       /me [get]
func (h *UserGRPCProxyHandler) GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	const op = "UserGRPCProxyHandler.GetCurrentUser"

	userID, err := contextUtils.GetUserIDFromContext(r)
	if err != nil {
		utils.SendError(r.Context(), op, w, http.StatusUnauthorized, err.Error())
		return
	}

	res, err := h.userClient.GetUserById(r.Context(), &gen.GetUserByIdReq{
		UserId: userID.String(),
	})
	if err != nil {
		utils.SendError(r.Context(), op, w, http.StatusNotFound, "user not found")
		return
	}

	user := mapProtoUserToDTO(res.User)
	utils.SendJSONResponse(r.Context(), op, w, http.StatusOK, user)
}

// GetUserByPhone получает информацию о пользователе по номеру телефона через gRPC
// @Summary      Получить информацию о пользователе по номеру телефона
// @Description  Возвращает полные данные о пользователе по указанному номеру телефона
// @Tags         user
// @Accept       json
// @Produce      json
// @Param X-CSRF-Token header string true "CSRF Token"
// @Security     ApiKeyAuth
// @Param        request body dto.GetUserByPhone  true  "Номер телефона"
// @Success      200  {object}  dto.User   "Информация о пользователе"
// @Failure      400  {object}  dto.ErrorDTO      "Неверный формат запроса"
// @Failure      401  {object}  dto.ErrorDTO      "Неавторизованный доступ"
// @Failure      404  {object}  dto.ErrorDTO      "Пользователь не найден"
// @Router       /user/by-phone [post]
func (h *UserGRPCProxyHandler) GetUserByPhone(w http.ResponseWriter, r *http.Request) {
	const op = "UserGRPCProxyHandler.GetUserByPhone"

	var req UserDTO.GetUserByPhone
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.SendError(r.Context(), op, w, http.StatusBadRequest, "Invalid request body")
		return
	}

	res, err := h.userClient.GetUserByPhone(r.Context(), &gen.GetUserByPhoneReq{
		PhoneNumber: req.PhoneNumber,
	})
	if err != nil {
		utils.SendError(r.Context(), op, w, http.StatusNotFound, "user not found")
		return
	}

	user := mapProtoUserToDTO(res.User)
	utils.SendJSONResponse(r.Context(), op, w, http.StatusOK, user)
}

// GetUserByUsername получает информацию о пользователе по username через gRPC
// @Summary      Получить информацию о пользователе по username
// @Description  Возвращает полные данные о пользователе по указанному username
// @Tags         user
// @Accept       json
// @Produce      json
// @Param X-CSRF-Token header string true "CSRF Token"
// @Security     ApiKeyAuth
// @Param        request body dto.GetUserByUsername  true  "Username пользователя"
// @Success      200  {object}  dto.User   "Информация о пользователе"
// @Failure      400  {object}  dto.ErrorDTO      "Неверный формат запроса"
// @Failure      401  {object}  dto.ErrorDTO      "Неавторизованный доступ"
// @Failure      404  {object}  dto.ErrorDTO      "Пользователь не найден"
// @Router       /user/by-username [post]
func (h *UserGRPCProxyHandler) GetUserByUsername(w http.ResponseWriter, r *http.Request) {
	const op = "UserGRPCProxyHandler.GetUserByUsername"

	var req UserDTO.GetUserByUsername
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.SendError(r.Context(), op, w, http.StatusBadRequest, "Invalid request body")
		return
	}

	res, err := h.userClient.GetUserByUsername(r.Context(), &gen.GetUserByUsernameReq{
		Username: req.Username,
	})
	if err != nil {
		utils.SendError(r.Context(), op, w, http.StatusNotFound, "user not found")
		return
	}

	user := mapProtoUserToDTO(res.User)
	utils.SendJSONResponse(r.Context(), op, w, http.StatusOK, user)
}

// UpdateUserInfo обновляет информацию о пользователе через gRPC
// @Summary      Обновить информацию о пользователе
// @Description  Обновляет имя, username или bio текущего пользователя
// @Tags         user
// @Accept       json
// @Produce      json
// @Param X-CSRF-Token header string true "CSRF Token"
// @Security     ApiKeyAuth
// @Param        user body dto.UpdateUserInfo  true  "Данные для обновления"
// @Success      200  "Информация успешно обновлена"
// @Failure      400  {object}  dto.ErrorDTO      "Неверный формат запроса"
// @Failure      401  {object}  dto.ErrorDTO      "Неавторизованный доступ"
// @Router       /me [patch]
func (h *UserGRPCProxyHandler) UpdateUserInfo(w http.ResponseWriter, r *http.Request) {
	const op = "UserGRPCProxyHandler.UpdateUserInfo"

	userID, err := contextUtils.GetUserIDFromContext(r)
	if err != nil {
		utils.SendError(r.Context(), op, w, http.StatusUnauthorized, err.Error())
		return
	}

	var req UserDTO.UpdateUserInfo
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.SendError(r.Context(), op, w, http.StatusBadRequest, "Invalid request body")
		return
	}

	grpcReq := &gen.UpdateUserInfoReq{
		UserId: userID.String(),
	}
	if req.Name != nil {
		grpcReq.Name = req.Name
	}
	if req.Username != nil {
		grpcReq.Username = req.Username
	}
	if req.Bio != nil {
		grpcReq.Bio = req.Bio
	}

	_, err = h.userClient.UpdateUserInfo(r.Context(), grpcReq)
	if err != nil {
		grpcUtils.HandleGRPCError(r.Context(), op, w, err)
		return
	}

	utils.SendJSONResponse(r.Context(), op, w, http.StatusOK, nil)
}

// UploadUserAvatar загружает аватар пользователя через gRPC
// @Summary      Загрузить аватар пользователя
// @Description  Загружает новый аватар для текущего пользователя
// @Tags         user
// @Accept       multipart/form-data
// @Produce      json
// @Param X-CSRF-Token header string true "CSRF Token"
// @Security     ApiKeyAuth
// @Param        avatar formData file true "Файл аватара"
// @Success      200  {object}  map[string]string  "URL загруженного аватара"
// @Failure      400  {object}  dto.ErrorDTO      "Неверный формат запроса"
// @Failure      401  {object}  dto.ErrorDTO      "Неавторизованный доступ"
// @Router       /user/avatar [post]
func (h *UserGRPCProxyHandler) UploadUserAvatar(w http.ResponseWriter, r *http.Request) {
	const op = "UserGRPCProxyHandler.UploadUserAvatar"
	logger := domains.GetLogger(r.Context()).WithField("op", op)

	userID, err := contextUtils.GetUserIDFromContext(r)
	if err != nil {
		utils.SendError(r.Context(), op, w, http.StatusUnauthorized, err.Error())
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

	res, err := h.userClient.UploadUserAvatar(r.Context(), &gen.UploadUserAvatarReq{
		UserId:      userID.String(),
		Data:        buf.Bytes(),
		Filename:    header.Filename,
		ContentType: header.Header.Get("Content-Type"),
	})
	if err != nil {
		logger.WithError(err).Error("failed to upload avatar")
		utils.SendError(r.Context(), op, w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SendJSONResponse(r.Context(), op, w, http.StatusOK, map[string]string{"avatar_url": res.AvatarUrl})
}

// GetUserAvatars получает аватарки нескольких пользователей
// @Summary      Получить аватарки пользователей
// @Description  Возвращает аватарки для списка пользователей по их ID
// @Tags         user
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        request  body      dto.GetAvatarsRequest   true  "Список ID пользователей"
// @Success      200      {object}  dto.GetAvatarsResponse  "Аватарки пользователей"
// @Failure      400      {object}  dto.ErrorDTO            "Некорректный запрос"
// @Failure      401      {object}  dto.ErrorDTO            "Неавторизованный доступ"
// @Router       /users/avatars/query [post]
func (h *UserGRPCProxyHandler) GetUserAvatars(w http.ResponseWriter, r *http.Request) {
	const op = "UserGRPCProxyHandler.GetUserAvatars"

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
			utils.SendError(r.Context(), op, w, http.StatusBadRequest, "invalid user id format: "+idStr)
			return
		}
	}

	request := &gen.GetUserAvatarsReq{UserIds: req.IDs}

	response, err := h.userClient.GetUserAvatars(r.Context(), request)
	if err != nil {
		utils.HandleGRPCError(r.Context(), w, err, op)
		return
	}

	utils.SendJSONResponse(r.Context(), op, w, http.StatusOK, dtoUtils.GetAvatarsResponse{Avatars: dtoUtils.StringMapToPointerMap(response.Avatars)})
}

func mapProtoUserToDTO(protoUser *gen.User) *UserDTO.User {
	createdAt, _ := time.Parse(time.RFC3339, protoUser.CreatedAt)
	updatedAt, _ := time.Parse(time.RFC3339, protoUser.UpdatedAt)

	var bio *string
	if protoUser.Bio != "" {
		bio = &protoUser.Bio
	}

	userID, _ := uuid.Parse(protoUser.GetId())

	return &UserDTO.User{
		ID:          userID,
		PhoneNumber: protoUser.PhoneNumber,
		Name:        protoUser.Name,
		Username:    protoUser.Username,
		Bio:         bio,
		AccountType: protoUser.AccountType,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
	}
}
