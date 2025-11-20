package chats

import (
	"time"

	"github.com/go-park-mail-ru/2025_2_Undefined/internal/repository/minio"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/chats-message/utils"
	dtoChats "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/dto/chats"
	dtoMessage "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/dto/message"
	gen "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/generated/chats"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Вспомогательные функции для работы со строками и указателями
func stringToPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func uuidToStringPtr(id *uuid.UUID) *string {
	if id == nil {
		return nil
	}
	s := id.String()
	return &s
}

// Вспомогательные функции для парсинга UUID с обработкой ошибок
func parseUUIDWithError(s string, fieldName string) (uuid.UUID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return uuid.UUID{}, status.Errorf(codes.InvalidArgument, "invalid %s: %v", fieldName, err)
	}
	return id, nil
}

func parseOptionalUUID(s *string) *uuid.UUID {
	if s == nil {
		return nil
	}
	id, err := uuid.Parse(*s)
	if err != nil {
		return nil
	}
	return &id
}

func ProtoMessageToDTO(msg *gen.Message) dtoMessage.MessageDTO {
	msgID, _ := uuid.Parse(msg.GetId())
	chatID, _ := uuid.Parse(msg.GetChatId())
	senderID := parseOptionalUUID(msg.SenderId)
	createdAt, _ := time.Parse(time.RFC3339, msg.GetCreatedAt())

	return dtoMessage.MessageDTO{
		ID:         msgID,
		ChatID:     chatID,
		SenderID:   senderID,
		SenderName: msg.GetSenderName(),
		Text:       msg.GetText(),
		CreatedAt:  createdAt,
		UpdatedAt:  nil, // Protobuf Message не имеет updated_at
		Type:       msg.GetType(),
	}
}

func DTOMessageToProto(msgDTO dtoMessage.MessageDTO) *gen.Message {
	return &gen.Message{
		Id:         msgDTO.ID.String(),
		ChatId:     msgDTO.ChatID.String(),
		SenderId:   uuidToStringPtr(msgDTO.SenderID),
		SenderName: msgDTO.SenderName,
		Text:       msgDTO.Text,
		CreatedAt:  msgDTO.CreatedAt.Format(time.RFC3339),
		Type:       msgDTO.Type,
	}
}

func DTOMessagesToProto(messages []dtoMessage.MessageDTO) []*gen.Message {
	result := make([]*gen.Message, len(messages))
	for i, msgDTO := range messages {
		result[i] = DTOMessageToProto(msgDTO)
	}
	return result
}

func ProtoUserInfoChatToDTO(member *gen.UserInfoChat) dtoChats.UserInfoChatDTO {
	userID, _ := uuid.Parse(member.GetUserId())

	return dtoChats.UserInfoChatDTO{
		UserId:   userID,
		UserName: member.GetUserName(),
		Role:     member.GetRole(),
	}
}

func DTOUserInfoChatToProto(memberDTO dtoChats.UserInfoChatDTO) *gen.UserInfoChat {
	return &gen.UserInfoChat{
		UserId:   memberDTO.UserId.String(),
		UserName: memberDTO.UserName,
		Role:     memberDTO.Role,
	}
}

func DTOMembersToProto(members []dtoChats.UserInfoChatDTO) []*gen.UserInfoChat {
	result := make([]*gen.UserInfoChat, len(members))
	for i, memberDTO := range members {
		result[i] = DTOUserInfoChatToProto(memberDTO)
	}
	return result
}

func ProtoChatToDTO(chat *gen.Chat) dtoChats.ChatViewInformationDTO {
	chatID, _ := uuid.Parse(chat.GetId())

	return dtoChats.ChatViewInformationDTO{
		ID:          chatID,
		Name:        chat.GetName(),
		Type:        chat.GetType(),
		LastMessage: ProtoMessageToDTO(chat.GetLastMessage()),
	}
}

func DTOChatViewToProto(chatDTO dtoChats.ChatViewInformationDTO) *gen.Chat {
	return &gen.Chat{
		Id:          chatDTO.ID.String(),
		Name:        chatDTO.Name,
		Type:        chatDTO.Type,
		LastMessage: DTOMessageToProto(chatDTO.LastMessage),
	}
}

func DTOChatsViewToProto(chats []dtoChats.ChatViewInformationDTO) []*gen.Chat {
	result := make([]*gen.Chat, len(chats))
	for i, chatDTO := range chats {
		result[i] = DTOChatViewToProto(chatDTO)
	}
	return result
}

func ProtoChatDetailedToDTO(chat *gen.ChatDetailedInformation) *dtoChats.ChatDetailedInformationDTO {
	chatID, _ := uuid.Parse(chat.GetId())

	messages := make([]dtoMessage.MessageDTO, len(chat.GetMessages()))
	for i, msg := range chat.GetMessages() {
		messages[i] = ProtoMessageToDTO(msg)
	}

	members := make([]dtoChats.UserInfoChatDTO, len(chat.GetMembers()))
	for i, member := range chat.GetMembers() {
		members[i] = ProtoUserInfoChatToDTO(member)
	}

	return &dtoChats.ChatDetailedInformationDTO{
		ID:          chatID,
		Name:        chat.GetName(),
		Description: chat.GetDescription(),
		IsAdmin:     chat.GetIsAdmin(),
		CanChat:     chat.GetCanChat(),
		IsMember:    chat.GetIsMember(),
		IsPrivate:   chat.GetIsPrivate(),
		Type:        chat.GetType(),
		Messages:    messages,
		Members:     members,
	}
}

func DTOChatDetailedToProto(chatDTO *dtoChats.ChatDetailedInformationDTO) *gen.ChatDetailedInformation {
	return &gen.ChatDetailedInformation{
		Id:          chatDTO.ID.String(),
		Name:        chatDTO.Name,
		Description: stringToPtr(chatDTO.Description),
		IsAdmin:     chatDTO.IsAdmin,
		CanChat:     chatDTO.CanChat,
		IsMember:    chatDTO.IsMember,
		IsPrivate:   chatDTO.IsPrivate,
		Type:        chatDTO.Type,
		Messages:    DTOMessagesToProto(chatDTO.Messages),
		Members:     DTOMembersToProto(chatDTO.Members),
	}
}

func ProtoAddMemberToDTO(member *gen.AddMember) (dtoChats.AddChatMemberDTO, error) {
	userID, err := uuid.Parse(member.GetUserId())
	if err != nil {
		return dtoChats.AddChatMemberDTO{}, err
	}

	return dtoChats.AddChatMemberDTO{
		UserId: userID,
		Role:   member.GetRole(),
	}, nil
}

func DTOAddChatMemberToProto(member dtoChats.AddChatMemberDTO) *gen.AddMember {
	return &gen.AddMember{
		UserId: member.UserId.String(),
		Role:   member.Role,
	}
}

func DTOAddChatMembersToProto(members []dtoChats.AddChatMemberDTO) []*gen.AddMember {
	result := make([]*gen.AddMember, len(members))
	for i, member := range members {
		result[i] = DTOAddChatMemberToProto(member)
	}
	return result
}

func ProtoAddMembersToDTO(members []*gen.AddMember) ([]dtoChats.AddChatMemberDTO, error) {
	result := make([]dtoChats.AddChatMemberDTO, len(members))

	for i, member := range members {
		memberID, err := uuid.Parse(member.GetUserId())
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid member id at index %d: %v", i, err)
		}

		result[i] = dtoChats.AddChatMemberDTO{
			UserId: memberID,
			Role:   member.GetRole(),
		}
	}

	return result, nil
}

func ProtoCreateChatToDTO(in *gen.CreateChatReq) (*dtoChats.ChatCreateInformationDTO, error) {
	name := in.GetName()
	typ := in.GetType()
	members := in.GetMembers()

	chatDTO := &dtoChats.ChatCreateInformationDTO{
		Name:    name,
		Type:    typ,
		Members: make([]dtoChats.AddChatMemberDTO, len(members)),
	}

	for i, member := range members {
		ID, err := uuid.Parse(member.GetUserId())
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid member id %s: %v", member.GetUserId(), err)
		}

		chatDTO.Members[i] = dtoChats.AddChatMemberDTO{
			UserId: ID,
			Role:   member.GetRole(),
		}
	}

	return chatDTO, nil
}

// ProtoMessageEventResToDTO конвертирует protobuf MessageEventRes в WebSocketMessageDTO
func ProtoMessageEventResToDTO(event *gen.MessageEventRes) dtoMessage.WebSocketMessageDTO {
	switch e := event.Event.(type) {
	case *gen.MessageEventRes_NewChatMessage:
		msg := ProtoMessageToDTO(e.NewChatMessage)
		chatID, _ := uuid.Parse(event.GetChatId())
		return dtoMessage.WebSocketMessageDTO{
			Type:   dtoMessage.WebSocketMessageTypeNewChatMessage,
			ChatID: chatID,
			Value:  msg,
		}

	case *gen.MessageEventRes_NewChatCreated:
		chat := ProtoChatToDTO(e.NewChatCreated)
		return dtoMessage.WebSocketMessageDTO{
			Type:   dtoMessage.WebSocketMessageTypeCreatedNewChat,
			ChatID: chat.ID,
			Value:  chat,
		}

	case *gen.MessageEventRes_EditChatMessage:
		messageID, _ := uuid.Parse(e.EditChatMessage.GetMessageId())
		chatID, _ := uuid.Parse(event.GetChatId())
		return dtoMessage.WebSocketMessageDTO{
			Type:   dtoMessage.WebSocketMessageTypeEditChatMessage,
			ChatID: chatID,
			Value: dtoMessage.EditMessageDTO{
				ID:        messageID,
				Text:      e.EditChatMessage.GetText(),
				UpdatedAt: e.EditChatMessage.UpdatedAt.AsTime(),
			},
		}

	case *gen.MessageEventRes_DeleteChatMessage:
		messageID, _ := uuid.Parse(e.DeleteChatMessage.GetMessageId())
		chatID, _ := uuid.Parse(event.GetChatId())
		return dtoMessage.WebSocketMessageDTO{
			Type:   dtoMessage.WebSocketMessageTypeDeleteChatMessage,
			ChatID: chatID,
			Value: dtoMessage.DeleteMessageDTO{
				ID: messageID,
			},
		}

	case *gen.MessageEventRes_UserJoined:
		chatID, _ := uuid.Parse(e.UserJoined.GetChatId())
		userID, _ := uuid.Parse(e.UserJoined.GetUserId())
		return dtoMessage.WebSocketMessageDTO{
			Type:   "user_joined",
			ChatID: chatID,
			Value: dtoMessage.UserJoinedDTO{
				UserID: userID,
				ChatID: chatID,
			},
		}

	default:
		return dtoMessage.WebSocketMessageDTO{
			Type: "unknown",
		}
	}
}

// decodeOrCast пытается получить значение нужного типа: сначала через type assertion, затем через декодирование
func decodeOrCast[T any](value any, dst *T) bool {
	// Попытка прямого type assertion
	if typed, ok := value.(T); ok {
		*dst = typed
		return true
	}

	// Попытка декодирования через JSON
	if err := utils.DecodeValue(value, dst); err == nil {
		return true
	}

	return false
}

// DTOWebSocketMessageToProto конвертирует WebSocketMessageDTO в protobuf MessageEventReq
func DTOWebSocketMessageToProto(userID uuid.UUID, wsMsg dtoMessage.WebSocketMessageDTO) *gen.MessageEventReq {
	userIDStr := userID.String()

	switch wsMsg.Type {
	case dtoMessage.WebSocketMessageTypeNewChatMessage:
		var createMsg dtoMessage.CreateMessageDTO
		if !decodeOrCast(wsMsg.Value, &createMsg) {
			return nil
		}

		return &gen.MessageEventReq{
			UserId: userIDStr,
			Event: &gen.MessageEventReq_NewChatMessage{
				NewChatMessage: &gen.CreateMessage{
					ChatId: createMsg.ChatId.String(),
					Text:   createMsg.Text,
				},
			},
		}

	case dtoMessage.WebSocketMessageTypeEditChatMessage:
		var editDTO dtoMessage.EditMessageDTO
		if !decodeOrCast(wsMsg.Value, &editDTO) {
			return nil
		}

		return &gen.MessageEventReq{
			UserId: userIDStr,
			Event: &gen.MessageEventReq_EditChatMessage{
				EditChatMessage: protoEditMessageToGen(wsMsg.ChatID, editDTO),
			},
		}

	case dtoMessage.WebSocketMessageTypeDeleteChatMessage:
		var deleteDTO dtoMessage.DeleteMessageDTO
		if !decodeOrCast(wsMsg.Value, &deleteDTO) {
			return nil
		}

		return &gen.MessageEventReq{
			UserId: userIDStr,
			Event: &gen.MessageEventReq_DeleteChatMessage{
				DeleteChatMessage: protoDeleteMessageToGen(wsMsg.ChatID, deleteDTO),
			},
		}
	}

	return nil
}

// DTOCreateMessageToProto конвертирует CreateMessageDTO в protobuf MessageEventReq
func DTOCreateMessageToProto(msg dtoMessage.CreateMessageDTO) *gen.MessageEventReq {
	return &gen.MessageEventReq{
		Event: &gen.MessageEventReq_NewChatMessage{
			NewChatMessage: &gen.CreateMessage{
				ChatId: msg.ChatId.String(),
				Text:   msg.Text,
			},
		},
	}
}

// ProtoMessageEventReqToDTO конвертирует protobuf MessageEventReq в WebSocketMessageDTO
func ProtoMessageEventReqToDTO(event *gen.MessageEventReq) (dtoMessage.WebSocketMessageDTO, error) {
	if event == nil || event.Event == nil {
		return dtoMessage.WebSocketMessageDTO{}, status.Error(codes.InvalidArgument, "event is nil")
	}

	switch e := event.Event.(type) {
	case *gen.MessageEventReq_NewChatMessage:
		return ProtoCreateMessageToDTO(e.NewChatMessage)

	case *gen.MessageEventReq_EditChatMessage:
		return ProtoEditMessageToDTO(e.EditChatMessage)

	case *gen.MessageEventReq_DeleteChatMessage:
		return ProtoDeleteMessageToDTO(e.DeleteChatMessage)

	default:
		return dtoMessage.WebSocketMessageDTO{}, status.Error(codes.InvalidArgument, "unknown event type")
	}
}

// ProtoCreateMessageToDTO конвертирует CreateMessage в WebSocketMessageDTO
func ProtoCreateMessageToDTO(msg *gen.CreateMessage) (dtoMessage.WebSocketMessageDTO, error) {
	if msg == nil {
		return dtoMessage.WebSocketMessageDTO{}, status.Error(codes.InvalidArgument, "create_message is nil")
	}

	chatID, err := parseUUIDWithError(msg.GetChatId(), "chat_id")
	if err != nil {
		return dtoMessage.WebSocketMessageDTO{}, err
	}

	return dtoMessage.WebSocketMessageDTO{
		Type:   dtoMessage.WebSocketMessageTypeNewChatMessage,
		ChatID: chatID,
		Value: dtoMessage.CreateMessageDTO{
			Text:      msg.GetText(),
			CreatedAt: time.Now(),
			ChatId:    chatID,
		},
	}, nil
}

// ProtoEditMessageToDTO конвертирует EditMessage в WebSocketMessageDTO
func ProtoEditMessageToDTO(msg *gen.EditMessage) (dtoMessage.WebSocketMessageDTO, error) {
	if msg == nil {
		return dtoMessage.WebSocketMessageDTO{}, status.Error(codes.InvalidArgument, "edit_message is nil")
	}

	messageID, err := parseUUIDWithError(msg.GetMessageId(), "message_id")
	if err != nil {
		return dtoMessage.WebSocketMessageDTO{}, err
	}
	// chat_id теперь должен быть в EditMessage, но в proto он только в MessageEventRes
	// Для универсальности, оставим ChatID пустым, использовать нужно из event.GetChatId()
	return dtoMessage.WebSocketMessageDTO{
		Type:   dtoMessage.WebSocketMessageTypeEditChatMessage,
		ChatID: uuid.Nil, // ChatID должен быть установлен снаружи, если нужно
		Value: dtoMessage.EditMessageDTO{
			ID:        messageID,
			Text:      msg.GetText(),
			UpdatedAt: msg.UpdatedAt.AsTime(),
		},
	}, nil
}

// ProtoDeleteMessageToDTO конвертирует DeleteMessage в WebSocketMessageDTO
func ProtoDeleteMessageToDTO(msg *gen.DeleteMessage) (dtoMessage.WebSocketMessageDTO, error) {
	if msg == nil {
		return dtoMessage.WebSocketMessageDTO{}, status.Error(codes.InvalidArgument, "delete_message is nil")
	}

	messageID, err := parseUUIDWithError(msg.GetMessageId(), "message_id")
	if err != nil {
		return dtoMessage.WebSocketMessageDTO{}, err
	}
	// chat_id теперь должен быть в DeleteMessage, но в proto он только в MessageEventRes
	return dtoMessage.WebSocketMessageDTO{
		Type:   dtoMessage.WebSocketMessageTypeDeleteChatMessage,
		ChatID: uuid.Nil, // ChatID должен быть установлен снаружи, если нужно
		Value: dtoMessage.DeleteMessageDTO{
			ID: messageID,
		},
	}, nil
}

// protoEditMessageToGen конвертирует EditMessageDTO в protobuf EditMessage
func protoEditMessageToGen(chatID uuid.UUID, editDTO dtoMessage.EditMessageDTO) *gen.EditMessage {
	return &gen.EditMessage{
		MessageId: editDTO.ID.String(),
		Text:      editDTO.Text,
		UpdatedAt: timestamppb.New(editDTO.UpdatedAt),
	}
}

// protoDeleteMessageToGen конвертирует DeleteMessageDTO в protobuf DeleteMessage
func protoDeleteMessageToGen(_ uuid.UUID, deleteDTO dtoMessage.DeleteMessageDTO) *gen.DeleteMessage {
	return &gen.DeleteMessage{
		MessageId: deleteDTO.ID.String(),
	}
}

// DTOWebSocketMessageToProtoEventRes конвертирует WebSocketMessageDTO в protobuf MessageEventRes
func DTOWebSocketMessageToProtoEventRes(wsMsg dtoMessage.WebSocketMessageDTO) (*gen.MessageEventRes, error) {
	switch wsMsg.Type {
	case dtoMessage.WebSocketMessageTypeNewChatMessage:
		if msgDTO, ok := wsMsg.Value.(dtoMessage.MessageDTO); ok {
			protoMsg := DTOMessageToProto(msgDTO)
			return &gen.MessageEventRes{
				Event: &gen.MessageEventRes_NewChatMessage{
					NewChatMessage: protoMsg,
				},
			}, nil
		}
		return nil, status.Errorf(codes.InvalidArgument, "invalid value type for new_chat_message: expected MessageDTO")

	case dtoMessage.WebSocketMessageTypeCreatedNewChat:
		if chatDTO, ok := wsMsg.Value.(dtoChats.ChatViewInformationDTO); ok {
			protoChat := DTOChatViewToProto(chatDTO)
			return &gen.MessageEventRes{
				Event: &gen.MessageEventRes_NewChatCreated{
					NewChatCreated: protoChat,
				},
			}, nil
		}
		return nil, status.Errorf(codes.InvalidArgument, "invalid value type for new_chat_created: expected ChatViewInformationDTO")

	case dtoMessage.WebSocketMessageTypeEditChatMessage:
		if editDTO, ok := wsMsg.Value.(dtoMessage.EditMessageDTO); ok {
			return &gen.MessageEventRes{
				Event: &gen.MessageEventRes_EditChatMessage{
					EditChatMessage: protoEditMessageToGen(wsMsg.ChatID, editDTO),
				},
			}, nil
		}
		return nil, status.Errorf(codes.InvalidArgument, "invalid value type for edit_chat_message: expected EditMessageDTO")

	case dtoMessage.WebSocketMessageTypeDeleteChatMessage:
		if deleteDTO, ok := wsMsg.Value.(dtoMessage.DeleteMessageDTO); ok {
			return &gen.MessageEventRes{
				Event: &gen.MessageEventRes_DeleteChatMessage{
					DeleteChatMessage: protoDeleteMessageToGen(wsMsg.ChatID, deleteDTO),
				},
			}, nil
		}
		return nil, status.Errorf(codes.InvalidArgument, "invalid value type for delete_chat_message: expected DeleteMessageDTO")

	case "user_joined":
		if userJoinedDTO, ok := wsMsg.Value.(dtoMessage.UserJoinedDTO); ok {
			return &gen.MessageEventRes{
				Event: &gen.MessageEventRes_UserJoined{
					UserJoined: &gen.UserJoined{
						ChatId: wsMsg.ChatID.String(),
						UserId: userJoinedDTO.UserID.String(),
					},
				},
			}, nil
		}
		return nil, status.Errorf(codes.InvalidArgument, "invalid value type for user_joined: expected UserJoinedDTO")

	default:
		return nil, status.Errorf(codes.InvalidArgument, "unknown websocket message type: %s", wsMsg.Type)
	}
}

// ProtoUploadChatAvatarReqToFileData конвертирует proto запрос в FileData для MinIO
func ProtoUploadChatAvatarReqToFileData(in *gen.UploadChatAvatarReq) minio.FileData {
	return minio.FileData{
		Name:        in.GetFilename(),
		Data:        in.GetData(),
		ContentType: in.GetContentType(),
	}
}
