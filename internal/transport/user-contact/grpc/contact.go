package grpc

import (
	"context"
	"errors"
	"time"

	"github.com/go-park-mail-ru/2025_2_Undefined/internal/models/domains"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/models/errs"
	ContactDTO "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/dto/contact"
	gen "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/generated/user"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (h *UserGRPCHandler) CreateContact(ctx context.Context, req *gen.CreateContactReq) (*emptypb.Empty, error) {
	const op = "UserGRPCHandler.CreateContact"
	logger := domains.GetLogger(ctx).WithField("op", op)

	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		logger.WithError(err).Error("invalid user ID")
		return nil, status.Error(codes.InvalidArgument, "invalid user ID")
	}

	contactUserID, err := uuid.Parse(req.ContactUserId)
	if err != nil {
		logger.WithError(err).Error("invalid contact user ID")
		return nil, status.Error(codes.InvalidArgument, "invalid contact user ID")
	}

	if userID == contactUserID {
		return nil, status.Error(codes.InvalidArgument, "cannot add yourself as contact")
	}

	contactReq := &ContactDTO.PostContactDTO{
		ContactUserID: contactUserID,
	}

	if err := h.contactUC.CreateContact(ctx, contactReq, userID); err != nil {
		logger.WithError(err).Error("failed to create contact")

		switch {
		case errors.Is(err, errs.ErrIsDuplicateKey), errors.Is(err, errs.ErrContactAlreadyExists):
			return nil, status.Error(codes.AlreadyExists, "contact already exists")
		case errors.Is(err, errs.ErrUserNotFound), errors.Is(err, errs.ErrNotFound):
			return nil, status.Error(codes.NotFound, "user not found")
		case errors.Is(err, errs.ErrBadRequest):
			return nil, status.Error(codes.InvalidArgument, err.Error())
		default:
			return nil, status.Error(codes.Internal, "failed to create contact")
		}
	}

	return &emptypb.Empty{}, nil
}

func (h *UserGRPCHandler) GetContacts(ctx context.Context, req *gen.GetContactsReq) (*gen.GetContactsRes, error) {
	const op = "UserGRPCHandler.GetContacts"
	logger := domains.GetLogger(ctx).WithField("op", op)

	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		logger.WithError(err).Error("invalid user ID")
		return nil, status.Error(codes.InvalidArgument, "invalid user ID")
	}

	contacts, err := h.contactUC.GetContacts(ctx, userID)
	if err != nil {
		logger.WithError(err).Error("failed to get contacts")
		return nil, status.Error(codes.Internal, err.Error())
	}

	var grpcContacts []*gen.Contact
	for _, c := range contacts {
		bio := ""
		if c.ContactUser.Bio != nil {
			bio = *c.ContactUser.Bio
		}

		grpcContacts = append(grpcContacts, &gen.Contact{
			Id:          c.ContactUser.ID.String(),
			PhoneNumber: c.ContactUser.PhoneNumber,
			Name:        c.ContactUser.Name,
			Username:    c.ContactUser.Username,
			Bio:         bio,
			AccountType: c.ContactUser.AccountType,
			CreatedAt:   c.CreatedAt.Format(time.RFC3339),
			UpdatedAt:   c.UpdatedAt.Format(time.RFC3339),
		})
	}

	return &gen.GetContactsRes{
		Contacts: grpcContacts,
	}, nil
}
