package messages

import (
	"context"
	"errors"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/go-park-mail-ru/2025_2_Undefined/internal/models/session"
	dto "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/dto/chats"
	dtoMessage "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/dto/message"
	sessionUtils "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/session"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/utils/response"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
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

type SessionRepository interface {
	GetSession(sessionID uuid.UUID) (*session.Session, error)
}

type MessageUsecase interface {
	AddMessage(msg dtoMessage.CreateMessageDTO, userId uuid.UUID) error
	SubscribeUserToChats(ctx context.Context, userId uuid.UUID, chatsDTO []dto.ChatViewInformationDTO) <-chan dtoMessage.MessageDTO
}

type ChatsService interface {
	GetChats(userId uuid.UUID) ([]dto.ChatViewInformationDTO, error)
}

type MessageHandler struct {
	messageUsecase    MessageUsecase
	chatsUsecase      ChatsService
	sessionRepository SessionRepository
}

func NewMessageHandler(messageUsecase MessageUsecase, chatsUsecase ChatsService, sessionRepository SessionRepository) *MessageHandler {
	return &MessageHandler{
		messageUsecase:    messageUsecase,
		chatsUsecase:      chatsUsecase,
		sessionRepository: sessionRepository,
	}
}

func (h *MessageHandler) HandleMessages(w http.ResponseWriter, r *http.Request) {
	userId, err := sessionUtils.GetUserIDFromSession(r, h.sessionRepository)
	if err != nil {
		response.SendError(w, http.StatusUnauthorized, err.Error())
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		response.SendError(w, http.StatusInternalServerError, "Failed to upgrade to WebSocket")
		return
	}
	defer conn.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Горутины для отправки и приёма сообщений по WebSocket.
	go h.sendMessages(cancel, conn, userId)
	go h.readMessages(ctx, cancel, conn, userId)

	<-ctx.Done()
}

func (h *MessageHandler) readMessages(ctx context.Context, cancel context.CancelFunc, conn *websocket.Conn, userId uuid.UUID) {
	defer cancel()

	chatsViewDTO, err := h.chatsUsecase.GetChats(userId)
	if err != nil {
		h.writeJSONErrorWebSocket(conn, err.Error())
		log.Printf("Failed to get chats for user %s: %v", userId.String(), err)
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
				log.Printf("Failed to write message to user with id %s: %v", userId.String(), err)
				return
			}

		case <-ctx.Done():
			return

		case <-time.After(30 * time.Second):
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Printf("Failed to send ping to user with id %s: %v", userId.String(), err)
				return
			}
		}
	}
}

func (h *MessageHandler) sendMessages(cancel context.CancelFunc, conn *websocket.Conn, userId uuid.UUID) {
	defer cancel()

	for {
		var msg dtoMessage.CreateMessageDTO
		if err := conn.ReadJSON(&msg); err != nil {
			// Ошибка чтения — закрываем соединение и отменяем контекст
			if shouldCloseConnection(err) {
				log.Printf("ws closing connection for user %s: %v", userId.String(), err)
				return
			}
			h.writeJSONErrorWebSocket(conn, err.Error())
			log.Printf("ws read error for user %s: %v", userId.String(), err)
			continue
		}

		// Валидация входящего сообщения
		if msg.ChatId == uuid.Nil {
			h.writeJSONErrorWebSocket(conn, "chat_id is required")
			continue
		}

		if msg.Text == "" {
			h.writeJSONErrorWebSocket(conn, "text is required")
			continue
		}

		if msg.CreatedAt.IsZero() {
			msg.CreatedAt = time.Now()
		}

		if err := h.messageUsecase.AddMessage(msg, userId); err != nil {
			log.Printf("failed to add message from user %s: %v", userId.String(), err)
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
