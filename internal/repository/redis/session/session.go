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
	const query = "ADD session"

	logger := domains.GetLogger(ctx).WithField("operation", op).WithField("user_id", userID.String())

	queryStatus := "success"
	defer func() {
		logger.Debugf("redis query: %s: status: %s", query, queryStatus)
	}()

	logger.Debugf("starting: %s", query)

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
		queryStatus = "fail"
		logger.WithError(err).Errorf("redis query: %s: marshal error: status: %s", query, queryStatus)
		return uuid.Nil, fmt.Errorf("%s: failed to marshal session data: %w", op, err)
	}

	sessionKey := fmt.Sprintf("%s:%s", sessionPrefix, sessionID.String())
	userSessionsKey := fmt.Sprintf("%s:%s", userSessionsPrefix, userID.String())

	pipe := r.client.Pipeline()

	pipe.Set(ctx, sessionKey, sessionJSON, r.ttl)
	pipe.SAdd(ctx, userSessionsKey, sessionID.String())
	pipe.Expire(ctx, userSessionsKey, r.ttl)

	_, err = pipe.Exec(ctx)
	if err != nil {
		queryStatus = "fail"
		logger.WithError(err).Errorf("redis query: %s: pipeline execution error: status: %s", query, queryStatus)
		return uuid.Nil, fmt.Errorf("%s: failed to execute redis pipeline: %w", op, err)
	}

	return sessionID, nil
}

func (r *SessionRepository) DeleteSession(ctx context.Context, sessionID uuid.UUID) error {
	const op = "SessionRepository.DeleteSession"
	const query = "DELETE session"

	logger := domains.GetLogger(ctx).WithField("operation", op).WithField("session_id", sessionID.String())

	queryStatus := "success"
	defer func() {
		logger.Debugf("redis query: %s: status: %s", query, queryStatus)
	}()

	logger.Debugf("starting: %s", query)

	sessionKey := fmt.Sprintf("%s:%s", sessionPrefix, sessionID.String())

	// Сначала получаем данные сессии, чтобы узнать user_id
	sessionJSON, err := r.client.Get(ctx, sessionKey).Result()
	if err != nil {
		if err == redis.Nil {
			queryStatus = "not found"
			logger.Debugf("redis query: %s: session not found: status: %s", query, queryStatus)
			return fmt.Errorf("%s: session not found", op)
		}
		queryStatus = "fail"
		logger.WithError(err).Errorf("redis query: %s: get session error: status: %s", query, queryStatus)
		return fmt.Errorf("%s: failed to get session: %w", op, err)
	}

	var data sessionData
	if err := json.Unmarshal([]byte(sessionJSON), &data); err != nil {
		queryStatus = "fail"
		logger.WithError(err).Errorf("redis query: %s: unmarshal error: status: %s", query, queryStatus)
		return fmt.Errorf("%s: failed to unmarshal session data: %w", op, err)
	}

	userSessionsKey := fmt.Sprintf("%s:%s", userSessionsPrefix, data.UserID.String())

	// Удаляем сессию и убираем её из списка пользователя
	pipe := r.client.Pipeline()
	pipe.Del(ctx, sessionKey)
	pipe.SRem(ctx, userSessionsKey, sessionID.String())

	_, err = pipe.Exec(ctx)
	if err != nil {
		queryStatus = "fail"
		logger.WithError(err).Errorf("redis query: %s: pipeline execution error: status: %s", query, queryStatus)
		return fmt.Errorf("%s: failed to execute redis pipeline: %w", op, err)
	}

	return nil
}

func (r *SessionRepository) UpdateSession(ctx context.Context, sessionID uuid.UUID) error {
	const op = "SessionRepository.UpdateSession"
	const query = "UPDATE session"

	logger := domains.GetLogger(ctx).WithField("operation", op).WithField("session_id", sessionID.String())

	queryStatus := "success"
	defer func() {
		logger.Debugf("redis query: %s: status: %s", query, queryStatus)
	}()

	logger.Debugf("starting: %s", query)

	sessionKey := fmt.Sprintf("%s:%s", sessionPrefix, sessionID.String())

	sessionJSON, err := r.client.Get(ctx, sessionKey).Result()
	if err != nil {
		if err == redis.Nil {
			queryStatus = "not found"
			logger.Debugf("redis query: %s: session not found: status: %s", query, queryStatus)
			return fmt.Errorf("%s: session not found", op)
		}
		queryStatus = "fail"
		logger.WithError(err).Errorf("redis query: %s: get session error: status: %s", query, queryStatus)
		return fmt.Errorf("%s: failed to get session: %w", op, err)
	}

	var data sessionData
	if err := json.Unmarshal([]byte(sessionJSON), &data); err != nil {
		queryStatus = "fail"
		logger.WithError(err).Errorf("redis query: %s: unmarshal error: status: %s", query, queryStatus)
		return fmt.Errorf("%s: failed to unmarshal session data: %w", op, err)
	}

	// Обновляем last_seen
	data.LastSeen = time.Now()

	updatedJSON, err := json.Marshal(data)
	if err != nil {
		queryStatus = "fail"
		logger.WithError(err).Errorf("redis query: %s: marshal error: status: %s", query, queryStatus)
		return fmt.Errorf("%s: failed to marshal updated session data: %w", op, err)
	}

	userSessionsKey := fmt.Sprintf("%s:%s", userSessionsPrefix, data.UserID.String())

	pipe := r.client.Pipeline()

	// Обновляем данные сессии и продлеваем TTL
	pipe.Set(ctx, sessionKey, updatedJSON, r.ttl)
	pipe.Expire(ctx, userSessionsKey, r.ttl)

	_, err = pipe.Exec(ctx)
	if err != nil {
		queryStatus = "fail"
		logger.WithError(err).Errorf("redis query: %s: pipeline execution error: status: %s", query, queryStatus)
		return fmt.Errorf("%s: failed to execute redis pipeline: %w", op, err)
	}

	return nil
}

func (r *SessionRepository) GetSession(ctx context.Context, sessionID uuid.UUID) (*models.Session, error) {
	const op = "SessionRepository.GetSession"
	const query = "GET session"

	logger := domains.GetLogger(ctx).WithField("operation", op).WithField("session_id", sessionID.String())

	queryStatus := "success"
	defer func() {
		logger.Debugf("redis query: %s: status: %s", query, queryStatus)
	}()

	logger.Debugf("starting: %s", query)

	sessionKey := fmt.Sprintf("%s:%s", sessionPrefix, sessionID.String())

	sessionJSON, err := r.client.Get(ctx, sessionKey).Result()
	if err != nil {
		if err == redis.Nil {
			queryStatus = "not found"
			logger.Debugf("redis query: %s: session not found: status: %s", query, queryStatus)
			return nil, fmt.Errorf("%s: session not found", op)
		}
		queryStatus = "fail"
		logger.WithError(err).Errorf("redis query: %s: get session error: status: %s", query, queryStatus)
		return nil, fmt.Errorf("%s: failed to get session: %w", op, err)
	}

	var data sessionData
	if err := json.Unmarshal([]byte(sessionJSON), &data); err != nil {
		queryStatus = "fail"
		logger.WithError(err).Errorf("redis query: %s: unmarshal error: status: %s", query, queryStatus)
		return nil, fmt.Errorf("%s: failed to unmarshal session data: %w", op, err)
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
	const query = "GET sessions by user"

	logger := domains.GetLogger(ctx).WithField("operation", op).WithField("user_id", userID.String())

	queryStatus := "success"
	defer func() {
		logger.Debugf("redis query: %s: status: %s", query, queryStatus)
	}()

	logger.Debugf("starting: %s", query)

	userSessionsKey := fmt.Sprintf("%s:%s", userSessionsPrefix, userID.String())

	// Получаем все ID сессий пользователя
	sessionIDs, err := r.client.SMembers(ctx, userSessionsKey).Result()
	if err != nil {
		queryStatus = "fail"
		logger.WithError(err).Errorf("redis query: %s: get session IDs error: status: %s", query, queryStatus)
		return nil, fmt.Errorf("%s: failed to get user session IDs: %w", op, err)
	}

	if len(sessionIDs) == 0 {
		logger.Debugf("redis query: %s: no sessions found: status: %s", query, queryStatus)
		return []*models.Session{}, nil
	}

	var sessions []*models.Session
	for _, sessionIDStr := range sessionIDs {
		sessionID, err := uuid.Parse(sessionIDStr)
		if err != nil {
			logger.WithError(err).Warnf("redis query: %s: invalid session ID %s: removing from set", query, sessionIDStr)
			r.client.SRem(ctx, userSessionsKey, sessionIDStr)
			continue
		}

		sess, err := r.GetSession(ctx, sessionID)
		if err != nil {
			logger.WithError(err).Warnf("redis query: %s: failed to get session %s: removing from set", query, sessionID)
			r.client.SRem(ctx, userSessionsKey, sessionIDStr)
			continue
		}

		sessions = append(sessions, sess)
	}

	return sessions, nil
}

func (r *SessionRepository) DeleteAllSessionWithoutCurrent(ctx context.Context, userID uuid.UUID, currentSessionID uuid.UUID) error {
	const op = "SessionRepository.DeleteAllSessionWithoutCurrent"
	const query = "DELETE sessions except current"

	logger := domains.GetLogger(ctx).WithField("operation", op).
		WithField("user_id", userID.String()).
		WithField("current_session_id", currentSessionID.String())

	queryStatus := "success"
	defer func() {
		logger.Debugf("redis query: %s: status: %s", query, queryStatus)
	}()

	logger.Debugf("starting: %s", query)

	userSessionsKey := fmt.Sprintf("%s:%s", userSessionsPrefix, userID.String())

	sessionIDs, err := r.client.SMembers(ctx, userSessionsKey).Result()
	if err != nil {
		queryStatus = "fail"
		logger.WithError(err).Errorf("redis query: %s: get session IDs error: status: %s", query, queryStatus)
		return fmt.Errorf("%s: failed to get user session IDs: %w", op, err)
	}

	for _, sessionIDStr := range sessionIDs {
		sessionID, err := uuid.Parse(sessionIDStr)
		if err != nil {
			logger.WithError(err).Warnf("redis query: %s: invalid session ID %s: removing from set", query, sessionIDStr)
			r.client.SRem(ctx, userSessionsKey, sessionIDStr)
			continue
		}

		if currentSessionID != sessionID {
			err = r.DeleteSession(ctx, sessionID)
			if err != nil {
				logger.WithError(err).Warnf("redis query: %s: failed to delete session %s", query, sessionID)
			}
		}
	}

	return nil
}
