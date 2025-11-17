package grpc

import (
	"context"

	"github.com/go-park-mail-ru/2025_2_Undefined/config"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/models/domains"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/models/errs"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/auth"
	AuthDTO "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/dto/auth"
	gen "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/generated/auth"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/utils/csrf"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/utils/validation"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type AuthGRPCHandler struct {
	gen.UnimplementedAuthServiceServer

	authUsecase    auth.IAuthUsecase
	sessionUsecase auth.ISessionUsecase
	csrfConfig     *config.CSRFConfig
}

func NewAuthGRPCHandler(uc auth.IAuthUsecase, sessionUC auth.ISessionUsecase, csrfConfig *config.CSRFConfig) *AuthGRPCHandler {
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
		logger.WithField("errors", validationErrors).Error("validation errors found")
		firstError := validationErrors[0]
		errorMsg := firstError.Field + ": " + firstError.Message
		return nil, status.Error(codes.InvalidArgument, errorMsg)
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
