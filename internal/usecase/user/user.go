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

	avatar_url, err := uc.fileStorage.GetOne(ctx, user.AvatarID)
	if err != nil {
		logger.Warningf("could not get avatar URL for user %s: %v", user.ID, err)
		avatar_url = ""
	}

	logger.Debugf("user avatar url is %s", avatar_url)

	userdto := &UserDto.User{
		ID:          user.ID,
		PhoneNumber: user.PhoneNumber,
		Name:        user.Name,
		Username:    user.Username,
		Bio:         user.Bio,
		AvatarURL:   avatar_url,
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

	avatar_url, err := uc.fileStorage.GetOne(ctx, user.AvatarID)
	if err != nil {
		logger.Warningf("could not get avatar URL for user %s: %v", user.ID, err)
		avatar_url = ""
	}

	userdto := &UserDto.User{
		ID:          user.ID,
		PhoneNumber: user.PhoneNumber,
		Name:        user.Name,
		Username:    user.Username,
		Bio:         user.Bio,
		AvatarURL:   avatar_url,
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

	avatar_url, err := uc.fileStorage.GetOne(ctx, user.AvatarID)
	if err != nil {
		logger.Warningf("could not get avatar URL for user %s: %v", user.ID, err)
		avatar_url = ""
	}

	userdto := &UserDto.User{
		ID:          user.ID,
		PhoneNumber: user.PhoneNumber,
		Name:        user.Name,
		Username:    user.Username,
		Bio:         user.Bio,
		AvatarURL:   avatar_url,
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
