package redis

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	models "github.com/go-park-mail-ru/2025_2_Undefined/internal/models/session"
	"github.com/gomodule/redigo/redis"
	"github.com/google/uuid"
)

const (
	sessionPrefix      = "session"
	userSessionsPrefix = "user_sessions"
)

type SessionRepository struct {
	pool *redis.Pool
	ttl  int // время жизни сессии в секундах
}

func New(pool *redis.Pool, sessionTTL time.Duration) *SessionRepository {
	return &SessionRepository{
		pool: pool,
		ttl:  int(sessionTTL.Seconds()),
	}
}

// sessionData структура для хранения в Redis
type sessionData struct {
	UserID    uuid.UUID `json:"user_id"`
	Device    string    `json:"device"`
	CreatedAt time.Time `json:"created_at"`
	LastSeen  time.Time `json:"last_seen"`
}

func (r *SessionRepository) AddSession(userID uuid.UUID, device string) (uuid.UUID, error) {
	const op = "SessionRepository.AddSession"

	sessionID := uuid.New()
	now := time.Now()

	sessionData := sessionData{
		UserID:    userID,
		Device:    device,
		CreatedAt: now,
		LastSeen:  now,
	}

	sessionJSON, err := json.Marshal(sessionData)
	if err != nil {
		wrappedErr := fmt.Errorf("%s: failed to marshal session data: %w", op, err)
		log.Printf("Error: %v", wrappedErr)
		return uuid.Nil, wrappedErr
	}

	conn := r.pool.Get()
	defer conn.Close()

	sessionKey := fmt.Sprintf("%s:%s", sessionPrefix, sessionID.String())
	userSessionsKey := fmt.Sprintf("%s:%s", userSessionsPrefix, userID.String())

	// Используем MULTI/EXEC для транзакции
	conn.Send("MULTI")
	conn.Send("SETEX", sessionKey, r.ttl, sessionJSON)
	conn.Send("SADD", userSessionsKey, sessionID.String())
	conn.Send("EXPIRE", userSessionsKey, r.ttl)
	_, err = conn.Do("EXEC")

	if err != nil {
		wrappedErr := fmt.Errorf("%s: failed to execute redis transaction: %w", op, err)
		log.Printf("Error: %v", wrappedErr)
		return uuid.Nil, wrappedErr
	}

	return sessionID, nil
}

func (r *SessionRepository) DeleteSession(sessionID uuid.UUID) error {
	const op = "SessionRepository.DeleteSession"

	conn := r.pool.Get()
	defer conn.Close()

	sessionKey := fmt.Sprintf("%s:%s", sessionPrefix, sessionID.String())

	// Сначала получаем данные сессии, чтобы узнать user_id
	sessionJSON, err := redis.String(conn.Do("GET", sessionKey))
	if err != nil {
		if err == redis.ErrNil {
			return fmt.Errorf("%s: session not found", op)
		}
		wrappedErr := fmt.Errorf("%s: failed to get session: %w", op, err)
		log.Printf("Error: %v", wrappedErr)
		return wrappedErr
	}

	var data sessionData
	if err := json.Unmarshal([]byte(sessionJSON), &data); err != nil {
		wrappedErr := fmt.Errorf("%s: failed to unmarshal session data: %w", op, err)
		log.Printf("Error: %v", wrappedErr)
		return wrappedErr
	}

	userSessionsKey := fmt.Sprintf("%s:%s", userSessionsPrefix, data.UserID.String())

	// Удаляем сессию и убираем её из списка пользователя
	conn.Send("MULTI")
	conn.Send("DEL", sessionKey)
	conn.Send("SREM", userSessionsKey, sessionID.String())
	_, err = conn.Do("EXEC")

	if err != nil {
		wrappedErr := fmt.Errorf("%s: failed to execute redis transaction: %w", op, err)
		log.Printf("Error: %v", wrappedErr)
		return wrappedErr
	}

	return nil
}

func (r *SessionRepository) UpdateSession(sessionID uuid.UUID) error {
	const op = "SessionRepository.UpdateSession"

	conn := r.pool.Get()
	defer conn.Close()

	sessionKey := fmt.Sprintf("%s:%s", sessionPrefix, sessionID.String())

	sessionJSON, err := redis.String(conn.Do("GET", sessionKey))
	if err != nil {
		if err == redis.ErrNil {
			return fmt.Errorf("%s: session not found", op)
		}
		wrappedErr := fmt.Errorf("%s: failed to get session: %w", op, err)
		log.Printf("Error: %v", wrappedErr)
		return wrappedErr
	}

	var data sessionData
	if err := json.Unmarshal([]byte(sessionJSON), &data); err != nil {
		wrappedErr := fmt.Errorf("%s: failed to unmarshal session data: %w", op, err)
		log.Printf("Error: %v", wrappedErr)
		return wrappedErr
	}

	// Обновляем last_seen
	data.LastSeen = time.Now()

	updatedJSON, err := json.Marshal(data)
	if err != nil {
		wrappedErr := fmt.Errorf("%s: failed to marshal updated session data: %w", op, err)
		log.Printf("Error: %v", wrappedErr)
		return wrappedErr
	}

	userSessionsKey := fmt.Sprintf("%s:%s", userSessionsPrefix, data.UserID.String())

	conn.Send("MULTI")
	conn.Send("SETEX", sessionKey, r.ttl, updatedJSON)
	conn.Send("EXPIRE", userSessionsKey, r.ttl)
	_, err = conn.Do("EXEC")

	if err != nil {
		wrappedErr := fmt.Errorf("%s: failed to execute redis transaction: %w", op, err)
		log.Printf("Error: %v", wrappedErr)
		return wrappedErr
	}

	return nil
}

func (r *SessionRepository) GetSession(sessionID uuid.UUID) (*models.Session, error) {
	const op = "SessionRepository.GetSession"

	conn := r.pool.Get()
	defer conn.Close()

	sessionKey := fmt.Sprintf("%s:%s", sessionPrefix, sessionID.String())

	sessionJSON, err := redis.String(conn.Do("GET", sessionKey))
	if err != nil {
		if err == redis.ErrNil {
			return nil, fmt.Errorf("%s: session not found", op)
		}
		wrappedErr := fmt.Errorf("%s: failed to get session: %w", op, err)
		log.Printf("Error: %v", wrappedErr)
		return nil, wrappedErr
	}

	var data sessionData
	if err := json.Unmarshal([]byte(sessionJSON), &data); err != nil {
		wrappedErr := fmt.Errorf("%s: failed to unmarshal session data: %w", op, err)
		log.Printf("Error: %v", wrappedErr)
		return nil, wrappedErr
	}

	sess := &models.Session{
		ID:         sessionID,
		UserID:     data.UserID,
		Device:     data.Device,
		Created_at: data.CreatedAt,
		Last_seen:  data.LastSeen,
	}

	return sess, nil
}

func (r *SessionRepository) GetSessionsByUserID(userID uuid.UUID) ([]*models.Session, error) {
	const op = "SessionRepository.GetSessionsByUserID"

	conn := r.pool.Get()
	defer conn.Close()

	userSessionsKey := fmt.Sprintf("%s:%s", userSessionsPrefix, userID.String())

	// Получаем все ID сессий пользователя
	sessionIDs, err := redis.Strings(conn.Do("SMEMBERS", userSessionsKey))
	if err != nil {
		wrappedErr := fmt.Errorf("%s: failed to get user session IDs: %w", op, err)
		log.Printf("Error: %v", wrappedErr)
		return nil, wrappedErr
	}

	if len(sessionIDs) == 0 {
		return []*models.Session{}, nil
	}

	var sessions []*models.Session
	for _, sessionIDStr := range sessionIDs {
		sessionID, err := uuid.Parse(sessionIDStr)
		if err != nil {
			log.Printf("%s: Warning: invalid session ID %s for user %s: %v", op, sessionIDStr, userID, err)
			conn.Do("SREM", userSessionsKey, sessionIDStr)
			continue
		}

		sess, err := r.GetSession(sessionID)
		if err != nil {
			log.Printf("%s: Warning: failed to get session %s for user %s: %v", op, sessionID, userID, err)
			conn.Do("SREM", userSessionsKey, sessionIDStr)
			continue
		}

		sessions = append(sessions, sess)
	}

	return sessions, nil
}

func (r *SessionRepository) DeleteAllSessionWithoutCurrent(userID uuid.UUID, currentSessionID uuid.UUID) error {
	const op = "SessionRepository.DeleteAllSessionWithoutCurrent"

	conn := r.pool.Get()
	defer conn.Close()

	userSessionsKey := fmt.Sprintf("%s:%s", userSessionsPrefix, userID.String())

	sessionIDs, err := redis.Strings(conn.Do("SMEMBERS", userSessionsKey))
	if err != nil {
		wrappedErr := fmt.Errorf("%s: failed to get user session IDs: %w", op, err)
		log.Printf("Error: %v", wrappedErr)
		return wrappedErr
	}

	for _, sessionIDStr := range sessionIDs {
		sessionID, err := uuid.Parse(sessionIDStr)
		if err != nil {
			log.Printf("%s: Warning: invalid session ID %s for user %s: %v", op, sessionIDStr, userID, err)
			conn.Do("SREM", userSessionsKey, sessionIDStr)
			continue
		}

		if currentSessionID != sessionID {
			err = r.DeleteSession(sessionID)
			if err != nil {
				log.Printf("%s: Warning: failed to delete session %s for user %s: %v", op, sessionID, userID, err)
			}
		}
	}
	return nil
}
