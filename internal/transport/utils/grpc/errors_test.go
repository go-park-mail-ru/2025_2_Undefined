package grpc

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestHandleGRPCError_InvalidArgument(t *testing.T) {
	w := httptest.NewRecorder()
	err := status.Error(codes.InvalidArgument, "invalid argument")

	HandleGRPCError(context.Background(), "test", w, err)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandleGRPCError_NotFound(t *testing.T) {
	w := httptest.NewRecorder()
	err := status.Error(codes.NotFound, "not found")

	HandleGRPCError(context.Background(), "test", w, err)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestHandleGRPCError_AlreadyExists(t *testing.T) {
	w := httptest.NewRecorder()
	err := status.Error(codes.AlreadyExists, "already exists")

	HandleGRPCError(context.Background(), "test", w, err)

	assert.Equal(t, http.StatusConflict, w.Code)
}

func TestHandleGRPCError_PermissionDenied(t *testing.T) {
	w := httptest.NewRecorder()
	err := status.Error(codes.PermissionDenied, "permission denied")

	HandleGRPCError(context.Background(), "test", w, err)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestHandleGRPCError_Unauthenticated(t *testing.T) {
	w := httptest.NewRecorder()
	err := status.Error(codes.Unauthenticated, "unauthenticated")

	HandleGRPCError(context.Background(), "test", w, err)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestHandleGRPCError_ResourceExhausted(t *testing.T) {
	w := httptest.NewRecorder()
	err := status.Error(codes.ResourceExhausted, "resource exhausted")

	HandleGRPCError(context.Background(), "test", w, err)

	assert.Equal(t, http.StatusTooManyRequests, w.Code)
}

func TestHandleGRPCError_Aborted(t *testing.T) {
	w := httptest.NewRecorder()
	err := status.Error(codes.Aborted, "aborted")

	HandleGRPCError(context.Background(), "test", w, err)

	assert.Equal(t, http.StatusConflict, w.Code)
}

func TestHandleGRPCError_Unimplemented(t *testing.T) {
	w := httptest.NewRecorder()
	err := status.Error(codes.Unimplemented, "unimplemented")

	HandleGRPCError(context.Background(), "test", w, err)

	assert.Equal(t, http.StatusNotImplemented, w.Code)
}

func TestHandleGRPCError_Unavailable(t *testing.T) {
	w := httptest.NewRecorder()
	err := status.Error(codes.Unavailable, "unavailable")

	HandleGRPCError(context.Background(), "test", w, err)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
}

func TestHandleGRPCError_DeadlineExceeded(t *testing.T) {
	w := httptest.NewRecorder()
	err := status.Error(codes.DeadlineExceeded, "deadline exceeded")

	HandleGRPCError(context.Background(), "test", w, err)

	assert.Equal(t, http.StatusGatewayTimeout, w.Code)
}

func TestHandleGRPCError_DefaultCase(t *testing.T) {
	w := httptest.NewRecorder()
	err := status.Error(codes.Unknown, "unknown error")

	HandleGRPCError(context.Background(), "test", w, err)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestHandleGRPCError_NonGRPCError(t *testing.T) {
	w := httptest.NewRecorder()
	err := errors.New("regular error")

	HandleGRPCError(context.Background(), "test", w, err)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
