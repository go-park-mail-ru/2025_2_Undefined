package chats

import (
	"context"

	"github.com/go-park-mail-ru/2025_2_Undefined/internal/repository/minio"
	dtoChats "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/dto/chats"
	dtoUtils "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/dto/utils"
	"github.com/google/uuid"
)

type ChatsUsecase interface {
	GetChats(ctx context.Context, userId uuid.UUID) ([]dtoChats.ChatViewInformationDTO, error)
	CreateChat(ctx context.Context, chatDTO dtoChats.ChatCreateInformationDTO) (uuid.UUID, error)
	GetInformationAboutChat(ctx context.Context, userId, chatId uuid.UUID) (*dtoChats.ChatDetailedInformationDTO, error)
	GetUsersDialog(ctx context.Context, user1ID, user2ID uuid.UUID) (*dtoUtils.IdDTO, error)
	AddUsersToChat(ctx context.Context, chatID, userID uuid.UUID, users []dtoChats.AddChatMemberDTO) error
	DeleteChat(ctx context.Context, userId, chatId uuid.UUID) error
	UpdateChat(ctx context.Context, userId, chatId uuid.UUID, name, description string) error
	GetChatAvatars(ctx context.Context, chatIDs []uuid.UUID) (map[string]*string, error)
	UploadChatAvatar(ctx context.Context, userID, chatID uuid.UUID, fileData minio.FileData) (string, error)
}
