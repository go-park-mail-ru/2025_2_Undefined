package grpc

import (
	"context"
	"net/http"

	utils "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/utils/response"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// HandleGRPCError конвертирует gRPC ошибки в HTTP статус коды и отправляет ответ
func HandleGRPCError(ctx context.Context, op string, w http.ResponseWriter, err error) {
	st, ok := status.FromError(err)
	if !ok {
		utils.SendError(ctx, op, w, http.StatusInternalServerError, "internal server error")
		return
	}

	switch st.Code() {
	case codes.InvalidArgument:
		utils.SendError(ctx, op, w, http.StatusBadRequest, st.Message())
	case codes.NotFound:
		utils.SendError(ctx, op, w, http.StatusNotFound, st.Message())
	case codes.AlreadyExists:
		utils.SendError(ctx, op, w, http.StatusConflict, st.Message())
	case codes.PermissionDenied:
		utils.SendError(ctx, op, w, http.StatusForbidden, st.Message())
	case codes.Unauthenticated:
		utils.SendError(ctx, op, w, http.StatusUnauthorized, st.Message())
	case codes.ResourceExhausted:
		utils.SendError(ctx, op, w, http.StatusTooManyRequests, st.Message())
	case codes.Aborted:
		utils.SendError(ctx, op, w, http.StatusConflict, st.Message())
	case codes.Unimplemented:
		utils.SendError(ctx, op, w, http.StatusNotImplemented, st.Message())
	case codes.Unavailable:
		utils.SendError(ctx, op, w, http.StatusServiceUnavailable, st.Message())
	case codes.DeadlineExceeded:
		utils.SendError(ctx, op, w, http.StatusGatewayTimeout, st.Message())
	default:
		utils.SendError(ctx, op, w, http.StatusInternalServerError, st.Message())
	}
}
