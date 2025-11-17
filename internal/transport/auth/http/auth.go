package transport

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-park-mail-ru/2025_2_Undefined/config"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/models/domains"
	AuthDTO "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/dto/auth"
	gen "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/generated/auth"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/utils/cookie"
	utils "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/utils/response"
	"github.com/mssola/user_agent"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AuthGRPCProxyHandler struct {
	authClient    gen.AuthServiceClient
	sessionConfig *config.SessionConfig
}

func NewAuthGRPCProxyHandler(authClient gen.AuthServiceClient, sessionConfig *config.SessionConfig) *AuthGRPCProxyHandler {
	return &AuthGRPCProxyHandler{
		authClient:    authClient,
		sessionConfig: sessionConfig,
	}
}

// getDeviceFromUserAgent извлекает информацию об устройстве из User-Agent заголовка
func getDeviceFromUserAgent(r *http.Request) string {
	userAgent := r.Header.Get("User-Agent")
	if userAgent == "" {
		return "Unknown Device"
	}

	ua := user_agent.New(userAgent)
	name, version := ua.Browser()
	os := ua.OS()

	if name == "" && version == "" {
		return "Unknown Device"
	}
	if os == "" {
		return fmt.Sprintf("%s %s", name, version)
	}

	return fmt.Sprintf("%s %s on %s", name, version, os)
}

// Register регистрирует нового пользователя через gRPC
// @Summary      Регистрация пользователя
// @Description  Регистрирует нового пользователя в системе через gRPC микросервис и создает сессию
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        user  body  dto.RegisterRequest  true  "Данные для регистрации"
// @Success      201   {object}  dto.AuthResponse  "Пользователь успешно зарегистрирован"
// @Failure      400   {object}  dto.ValidationErrorsDTO  "Ошибки валидации"
// @Router       /register [post]
func (h *AuthGRPCProxyHandler) Register(w http.ResponseWriter, r *http.Request) {
	const op = "AuthGRPCProxyHandler.Register"
	logger := domains.GetLogger(r.Context()).WithField("op", op)

	var req AuthDTO.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.WithError(err).Error("failed to decode request")
		utils.SendError(r.Context(), op, w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Получаем информацию об устройстве из User-Agent
	device := getDeviceFromUserAgent(r)

	// Вызываем gRPC сервис
	res, err := h.authClient.Register(r.Context(), &gen.RegisterReq{
		PhoneNumber: req.PhoneNumber,
		Password:    req.Password,
		Name:        req.Name,
		Device:      device,
	})
	if err != nil {
		logger.WithError(err).Error("grpc register failed")
		handleGRPCError(r.Context(), op, w, err)
		return
	}

	// Устанавливаем cookie с session_id
	cookie.Set(w, res.SessionId, h.sessionConfig.Signature)

	// Возвращаем CSRF токен
	response := AuthDTO.AuthResponse{
		CSRFToken: res.CsrfToken,
	}

	utils.SendJSONResponse(r.Context(), op, w, http.StatusCreated, response)
}

// Login аутентифицирует пользователя через gRPC
// @Summary      Аутентификация пользователя
// @Description  Аутентифицирует пользователя по номеру телефона и паролю через gRPC микросервис, создает сессию
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        credentials  body  dto.LoginRequest  true  "Креденшиалы для входа"
// @Success      200  {object}  dto.AuthResponse  "Вход выполнен успешно"
// @Failure      400  {object}  dto.ValidationErrorsDTO  "Ошибки валидации"
// @Failure      401  {object}  dto.ErrorDTO  "Неверные креденшиалы"
// @Router       /login [post]
func (h *AuthGRPCProxyHandler) Login(w http.ResponseWriter, r *http.Request) {
	const op = "AuthGRPCProxyHandler.Login"
	logger := domains.GetLogger(r.Context()).WithField("op", op)

	var req AuthDTO.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.WithError(err).Error("failed to decode request")
		utils.SendError(r.Context(), op, w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Получаем информацию об устройстве из User-Agent
	device := getDeviceFromUserAgent(r)

	// Вызываем gRPC сервис
	res, err := h.authClient.Login(r.Context(), &gen.LoginReq{
		PhoneNumber: req.PhoneNumber,
		Password:    req.Password,
		Device:      device,
	})
	if err != nil {
		logger.WithError(err).Error("grpc login failed")
		handleGRPCError(r.Context(), op, w, err)
		return
	}

	// Устанавливаем cookie с session_id
	cookie.Set(w, res.SessionId, h.sessionConfig.Signature)

	// Возвращаем CSRF токен
	response := AuthDTO.AuthResponse{
		CSRFToken: res.CsrfToken,
	}

	utils.SendJSONResponse(r.Context(), op, w, http.StatusOK, response)
}

// Logout завершает сессию пользователя через gRPC
// @Summary      Выход из системы
// @Description  Аннулирует текущую сессию через gRPC микросервис и удаляет cookie
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param X-CSRF-Token header string true "CSRF Token"
// @Security     ApiKeyAuth
// @Success      200  "Logout successful"
// @Failure      401  {object}  dto.ErrorDTO  "Неавторизованный доступ"
// @Router       /logout [post]
func (h *AuthGRPCProxyHandler) Logout(w http.ResponseWriter, r *http.Request) {
	const op = "AuthGRPCProxyHandler.Logout"
	logger := domains.GetLogger(r.Context()).WithField("op", op)

	// Получаем ID сессии из cookie
	sessionCookie, err := r.Cookie(h.sessionConfig.Signature)
	if err != nil {
		logger.WithError(err).Error("session cookie not found")
		utils.SendError(r.Context(), op, w, http.StatusUnauthorized, "session not found")
		return
	}

	// Вызываем gRPC сервис
	_, err = h.authClient.Logout(r.Context(), &gen.LogoutReq{
		SessionId: sessionCookie.Value,
	})
	if err != nil {
		logger.WithError(err).Error("grpc logout failed")
		handleGRPCError(r.Context(), op, w, err)
		return
	}

	// Удаляем cookie
	cookie.Unset(w, h.sessionConfig.Signature)

	utils.SendJSONResponse(r.Context(), op, w, http.StatusOK, nil)
}

// handleGRPCError обрабатывает ошибки gRPC и преобразует их в HTTP статусы
func handleGRPCError(ctx context.Context, op string, w http.ResponseWriter, err error) {
	logger := domains.GetLogger(ctx).WithField("op", op)
	logger.WithError(err).Error("gRPC error occurred")

	if st, ok := status.FromError(err); ok {
		switch st.Code() {
		case codes.InvalidArgument:
			utils.SendError(ctx, op, w, http.StatusBadRequest, st.Message())
			return
		case codes.Unauthenticated:
			utils.SendError(ctx, op, w, http.StatusUnauthorized, st.Message())
			return
		case codes.NotFound:
			utils.SendError(ctx, op, w, http.StatusNotFound, st.Message())
			return
		case codes.AlreadyExists:
			utils.SendError(ctx, op, w, http.StatusConflict, st.Message())
			return
		case codes.PermissionDenied:
			utils.SendError(ctx, op, w, http.StatusForbidden, st.Message())
			return
		case codes.Internal:
			utils.SendError(ctx, op, w, http.StatusInternalServerError, st.Message())
			return
		}
	}

	utils.SendError(ctx, op, w, http.StatusBadRequest, err.Error())
}
