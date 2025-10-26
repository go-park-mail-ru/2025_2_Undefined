package messages

import (
	"context"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/go-park-mail-ru/2025_2_Undefined/internal/models/domains"
	dto "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/dto/chats"
	dtoMessage "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/dto/message"
	_ "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/dto/utils"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/utils/response"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Для разработки
		// Раскомментить :)
		/* 		origin := r.Header.Get("Origin")
		   		// Разрешенные origins для разработки
		   		allowedOrigins := []string{
		   			"http://localhost:3000",
		   			"http://localhost:8080",
		   		}

		   		for _, allowed := range allowedOrigins {
		   			if origin == allowed {
		   				return true
		   			}
		   		}
		   		return false */
	},
}

type SessionUtilsI interface {
	GetUserIDFromSession(r *http.Request) (uuid.UUID, error)
}

type MessageUsecase interface {
	AddMessage(msg dtoMessage.CreateMessageDTO, userId uuid.UUID) error
	SubscribeUserToChats(ctx context.Context, userId uuid.UUID, chatsDTO []dto.ChatViewInformationDTO) <-chan dtoMessage.MessageDTO
}

type ChatsService interface {
	GetChats(userId uuid.UUID) ([]dto.ChatViewInformationDTO, error)
}

type MessageHandler struct {
	messageUsecase MessageUsecase
	chatsUsecase   ChatsService
	sessionUtils   SessionUtilsI
}

func NewMessageHandler(messageUsecase MessageUsecase, chatsUsecase ChatsService, sessionUtils SessionUtilsI) *MessageHandler {
	return &MessageHandler{
		messageUsecase: messageUsecase,
		chatsUsecase:   chatsUsecase,
		sessionUtils:   sessionUtils,
	}
}

// HandleMessages устанавливает WebSocket соединение для обмена сообщениями
// @Summary      Установить WebSocket соединение для сообщений
// @Description  Устанавливает WebSocket соединение для отправки и получения сообщений в реальном времени.
// @Description  После установки соединения клиент может отправлять сообщения в формате CreateMessageDTO
// @Description  и получать уведомления о новых сообщениях в формате MessageDTO.
// @Description
// @Description  **Протокол WebSocket:**
// @Description
// @Description  **Отправка сообщения (клиент → сервер):**
// @Description  ```json
// @Description  {
// @Description    "text": "Текст сообщения",
// @Description    "created_at": "2025-01-15T10:30:00Z",
// @Description    "chat_id": "123e4567-e89b-12d3-a456-426614174000"
// @Description  }
// @Description  ```
// @Description
// @Description  **Получение сообщения (сервер → клиент):**
// @Description  ```json
// @Description  {
// @Description    "sender_name": "Имя отправителя",
// @Description    "sender_avatar": "https://example.com/avatar.jpg",
// @Description    "text": "Текст сообщения",
// @Description    "created_at": "2025-01-15T10:30:00Z",
// @Description    "chat_id": "123e4567-e89b-12d3-a456-426614174000"
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
// @Router       /ws/messages [get]
func (h *MessageHandler) HandleMessages(w http.ResponseWriter, r *http.Request) {
	const op = "MessageHandler.HandleMessages"
	userId, err := h.sessionUtils.GetUserIDFromSession(r)
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
	go h.sendMessages(ctx, cancel, conn, userId)
	go h.readMessages(ctx, cancel, conn, userId)

	<-ctx.Done()
}

func (h *MessageHandler) readMessages(ctx context.Context, cancel context.CancelFunc, conn *websocket.Conn, userId uuid.UUID) {
	defer cancel()

	logger := domains.GetLogger(ctx).WithFields(logrus.Fields{
		"operation": "MessageHandler.readMessages",
		"user_id":   userId.String(),
	})

	chatsViewDTO, err := h.chatsUsecase.GetChats(userId)
	if err != nil {
		h.writeJSONErrorWebSocket(conn, err.Error())
		logger.WithError(err).Error("Failed to get chats for user")
		return
	}

	// Подписываемся на получение сообщений из всех чатов для данного пользователя
	ch := h.messageUsecase.SubscribeUserToChats(ctx, userId, chatsViewDTO)

	for {
		select {
		case msg, ok := <-ch:
			if !ok {
				conn.WriteMessage(websocket.CloseMessage,
					websocket.FormatCloseMessage(websocket.CloseGoingAway, "Server shutting down"))
				return
			}

			if err := conn.WriteJSON(msg); err != nil {
				logger.WithError(err).Error("Failed to write message to user")
				return
			}

		case <-ctx.Done():
			return

		case <-time.After(30 * time.Second):
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				logger.WithError(err).Error("Failed to send ping to user")
				return
			}
		}
	}
}

func (h *MessageHandler) sendMessages(ctx context.Context, cancel context.CancelFunc, conn *websocket.Conn, userId uuid.UUID) {
	defer cancel()

	logger := domains.GetLogger(ctx).WithFields(logrus.Fields{
		"operation": "MessageHandler.sendMessages",
		"user_id":   userId.String(),
	})

	for {
		var msg dtoMessage.CreateMessageDTO
		if err := conn.ReadJSON(&msg); err != nil {
			// Ошибка чтения — закрываем соединение и отменяем контекст
			if shouldCloseConnection(err) {
				logger.WithError(err).Info("WebSocket connection closing")
				return
			}
			h.writeJSONErrorWebSocket(conn, err.Error())
			logger.WithError(err).Error("WebSocket read error")
			continue
		}

		// Валидация входящего сообщения
		if msg.ChatId == uuid.Nil {
			logger.Error("chat_id is required")
			h.writeJSONErrorWebSocket(conn, "chat_id is required")
			continue
		}

		if msg.Text == "" {
			logger.Error("text is required")
			h.writeJSONErrorWebSocket(conn, "text is required")
			continue
		}

		if msg.CreatedAt.IsZero() {
			msg.CreatedAt = time.Now()
		}

		if err := h.messageUsecase.AddMessage(msg, userId); err != nil {
			logger.WithError(err).Error("Failed to add message")
			h.writeJSONErrorWebSocket(conn, err.Error())
			continue
		}
	}
}

func (h *MessageHandler) writeJSONErrorWebSocket(conn *websocket.Conn, str string) {
	_ = conn.WriteJSON(map[string]string{"error": str})
}

func shouldCloseConnection(err error) bool {
	// Все случаи когда соединение УЖЕ закрыто или должно быть закрыто
	return websocket.IsUnexpectedCloseError(err,
		websocket.CloseGoingAway,               // Клиент ушел (нормально)
		websocket.CloseAbnormalClosure,         // Соединение разорвано (авария)
		websocket.CloseInvalidFramePayloadData, // Поврежденные данные
		websocket.CloseProtocolError,           // Ошибка протокола
		websocket.CloseMessageTooBig,           // Слишком большое сообщение
		websocket.CloseUnsupportedData,         // Неподдерживаемый тип данных
		websocket.ClosePolicyViolation,         // Нарушение политики
		websocket.CloseInternalServerErr) ||    // Внутренняя ошибка сервера
		websocket.IsCloseError(err, websocket.CloseNormalClosure) || // Нормальное закрытие
		errors.Is(err, io.EOF) // TCP соединение разорвано
}
