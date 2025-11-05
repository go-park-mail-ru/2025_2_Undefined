package message

import (
	"context"

	dtoChats "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/dto/chats"
	dtoMessage "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/dto/message"
	"github.com/google/uuid"
)

type MessageUsecase interface {
	AddMessage(ctx context.Context, msg dtoMessage.CreateMessageDTO, userID uuid.UUID) error
	SubscribeConnectionToChats(ctx context.Context, connectionID uuid.UUID, chatsDTO []dtoChats.ChatViewInformationDTO) <-chan dtoMessage.MessageDTO
}
