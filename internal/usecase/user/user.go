package user

import (
	"context"

	"github.com/go-park-mail-ru/2025_2_Undefined/internal/models/domains"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/models/errs"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/repository/minio"
	UserDto "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/dto/user"
	InterfaceFileStorage "github.com/go-park-mail-ru/2025_2_Undefined/internal/usecase/interface/storage"
	InterfaceUserRepository "github.com/go-park-mail-ru/2025_2_Undefined/internal/usecase/interface/user"
	"github.com/google/uuid"
)

type UserUsecase struct {
	userrepo    InterfaceUserRepository.UserRepository
	fileStorage InterfaceFileStorage.FileStorage
}

func New(userrepo InterfaceUserRepository.UserRepository, fileStorage InterfaceFileStorage.FileStorage) *UserUsecase {
	return &UserUsecase{
		userrepo:    userrepo,
		fileStorage: fileStorage,
	}
}

func (uc *UserUsecase) GetUserById(ctx context.Context, id uuid.UUID) (*UserDto.User, error) {
	const op = "AuthUsecase.GetUserById"

	logger := domains.GetLogger(ctx).WithField("operation", op)

	user, err := uc.userrepo.GetUserByID(ctx, id)
	if err != nil {
		logger.WithError(err).Error(errs.ErrUserNotFound)
		return nil, errs.ErrUserNotFound
	}

	userdto := &UserDto.User{
		ID:          user.ID,
		PhoneNumber: user.PhoneNumber,
		Name:        user.Name,
		Username:    user.Username,
		Bio:         user.Bio,
		AccountType: user.AccountType,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
	}

	return userdto, nil
}

func (uc *UserUsecase) GetUserByPhone(ctx context.Context, phone string) (*UserDto.User, error) {
	const op = "AuthUsecase.GetUserByPhone"

	logger := domains.GetLogger(ctx).WithField("operation", op)

	user, err := uc.userrepo.GetUserByPhone(ctx, phone)
	if err != nil {
		logger.WithError(err).Error(errs.ErrUserNotFound)
		return nil, errs.ErrUserNotFound
	}

	userdto := &UserDto.User{
		ID:          user.ID,
		PhoneNumber: user.PhoneNumber,
		Name:        user.Name,
		Username:    user.Username,
		Bio:         user.Bio,
		AccountType: user.AccountType,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
	}

	return userdto, nil
}

func (uc *UserUsecase) GetUserByUsername(ctx context.Context, username string) (*UserDto.User, error) {
	const op = "AuthUsecase.GetUserByPhone"

	logger := domains.GetLogger(ctx).WithField("operation", op)

	user, err := uc.userrepo.GetUserByUsername(ctx, username)
	if err != nil {
		logger.WithError(err).Error(errs.ErrUserNotFound)
		return nil, errs.ErrUserNotFound
	}

	userdto := &UserDto.User{
		ID:          user.ID,
		PhoneNumber: user.PhoneNumber,
		Name:        user.Name,
		Username:    user.Username,
		Bio:         user.Bio,
		AccountType: user.AccountType,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
	}

	return userdto, nil
}

func (uc *UserUsecase) UploadUserAvatar(ctx context.Context, userID uuid.UUID, data []byte, filename, contentType string) (string, error) {
	const op = "UserUsecase.UploadUserAvatar"

	logger := domains.GetLogger(ctx).WithField("operation", op)

	avatarID := uuid.New()

	file := minio.FileData{
		Name:        filename,
		Data:        data,
		ContentType: contentType,
	}

	avatar_url, err := uc.fileStorage.CreateOne(ctx, file, avatarID)
	if err != nil {
		logger.WithError(err).Error("could not upload user avatar to file storage")
		return "", err
	}

	err = uc.userrepo.UpdateUserAvatar(ctx, userID, avatarID, int64(len(data)))
	if err != nil {
		logger.WithError(err).Error("could not update user avatar")
		return "", err
	}

	return avatar_url, nil
}

func (uc *UserUsecase) UpdateUserInfo(ctx context.Context, userID uuid.UUID, name *string, username *string, bio *string) error {
	const op = "UserUsecase.UpdateUserInfo"

	logger := domains.GetLogger(ctx).WithField("operation", op)

	err := uc.userrepo.UpdateUserInfo(ctx, userID, name, username, bio)
	if err != nil {
		logger.WithError(err).Error("could not update user info")
		return err
	}

	return nil
}

func (uc *UserUsecase) GetUserAvatars(ctx context.Context, userIDs []uuid.UUID) (map[string]*string, error) {
	const op = "UserUsecase.GetUserAvatars"

	logger := domains.GetLogger(ctx).WithField("operation", op).WithField("user_ids_count", len(userIDs))
	logger.Debug("Starting usecase operation: get user avatars")

	// Инициализируем карту для всех запрошенных ID со значением nil
	avatars := make(map[string]*string, len(userIDs))
	for _, userID := range userIDs {
		avatars[userID.String()] = nil
	}

	avatarsIDs, err := uc.userrepo.GetUserAvatars(ctx, userIDs)
	if err != nil {
		logger.WithError(err).Error("Failed to get user avatars from repository")
		return nil, err
	}

	for userID, attachmentID := range avatarsIDs {
		url, err := uc.fileStorage.GetOne(ctx, &attachmentID)
		if err != nil {
			avatars[userID] = nil
		} else {
			u := url
			avatars[userID] = &u
		}
	}

	logger.WithField("avatars_count", len(avatars)).Info("Usecase operation completed successfully: user avatars retrieved")
	return avatars, nil
}
