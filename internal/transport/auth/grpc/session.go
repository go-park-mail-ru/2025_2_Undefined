package grpc

import (
	"context"

	"github.com/go-park-mail-ru/2025_2_Undefined/internal/models/domains"
	gen "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/generated/auth"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

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
