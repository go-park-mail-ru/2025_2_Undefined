package grpc

import (
	"context"
	"errors"
	"time"

	"github.com/go-park-mail-ru/2025_2_Undefined/internal/models/domains"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/models/errs"
	gen "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/generated/user"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/user-contact"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/utils/validation"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type UserGRPCHandler struct {
	gen.UnimplementedUserServiceServer

	userUC    user.IUserUsecase
	contactUC user.IContactUsecase
}

func NewUserGRPCHandler(userUC user.IUserUsecase, contactUC user.IContactUsecase) *UserGRPCHandler {
	return &UserGRPCHandler{
		userUC:    userUC,
		contactUC: contactUC,
	}
}

func (h *UserGRPCHandler) GetUserById(ctx context.Context, req *gen.GetUserByIdReq) (*gen.GetUserByIdRes, error) {
	const op = "UserGRPCHandler.GetUserById"
	logger := domains.GetLogger(ctx).WithField("op", op)

	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		logger.WithError(err).Error("invalid user ID")
		return nil, status.Error(codes.InvalidArgument, "invalid user ID")
	}

	user, err := h.userUC.GetUserById(ctx, userID)
	if err != nil {
		logger.WithError(err).Error("failed to get user")
		return nil, status.Error(codes.NotFound, "user not found")
	}

	bio := ""
	if user.Bio != nil {
		bio = *user.Bio
	}

	return &gen.GetUserByIdRes{
		User: &gen.User{
			Id:          user.ID.String(),
			PhoneNumber: user.PhoneNumber,
			Name:        user.Name,
			Username:    user.Username,
			Bio:         bio,
			AvatarUrl:   user.AvatarURL,
			AccountType: user.AccountType,
			CreatedAt:   user.CreatedAt.Format(time.RFC3339),
			UpdatedAt:   user.UpdatedAt.Format(time.RFC3339),
		},
	}, nil
}

func (h *UserGRPCHandler) GetUserByPhone(ctx context.Context, req *gen.GetUserByPhoneReq) (*gen.GetUserByPhoneRes, error) {
	const op = "UserGRPCHandler.GetUserByPhone"
	logger := domains.GetLogger(ctx).WithField("op", op)

	user, err := h.userUC.GetUserByPhone(ctx, req.PhoneNumber)
	if err != nil {
		logger.WithError(err).Error("failed to get user")
		return nil, status.Error(codes.NotFound, "user not found")
	}

	bio := ""
	if user.Bio != nil {
		bio = *user.Bio
	}

	return &gen.GetUserByPhoneRes{
		User: &gen.User{
			Id:          user.ID.String(),
			PhoneNumber: user.PhoneNumber,
			Name:        user.Name,
			Username:    user.Username,
			Bio:         bio,
			AvatarUrl:   user.AvatarURL,
			AccountType: user.AccountType,
			CreatedAt:   user.CreatedAt.Format(time.RFC3339),
			UpdatedAt:   user.UpdatedAt.Format(time.RFC3339),
		},
	}, nil
}

func (h *UserGRPCHandler) GetUserByUsername(ctx context.Context, req *gen.GetUserByUsernameReq) (*gen.GetUserByUsernameRes, error) {
	const op = "UserGRPCHandler.GetUserByUsername"
	logger := domains.GetLogger(ctx).WithField("op", op)

	user, err := h.userUC.GetUserByUsername(ctx, req.Username)
	if err != nil {
		logger.WithError(err).Error("failed to get user")
		return nil, status.Error(codes.NotFound, "user not found")
	}

	bio := ""
	if user.Bio != nil {
		bio = *user.Bio
	}

	return &gen.GetUserByUsernameRes{
		User: &gen.User{
			Id:          user.ID.String(),
			PhoneNumber: user.PhoneNumber,
			Name:        user.Name,
			Username:    user.Username,
			Bio:         bio,
			AvatarUrl:   user.AvatarURL,
			AccountType: user.AccountType,
			CreatedAt:   user.CreatedAt.Format(time.RFC3339),
			UpdatedAt:   user.UpdatedAt.Format(time.RFC3339),
		},
	}, nil
}

func (h *UserGRPCHandler) UpdateUserInfo(ctx context.Context, req *gen.UpdateUserInfoReq) (*emptypb.Empty, error) {
	const op = "UserGRPCHandler.UpdateUserInfo"
	logger := domains.GetLogger(ctx).WithField("op", op)

	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		logger.WithError(err).Error("invalid user ID")
		return nil, status.Error(codes.InvalidArgument, "invalid user ID")
	}

	var name, username, bio *string
	if req.Name != nil {
		if !validation.ValidateName(*req.Name) {
			return nil, status.Error(codes.InvalidArgument, "name must be 1-20 characters")
		}
		name = req.Name
	}
	if req.Username != nil {
		if !validation.ValidateUsername(*req.Username) {
			return nil, status.Error(codes.InvalidArgument, "username must be 3-20 characters and contain only Latin letters, digits, and underscores")
		}
		username = req.Username
	}
	if req.Bio != nil {
		if len(*req.Bio) > 200 {
			return nil, status.Error(codes.InvalidArgument, "bio must not exceed 200 characters")
		}
		bio = req.Bio
	}

	if err := h.userUC.UpdateUserInfo(ctx, userID, name, username, bio); err != nil {
		logger.WithError(err).Error("failed to update user info")

		switch {
		case errors.Is(err, errs.ErrIsDuplicateKey):
			return nil, status.Error(codes.AlreadyExists, "username already exist")
		case errors.Is(err, errs.ErrUserNotFound), errors.Is(err, errs.ErrNotFound):
			return nil, status.Error(codes.NotFound, "user not found")
		case errors.Is(err, errs.ErrBadRequest):
			return nil, status.Error(codes.InvalidArgument, err.Error())
		default:
			return nil, status.Error(codes.Internal, "failed to update user info")
		}
	}

	return &emptypb.Empty{}, nil
}

func (h *UserGRPCHandler) UploadUserAvatar(ctx context.Context, req *gen.UploadUserAvatarReq) (*gen.UploadUserAvatarRes, error) {
	const op = "UserGRPCHandler.UploadUserAvatar"
	logger := domains.GetLogger(ctx).WithField("op", op)

	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		logger.WithError(err).Error("invalid user ID")
		return nil, status.Error(codes.InvalidArgument, "invalid user ID")
	}

	avatarURL, err := h.userUC.UploadUserAvatar(ctx, userID, req.Data, req.Filename, req.ContentType)
	if err != nil {
		logger.WithError(err).Error("failed to upload avatar")
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &gen.UploadUserAvatarRes{
		AvatarUrl: avatarURL,
	}, nil
}
