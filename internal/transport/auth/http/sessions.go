package transport

import (
	"encoding/json"
	"net/http"

	"github.com/go-park-mail-ru/2025_2_Undefined/internal/models/domains"
	gen "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/generated/auth"
	grpcUtils "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/utils/grpc"
	utils "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/utils/response"
)

// GetSessionsByUser получает все сессии текущего пользователя через gRPC
// @Summary      Получить список сессий пользователя
// @Description  Возвращает все активные сессии текущего авторизованного пользователя
// @Tags         auth
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Success      200  {array}   dto.Session  "Список сессий пользователя"
// @Failure      401  {object}  dto.ErrorDTO     "Неавторизованный доступ"
// @Failure      500  {object}  dto.ErrorDTO     "Внутренняя ошибка сервера"
// @Router       /sessions [get]
func (h *AuthGRPCProxyHandler) GetSessionsByUser(w http.ResponseWriter, r *http.Request) {
	const op = "AuthGRPCProxyHandler.GetSessionsByUser"
	logger := domains.GetLogger(r.Context()).WithField("op", op)

	userIDVal := r.Context().Value(domains.UserIDKey{})
	if userIDVal == nil {
		utils.SendError(r.Context(), op, w, http.StatusUnauthorized, "user_id not found in context")
		return
	}

	userID, ok := userIDVal.(string)
	if !ok {
		utils.SendError(r.Context(), op, w, http.StatusUnauthorized, "invalid user_id in context")
		return
	}

	res, err := h.authClient.GetSessionsByUserID(r.Context(), &gen.GetSessionsByUserIDReq{
		UserId: userID,
	})
	if err != nil {
		logger.WithError(err).Error("grpc GetSessionsByUserID failed")
		grpcUtils.HandleGRPCError(r.Context(), op, w, err)
		return
	}

	utils.SendJSONResponse(r.Context(), op, w, http.StatusOK, res.Sessions)
}

// DeleteSession удаляет конкретную сессию пользователя через gRPC
// @Summary      удалить сессию пользователя
// @Description  удалить конкретную сессию пользователя
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param X-CSRF-Token header string true "CSRF Token"
// @Security     ApiKeyAuth
// @Param        session body dto.DeleteSession  true "Сессию которую надо удалить"
// @Success      200  "сессия удалена"
// @Failure      401  {object}  dto.ErrorDTO     "Неавторизованный доступ"
// @Failure      404  {object}  dto.ErrorDTO     "Не удалось удалить"
// @Router       /session [delete]
func (h *AuthGRPCProxyHandler) DeleteSession(w http.ResponseWriter, r *http.Request) {
	const op = "AuthGRPCProxyHandler.DeleteSession"
	logger := domains.GetLogger(r.Context()).WithField("op", op)

	userIDVal := r.Context().Value(domains.UserIDKey{})
	if userIDVal == nil {
		utils.SendError(r.Context(), op, w, http.StatusUnauthorized, "user_id not found in context")
		return
	}

	userID, ok := userIDVal.(string)
	if !ok {
		utils.SendError(r.Context(), op, w, http.StatusUnauthorized, "invalid user_id in context")
		return
	}

	var req struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.SendError(r.Context(), op, w, http.StatusBadRequest, "Invalid request body")
		return
	}

	_, err := h.authClient.DeleteSession(r.Context(), &gen.DeleteSessionReq{
		UserId:    userID,
		SessionId: req.ID,
	})
	if err != nil {
		logger.WithError(err).Error("grpc DeleteSession failed")
		grpcUtils.HandleGRPCError(r.Context(), op, w, err)
		return
	}

	utils.SendJSONResponse(r.Context(), op, w, http.StatusOK, nil)
}

// DeleteAllSessionsExceptCurrent удаляет все сессии пользователя кроме текущей через gRPC
// @Summary      удалить сессии пользователя, кроме текущей
// @Description  удалить сессии пользователя, кроме текущей
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param X-CSRF-Token header string true "CSRF Token"
// @Security     ApiKeyAuth
// @Success      200  "сессии удалены"
// @Failure      401  {object}  dto.ErrorDTO     "Неавторизованный доступ"
// @Failure      404  {object}  dto.ErrorDTO     "Не удалось удалить сессии"
// @Router       /sessions [delete]
func (h *AuthGRPCProxyHandler) DeleteAllSessionsExceptCurrent(w http.ResponseWriter, r *http.Request) {
	const op = "AuthGRPCProxyHandler.DeleteAllSessionsExceptCurrent"
	logger := domains.GetLogger(r.Context()).WithField("op", op)

	userIDVal := r.Context().Value(domains.UserIDKey{})
	if userIDVal == nil {
		utils.SendError(r.Context(), op, w, http.StatusUnauthorized, "user_id not found in context")
		return
	}

	userID, ok := userIDVal.(string)
	if !ok {
		utils.SendError(r.Context(), op, w, http.StatusUnauthorized, "invalid user_id in context")
		return
	}

	sessionCookie, err := r.Cookie(h.sessionConfig.Signature)
	if err != nil {
		logger.WithError(err).Error("session cookie not found")
		utils.SendError(r.Context(), op, w, http.StatusUnauthorized, "session not found")
		return
	}

	_, err = h.authClient.DeleteAllSessionsExceptCurrent(r.Context(), &gen.DeleteAllSessionsExceptCurrentReq{
		UserId:           userID,
		CurrentSessionId: sessionCookie.Value,
	})
	if err != nil {
		logger.WithError(err).Error("grpc DeleteAllSessionsExceptCurrent failed")
		grpcUtils.HandleGRPCError(r.Context(), op, w, err)
		return
	}

	utils.SendJSONResponse(r.Context(), op, w, http.StatusOK, nil)
}
