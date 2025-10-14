package repository

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/go-park-mail-ru/2025_2_Undefined/internal/models/session"
	"github.com/google/uuid"
)

const (
	addSessionQuery = `
		INSERT INTO session (id, user_id, device, created_at, last_seen)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id`

	deleteSessionQuery = `
		DELETE FROM session 
		WHERE id = $1`

	updateSessionQuery = `
        UPDATE session 
        SET last_seen = $1 
        WHERE id = $2`

	getSessionQuery = `
		SELECT id, user_id, device, created_at, last_seen 
		FROM session 
		WHERE id = $1`
)

type SessionRepository struct {
	db *sql.DB
}

func New(db *sql.DB) *SessionRepository {
	return &SessionRepository{
		db: db,
	}
}

func (r *SessionRepository) AddSession(UserID uuid.UUID, device string) (uuid.UUID, error) {
	const op = "SessionRepository.AddSession"
	NewSession := &session.Session{
		ID:         uuid.New(),
		UserID:     UserID,
		Device:     device,
		Created_at: time.Now(),
		Last_seen:  time.Now(),
	}

	err := r.db.QueryRow(addSessionQuery,
		NewSession.ID, NewSession.UserID, NewSession.Device, NewSession.Created_at, NewSession.Last_seen).
		Scan(&NewSession.ID)

	if err != nil || NewSession.ID.String() == "" {
		wrappedErr := fmt.Errorf("%s: %w", op, err)
		log.Printf("Error: %v", wrappedErr)
		return uuid.Nil, wrappedErr
	}
	return NewSession.ID, nil
}

func (r *SessionRepository) DeleteSession(SessionID uuid.UUID) error {
	const op = "SessionRepository.DeleteSession"

	result, err := r.db.Exec(deleteSessionQuery, SessionID)
	if err != nil {
		wrappedErr := fmt.Errorf("%s: %w", op, err)
		log.Printf("Error: %v", wrappedErr)
		return wrappedErr
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		wrappedErr := fmt.Errorf("%s: failed to get rows affected: %w", op, err)
		log.Printf("Error: %v", wrappedErr)
		return wrappedErr
	}

	if rowsAffected == 0 {
		err := fmt.Errorf("%s: session not found", op)
		log.Printf("Error: %v", err)
		return err
	}

	return nil
}

func (r *SessionRepository) UpdateSession(sessionID uuid.UUID) error {
	const op = "SessionRepository.UpdateSession"

	_, err := r.db.Exec(updateSessionQuery, time.Now(), sessionID)
	if err != nil {
		wrappedErr := fmt.Errorf("%s: %w", op, err)
		log.Printf("Error: %v", wrappedErr)
		return wrappedErr
	}

	return nil
}

func (r *SessionRepository) GetSession(sessionID uuid.UUID) (*session.Session, error) {
	const op = "SessionRepository.GetSession"
	
	var sess session.Session
	err := r.db.QueryRow(getSessionQuery, sessionID).Scan(
		&sess.ID, &sess.UserID, &sess.Device, &sess.Created_at, &sess.Last_seen)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("%s: session not found", op)
		}
		wrappedErr := fmt.Errorf("%s: %w", op, err)
		log.Printf("Error: %v", wrappedErr)
		return nil, wrappedErr
	}
	
	return &sess, nil
}
