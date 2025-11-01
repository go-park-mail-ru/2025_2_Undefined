package chats

import (
	"context"

	dto "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/dto/chats"
	"github.com/google/uuid"
)

type ChatsUsecase interface {
	GetChats(ctx context.Context, userId uuid.UUID) ([]dto.ChatViewInformationDTO, error)
	CreateChat(ctx context.Context, chatDTO dto.ChatCreateInformationDTO) (uuid.UUID, error)
	GetInformationAboutChat(ctx context.Context, userId, chatId uuid.UUID) (*dto.ChatDetailedInformationDTO, error)
	AddUsersToChat(ctx context.Context, chatID uuid.UUID, users []dto.AddChatMemberDTO) error
}
