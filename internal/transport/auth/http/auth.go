package transport

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-park-mail-ru/2025_2_Undefined/config"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/models/domains"
	AuthDTO "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/dto/auth"
	gen "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/generated/auth"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/utils/cookie"
	grpcUtils "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/utils/grpc"
	utils "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/utils/response"
	"github.com/mssola/user_agent"
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

	device := getDeviceFromUserAgent(r)

	res, err := h.authClient.Register(r.Context(), &gen.RegisterReq{
		PhoneNumber: req.PhoneNumber,
		Password:    req.Password,
		Name:        req.Name,
		Device:      device,
	})
	if err != nil {
		logger.WithError(err).Error("grpc register failed")
		grpcUtils.HandleGRPCError(r.Context(), op, w, err)
		return
	}

	cookie.Set(w, res.SessionId, h.sessionConfig.Signature)

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

	device := getDeviceFromUserAgent(r)

	res, err := h.authClient.Login(r.Context(), &gen.LoginReq{
		PhoneNumber: req.PhoneNumber,
		Password:    req.Password,
		Device:      device,
	})
	if err != nil {
		logger.WithError(err).Error("grpc login failed")
		grpcUtils.HandleGRPCError(r.Context(), op, w, err)
		return
	}

	cookie.Set(w, res.SessionId, h.sessionConfig.Signature)

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

	sessionCookie, err := r.Cookie(h.sessionConfig.Signature)
	if err != nil {
		logger.WithError(err).Error("session cookie not found")
		utils.SendError(r.Context(), op, w, http.StatusUnauthorized, "session not found")
		return
	}

	_, err = h.authClient.Logout(r.Context(), &gen.LogoutReq{
		SessionId: sessionCookie.Value,
	})
	if err != nil {
		logger.WithError(err).Error("grpc logout failed")
		grpcUtils.HandleGRPCError(r.Context(), op, w, err)
		return
	}

	cookie.Unset(w, h.sessionConfig.Signature)

	utils.SendJSONResponse(r.Context(), op, w, http.StatusOK, nil)
}
