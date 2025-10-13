package repository

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	models "github.com/go-park-mail-ru/2025_2_Undefined/internal/models/chats"
	"github.com/google/uuid"
)

const (
	getChatsQuery = `
		SELECT c.id, c.chat_type::text, c.name, c.description 
		FROM chat c
		JOIN chat_member cm ON cm.chat_id = c.id
		WHERE cm.user_id = $1`

	getLastMessagesOfChatsQuery = `
		SELECT DISTINCT ON (m.chat_id) m.id, m.chat_id, m.user_id, m.text, m.created_at, m.message_type::text
		FROM message m
		JOIN chat_member cm ON cm.chat_id = m.chat_id
		WHERE cm.user_id = $1
		ORDER BY m.chat_id, m.created_at DESC`

	getChatQuery = `
		SELECT c.id, c.chat_type::text, c.name, c.description 
		FROM chat c
		JOIN chat_member cm ON cm.chat_id = c.id
		WHERE cm.user_id = $1 AND c.id = $2`

	getUsersOfChat = `
		SELECT user_id, chat_id, chat_member_role::text
		FROM chat_member
		WHERE chat_id = $1`

	getMessagesOfChatQuery = `
		SELECT id, chat_id, user_id, text, created_at, message_type::text
		FROM message
		WHERE chat_id = $1
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $2`

	getUserInfo = `
		SELECT user_id, chat_id, chat_member_role::text
		FROM chat_member
		WHERE user_id = $1 AND chat_id = $2`
)

type ChatsRepository struct {
	db *sql.DB
}

func NewChatsRepository(db *sql.DB) *ChatsRepository {
	return &ChatsRepository{
		db: db,
	}
}

func (r *ChatsRepository) GetChats(userId uuid.UUID) ([]models.Chat, error) {
	const op = "ChatsRepository.GetChats"
	rows, err := r.db.Query(getChatsQuery, userId)
	if err != nil {
		wrappedErr := fmt.Errorf("%s: %w", op, err)
		log.Printf("Error: %v", wrappedErr)
		return nil, err
	}
	defer rows.Close()

	result := make([]models.Chat, 0)
	for rows.Next() {
		var chat models.Chat
		if err := rows.Scan(&chat.ID, &chat.Type, &chat.Name, &chat.Description); err != nil {
			wrappedErr := fmt.Errorf("%s: %w", op, err)
			log.Printf("Error: %v", wrappedErr)
			return nil, err
		}

		result = append(result, chat)
	}

	return result, nil
}

func (r *ChatsRepository) GetLastMessagesOfChats(userId uuid.UUID) ([]models.Message, error) {
	const op = "ChatsRepository.GetLastMessagesOfChats"
	rows, err := r.db.Query(getLastMessagesOfChatsQuery, userId)
	if err != nil {
		wrappedErr := fmt.Errorf("%s: %w", op, err)
		log.Printf("Error: %v", wrappedErr)
		return nil, err
	}
	defer rows.Close()

	result := make([]models.Message, 0)
	for rows.Next() {
		var message models.Message
		if err := rows.Scan(&message.ID, &message.ChatID, &message.UserID,
			&message.Text, &message.CreatedAt, &message.Type); err != nil {
			wrappedErr := fmt.Errorf("%s: %w", op, err)
			log.Printf("Error: %v", wrappedErr)
			return nil, err
		}

		result = append(result, message)
	}

	return result, nil
}

func (r *ChatsRepository) GetChat(userId, chatId uuid.UUID) (*models.Chat, error) {
	const op = "ChatsRepository.GetChat"
	chat := &models.Chat{}

	err := r.db.QueryRow(getChatQuery, userId, chatId).
		Scan(&chat.ID, &chat.Type, &chat.Name, &chat.Description)
	if err != nil {
		wrappedErr := fmt.Errorf("%s: %w", op, err)
		log.Printf("Error: %v", wrappedErr)
		return nil, err
	}

	return chat, nil
}

func (r *ChatsRepository) GetUsersOfChat(chatId uuid.UUID) ([]models.UserInfo, error) {
	const op = "ChatsRepository.GetUsersOfChat"
	rows, err := r.db.Query(getUsersOfChat, chatId)
	if err != nil {
		wrappedErr := fmt.Errorf("%s: %w", op, err)
		log.Printf("Error: %v", wrappedErr)
		return nil, err
	}
	defer rows.Close()

	result := make([]models.UserInfo, 0)
	for rows.Next() {
		var userInfo models.UserInfo
		if err := rows.Scan(&userInfo.UserID, &userInfo.ChatID, &userInfo.Role); err != nil {
			wrappedErr := fmt.Errorf("%s: %w", op, err)
			log.Printf("Error: %v", wrappedErr)
			return nil, err
		}

		result = append(result, userInfo)
	}

	return result, nil
}

func (r *ChatsRepository) GetMessagesOfChat(chatId uuid.UUID, offset, limit int) ([]models.Message, error) {
	const op = "ChatsRepository.GetMessagesOfChats"
	rows, err := r.db.Query(getMessagesOfChatQuery, chatId, offset, limit)
	if err != nil {
		wrappedErr := fmt.Errorf("%s: %w", op, err)
		log.Printf("Error: %v", wrappedErr)
		return nil, err
	}
	defer rows.Close()

	result := make([]models.Message, 0)
	for rows.Next() {
		var message models.Message
		if err := rows.Scan(&message.ID, &message.ChatID, &message.UserID,
			&message.Text, &message.CreatedAt, &message.Type); err != nil {
			wrappedErr := fmt.Errorf("%s: %w", op, err)
			log.Printf("Error: %v", wrappedErr)
			return nil, err
		}

		result = append(result, message)
	}

	return result, nil
}

func (r *ChatsRepository) CreateChat(chat models.Chat, usersInfo []models.UserInfo, usersNames []string) error {
	const op = "ChatsRepository.CreateChat"
	if len(usersInfo) != len(usersNames) || len(usersInfo) == 0 {
		err := fmt.Errorf("invalid input: usersInfo and usersNames must have the same non-zero length")
		wrappedErr := fmt.Errorf("%s: %w", op, err)
		log.Printf("Error: %v", wrappedErr)
		return wrappedErr
	}

	tx, err := r.db.Begin()
	if err != nil {
		wrappedErr := fmt.Errorf("%s: %w", op, err)
		log.Printf("Error: %v", wrappedErr)
		return wrappedErr
	}
	defer tx.Rollback()

	// 1. Вставка чата
	_, err = tx.Exec(`INSERT INTO chat (id, chat_type, name, description) 
        VALUES ($1, $2::chat_type_enum, $3, $4)`,
		chat.ID, chat.Type, chat.Name, chat.Description)
	if err != nil {
		wrappedErr := fmt.Errorf("%s: failed to insert chat: %w", op, err)
		log.Printf("Error: %v", wrappedErr)
		return wrappedErr
	}

	// 2. Вставка участников чата
	query := `INSERT INTO chat_member (user_id, chat_id, chat_member_role) VALUES `
	values := []interface{}{}
	placeholders := []string{}

	for _, userInfo := range usersInfo {
		placeholders = append(placeholders, fmt.Sprintf("($%d, $%d, $%d::chat_member_role_enum)",
			len(values)+1, len(values)+2, len(values)+3))
		values = append(values, userInfo.UserID, chat.ID, userInfo.Role)
	}

	query += strings.Join(placeholders, ", ")
	_, err = tx.Exec(query, values...)
	if err != nil {
		wrappedErr := fmt.Errorf("%s: failed to insert chat members: %w", op, err)
		log.Printf("Error: %v", wrappedErr)
		return wrappedErr
	}

	// 3. Вставка системных сообщений
	query = `INSERT INTO message (chat_id, user_id, text, message_type) VALUES `
	values = []interface{}{}
	placeholders = []string{}

	for i, userName := range usersNames {

		placeholders = append(placeholders, fmt.Sprintf("($%d, $%d, $%d, $%d::message_type_enum)",
			len(values)+1, len(values)+2, len(values)+3, len(values)+4))
		text := fmt.Sprintf("Пользователь %s вступил в чат", userName)
		values = append(values, chat.ID, usersInfo[i].UserID, text, "system")

	}

	if len(placeholders) > 0 {
		query += strings.Join(placeholders, ", ")
		_, err = tx.Exec(query, values...)
		if err != nil {
			wrappedErr := fmt.Errorf("%s: failed to insert messages: %w", op, err)
			log.Printf("Error: %v", wrappedErr)
			return wrappedErr
		}
	}

	if err := tx.Commit(); err != nil {
		wrappedErr := fmt.Errorf("%s: failed to commit transaction: %w", op, err)
		log.Printf("Error: %v", wrappedErr)
		return wrappedErr
	}

	return nil
}

func (r *ChatsRepository) GetUserInfo(userId, chatId uuid.UUID) (*models.UserInfo, error) {
	const op = "ChatsRepository.GetUserInfo"
	userInfo := &models.UserInfo{}

	err := r.db.QueryRow(getUserInfo, userId, chatId).
		Scan(&userInfo.UserID, &userInfo.ChatID, &userInfo.Role)
	if err != nil {
		wrappedErr := fmt.Errorf("%s: %w", op, err)
		log.Printf("Error: %v", wrappedErr)
		return nil, err
	}

	return userInfo, nil
}
