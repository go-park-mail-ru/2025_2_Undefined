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
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/utils/response"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Для разработки
		// Раскомментируйте в продакшене:
		/*
			origin := r.Header.Get("Origin")
			allowedOrigins := []string{
				"http://localhost:3000",
				"http://localhost:8080",
			}
			for _, allowed := range allowedOrigins {
				if origin == allowed {
					return true
				}
			}
			return false
		*/
	},
}

type SessionUtilsI interface {
	GetUserIDFromSession(r *http.Request) (uuid.UUID, error)
}

type MessageUsecase interface {
	AddMessage(ctx context.Context, msg dtoMessage.CreateMessageDTO, userId uuid.UUID) error
	SubscribeUserToChats(ctx context.Context, userId uuid.UUID, chatsDTO []dto.ChatViewInformationDTO) <-chan dtoMessage.MessageDTO
}

type ChatsService interface {
	GetChats(ctx context.Context, userId uuid.UUID) ([]dto.ChatViewInformationDTO, error)
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
// @Description  ...
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

	chatsViewDTO, err := h.chatsUsecase.GetChats(ctx, userId)
	if err != nil {
		h.writeJSONErrorWebSocket(conn, err.Error())
		logger.WithError(err).Error("Failed to get chats for user")
		return
	}

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
			// Любая ошибка при чтении — завершаем горутину
			if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) ||
				websocket.IsUnexpectedCloseError(err) ||
				errors.Is(err, io.EOF) {
				logger.WithError(err).Info("WebSocket closed by client")
			} else {
				logger.WithError(err).Error("Unexpected WebSocket read error")
			}
			return // 🔥 КРИТИЧЕСКИ ВАЖНО: НЕЛЬЗЯ ЧИТАТЬ ПОВТОРНО!
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

		if err := h.messageUsecase.AddMessage(ctx, msg, userId); err != nil {
			logger.WithError(err).Error("Failed to add message")
			h.writeJSONErrorWebSocket(conn, err.Error())
			// Ошибки бизнес-логики не разрывают соединение
			continue
		}
	}
}

func (h *MessageHandler) writeJSONErrorWebSocket(conn *websocket.Conn, str string) {
	_ = conn.WriteJSON(map[string]string{"error": str})
}