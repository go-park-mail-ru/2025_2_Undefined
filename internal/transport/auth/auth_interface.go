package auth

import (
	"context"

	AuthDTO "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/dto/auth"
	dto "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/dto/utils"
	"github.com/google/uuid"
)

//go:generate mockgen -source=auth_interface.go -destination=../../usecase/mocks/mock_auth_usecase_mock.go -package=mocks IAuthUsecase
type IAuthUsecase interface {
	Register(ctx context.Context, req *AuthDTO.RegisterRequest, device string) (uuid.UUID, *dto.ValidationErrorsDTO)
	Login(ctx context.Context, req *AuthDTO.LoginRequest, device string) (uuid.UUID, error)
	Logout(ctx context.Context, SessionID uuid.UUID) error
}
