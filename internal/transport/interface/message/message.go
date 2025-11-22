package message

import (
	"context"

	dtoChats "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/dto/chats"
	dtoMessage "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/dto/message"
	"github.com/google/uuid"
)

type MessageUsecase interface {
	AddMessage(ctx context.Context, message dtoMessage.CreateMessageDTO, userID uuid.UUID) error
	EditMessage(ctx context.Context, message dtoMessage.EditMessageDTO, userID uuid.UUID) error
	DeleteMessage(ctx context.Context, message dtoMessage.DeleteMessageDTO, userID uuid.UUID) error
	SubscribeConnectionToChats(ctx context.Context, connectionID, userID uuid.UUID, chatsDTO []dtoChats.ChatViewInformationDTO) <-chan dtoMessage.WebSocketMessageDTO
	SubscribeUsersOnChat(ctx context.Context, chatID uuid.UUID, members []dtoChats.AddChatMemberDTO) error
}
