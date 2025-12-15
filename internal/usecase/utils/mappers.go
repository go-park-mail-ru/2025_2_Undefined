package utils

import (
	"context"

	"github.com/go-park-mail-ru/2025_2_Undefined/internal/models/domains"

	modelsAttachment "github.com/go-park-mail-ru/2025_2_Undefined/internal/models/attachment"
	modelsMessage "github.com/go-park-mail-ru/2025_2_Undefined/internal/models/message"
	dtoMessage "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/dto/message"
	interfaceFileStorage "github.com/go-park-mail-ru/2025_2_Undefined/internal/usecase/interface/storage"
)

// ConvertMessageToDTO преобразует модель Message в MessageDTO с вложениями
func ConvertMessageToDTO(ctx context.Context, msg modelsMessage.Message, fileStorage interfaceFileStorage.FileStorage) dtoMessage.MessageDTO {
	var attachmentDTO *dtoMessage.AttachmentDTO

	if msg.Attachment != nil {
		var fileURL string

		// Для стикеров используем FileName (ID стикера), для остальных - URL из MinIO
		if msg.Attachment.Type != nil && *msg.Attachment.Type == modelsAttachment.AttachmentTypeSticker {
			fileURL = msg.Attachment.FileName
		} else {
			attachmentURL, err := fileStorage.GetOne(ctx, &msg.Attachment.ID)
			if err != nil {
				domains.GetLogger(ctx).WithError(err).Warningf("could not get url of file with id %s", msg.Attachment.ID.String())
				attachmentURL = "" // fallback
			}
			fileURL = attachmentURL
		}

		attachmentDTO = &dtoMessage.AttachmentDTO{
			ID:       &msg.Attachment.ID,
			Type:     msg.Attachment.Type,
			FileURL:  fileURL,
			Duration: msg.Attachment.Duration,
		}
	}

	return dtoMessage.MessageDTO{
		ID:         msg.ID,
		SenderID:   msg.UserID,
		SenderName: msg.UserName,
		Text:       msg.Text,
		CreatedAt:  msg.CreatedAt,
		UpdatedAt:  msg.UpdatedAt,
		ChatID:     msg.ChatID,
		Type:       msg.Type,
		Attachment: attachmentDTO,
	}
}
