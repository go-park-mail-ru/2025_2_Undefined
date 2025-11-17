package grpc

import (
	"context"

	"github.com/go-park-mail-ru/2025_2_Undefined/config"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/models/domains"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/models/errs"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/auth"
	AuthDTO "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/dto/auth"
	sessionDTO "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/dto/session"
	gen "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/generated/auth"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/utils/csrf"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/utils/validation"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type SessionUsecase interface {
	GetSession(ctx context.Context, sessionID uuid.UUID) (*sessionDTO.Session, error)
	GetSessionsByUserID(ctx context.Context, userID uuid.UUID) ([]*sessionDTO.Session, error)
	UpdateSession(ctx context.Context, sessionID uuid.UUID) error
	DeleteSession(ctx context.Context, userID uuid.UUID, sessionID uuid.UUID) error
	DeleteAllSessionWithoutCurrent(ctx context.Context, userID uuid.UUID, currentSessionID uuid.UUID) error
}

type AuthGRPCHandler struct {
	gen.UnimplementedAuthServiceServer

	authUsecase    auth.IAuthUsecase
	sessionUsecase SessionUsecase
	csrfConfig     *config.CSRFConfig
}

func NewAuthGRPCHandler(uc auth.IAuthUsecase, sessionUC SessionUsecase, csrfConfig *config.CSRFConfig) *AuthGRPCHandler {
	return &AuthGRPCHandler{
		authUsecase:    uc,
		sessionUsecase: sessionUC,
		csrfConfig:     csrfConfig,
	}
}

func (h *AuthGRPCHandler) Register(ctx context.Context, in *gen.RegisterReq) (*gen.RegisterRes, error) {
	const op = "AuthGRPCHandler.Register"
	logger := domains.GetLogger(ctx).WithField("op", op)

	request := &AuthDTO.RegisterRequest{
		PhoneNumber: in.PhoneNumber,
		Password:    in.Password,
		Name:        in.Name,
	}

	// Валидация
	validationErrors := validation.ValidateRegisterRequest(request)
	if len(validationErrors) > 0 {
		logger.Error("validation errors found")
		return nil, status.Error(codes.InvalidArgument, "validation failed")
	}

	device := in.Device

	sessionID, validationErr := h.authUsecase.Register(ctx, request, device)
	if validationErr != nil {
		logger.WithError(errs.ErrBadRequest).Error("registration failed")
		// Проверяем, является ли ошибка конфликтом (пользователь уже существует)
		if len(validationErr.Errors) > 0 {
			for _, e := range validationErr.Errors {
				if e.Field == "phone_number" && e.Message == errs.ValidateUserAlreadyExists {
					return nil, status.Error(codes.AlreadyExists, errs.ValidateUserAlreadyExists)
				}
			}
		}
		return nil, status.Error(codes.InvalidArgument, validationErr.Message)
	}

	if sessionID == uuid.Nil {
		logger.Error("session ID is nil")
		return nil, status.Error(codes.Internal, "failed to create session")
	}

	// Генерируем CSRF токен
	csrfToken := csrf.GenerateCSRFToken(sessionID.String(), h.csrfConfig.Secret)

	return &gen.RegisterRes{SessionId: sessionID.String(), CsrfToken: csrfToken}, nil
}

func (h *AuthGRPCHandler) Login(ctx context.Context, in *gen.LoginReq) (*gen.LoginRes, error) {
	const op = "AuthGRPCHandler.Login"
	logger := domains.GetLogger(ctx).WithField("op", op)

	request := &AuthDTO.LoginRequest{
		PhoneNumber: in.PhoneNumber,
		Password:    in.Password,
	}

	// Валидация
	validationErrors := validation.ValidateLoginRequest(request)
	if len(validationErrors) > 0 {
		logger.Error("validation errors found")
		return nil, status.Error(codes.InvalidArgument, "validation failed")
	}

	device := in.Device

	sessionID, err := h.authUsecase.Login(ctx, request, device)
	if err != nil {
		logger.WithError(err).Error("login failed")
		return nil, status.Error(codes.Unauthenticated, errs.ErrInvalidCredentials.Error())
	}

	// Генерируем CSRF токен
	csrfToken := csrf.GenerateCSRFToken(sessionID.String(), h.csrfConfig.Secret)

	return &gen.LoginRes{SessionId: sessionID.String(), CsrfToken: csrfToken}, nil
}

func (h *AuthGRPCHandler) Logout(ctx context.Context, req *gen.LogoutReq) (*emptypb.Empty, error) {
	const op = "AuthGRPCHandler.Logout"
	logger := domains.GetLogger(ctx).WithField("op", op)

	sessionID, err := uuid.Parse(req.SessionId)
	if err != nil {
		logger.WithError(err).Error("invalid session ID")
		return nil, status.Error(codes.InvalidArgument, "invalid session ID")
	}

	if err := h.authUsecase.Logout(ctx, sessionID); err != nil {
		logger.WithError(err).Error("logout failed")
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &emptypb.Empty{}, nil
}

func (h *AuthGRPCHandler) ValidateSession(ctx context.Context, req *gen.ValidateSessionReq) (*gen.ValidateSessionRes, error) {
	const op = "AuthGRPCHandler.ValidateSession"
	logger := domains.GetLogger(ctx).WithField("op", op)

	sessionID, err := uuid.Parse(req.SessionId)
	if err != nil {
		logger.WithError(err).Error("invalid session ID")
		return &gen.ValidateSessionRes{Valid: false}, nil
	}

	session, err := h.sessionUsecase.GetSession(ctx, sessionID)
	if err != nil {
		logger.WithError(err).Debug("session not found or invalid")
		return &gen.ValidateSessionRes{Valid: false}, nil
	}

	// Обновляем время последней активности
	if err := h.sessionUsecase.UpdateSession(ctx, sessionID); err != nil {
		logger.WithError(err).Warn("failed to update session")
	}

	return &gen.ValidateSessionRes{
		Valid:  true,
		UserId: session.UserID.String(),
	}, nil
}

func (h *AuthGRPCHandler) GetSessionsByUserID(ctx context.Context, req *gen.GetSessionsByUserIDReq) (*gen.GetSessionsByUserIDRes, error) {
	const op = "AuthGRPCHandler.GetSessionsByUserID"
	logger := domains.GetLogger(ctx).WithField("op", op)

	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		logger.WithError(err).Error("invalid user ID")
		return nil, status.Error(codes.InvalidArgument, "invalid user ID")
	}

	sessions, err := h.sessionUsecase.GetSessionsByUserID(ctx, userID)
	if err != nil {
		logger.WithError(err).Error("failed to get sessions")
		return nil, status.Error(codes.Internal, err.Error())
	}

	var grpcSessions []*gen.Session
	for _, s := range sessions {
		grpcSessions = append(grpcSessions, &gen.Session{
			Id:        s.ID.String(),
			UserId:    s.UserID.String(),
			Device:    s.Device,
			CreatedAt: s.Created_at.Format("2006-01-02 15:04:05"),
			LastSeen:  s.Last_seen.Format("2006-01-02 15:04:05"),
		})
	}

	return &gen.GetSessionsByUserIDRes{Sessions: grpcSessions}, nil
}

func (h *AuthGRPCHandler) DeleteSession(ctx context.Context, req *gen.DeleteSessionReq) (*emptypb.Empty, error) {
	const op = "AuthGRPCHandler.DeleteSession"
	logger := domains.GetLogger(ctx).WithField("op", op)

	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		logger.WithError(err).Error("invalid user ID")
		return nil, status.Error(codes.InvalidArgument, "invalid user ID")
	}

	sessionID, err := uuid.Parse(req.SessionId)
	if err != nil {
		logger.WithError(err).Error("invalid session ID")
		return nil, status.Error(codes.InvalidArgument, "invalid session ID")
	}

	if err := h.sessionUsecase.DeleteSession(ctx, userID, sessionID); err != nil {
		logger.WithError(err).Error("failed to delete session")
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &emptypb.Empty{}, nil
}

func (h *AuthGRPCHandler) DeleteAllSessionsExceptCurrent(ctx context.Context, req *gen.DeleteAllSessionsExceptCurrentReq) (*emptypb.Empty, error) {
	const op = "AuthGRPCHandler.DeleteAllSessionsExceptCurrent"
	logger := domains.GetLogger(ctx).WithField("op", op)

	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		logger.WithError(err).Error("invalid user ID")
		return nil, status.Error(codes.InvalidArgument, "invalid user ID")
	}

	currentSessionID, err := uuid.Parse(req.CurrentSessionId)
	if err != nil {
		logger.WithError(err).Error("invalid current session ID")
		return nil, status.Error(codes.InvalidArgument, "invalid current session ID")
	}

	if err := h.sessionUsecase.DeleteAllSessionWithoutCurrent(ctx, userID, currentSessionID); err != nil {
		logger.WithError(err).Error("failed to delete sessions")
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &emptypb.Empty{}, nil
}
