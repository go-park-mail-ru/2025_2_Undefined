package chats

import (
	"testing"
	"time"

	dtoChats "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/dto/chats"
	dtoMessage "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/dto/message"
	gen "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/generated/chats"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestStringToPtr(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected *string
	}{
		{"EmptyString", "", nil},
		{"NonEmptyString", "test", func() *string { s := "test"; return &s }()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := stringToPtr(tt.input)
			if tt.expected == nil {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
				assert.Equal(t, *tt.expected, *result)
			}
		})
	}
}

func TestUuidToStringPtr(t *testing.T) {
	tests := []struct {
		name     string
		input    *uuid.UUID
		expected *string
	}{
		{"NilUUID", nil, nil},
		{"ValidUUID", func() *uuid.UUID { id := uuid.New(); return &id }(), func() *string { id := uuid.New(); s := id.String(); return &s }()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := uuidToStringPtr(tt.input)
			if tt.expected == nil {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
			}
		})
	}
}

func TestParseUUIDWithError(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		fieldName string
		wantErr   bool
	}{
		{"ValidUUID", uuid.New().String(), "test_field", false},
		{"InvalidUUID", "invalid-uuid", "test_field", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseUUIDWithError(tt.input, tt.fieldName)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestParseOptionalUUID(t *testing.T) {
	validUUIDStr := uuid.New().String()
	invalidUUIDStr := "invalid"

	tests := []struct {
		name     string
		input    *string
		expected bool
	}{
		{"NilString", nil, false},
		{"ValidUUID", &validUUIDStr, true},
		{"InvalidUUID", &invalidUUIDStr, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseOptionalUUID(tt.input)
			if tt.expected {
				assert.NotNil(t, result)
			} else {
				assert.Nil(t, result)
			}
		})
	}
}

func TestProtoMessageToDTO(t *testing.T) {
	msgID := uuid.New()
	chatID := uuid.New()
	senderID := uuid.New()
	senderIDStr := senderID.String()
	senderName := "Test User"
	createdAt := time.Now().Format(time.RFC3339)

	protoMsg := &gen.Message{
		Id:         msgID.String(),
		ChatId:     chatID.String(),
		SenderId:   &senderIDStr,
		SenderName: senderName,
		Text:       "Test message",
		CreatedAt:  createdAt,
		Type:       "text",
	}

	result := ProtoMessageToDTO(protoMsg)

	assert.Equal(t, msgID, result.ID)
	assert.Equal(t, chatID, result.ChatID)
	assert.NotNil(t, result.SenderID)
	assert.Equal(t, "Test message", result.Text)
	assert.Equal(t, "text", result.Type)
}

func TestDTOMessageToProto(t *testing.T) {
	msgID := uuid.New()
	chatID := uuid.New()
	senderID := uuid.New()
	senderName := "Test User"
	createdAt := time.Now()

	dtoMsg := dtoMessage.MessageDTO{
		ID:         msgID,
		ChatID:     chatID,
		SenderID:   &senderID,
		SenderName: &senderName,
		Text:       "Test message",
		CreatedAt:  createdAt,
		Type:       "text",
	}

	result := DTOMessageToProto(dtoMsg)

	assert.Equal(t, msgID.String(), result.Id)
	assert.Equal(t, chatID.String(), result.ChatId)
	assert.Equal(t, senderName, result.SenderName)
	assert.Equal(t, "Test message", result.Text)
	assert.Equal(t, "text", result.Type)
}

func TestDTOMessagesToProto(t *testing.T) {
	msg1ID := uuid.New()
	msg2ID := uuid.New()
	chatID := uuid.New()

	dtoMessages := []dtoMessage.MessageDTO{
		{
			ID:        msg1ID,
			ChatID:    chatID,
			Text:      "Message 1",
			CreatedAt: time.Now(),
			Type:      "text",
		},
		{
			ID:        msg2ID,
			ChatID:    chatID,
			Text:      "Message 2",
			CreatedAt: time.Now(),
			Type:      "text",
		},
	}

	result := DTOMessagesToProto(dtoMessages)

	assert.Len(t, result, 2)
	assert.Equal(t, msg1ID.String(), result[0].Id)
	assert.Equal(t, msg2ID.String(), result[1].Id)
}

func TestProtoUserInfoChatToDTO(t *testing.T) {
	userID := uuid.New()
	member := &gen.UserInfoChat{
		UserId:   userID.String(),
		UserName: "Test User",
		Role:     "admin",
	}

	result := ProtoUserInfoChatToDTO(member)

	assert.Equal(t, userID, result.UserId)
	assert.Equal(t, "Test User", result.UserName)
	assert.Equal(t, "admin", result.Role)
}

func TestDTOUserInfoChatToProto(t *testing.T) {
	userID := uuid.New()
	memberDTO := dtoChats.UserInfoChatDTO{
		UserId:   userID,
		UserName: "Test User",
		Role:     "admin",
	}

	result := DTOUserInfoChatToProto(memberDTO)

	assert.Equal(t, userID.String(), result.UserId)
	assert.Equal(t, "Test User", result.UserName)
	assert.Equal(t, "admin", result.Role)
}

func TestDTOMembersToProto(t *testing.T) {
	user1ID := uuid.New()
	user2ID := uuid.New()

	members := []dtoChats.UserInfoChatDTO{
		{UserId: user1ID, UserName: "User 1", Role: "admin"},
		{UserId: user2ID, UserName: "User 2", Role: "member"},
	}

	result := DTOMembersToProto(members)

	assert.Len(t, result, 2)
	assert.Equal(t, user1ID.String(), result[0].UserId)
	assert.Equal(t, user2ID.String(), result[1].UserId)
}

func TestProtoChatToDTO(t *testing.T) {
	chatID := uuid.New()
	msgID := uuid.New()
	msgChatID := uuid.New()
	createdAt := time.Now().Format(time.RFC3339)

	protoChat := &gen.Chat{
		Id:   chatID.String(),
		Name: "Test Chat",
		Type: "group",
		LastMessage: &gen.Message{
			Id:        msgID.String(),
			ChatId:    msgChatID.String(),
			Text:      "Last message",
			CreatedAt: createdAt,
			Type:      "text",
		},
	}

	result := ProtoChatToDTO(protoChat)

	assert.Equal(t, chatID, result.ID)
	assert.Equal(t, "Test Chat", result.Name)
	assert.Equal(t, "group", result.Type)
	assert.Equal(t, msgID, result.LastMessage.ID)
}

func TestDTOChatViewToProto(t *testing.T) {
	chatID := uuid.New()
	msgID := uuid.New()

	chatDTO := dtoChats.ChatViewInformationDTO{
		ID:   chatID,
		Name: "Test Chat",
		Type: "group",
		LastMessage: dtoMessage.MessageDTO{
			ID:        msgID,
			ChatID:    chatID,
			Text:      "Last message",
			CreatedAt: time.Now(),
			Type:      "text",
		},
	}

	result := DTOChatViewToProto(chatDTO)

	assert.Equal(t, chatID.String(), result.Id)
	assert.Equal(t, "Test Chat", result.Name)
	assert.Equal(t, "group", result.Type)
	assert.NotNil(t, result.LastMessage)
}

func TestDTOChatsViewToProto(t *testing.T) {
	chat1ID := uuid.New()
	chat2ID := uuid.New()
	msgID := uuid.New()

	chats := []dtoChats.ChatViewInformationDTO{
		{
			ID:   chat1ID,
			Name: "Chat 1",
			Type: "group",
			LastMessage: dtoMessage.MessageDTO{
				ID:        msgID,
				ChatID:    chat1ID,
				Text:      "Message 1",
				CreatedAt: time.Now(),
				Type:      "text",
			},
		},
		{
			ID:   chat2ID,
			Name: "Chat 2",
			Type: "dialog",
			LastMessage: dtoMessage.MessageDTO{
				ID:        msgID,
				ChatID:    chat2ID,
				Text:      "Message 2",
				CreatedAt: time.Now(),
				Type:      "text",
			},
		},
	}

	result := DTOChatsViewToProto(chats)

	assert.Len(t, result, 2)
	assert.Equal(t, chat1ID.String(), result[0].Id)
	assert.Equal(t, chat2ID.String(), result[1].Id)
}

func TestProtoSearchChatsResToDTO(t *testing.T) {
	chatID := uuid.New()
	msgID := uuid.New()
	createdAt := time.Now().Format(time.RFC3339)

	res := &gen.GetChatsRes{
		Chats: []*gen.Chat{
			{
				Id:   chatID.String(),
				Name: "Test Chat",
				Type: "group",
				LastMessage: &gen.Message{
					Id:        msgID.String(),
					ChatId:    chatID.String(),
					Text:      "Last message",
					CreatedAt: createdAt,
					Type:      "text",
				},
			},
		},
	}

	result := ProtoSearchChatsResToDTO(res)

	assert.Len(t, result, 1)
	assert.Equal(t, chatID, result[0].ID)
	assert.Equal(t, "Test Chat", result[0].Name)
}

func TestProtoSearchChatsResToDTO_NilResponse(t *testing.T) {
	result := ProtoSearchChatsResToDTO(nil)
	assert.Empty(t, result)
}

func TestProtoAddMemberToDTO(t *testing.T) {
	userID := uuid.New()
	member := &gen.AddMember{
		UserId: userID.String(),
		Role:   "admin",
	}

	result, err := ProtoAddMemberToDTO(member)

	assert.NoError(t, err)
	assert.Equal(t, userID, result.UserId)
	assert.Equal(t, "admin", result.Role)
}

func TestProtoAddMemberToDTO_InvalidUUID(t *testing.T) {
	member := &gen.AddMember{
		UserId: "invalid-uuid",
		Role:   "admin",
	}

	_, err := ProtoAddMemberToDTO(member)
	assert.Error(t, err)
}

func TestDTOAddChatMemberToProto(t *testing.T) {
	userID := uuid.New()
	member := dtoChats.AddChatMemberDTO{
		UserId: userID,
		Role:   "admin",
	}

	result := DTOAddChatMemberToProto(member)

	assert.Equal(t, userID.String(), result.UserId)
	assert.Equal(t, "admin", result.Role)
}

func TestDTOAddChatMembersToProto(t *testing.T) {
	user1ID := uuid.New()
	user2ID := uuid.New()

	members := []dtoChats.AddChatMemberDTO{
		{UserId: user1ID, Role: "admin"},
		{UserId: user2ID, Role: "member"},
	}

	result := DTOAddChatMembersToProto(members)

	assert.Len(t, result, 2)
	assert.Equal(t, user1ID.String(), result[0].UserId)
	assert.Equal(t, user2ID.String(), result[1].UserId)
}

func TestProtoAddMembersToDTO(t *testing.T) {
	user1ID := uuid.New()
	user2ID := uuid.New()

	members := []*gen.AddMember{
		{UserId: user1ID.String(), Role: "admin"},
		{UserId: user2ID.String(), Role: "member"},
	}

	result, err := ProtoAddMembersToDTO(members)

	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, user1ID, result[0].UserId)
	assert.Equal(t, user2ID, result[1].UserId)
}

func TestProtoAddMembersToDTO_InvalidUUID(t *testing.T) {
	members := []*gen.AddMember{
		{UserId: "invalid-uuid", Role: "admin"},
	}

	_, err := ProtoAddMembersToDTO(members)
	assert.Error(t, err)
}

func TestProtoCreateChatToDTO(t *testing.T) {
	userID := uuid.New()
	req := &gen.CreateChatReq{
		Name: "Test Chat",
		Type: "group",
		Members: []*gen.AddMember{
			{UserId: userID.String(), Role: "admin"},
		},
	}

	result, err := ProtoCreateChatToDTO(req)

	assert.NoError(t, err)
	assert.Equal(t, "Test Chat", result.Name)
	assert.Equal(t, "group", result.Type)
	assert.Len(t, result.Members, 1)
	assert.Equal(t, userID, result.Members[0].UserId)
}

func TestProtoCreateChatToDTO_InvalidMemberID(t *testing.T) {
	req := &gen.CreateChatReq{
		Name: "Test Chat",
		Type: "group",
		Members: []*gen.AddMember{
			{UserId: "invalid-uuid", Role: "admin"},
		},
	}

	_, err := ProtoCreateChatToDTO(req)
	assert.Error(t, err)
}

func TestProtoSearchMessagesResToDTO(t *testing.T) {
	msgID := uuid.New()
	chatID := uuid.New()
	createdAt := time.Now().Format(time.RFC3339)

	res := &gen.SearchMessagesRes{
		Messages: []*gen.Message{
			{
				Id:        msgID.String(),
				ChatId:    chatID.String(),
				Text:      "Test message",
				CreatedAt: createdAt,
				Type:      "text",
			},
		},
	}

	result := ProtoSearchMessagesResToDTO(res)

	assert.Len(t, result, 1)
	assert.Equal(t, msgID, result[0].ID)
	assert.Equal(t, "Test message", result[0].Text)
}

func TestProtoSearchMessagesResToDTO_NilResponse(t *testing.T) {
	result := ProtoSearchMessagesResToDTO(nil)
	assert.Empty(t, result)
}

func TestDTOMessagesToProtoMessage(t *testing.T) {
	msgID := uuid.New()
	chatID := uuid.New()

	messages := []dtoMessage.MessageDTO{
		{
			ID:        msgID,
			ChatID:    chatID,
			Text:      "Test message",
			CreatedAt: time.Now(),
			Type:      "text",
		},
	}

	result := DTOMessagesToProtoMessage(messages)

	assert.Len(t, result, 1)
	assert.Equal(t, msgID.String(), result[0].Id)
	assert.Equal(t, "Test message", result[0].Text)
}
