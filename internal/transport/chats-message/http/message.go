package chats

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-park-mail-ru/2025_2_Undefined/internal/models/domains"
	mappers "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/chats-message/mappers"
	dtoMessage "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/dto/message"
	gen "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/generated/chats"
	contextUtils "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/utils/context"
	utils "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/utils/response"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		origin := r.Header.Get("Origin")

		// Если это режим разработки, разрешаем любой источник
		if isDevelopment := os.Getenv("ENVIRONMENT"); isDevelopment == "development" || isDevelopment == "dev" {
			return true
		}

		// В продакшене разрешаем только определенные origins
		allowedOrigins := []string{
			"https://100gramm.online",
			"https://www.100gramm.online",
			"http://localhost:3000", // для локальной разработки
			"http://localhost:8080", // для локальной разработки
		}

		for _, allowed := range allowedOrigins {
			if origin == allowed {
				return true
			}
		}
		return false
	},
}

// HandleMessages устанавливает WebSocket соединение для обмена сообщениями
// @Summary      Установить WebSocket соединение для сообщений
// @Description  Устанавливает WebSocket соединение для отправки и получения сообщений в реальном времени.
// @Description
// @Description  **Протокол WebSocket:**
// @Description
// @Description  **1. Создание нового сообщения (клиент → сервер):**
// @Description  ```json
// @Description  {
// @Description    "type": "new_message",
// @Description    "chat_id": "123e4567-e89b-12d3-a456-426614174000",
// @Description    "value": {
// @Description      "text": "Текст сообщения",
// @Description      "created_at": "2025-01-15T10:30:00Z",
// @Description      "chat_id": "123e4567-e89b-12d3-a456-426614174000"
// @Description    }
// @Description  }
// @Description  ```
// @Description
// @Description  **1.1. Создание сообщения с вложением (клиент → сервер):**
// @Description  Для отправки файлов сначала загрузите файл через POST /messages/attachment, получите attachment_id, затем:
// @Description  ```json
// @Description  {
// @Description    "type": "new_message",
// @Description    "chat_id": "123e4567-e89b-12d3-a456-426614174000",
// @Description    "value": {
// @Description      "text": "Текст к вложению (опционально)",
// @Description      "created_at": "2025-01-15T10:30:00Z",
// @Description      "chat_id": "123e4567-e89b-12d3-a456-426614174000",
// @Description      "attachment": {
// @Description        "attachment_id": "550e8400-e29b-41d4-a716-446655440000",
// @Description        "type": "image", // "image", "document", "audio", "video", "sticker", "voice", "video_note"
// @Description        "duration": 45 // для audio/voice/video_note - длительность в секундах (опционально)
// @Description      }
// @Description    }
// @Description  }
// @Description  ```
// @Description  Для стикеров используйте sticker_id вместо attachment_id. Поля attachment_id, type и file_url возвращаются из POST /messages/attachment.
// @Description
// @Description  **2. Редактирование сообщения (клиент → сервер):**
// @Description  ```json
// @Description  {
// @Description    "type": "edit_message",
// @Description    "chat_id": "123e4567-e89b-12d3-a456-426614174000",
// @Description    "value": {
// @Description      "id": "456e4567-e89b-12d3-a456-426614174001",
// @Description      "text": "Обновленный текст",
// @Description      "updated_at": "2025-01-15T10:35:00Z" // По желанию
// @Description    }
// @Description  }
// @Description  ```
// @Description
// @Description  **3. Удаление сообщения (клиент → сервер):**
// @Description  ```json
// @Description  {
// @Description    "type": "delete_message",
// @Description    "chat_id": "123e4567-e89b-12d3-a456-426614174000",
// @Description    "value": {
// @Description      "id": "456e4567-e89b-12d3-a456-426614174001"
// @Description    }
// @Description  }
// @Description  ```
// @Description
// @Description  **Получение событий (сервер → клиент):**
// @Description
// @Description  **Новое сообщение:**
// @Description  ```json
// @Description  {
// @Description    "type": "new_message",
// @Description    "chat_id": "123e4567-e89b-12d3-a456-426614174000",
// @Description    "value": {
// @Description      "id": "789e4567-e89b-12d3-a456-426614174002",
// @Description      "sender_id": "321e4567-e89b-12d3-a456-426614174003",
// @Description      "sender_name": "Иван Иванов",
// @Description      "text": "Текст сообщения",
// @Description      "created_at": "2025-01-15T10:30:00Z",
// @Description      "chat_id": "123e4567-e89b-12d3-a456-426614174000",
// @Description      "type": "user"
// @Description    }
// @Description  }
// @Description  ```
// @Description
// @Description  **Редактирование сообщения:**
// @Description  ```json
// @Description  {
// @Description    "type": "edit_message",
// @Description    "chat_id": "123e4567-e89b-12d3-a456-426614174000",
// @Description    "value": {
// @Description      "id": "456e4567-e89b-12d3-a456-426614174001",
// @Description      "text": "Обновленный текст",
// @Description      "updated_at": "2025-01-15T10:35:00Z"
// @Description    }
// @Description  }
// @Description  ```
// @Description
// @Description  **Удаление сообщения:**
// @Description  ```json
// @Description  {
// @Description    "type": "delete_message",
// @Description    "chat_id": "123e4567-e89b-12d3-a456-426614174000",
// @Description    "value": {
// @Description      "id": "456e4567-e89b-12d3-a456-426614174001"
// @Description    }
// @Description  }
// @Description  ```
// @Description
// @Description  **Создан новый чат:**
// @Description  ```json
// @Description  {
// @Description    "type": "chat_created",
// @Description    "chat_id": "123e4567-e89b-12d3-a456-426614174000",
// @Description    "type": "dialog | group | channel",
// @Description    "value": {
// @Description      "id": "123e4567-e89b-12d3-a456-426614174000",
// @Description      "name": "Название чата",
// @Description      "last_message": {
// @Description          "id": "789e4567-e89b-12d3-a456-426614174002",
// @Description           "sender_id": "321e4567-e89b-12d3-a456-426614174003",
// @Description           "sender_name": "Иван Иванов",
// @Description           "text": "Текст сообщения",
// @Description           "created_at": "2025-01-15T10:30:00Z",
// @Description           "chat_id": "123e4567-e89b-12d3-a456-426614174000",
// @Description           "type": "user"
// @Description           }
// @Description    }
// @Description  }
// @Description  ```
// @Description
// @Description  **Обработка ошибок:**
// @Description  ```json
// @Description  {
// @Description    "error": "Описание ошибки"
// @Description  }
// @Description  ```
// @Tags         messages
// @Accept       json
// @Produce      json
// @Security     Cookie
// @Success      101  "WebSocket соединение установлено"
// @Failure      401  {object}  dto.ErrorDTO  "Пользователь не авторизован"
// @Failure      500  {object}  dto.ErrorDTO  "Ошибка сервера при установке WebSocket соединения"
// @Router       /message/ws [get]
func (h *ChatsGRPCProxyHandler) HandleMessages(w http.ResponseWriter, r *http.Request) {
	const op = "ChatsGRPCProxyHandler.HandleMessages"
	userID, err := contextUtils.GetUserIDFromContext(r)
	if err != nil {
		utils.SendError(r.Context(), op, w, http.StatusUnauthorized, err.Error())
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		utils.SendError(r.Context(), op, w, http.StatusInternalServerError, err.Error())
		return
	}
	defer conn.Close()

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	// Горутины для отправки и приёма сообщений по WebSocket.
	go h.sendMessages(ctx, cancel, conn, userID)
	go h.readMessages(ctx, cancel, conn, userID)

	<-ctx.Done()
}

func (h *ChatsGRPCProxyHandler) readMessages(ctx context.Context, cancel context.CancelFunc, conn *websocket.Conn, userID uuid.UUID) {
	const op = "ChatsGRPCProxyHandler.readMessages"
	defer cancel()

	logger := domains.GetLogger(ctx).WithField("operation", op).WithField("user_id", userID.String())

	logger.Debugf("Start read messages from user %s ", userID)

	// Подписываемся на получение сообщений из всех чатов для данного пользователя
	stream, err := h.messageClient.StreamMessagesForUser(ctx, &gen.StreamMessagesForUserReq{
		UserId: userID.String(),
	})
	if err != nil {
		logger.WithError(err).Error("Error getting stream messages")
		h.writeJSONErrorWebSocket(conn, "error getting stream messages")
		return
	}

	pingTicker := time.NewTicker(30 * time.Second) // Интервал пинга
	defer pingTicker.Stop()

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-pingTicker.C:
				if err := conn.WriteControl(websocket.PingMessage, nil, time.Now().Add(10*time.Second)); err != nil {
					logger.WithError(err).Error("Failed to send ping")
					cancel()
					return
				}
			}
		}
	}()

	for {
		protoMessage, err := stream.Recv()
		if err == io.EOF {
			logger.Info("Stream ended")
			break
		}

		if err != nil {
			logger.WithError(err).Error("Error receiving message from stream")
			h.writeJSONErrorWebSocket(conn, "error receiving message from stream")
			return
		}

		dtoMessage := mappers.ProtoMessageEventResToDTO(protoMessage)
		logger.Debugf("Received message: %v from user %s", dtoMessage, userID)

		if err := conn.WriteJSON(dtoMessage); err != nil {
			logger.WithError(err).Error("Failed to write message to user")
			return
		}
	}
}

func (h *ChatsGRPCProxyHandler) sendMessages(ctx context.Context, cancel context.CancelFunc, conn *websocket.Conn, userID uuid.UUID) {
	const op = "ChatsGRPCProxyHandler.sendMessages"
	defer cancel()

	logger := domains.GetLogger(ctx).WithField("user_id", userID.String()).WithField("operation", op)

	logger.Debugf("start send messages from user %s ", userID)

	for {
		var msg dtoMessage.WebSocketMessageDTO
		if err := conn.ReadJSON(&msg); err != nil {
			// Ошибка чтения — закрываем соединение и отменяем контекст
			if shouldCloseConnection(err) {
				logger.WithError(err).Info("WebSocket connection closing")
				_ = conn.Close() // Закрываем соединение
				return
			}
			h.writeJSONErrorWebSocket(conn, err.Error())
			logger.WithError(err).Error("WebSocket read error")
			continue
		}

		_, err := h.messageClient.HandleSendMessage(ctx, mappers.DTOWebSocketMessageToProto(userID, msg))
		if err != nil {
			logger.WithError(err).Error("Failed to send message")
			h.writeJSONErrorWebSocket(conn, err.Error())
			continue
		}
	}
}

func (h *ChatsGRPCProxyHandler) writeJSONErrorWebSocket(conn *websocket.Conn, str string) {
	_ = conn.WriteJSON(map[string]string{"error": str})
}

func shouldCloseConnection(err error) bool {
	// Все случаи когда соединение УЖЕ закрыто или должно быть закрыто
	if err == nil {
		return false
	}

	if errors.Is(err, net.ErrClosed) {
		return true
	}

	if strings.Contains(err.Error(), "use of closed network connection") {
		return true
	}

	if errors.Is(err, io.EOF) {
		return true
	}

	return websocket.IsUnexpectedCloseError(err)
}

// SearchMessages выполняет поиск сообщений в чате по текстовому запросу
// @Summary      Поиск сообщений в чате
// @Description  Выполняет поиск сообщений в указанном чате по текстовому запросу.
// @Tags         messages
// @Accept       json
// @Produce      json
// @Security     Cookie
// @Param        chat_id  path      string  true  "ID чата"
// @Param        text     query     string  true  "Текстовый запрос для поиска сообщений"
// @Success      200      {array}   dto.MessageDTO  "Список найденных сообщений"
// @Failure      400      {object}  dto.ErrorDTO    "Некорректный запрос (например, отсутствует текстовый запрос)"
// @Failure      401      {object}  dto.ErrorDTO    "Пользователь не авторизован"
// @Failure      500      {object}  dto.ErrorDTO    "Ошибка сервера при поиске сообщений"
// @Router       /chats/{chat_id}/messages/search [get]
func (h *ChatsGRPCProxyHandler) SearchMessages(w http.ResponseWriter, r *http.Request) {
	const op = "ChatsGRPCProxyHandler.SearchMessages"

	vars := mux.Vars(r)
	chatIDStr := vars["chat_id"]

	chatID, err := uuid.Parse(chatIDStr)
	if err != nil {
		utils.SendError(r.Context(), op, w, http.StatusBadRequest, "bad format chat_id in query")
		return
	}

	queryValues := r.URL.Query()
	textQuery := queryValues.Get("text")
	if textQuery == "" {
		utils.SendError(r.Context(), op, w, http.StatusBadRequest, "text in query is required")
		return
	}

	userID, err := contextUtils.GetUserIDFromContext(r)
	if err != nil {
		utils.SendError(r.Context(), op, w, http.StatusUnauthorized, err.Error())
		return
	}

	protoReq := &gen.SearchMessagesReq{
		UserId: userID.String(),
		ChatId: chatID.String(),
		Text:   textQuery,
	}

	protoRes, err := h.messageClient.SearchMessages(r.Context(), protoReq)
	if err != nil {
		utils.SendError(r.Context(), op, w, http.StatusInternalServerError, "failed to search messages")
		return
	}

	dtoRes := mappers.ProtoSearchMessagesResToDTO(protoRes)

	utils.SendJSONResponse(r.Context(), op, w, http.StatusOK, dtoRes)
}

// UploadAttachment загружает файл
// @Summary      Загрузить файл
// @Description  Загружает файл для последующей отправки в сообщении. Тип вложения (image, document, audio, video) определяется автоматически по Content-Type файла.
// @Tags         messages
// @Accept       multipart/form-data
// @Produce      json
// @Param X-CSRF-Token header string true "CSRF Token"
// @Security     ApiKeyAuth
// @Param        chat_id formData string true "ID чата"
// @Param        file formData file true "Файл вложения"
// @Param        duration formData int false "Длительность в секундах (для audio/voice/video_note)"
// @Success      200  {object}  dto.AttachmentDTO  "Информация о загруженном вложении с автоматически определенным типом"
// @Failure      400  {object}  dto.ErrorDTO       "Неверный формат запроса"
// @Failure      401  {object}  dto.ErrorDTO       "Неавторизованный доступ"
// @Router       /message/attachment [post]
func (h *ChatsGRPCProxyHandler) UploadAttachment(w http.ResponseWriter, r *http.Request) {
	const op = "ChatsGRPCProxyHandler.UploadAttachment"
	logger := domains.GetLogger(r.Context()).WithField("op", op)

	userID, err := contextUtils.GetUserIDFromContext(r)
	if err != nil {
		utils.SendError(r.Context(), op, w, http.StatusUnauthorized, err.Error())
		return
	}

	if err := r.ParseMultipartForm(20 << 20); err != nil { // 20 MB
		logger.WithError(err).Error("failed to parse multipart form")
		utils.SendError(r.Context(), op, w, http.StatusBadRequest, "failed to parse form")
		return
	}

	chatIDStr := r.FormValue("chat_id")
	durationStr := r.FormValue("duration")

	chatID, err := uuid.Parse(chatIDStr)
	if err != nil {
		utils.SendError(r.Context(), op, w, http.StatusBadRequest, "chat_id is required in valid format")
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		logger.WithError(err).Error("failed to get file")
		utils.SendError(r.Context(), op, w, http.StatusBadRequest, "file is required")
		return
	}
	defer file.Close()

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, file); err != nil {
		logger.WithError(err).Error("failed to read file")
		utils.SendError(r.Context(), op, w, http.StatusInternalServerError, "failed to read file")
		return
	}

	var durationPtr *int32
	if durationStr != "" {
		d, err := strconv.ParseInt(durationStr, 10, 32)
		if err == nil {
			d32 := int32(d)
			durationPtr = &d32
		}
	}

	grpcReq := &gen.UploadAttachmentReq{
		UserId:      userID.String(),
		ChatId:      chatID.String(),
		Data:        buf.Bytes(),
		Filename:    header.Filename,
		ContentType: header.Header.Get("Content-Type"),
		Duration:    durationPtr,
	}

	response, err := h.messageClient.UploadAttachment(r.Context(), grpcReq)
	if err != nil {
		logger.WithError(err).Error("failed to upload attachment via grpc")
		utils.HandleGRPCError(r.Context(), w, err, op)
		return
	}

	dto := mappers.ProtoUploadAttachmentResToDTO(response)

	utils.SendJSONResponse(r.Context(), op, w, http.StatusOK, dto)
}
