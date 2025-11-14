package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-park-mail-ru/2025_2_Undefined/internal/models/domains"
	models "github.com/go-park-mail-ru/2025_2_Undefined/internal/models/session"
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

func (r *SessionRepository) AddSession(ctx context.Context, userID uuid.UUID, device string) (uuid.UUID, error) {
	const op = "SessionRepository.AddSession"

	logger := domains.GetLogger(ctx).WithField("operation", op)

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
		logger.WithError(wrappedErr).Error("failed to marshal session data")
		return uuid.Nil, wrappedErr
	}

	sessionKey := fmt.Sprintf("%s:%s", sessionPrefix, sessionID.String())
	userSessionsKey := fmt.Sprintf("%s:%s", userSessionsPrefix, userID.String())

	pipe := r.client.Pipeline()

	pipe.Set(ctx, sessionKey, sessionJSON, r.ttl)
	pipe.SAdd(ctx, userSessionsKey, sessionID.String())
	pipe.Expire(ctx, userSessionsKey, r.ttl)

	_, err = pipe.Exec(ctx)
	if err != nil {
		wrappedErr := fmt.Errorf("%s: failed to execute redis pipeline: %w", op, err)
		logger.WithError(wrappedErr).Error("failed to execute redis pipeline")
		return uuid.Nil, wrappedErr
	}

	return sessionID, nil
}

func (r *SessionRepository) DeleteSession(ctx context.Context, sessionID uuid.UUID) error {
	const op = "SessionRepository.DeleteSession"

	logger := domains.GetLogger(ctx).WithField("operation", op)
	sessionKey := fmt.Sprintf("%s:%s", sessionPrefix, sessionID.String())

	// Сначала получаем данные сессии, чтобы узнать user_id
	sessionJSON, err := r.client.Get(ctx, sessionKey).Result()
	if err != nil {
		if err == redis.Nil {
			return fmt.Errorf("%s: session not found", op)
		}
		wrappedErr := fmt.Errorf("%s: failed to get session: %w", op, err)
		logger.WithError(wrappedErr).Error("Redis operation failed")
		return wrappedErr
	}

	var data sessionData
	if err := json.Unmarshal([]byte(sessionJSON), &data); err != nil {
		wrappedErr := fmt.Errorf("%s: failed to unmarshal session data: %w", op, err)
		logger.WithError(wrappedErr).Error("Redis operation failed")
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
		logger.WithError(wrappedErr).Error("Redis operation failed")
		return wrappedErr
	}

	return nil
}

func (r *SessionRepository) UpdateSession(ctx context.Context, sessionID uuid.UUID) error {
	const op = "SessionRepository.UpdateSession"

	logger := domains.GetLogger(ctx).WithField("operation", op)
	sessionKey := fmt.Sprintf("%s:%s", sessionPrefix, sessionID.String())

	sessionJSON, err := r.client.Get(ctx, sessionKey).Result()
	if err != nil {
		if err == redis.Nil {
			return fmt.Errorf("%s: session not found", op)
		}
		wrappedErr := fmt.Errorf("%s: failed to get session: %w", op, err)
		logger.WithError(wrappedErr).Error("Redis operation failed")
		return wrappedErr
	}

	var data sessionData
	if err := json.Unmarshal([]byte(sessionJSON), &data); err != nil {
		wrappedErr := fmt.Errorf("%s: failed to unmarshal session data: %w", op, err)
		logger.WithError(wrappedErr).Error("Redis operation failed")
		return wrappedErr
	}

	// Обновляем last_seen
	data.LastSeen = time.Now()

	updatedJSON, err := json.Marshal(data)
	if err != nil {
		wrappedErr := fmt.Errorf("%s: failed to marshal updated session data: %w", op, err)
		logger.WithError(wrappedErr).Error("Redis operation failed")
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
		logger.WithError(wrappedErr).Error("Redis operation failed")
		return wrappedErr
	}

	return nil
}

func (r *SessionRepository) GetSession(ctx context.Context, sessionID uuid.UUID) (*models.Session, error) {
	const op = "SessionRepository.GetSession"

	logger := domains.GetLogger(ctx).WithField("operation", op)
	sessionKey := fmt.Sprintf("%s:%s", sessionPrefix, sessionID.String())

	sessionJSON, err := r.client.Get(ctx, sessionKey).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("%s: session not found", op)
		}
		wrappedErr := fmt.Errorf("%s: failed to get session: %w", op, err)
		logger.WithError(wrappedErr).Error("Redis operation failed")
		return nil, wrappedErr
	}

	var data sessionData
	if err := json.Unmarshal([]byte(sessionJSON), &data); err != nil {
		wrappedErr := fmt.Errorf("%s: failed to unmarshal session data: %w", op, err)
		logger.WithError(wrappedErr).Error("Redis operation failed")
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

func (r *SessionRepository) GetSessionsByUserID(ctx context.Context, userID uuid.UUID) ([]*models.Session, error) {
	const op = "SessionRepository.GetSessionsByUserID"

	logger := domains.GetLogger(ctx).WithField("operation", op)
	userSessionsKey := fmt.Sprintf("%s:%s", userSessionsPrefix, userID.String())

	// Получаем все ID сессий пользователя
	sessionIDs, err := r.client.SMembers(ctx, userSessionsKey).Result()
	if err != nil {
		wrappedErr := fmt.Errorf("%s: failed to get user session IDs: %w", op, err)
		logger.WithError(wrappedErr).Error("Redis operation failed")
		return nil, wrappedErr
	}

	if len(sessionIDs) == 0 {
		return []*models.Session{}, nil
	}

	var sessions []*models.Session
	for _, sessionIDStr := range sessionIDs {
		sessionID, err := uuid.Parse(sessionIDStr)
		if err != nil {
			logger.WithError(err).Warnf("invalid session ID %s for user %s", sessionIDStr, userID)
			r.client.SRem(ctx, userSessionsKey, sessionIDStr)
			continue
		}

		sess, err := r.GetSession(ctx, sessionID)
		if err != nil {
			logger.WithError(err).Warnf("failed to get session %s for user %s", sessionID, userID)
			r.client.SRem(ctx, userSessionsKey, sessionIDStr)
			continue
		}

		sessions = append(sessions, sess)
	}

	return sessions, nil
}

func (r *SessionRepository) DeleteAllSessionWithoutCurrent(ctx context.Context, userID uuid.UUID, currentSessionID uuid.UUID) error {
	const op = "SessionRepository.DeleteAllSessionWithoutCurrent"

	logger := domains.GetLogger(ctx).WithField("operation", op)
	userSessionsKey := fmt.Sprintf("%s:%s", userSessionsPrefix, userID.String())

	sessionIDs, err := r.client.SMembers(ctx, userSessionsKey).Result()
	if err != nil {
		wrappedErr := fmt.Errorf("%s: failed to get user session IDs: %w", op, err)
		logger.WithError(wrappedErr).Error("Redis operation failed")
		return wrappedErr
	}

	for _, sessionIDStr := range sessionIDs {
		sessionID, err := uuid.Parse(sessionIDStr)
		if err != nil {
			logger.WithError(err).Warnf("invalid session ID %s for user %s", sessionIDStr, userID)
			r.client.SRem(ctx, userSessionsKey, sessionIDStr)
			continue
		}

		if currentSessionID != sessionID {
			err = r.DeleteSession(ctx, sessionID)
			if err != nil {
				logger.WithError(err).Warnf("failed to delete session %s for user %s", sessionID, userID)
			}
		}
	}
	return nil
}
