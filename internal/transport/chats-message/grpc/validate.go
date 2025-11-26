package chats

import (
	"errors"
	"strings"

	gen "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/generated/chats"
	"github.com/google/uuid"
)

func validateChatCreateDTO(in *gen.CreateChatReq) error {
	name := in.GetName()
	if strings.TrimSpace(name) == "" {
		return errors.New("name is required and cannot be empty")
	}

	t := in.GetType()
	if strings.TrimSpace(t) == "" {
		return errors.New("type is required and cannot be empty")
	}

	members := in.GetMembers()
	err := validateAddMembers(members)
	if err != nil {
		return err
	}

	return nil
}

func validateAddMembers(members []*gen.AddMember) error {
	if len(members) == 0 {
		return errors.New("members field is required and cannot be empty")
	}

	userIds := make(map[uuid.UUID]bool)
	for _, member := range members {
		uidStr := member.GetUserId()
		if strings.TrimSpace(uidStr) == "" {
			return errors.New("user_id is required for all users")
		}

		uid, err := uuid.Parse(uidStr)
		if err != nil {
			return errors.New("user_id must be a valid uuid")
		}

		if userIds[uid] {
			return errors.New("duplicate user_id found in request")
		}
		userIds[uid] = true

		role := member.GetRole()
		if strings.TrimSpace(role) == "" {
			return errors.New("role is required for all users")
		}

		if err := validateRole(role); err != nil {
			return err
		}
	}

	return nil
}

func validateRole(role string) error {
	if role != "admin" && role != "writer" && role != "viewer" {
		return errors.New("role must be one of: admin, writer, viewer")
	}
	return nil
}

func validateChatUpdateReq(in *gen.UpdateChatReq) error {
	// Проверяем, что хотя бы одно поле для обновления передано
	if in.Name == nil && in.Description == nil {
		return errors.New("at least one field (name or description) must be provided for update")
	}

	if in.Name != nil {
		name := in.GetName()
		if strings.TrimSpace(name) == "" {
			return errors.New("name cannot be empty")
		}
	}

	if in.Description != nil {
		description := in.GetDescription()
		if strings.TrimSpace(description) == "" {
			return errors.New("description cannot be empty")
		}
	}

	return nil
}

func validateSendMessageReq(in *gen.CreateMessage) error {
	chatID := in.GetChatId()
	if strings.TrimSpace(chatID) == "" {
		return errors.New("chat_id is required")
	}

	if _, err := uuid.Parse(chatID); err != nil {
		return errors.New("chat_id must be a valid uuid")
	}

	text := in.GetText()
	if strings.TrimSpace(text) == "" {
		return errors.New("text is required and cannot be empty")
	}

	return nil
}

func validateEditMessageReq(in *gen.EditMessage) error {
	messageID := in.GetMessageId()
	if strings.TrimSpace(messageID) == "" {
		return errors.New("message_id is required")
	}

	if _, err := uuid.Parse(messageID); err != nil {
		return errors.New("message_id must be a valid uuid")
	}

	text := in.GetText()
	if strings.TrimSpace(text) == "" {
		return errors.New("text is required and cannot be empty")
	}

	return nil
}

func validateMessageEventReq(in *gen.MessageEventReq) error {
	if in.Event == nil {
		return errors.New("type is required")
	}

	if in.UserId == "" {
		return errors.New("user_id is required")
	}

	_, err := uuid.Parse(in.UserId)
	if err != nil {
		return errors.New("wrong format of user_id")
	}

	switch e := in.Event.(type) {
	case *gen.MessageEventReq_NewChatMessage:
		return validateCreateMessage(e.NewChatMessage)
	case *gen.MessageEventReq_EditChatMessage:
		return validateEditMessage(e.EditChatMessage)
	case *gen.MessageEventReq_DeleteChatMessage:
		return validateDeleteMessage(e.DeleteChatMessage)
	default:
		return errors.New("unknown event type")
	}
}

func validateCreateMessage(msg *gen.CreateMessage) error {
	if msg == nil {
		return errors.New("create_message is required")
	}

	chatID := msg.GetChatId()
	if strings.TrimSpace(chatID) == "" {
		return errors.New("chat_id is required")
	}

	if _, err := uuid.Parse(chatID); err != nil {
		return errors.New("chat_id must be a valid uuid")
	}

	text := msg.GetText()
	if strings.TrimSpace(text) == "" {
		return errors.New("text is required and cannot be empty")
	}

	return nil
}

func validateEditMessage(msg *gen.EditMessage) error {
	if msg == nil {
		return errors.New("edit_message is required")
	}

	messageID := msg.GetMessageId()
	if strings.TrimSpace(messageID) == "" {
		return errors.New("message_id is required")
	}

	if _, err := uuid.Parse(messageID); err != nil {
		return errors.New("message_id must be a valid uuid")
	}

	text := msg.GetText()
	if strings.TrimSpace(text) == "" {
		return errors.New("text is required and cannot be empty")
	}

	return nil
}

func validateDeleteMessage(msg *gen.DeleteMessage) error {
	if msg == nil {
		return errors.New("delete_message is required")
	}

	messageID := msg.GetMessageId()
	if strings.TrimSpace(messageID) == "" {
		return errors.New("message_id is required")
	}

	if _, err := uuid.Parse(messageID); err != nil {
		return errors.New("message_id must be a valid uuid")
	}

	return nil
}
