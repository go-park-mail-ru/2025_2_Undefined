package chats

import (
	"testing"

	gen "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/generated/chats"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestValidateChatCreateDTO(t *testing.T) {
	tests := []struct {
		name    string
		input   *gen.CreateChatReq
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid chat create request",
			input: &gen.CreateChatReq{
				Name: "Test Chat",
				Type: "group",
				Members: []*gen.AddMember{
					{UserId: uuid.New().String(), Role: "admin"},
					{UserId: uuid.New().String(), Role: "writer"},
				},
			},
			wantErr: false,
		},
		{
			name: "empty name",
			input: &gen.CreateChatReq{
				Name: "",
				Type: "group",
				Members: []*gen.AddMember{
					{UserId: uuid.New().String(), Role: "admin"},
				},
			},
			wantErr: true,
			errMsg:  "name is required and cannot be empty",
		},
		{
			name: "whitespace only name",
			input: &gen.CreateChatReq{
				Name: "   ",
				Type: "group",
				Members: []*gen.AddMember{
					{UserId: uuid.New().String(), Role: "admin"},
				},
			},
			wantErr: true,
			errMsg:  "name is required and cannot be empty",
		},
		{
			name: "empty type",
			input: &gen.CreateChatReq{
				Name: "Test Chat",
				Type: "",
				Members: []*gen.AddMember{
					{UserId: uuid.New().String(), Role: "admin"},
				},
			},
			wantErr: true,
			errMsg:  "type is required and cannot be empty",
		},
		{
			name: "empty members",
			input: &gen.CreateChatReq{
				Name:    "Test Chat",
				Type:    "group",
				Members: []*gen.AddMember{},
			},
			wantErr: true,
			errMsg:  "members field is required and cannot be empty",
		},
		{
			name: "nil members",
			input: &gen.CreateChatReq{
				Name:    "Test Chat",
				Type:    "group",
				Members: nil,
			},
			wantErr: true,
			errMsg:  "members field is required and cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateChatCreateDTO(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateAddMembers(t *testing.T) {
	validUUID1 := uuid.New().String()
	validUUID2 := uuid.New().String()

	tests := []struct {
		name    string
		input   []*gen.AddMember
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid members",
			input: []*gen.AddMember{
				{UserId: validUUID1, Role: "admin"},
				{UserId: validUUID2, Role: "writer"},
			},
			wantErr: false,
		},
		{
			name:    "empty members list",
			input:   []*gen.AddMember{},
			wantErr: true,
			errMsg:  "members field is required and cannot be empty",
		},
		{
			name:    "nil members list",
			input:   nil,
			wantErr: true,
			errMsg:  "members field is required and cannot be empty",
		},
		{
			name: "empty user_id",
			input: []*gen.AddMember{
				{UserId: "", Role: "admin"},
			},
			wantErr: true,
			errMsg:  "user_id is required for all users",
		},
		{
			name: "invalid uuid format",
			input: []*gen.AddMember{
				{UserId: "invalid-uuid", Role: "admin"},
			},
			wantErr: true,
			errMsg:  "user_id must be a valid uuid",
		},
		{
			name: "duplicate user_id",
			input: []*gen.AddMember{
				{UserId: validUUID1, Role: "admin"},
				{UserId: validUUID1, Role: "writer"},
			},
			wantErr: true,
			errMsg:  "duplicate user_id found in request",
		},
		{
			name: "empty role",
			input: []*gen.AddMember{
				{UserId: validUUID1, Role: ""},
			},
			wantErr: true,
			errMsg:  "role is required for all users",
		},
		{
			name: "invalid role",
			input: []*gen.AddMember{
				{UserId: validUUID1, Role: "invalid_role"},
			},
			wantErr: true,
			errMsg:  "role must be one of: admin, writer, viewer",
		},
		{
			name: "all valid roles",
			input: []*gen.AddMember{
				{UserId: uuid.New().String(), Role: "admin"},
				{UserId: uuid.New().String(), Role: "writer"},
				{UserId: uuid.New().String(), Role: "viewer"},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateAddMembers(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateRole(t *testing.T) {
	tests := []struct {
		name    string
		role    string
		wantErr bool
	}{
		{
			name:    "admin role",
			role:    "admin",
			wantErr: false,
		},
		{
			name:    "writer role",
			role:    "writer",
			wantErr: false,
		},
		{
			name:    "viewer role",
			role:    "viewer",
			wantErr: false,
		},
		{
			name:    "invalid role",
			role:    "invalid",
			wantErr: true,
		},
		{
			name:    "empty role",
			role:    "",
			wantErr: true,
		},
		{
			name:    "mixed case role",
			role:    "Admin",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateRole(tt.role)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateChatUpdateReq(t *testing.T) {
	namePtr := func(s string) *string { return &s }
	descPtr := func(s string) *string { return &s }

	tests := []struct {
		name    string
		input   *gen.UpdateChatReq
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid name update",
			input: &gen.UpdateChatReq{
				ChatId: uuid.New().String(),
				UserId: uuid.New().String(),
				Name:   namePtr("New Name"),
			},
			wantErr: false,
		},
		{
			name: "valid description update",
			input: &gen.UpdateChatReq{
				ChatId:      uuid.New().String(),
				UserId:      uuid.New().String(),
				Description: descPtr("New Description"),
			},
			wantErr: false,
		},
		{
			name: "valid both fields update",
			input: &gen.UpdateChatReq{
				ChatId:      uuid.New().String(),
				UserId:      uuid.New().String(),
				Name:        namePtr("New Name"),
				Description: descPtr("New Description"),
			},
			wantErr: false,
		},
		{
			name: "no fields to update",
			input: &gen.UpdateChatReq{
				ChatId: uuid.New().String(),
				UserId: uuid.New().String(),
			},
			wantErr: true,
			errMsg:  "at least one field (name or description) must be provided for update",
		},
		{
			name: "empty name",
			input: &gen.UpdateChatReq{
				ChatId: uuid.New().String(),
				UserId: uuid.New().String(),
				Name:   namePtr(""),
			},
			wantErr: true,
			errMsg:  "name cannot be empty",
		},
		{
			name: "whitespace only name",
			input: &gen.UpdateChatReq{
				ChatId: uuid.New().String(),
				UserId: uuid.New().String(),
				Name:   namePtr("   "),
			},
			wantErr: true,
			errMsg:  "name cannot be empty",
		},
		{
			name: "empty description",
			input: &gen.UpdateChatReq{
				ChatId:      uuid.New().String(),
				UserId:      uuid.New().String(),
				Description: descPtr(""),
			},
			wantErr: true,
			errMsg:  "description cannot be empty",
		},
		{
			name: "whitespace only description",
			input: &gen.UpdateChatReq{
				ChatId:      uuid.New().String(),
				UserId:      uuid.New().String(),
				Description: descPtr("   "),
			},
			wantErr: true,
			errMsg:  "description cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateChatUpdateReq(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateMessageEventReq(t *testing.T) {
	validUserID := uuid.New().String()
	validChatID := uuid.New().String()
	validMessageID := uuid.New().String()

	tests := []struct {
		name    string
		input   *gen.MessageEventReq
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid create message event",
			input: &gen.MessageEventReq{
				UserId: validUserID,
				Event: &gen.MessageEventReq_NewChatMessage{
					NewChatMessage: &gen.CreateMessage{
						ChatId: validChatID,
						Text:   "Hello, World!",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "valid edit message event",
			input: &gen.MessageEventReq{
				UserId: validUserID,
				Event: &gen.MessageEventReq_EditChatMessage{
					EditChatMessage: &gen.EditMessage{
						MessageId: validMessageID,
						Text:      "Updated text",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "valid delete message event",
			input: &gen.MessageEventReq{
				UserId: validUserID,
				Event: &gen.MessageEventReq_DeleteChatMessage{
					DeleteChatMessage: &gen.DeleteMessage{
						MessageId: validMessageID,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "nil event",
			input: &gen.MessageEventReq{
				UserId: validUserID,
				Event:  nil,
			},
			wantErr: true,
			errMsg:  "type is required",
		},
		{
			name: "empty user_id",
			input: &gen.MessageEventReq{
				UserId: "",
				Event: &gen.MessageEventReq_NewChatMessage{
					NewChatMessage: &gen.CreateMessage{
						ChatId: validChatID,
						Text:   "Hello",
					},
				},
			},
			wantErr: true,
			errMsg:  "user_id is required",
		},
		{
			name: "invalid user_id format",
			input: &gen.MessageEventReq{
				UserId: "invalid-uuid",
				Event: &gen.MessageEventReq_NewChatMessage{
					NewChatMessage: &gen.CreateMessage{
						ChatId: validChatID,
						Text:   "Hello",
					},
				},
			},
			wantErr: true,
			errMsg:  "wrong format of user_id",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateMessageEventReq(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateCreateMessage(t *testing.T) {
	validChatID := uuid.New().String()

	tests := []struct {
		name    string
		input   *gen.CreateMessage
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid create message",
			input: &gen.CreateMessage{
				ChatId: validChatID,
				Text:   "Hello, World!",
			},
			wantErr: false,
		},
		{
			name:    "nil message",
			input:   nil,
			wantErr: true,
			errMsg:  "create_message is required",
		},
		{
			name: "empty chat_id",
			input: &gen.CreateMessage{
				ChatId: "",
				Text:   "Hello",
			},
			wantErr: true,
			errMsg:  "chat_id is required",
		},
		{
			name: "invalid chat_id format",
			input: &gen.CreateMessage{
				ChatId: "invalid-uuid",
				Text:   "Hello",
			},
			wantErr: true,
			errMsg:  "chat_id must be a valid uuid",
		},
		{
			name: "empty text",
			input: &gen.CreateMessage{
				ChatId: validChatID,
				Text:   "",
			},
			wantErr: true,
			errMsg:  "text is required and cannot be empty",
		},
		{
			name: "whitespace only text",
			input: &gen.CreateMessage{
				ChatId: validChatID,
				Text:   "   ",
			},
			wantErr: true,
			errMsg:  "text is required and cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateCreateMessage(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateEditMessage(t *testing.T) {
	validMessageID := uuid.New().String()

	tests := []struct {
		name    string
		input   *gen.EditMessage
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid edit message",
			input: &gen.EditMessage{
				MessageId: validMessageID,
				Text:      "Updated text",
			},
			wantErr: false,
		},
		{
			name:    "nil message",
			input:   nil,
			wantErr: true,
			errMsg:  "edit_message is required",
		},
		{
			name: "empty message_id",
			input: &gen.EditMessage{
				MessageId: "",
				Text:      "Updated text",
			},
			wantErr: true,
			errMsg:  "message_id is required",
		},
		{
			name: "invalid message_id format",
			input: &gen.EditMessage{
				MessageId: "invalid-uuid",
				Text:      "Updated text",
			},
			wantErr: true,
			errMsg:  "message_id must be a valid uuid",
		},
		{
			name: "empty text",
			input: &gen.EditMessage{
				MessageId: validMessageID,
				Text:      "",
			},
			wantErr: true,
			errMsg:  "text is required and cannot be empty",
		},
		{
			name: "whitespace only text",
			input: &gen.EditMessage{
				MessageId: validMessageID,
				Text:      "   ",
			},
			wantErr: true,
			errMsg:  "text is required and cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateEditMessage(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateDeleteMessage(t *testing.T) {
	validMessageID := uuid.New().String()

	tests := []struct {
		name    string
		input   *gen.DeleteMessage
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid delete message",
			input: &gen.DeleteMessage{
				MessageId: validMessageID,
			},
			wantErr: false,
		},
		{
			name:    "nil message",
			input:   nil,
			wantErr: true,
			errMsg:  "delete_message is required",
		},
		{
			name: "empty message_id",
			input: &gen.DeleteMessage{
				MessageId: "",
			},
			wantErr: true,
			errMsg:  "message_id is required",
		},
		{
			name: "invalid message_id format",
			input: &gen.DeleteMessage{
				MessageId: "invalid-uuid",
			},
			wantErr: true,
			errMsg:  "message_id must be a valid uuid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateDeleteMessage(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateSendMessageReq(t *testing.T) {
	validChatID := uuid.New().String()

	tests := []struct {
		name    string
		input   *gen.CreateMessage
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid send message request",
			input: &gen.CreateMessage{
				ChatId: validChatID,
				Text:   "Test message",
			},
			wantErr: false,
		},
		{
			name: "empty chat_id",
			input: &gen.CreateMessage{
				ChatId: "",
				Text:   "Test message",
			},
			wantErr: true,
			errMsg:  "chat_id is required",
		},
		{
			name: "invalid chat_id format",
			input: &gen.CreateMessage{
				ChatId: "not-a-uuid",
				Text:   "Test message",
			},
			wantErr: true,
			errMsg:  "chat_id must be a valid uuid",
		},
		{
			name: "empty text",
			input: &gen.CreateMessage{
				ChatId: validChatID,
				Text:   "",
			},
			wantErr: true,
			errMsg:  "text is required and cannot be empty",
		},
		{
			name: "whitespace only text",
			input: &gen.CreateMessage{
				ChatId: validChatID,
				Text:   "   ",
			},
			wantErr: true,
			errMsg:  "text is required and cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateSendMessageReq(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateEditMessageReq(t *testing.T) {
	validMessageID := uuid.New().String()

	tests := []struct {
		name    string
		input   *gen.EditMessage
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid edit message request",
			input: &gen.EditMessage{
				MessageId: validMessageID,
				Text:      "Updated message",
			},
			wantErr: false,
		},
		{
			name: "empty message_id",
			input: &gen.EditMessage{
				MessageId: "",
				Text:      "Updated message",
			},
			wantErr: true,
			errMsg:  "message_id is required",
		},
		{
			name: "invalid message_id format",
			input: &gen.EditMessage{
				MessageId: "not-a-uuid",
				Text:      "Updated message",
			},
			wantErr: true,
			errMsg:  "message_id must be a valid uuid",
		},
		{
			name: "empty text",
			input: &gen.EditMessage{
				MessageId: validMessageID,
				Text:      "",
			},
			wantErr: true,
			errMsg:  "text is required and cannot be empty",
		},
		{
			name: "whitespace only text",
			input: &gen.EditMessage{
				MessageId: validMessageID,
				Text:      "   ",
			},
			wantErr: true,
			errMsg:  "text is required and cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateEditMessageReq(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
