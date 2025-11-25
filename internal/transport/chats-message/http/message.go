package chats

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-park-mail-ru/2025_2_Undefined/internal/models/domains"
	mappers "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/chats-message/mappers"
	dtoMessage "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/dto/message"
	gen "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/generated/chats"
	contextUtils "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/utils/context"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/utils/response"
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
		response.SendError(r.Context(), op, w, http.StatusUnauthorized, err.Error())
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		response.SendError(r.Context(), op, w, http.StatusInternalServerError, "Failed to upgrade to WebSocket")
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
		response.SendError(r.Context(), op, w, http.StatusUnauthorized, err.Error())
		return
	}

	protoReq := &gen.SearchMessagesReq{
		UserId: userID.String(),
		ChatId: chatID.String(),
		Text:   textQuery,
	}

	protoRes, err := h.messageClient.SearchMessages(r.Context(), protoReq)
	if err != nil {
		response.SendError(r.Context(), op, w, http.StatusInternalServerError, "failed to search messages")
		return
	}

	dtoRes := mappers.ProtoSearchMessagesResToDTO(protoRes)

	response.SendJSONResponse(r.Context(), op, w, http.StatusOK, dtoRes)
}
