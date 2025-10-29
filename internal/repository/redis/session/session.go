package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/go-park-mail-ru/2025_2_Undefined/internal/models/session"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const (
	sessionPrefix      = "session"
	userSessionsPrefix = "user_sessions"
)

type SessionRepository struct {
	client *redis.Client
	ttl    time.Duration // время жизни сессии
}

func New(client *redis.Client, sessionTTL time.Duration) *SessionRepository {
	return &SessionRepository{
		client: client,
		ttl:    sessionTTL,
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

	ctx := context.Background()
	sessionKey := fmt.Sprintf("%s:%s", sessionPrefix, sessionID.String())
	userSessionsKey := fmt.Sprintf("%s:%s", userSessionsPrefix, userID.String())

	pipe := r.client.Pipeline()

	pipe.Set(ctx, sessionKey, sessionJSON, r.ttl)
	pipe.SAdd(ctx, userSessionsKey, sessionID.String())
	pipe.Expire(ctx, userSessionsKey, r.ttl)

	_, err = pipe.Exec(ctx)
	if err != nil {
		wrappedErr := fmt.Errorf("%s: failed to execute redis pipeline: %w", op, err)
		log.Printf("Error: %v", wrappedErr)
		return uuid.Nil, wrappedErr
	}

	return sessionID, nil
}

func (r *SessionRepository) DeleteSession(sessionID uuid.UUID) error {
	const op = "SessionRepository.DeleteSession"

	ctx := context.Background()
	sessionKey := fmt.Sprintf("%s:%s", sessionPrefix, sessionID.String())

	// Сначала получаем данные сессии, чтобы узнать user_id
	sessionJSON, err := r.client.Get(ctx, sessionKey).Result()
	if err != nil {
		if err == redis.Nil {
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
	pipe := r.client.Pipeline()
	pipe.Del(ctx, sessionKey)
	pipe.SRem(ctx, userSessionsKey, sessionID.String())

	_, err = pipe.Exec(ctx)
	if err != nil {
		wrappedErr := fmt.Errorf("%s: failed to execute redis pipeline: %w", op, err)
		log.Printf("Error: %v", wrappedErr)
		return wrappedErr
	}

	return nil
}

func (r *SessionRepository) UpdateSession(sessionID uuid.UUID) error {
	const op = "SessionRepository.UpdateSession"

	ctx := context.Background()
	sessionKey := fmt.Sprintf("%s:%s", sessionPrefix, sessionID.String())

	sessionJSON, err := r.client.Get(ctx, sessionKey).Result()
	if err != nil {
		if err == redis.Nil {
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

	pipe := r.client.Pipeline()

	// Обновляем данные сессии и продлеваем TTL
	pipe.Set(ctx, sessionKey, updatedJSON, r.ttl)
	pipe.Expire(ctx, userSessionsKey, r.ttl)

	_, err = pipe.Exec(ctx)
	if err != nil {
		wrappedErr := fmt.Errorf("%s: failed to execute redis pipeline: %w", op, err)
		log.Printf("Error: %v", wrappedErr)
		return wrappedErr
	}

	return nil
}

func (r *SessionRepository) GetSession(sessionID uuid.UUID) (*models.Session, error) {
	const op = "SessionRepository.GetSession"

	ctx := context.Background()
	sessionKey := fmt.Sprintf("%s:%s", sessionPrefix, sessionID.String())

	sessionJSON, err := r.client.Get(ctx, sessionKey).Result()
	if err != nil {
		if err == redis.Nil {
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

	ctx := context.Background()
	userSessionsKey := fmt.Sprintf("%s:%s", userSessionsPrefix, userID.String())

	// Получаем все ID сессий пользователя
	sessionIDs, err := r.client.SMembers(ctx, userSessionsKey).Result()
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
			r.client.SRem(ctx, userSessionsKey, sessionIDStr)
			continue
		}

		sess, err := r.GetSession(sessionID)
		if err != nil {
			log.Printf("%s: Warning: failed to get session %s for user %s: %v", op, sessionID, userID, err)
			r.client.SRem(ctx, userSessionsKey, sessionIDStr)
			continue
		}

		sessions = append(sessions, sess)
	}

	return sessions, nil
}
