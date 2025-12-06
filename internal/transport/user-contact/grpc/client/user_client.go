package client

import (
	"context"

	"github.com/go-park-mail-ru/2025_2_Undefined/internal/models/domains"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/models/errs"
	UserModels "github.com/go-park-mail-ru/2025_2_Undefined/internal/models/user"
	gen "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/generated/user"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// UserServiceClient - gRPC клиент для взаимодействия с user_service
type UserServiceClient struct {
	client gen.UserServiceClient
	conn   *grpc.ClientConn
}

func NewUserServiceClient(addr string) (*UserServiceClient, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	client := gen.NewUserServiceClient(conn)
	return &UserServiceClient{
		client: client,
		conn:   conn,
	}, nil
}

func (c *UserServiceClient) Close() error {
	return c.conn.Close()
}

func (c *UserServiceClient) GetUserByID(ctx context.Context, id uuid.UUID) (*UserModels.User, error) {
	const op = "UserServiceClient.GetUserByID"
	logger := domains.GetLogger(ctx).WithField("operation", op)

	req := &gen.GetUserByIdReq{
		UserId: id.String(),
	}

	resp, err := c.client.GetUserById(ctx, req)
	if err != nil {
		logger.WithError(err).Errorf("failed to get user by id: %s", id)
		return nil, errs.ErrNotFound
	}

	if resp.User == nil {
		return nil, errs.ErrNotFound
	}

	userID, err := uuid.Parse(resp.User.Id)
	if err != nil {
		logger.WithError(err).Error("failed to parse user id")
		return nil, errs.ErrInternalServerError
	}

	user := &UserModels.User{
		ID:          userID,
		Name:        resp.User.Name,
		Username:    resp.User.Username,
		PhoneNumber: resp.User.PhoneNumber,
		Bio:         &resp.User.Bio,
		AccountType: resp.User.AccountType,
	}

	return user, nil
}

func (c *UserServiceClient) GetUsersNames(ctx context.Context, usersIds []uuid.UUID) ([]string, error) {
	const op = "UserServiceClient.GetUsersNames"
	logger := domains.GetLogger(ctx).WithField("operation", op)

	names := make([]string, 0, len(usersIds))

	for _, userID := range usersIds {
		user, err := c.GetUserByID(ctx, userID)
		if err != nil {
			logger.WithError(err).Warningf("failed to get user name for id: %s", userID)
			names = append(names, "Unknown")
			continue
		}
		names = append(names, user.Name)
	}

	return names, nil
}

func (c *UserServiceClient) GetUserByPhone(ctx context.Context, phone string) (*UserModels.User, error) {
	const op = "UserServiceClient.GetUserByPhone"
	logger := domains.GetLogger(ctx).WithField("operation", op)

	req := &gen.GetUserByPhoneReq{
		PhoneNumber: phone,
	}

	resp, err := c.client.GetUserByPhone(ctx, req)
	if err != nil {
		logger.WithError(err).Errorf("failed to get user by phone: %s", phone)
		return nil, errs.ErrNotFound
	}

	if resp.User == nil {
		return nil, errs.ErrNotFound
	}

	userID, err := uuid.Parse(resp.User.Id)
	if err != nil {
		logger.WithError(err).Error("failed to parse user id")
		return nil, errs.ErrInternalServerError
	}

	user := &UserModels.User{
		ID:          userID,
		Name:        resp.User.Name,
		Username:    resp.User.Username,
		PhoneNumber: resp.User.PhoneNumber,
		Bio:         &resp.User.Bio,
		AccountType: resp.User.AccountType,
	}

	return user, nil
}
