package chats

import (
	"context"

	"github.com/go-park-mail-ru/2025_2_Undefined/internal/models/domains"
	mappers "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/chats-message/mappers"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/chats-message/utils"
	dtoMessage "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/dto/message"
	gen "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/generated/chats"
	chatsInterface "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/interface/chats"
	messageInterface "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/interface/message"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type MessageGRPCHandler struct {
	gen.UnimplementedMessageServiceServer

	messageUsecase messageInterface.MessageUsecase
	chatsUsecase   chatsInterface.ChatsUsecase
}

func NewMessageGRPCHandler(messageUC messageInterface.MessageUsecase, chatsUC chatsInterface.ChatsUsecase) *MessageGRPCHandler {
	return &MessageGRPCHandler{
		messageUsecase: messageUC,
		chatsUsecase:   chatsUC,
	}
}

func (h *MessageGRPCHandler) StreamMessagesForUser(in *gen.StreamMessagesForUserReq, stream gen.MessageService_StreamMessagesForUserServer) error {
	const op = "MessageGRPCHandler.StreamMessagesForUser"
	logger := domains.GetLogger(stream.Context()).WithField("operation", op)

	userID, err := uuid.Parse(in.GetUserId())
	if err != nil {
		logger.WithError(err).Errorf("error parsing senderId: %s", in.GetUserId())
		return status.Error(codes.InvalidArgument, "wrong sender id format")
	}

	logger.Debugf("start stream messages from user %s ", userID)

	chatsViewDTO, err := h.chatsUsecase.GetChats(stream.Context(), userID)
	if err != nil {
		logger.WithError(err).Error("Failed to get chats for user")
		return status.Error(codes.InvalidArgument, "can't get chats for user")
	}

	connectionID := uuid.New()
	msgChan := h.messageUsecase.SubscribeConnectionToChats(stream.Context(), connectionID, userID, chatsViewDTO)

	for {
		select {
		case <-stream.Context().Done():
			return stream.Context().Err()
		case msg, ok := <-msgChan:
			if !ok {
				return nil
			}

			protoMsg, err := mappers.DTOWebSocketMessageToProtoEventRes(msg)
			if err != nil {
				logger.WithError(err).Error("error converting dto to proto")
				continue
			}

			err = stream.Send(protoMsg)
			if err != nil {
				logger.WithError(err).Error("error sending proto message")
				continue
			}
		}
	}
}

func (h *MessageGRPCHandler) HandleSendMessage(ctx context.Context, in *gen.MessageEventReq) (*emptypb.Empty, error) {
	const op = "MessageGRPCHandler.HandleSendMessage"
	logger := domains.GetLogger(ctx).WithField("operation", op)

	if err := validateMessageEventReq(in); err != nil {
		logger.WithError(err).Error(err.Error())
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	userID, err := uuid.Parse(in.GetUserId())
	if err != nil {
		logger.WithError(err).Error("can't parse user_id")
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	websocketMessageDTO, err := mappers.ProtoMessageEventReqToDTO(in)
	if err != nil {
		logger.WithError(err).Error("error converting proto to dto")
		return nil, status.Error(codes.InvalidArgument, "error converting proto to dto")
	}

	// Обработка разных типов сообщений
	var processingErr error
	switch websocketMessageDTO.Type {
	case dtoMessage.WebSocketMessageTypeNewChatMessage:
		var messageDTO dtoMessage.CreateMessageDTO
		if err := utils.DecodeValue(websocketMessageDTO.Value, &messageDTO); err != nil {
			logger.Errorf("can't parse new_message: %v, err: %v", websocketMessageDTO.Value, err)
			return nil, status.Error(codes.InvalidArgument, "can't parse message")
		}
		processingErr = h.createMessage(ctx, userID, messageDTO)

	case dtoMessage.WebSocketMessageTypeEditChatMessage:
		var messageDTO dtoMessage.EditMessageDTO
		if err := utils.DecodeValue(websocketMessageDTO.Value, &messageDTO); err != nil {
			logger.Errorf("can't parse edit_message: %v, err: %v", websocketMessageDTO.Value, err)
			return nil, status.Error(codes.InvalidArgument, "can't parse message")
		}
		processingErr = h.editMessage(ctx, userID, messageDTO)

	case dtoMessage.WebSocketMessageTypeDeleteChatMessage:
		var messageDTO dtoMessage.DeleteMessageDTO
		if err := utils.DecodeValue(websocketMessageDTO.Value, &messageDTO); err != nil {
			logger.Errorf("can't parse delete_message: %v, err: %v", websocketMessageDTO.Value, err)
			return nil, status.Error(codes.InvalidArgument, "can't parse message")
		}
		processingErr = h.deleteMessage(ctx, userID, messageDTO)
	}

	if processingErr != nil {
		logger.WithError(processingErr).Errorf("failed to process message type: %s", websocketMessageDTO.Type)
		return nil, status.Error(codes.InvalidArgument, processingErr.Error())
	}

	return &emptypb.Empty{}, nil
}

func (h *MessageGRPCHandler) createMessage(ctx context.Context, userID uuid.UUID, message dtoMessage.CreateMessageDTO) error {
	return h.messageUsecase.AddMessage(ctx, message, userID)
}

func (h *MessageGRPCHandler) editMessage(ctx context.Context, userID uuid.UUID, message dtoMessage.EditMessageDTO) error {
	return h.messageUsecase.EditMessage(ctx, message, userID)
}

func (h *MessageGRPCHandler) deleteMessage(ctx context.Context, userID uuid.UUID, message dtoMessage.DeleteMessageDTO) error {
	return h.messageUsecase.DeleteMessage(ctx, message, userID)
}
