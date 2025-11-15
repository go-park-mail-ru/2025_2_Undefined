package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	appealModels "github.com/go-park-mail-ru/2025_2_Undefined/internal/models/appeal"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/models/domains"
	"github.com/google/uuid"
)

const (
	createAppealQuery = `
		INSERT INTO appeal (user_id, title, status, category, assigned_to) 
		VALUES ($1, $2, $3, $4, $5) 
		RETURNING id, created_at, updated_at`

	createAnonymousAppealQuery = `
		INSERT INTO appeal (anonym_id, anonym_email, anonym_contact, title, status, category, assigned_to) 
		VALUES ($1, $2, $3, $4, $5, $6, $7) 
		RETURNING id, created_at, updated_at`

	getAppealByIDQuery = `
		SELECT id, user_id, anonym_id, anonym_email, anonym_contact, title, status, category, assigned_to, created_at, updated_at 
		FROM appeal 
		WHERE id = $1`

	getAppealByAnonymIDQuery = `
		SELECT id, user_id, anonym_id, anonym_email, anonym_contact, title, status, category, assigned_to, created_at, updated_at 
		FROM appeal 
		WHERE anonym_id = $1 
		ORDER BY created_at DESC`

	getAppealsByUserIDQuery = `
		SELECT id, user_id, anonym_id, anonym_email, anonym_contact, title, status, category, assigned_to, created_at, updated_at 
		FROM appeal 
		WHERE user_id = $1 
		ORDER BY created_at DESC`

	getAppealsWithFiltersQuery = `
		SELECT id, user_id, anonym_id, anonym_email, anonym_contact, title, status, category, assigned_to, created_at, updated_at 
		FROM appeal 
		WHERE ($1 = '' OR status = $1) 
		  AND ($2 = '' OR category = $2)
		ORDER BY created_at DESC 
		LIMIT $3 OFFSET $4`

	updateAppealStatusQuery = `
		UPDATE appeal 
		SET status = $2, updated_at = NOW() 
		WHERE id = $1`

	updateAppealCategoryQuery = `
		UPDATE appeal 
		SET category = $2, updated_at = NOW() 
		WHERE id = $1`

	updateAppealTitleQuery = `
		UPDATE appeal 
		SET title = $2, updated_at = NOW() 
		WHERE id = $1`

	updateAppealAssignmentQuery = `
		UPDATE appeal 
		SET assigned_to = $2, updated_at = NOW() 
		WHERE id = $1`

	getAppealsByAssignedRoleQuery = `
		SELECT id, user_id, anonym_id, anonym_email, anonym_contact, title, status, category, assigned_to, created_at, updated_at 
		FROM appeal 
		WHERE assigned_to = $1 
		ORDER BY created_at DESC 
		LIMIT $2 OFFSET $3`

	createAppealMessageQuery = `
		INSERT INTO message_appeal (appeal_id, text, sender_id) 
		VALUES ($1, $2, $3) 
		RETURNING id, created_at, updated_at`

	createAnonymousAppealMessageQuery = `
		INSERT INTO message_appeal (appeal_id, text, sender_anonym_id) 
		VALUES ($1, $2, $3) 
		RETURNING id, created_at, updated_at`

	getAppealMessagesQuery = `
		SELECT id, appeal_id, text, sender_id, sender_anonym_id, created_at, updated_at 
		FROM message_appeal 
		WHERE appeal_id = $1 
		ORDER BY created_at ASC`

	getUserRoleQuery = `
		SELECT role 
		FROM appeal_roles 
		WHERE user_id = $1`

	updateUserRoleQuery = `
		INSERT INTO appeal_roles (user_id, role) 
		VALUES ($1, $2) 
		ON CONFLICT (user_id) 
		DO UPDATE SET role = $2, updated_at = NOW()`

	deleteUserRoleQuery = `
		DELETE FROM appeal_roles 
		WHERE user_id = $1`

	getAppealsByRoleQuery = `
		SELECT id, user_id, anonym_id, anonym_email, anonym_contact, title, status, category, created_at, updated_at 
		FROM appeal 
		ORDER BY created_at DESC 
		LIMIT $1 OFFSET $2`

	assignAppealToSupportQuery = `
		UPDATE appeal 
		SET status = $2, updated_at = NOW() 
		WHERE id = $1`

	// Статистика обращений
	getAppealsStatsByStatusQuery = `
		SELECT status, COUNT(*) as count 
		FROM appeal 
		GROUP BY status`

	getAppealsStatsByCategoryQuery = `
		SELECT category, COUNT(*) as count 
		FROM appeal 
		GROUP BY category`

	getAppealsStatsByDateRangeQuery = `
		SELECT DATE(created_at) as date, COUNT(*) as count 
		FROM appeal 
		WHERE created_at >= $1 AND created_at <= $2 
		GROUP BY DATE(created_at) 
		ORDER BY date`

	getAppealsStatsTotalQuery = `
		SELECT 
			COUNT(*) as total_appeals,
			COUNT(CASE WHEN status = 'open' THEN 1 END) as open_appeals,
			COUNT(CASE WHEN status = 'in_work' THEN 1 END) as in_work_appeals,
			COUNT(CASE WHEN status = 'closed' THEN 1 END) as closed_appeals,
			COUNT(CASE WHEN category = 'bug' THEN 1 END) as bug_reports,
			COUNT(CASE WHEN category = 'feature' THEN 1 END) as feature_requests,
			COUNT(CASE WHEN category = 'claim' THEN 1 END) as claims,
			COUNT(CASE WHEN category = 'other' THEN 1 END) as others
		FROM appeal`
)

type AppealRepository struct {
	db *sql.DB
}

func New(db *sql.DB) *AppealRepository {
	return &AppealRepository{
		db: db,
	}
}

func (r *AppealRepository) CreateAppeal(ctx context.Context, appeal *appealModels.Appeal) error {
	const op = "AppealRepository.CreateAppeal"

	logger := domains.GetLogger(ctx).WithField("operation", op)
	logger.Debug("Starting database operation: create appeal")

	if appeal.UserID != nil {
		// Создание обращения для авторизованного пользователя
		err := r.db.QueryRowContext(ctx, createAppealQuery, appeal.UserID, appeal.Title, appeal.Status, appeal.Category, appeal.AssignedTo).
			Scan(&appeal.ID, &appeal.CreatedAt, &appeal.UpdatedAt)

		if err != nil {
			logger.WithError(err).Error("Database operation failed: create appeal query")
			return err
		}
	} else {
		// Создание анонимного обращения
		err := r.db.QueryRowContext(ctx, createAnonymousAppealQuery,
			appeal.AnonymID, appeal.AnonymEmail, appeal.AnonymContact, appeal.Title, appeal.Status, appeal.Category, appeal.AssignedTo).
			Scan(&appeal.ID, &appeal.CreatedAt, &appeal.UpdatedAt)

		if err != nil {
			logger.WithError(err).Error("Database operation failed: create anonymous appeal query")
			return err
		}
	}

	logger.WithField("appeal_id", appeal.ID.String()).Info("Database operation completed successfully: appeal created")
	return nil
}

func (r *AppealRepository) GetAppealByID(ctx context.Context, id uuid.UUID) (*appealModels.Appeal, error) {
	const op = "AppealRepository.GetAppealByID"

	logger := domains.GetLogger(ctx).WithField("operation", op).WithField("appeal_id", id.String())
	logger.Debug("Starting database operation: get appeal by ID")

	var appeal appealModels.Appeal
	err := r.db.QueryRowContext(ctx, getAppealByIDQuery, id).
		Scan(&appeal.ID, &appeal.UserID, &appeal.AnonymID, &appeal.AnonymEmail, &appeal.AnonymContact,
			&appeal.Title, &appeal.Status, &appeal.Category, &appeal.AssignedTo, &appeal.CreatedAt, &appeal.UpdatedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			logger.Debug("Database operation completed: appeal not found")
			return nil, errors.New("appeal not found")
		}
		logger.WithError(err).Error("Database operation failed: get appeal by ID query")
		return nil, err
	}

	logger.Info("Database operation completed successfully: appeal found by ID")
	return &appeal, nil
}

func (r *AppealRepository) GetAppealsByUserID(ctx context.Context, userID uuid.UUID) ([]*appealModels.Appeal, error) {
	const op = "AppealRepository.GetAppealsByUserID"

	logger := domains.GetLogger(ctx).WithField("operation", op).WithField("user_id", userID.String())
	logger.Debug("Starting database operation: get appeals by user ID")

	rows, err := r.db.QueryContext(ctx, getAppealsByUserIDQuery, userID)
	if err != nil {
		logger.WithError(err).Error("Database operation failed: get appeals by user ID query")
		return nil, err
	}
	defer rows.Close()

	var appeals []*appealModels.Appeal
	for rows.Next() {
		var appeal appealModels.Appeal
		err := rows.Scan(&appeal.ID, &appeal.UserID, &appeal.AnonymID, &appeal.AnonymEmail, &appeal.AnonymContact,
			&appeal.Title, &appeal.Status, &appeal.Category, &appeal.AssignedTo, &appeal.CreatedAt, &appeal.UpdatedAt)
		if err != nil {
			logger.WithError(err).Error("Database operation failed: scan appeal row")
			return nil, err
		}
		appeals = append(appeals, &appeal)
	}

	if err = rows.Err(); err != nil {
		logger.WithError(err).Error("Database operation failed: rows iteration error")
		return nil, err
	}

	logger.WithField("appeals_count", len(appeals)).Info("Database operation completed successfully: appeals retrieved by user ID")
	return appeals, nil
}

func (r *AppealRepository) GetAppealsWithFilters(ctx context.Context, status, category string, limit, offset int) ([]*appealModels.Appeal, error) {
	const op = "AppealRepository.GetAppealsWithFilters"

	logger := domains.GetLogger(ctx).WithField("operation", op).
		WithField("status", status).WithField("category", category).
		WithField("limit", limit).WithField("offset", offset)
	logger.Debug("Starting database operation: get appeals with filters")

	rows, err := r.db.QueryContext(ctx, getAppealsWithFiltersQuery, status, category, limit, offset)
	if err != nil {
		logger.WithError(err).Error("Database operation failed: get appeals with filters query")
		return nil, err
	}
	defer rows.Close()

	var appeals []*appealModels.Appeal
	for rows.Next() {
		var appeal appealModels.Appeal
		err := rows.Scan(&appeal.ID, &appeal.UserID, &appeal.AnonymID, &appeal.AnonymEmail, &appeal.AnonymContact,
			&appeal.Title, &appeal.Status, &appeal.Category, &appeal.AssignedTo, &appeal.CreatedAt, &appeal.UpdatedAt)
		if err != nil {
			logger.WithError(err).Error("Database operation failed: scan appeal row")
			return nil, err
		}
		appeals = append(appeals, &appeal)
	}

	if err = rows.Err(); err != nil {
		logger.WithError(err).Error("Database operation failed: rows iteration error")
		return nil, err
	}

	logger.WithField("appeals_count", len(appeals)).Info("Database operation completed successfully: appeals retrieved with filters")
	return appeals, nil
}

func (r *AppealRepository) UpdateAppealStatus(ctx context.Context, id uuid.UUID, status string) error {
	const op = "AppealRepository.UpdateAppealStatus"

	logger := domains.GetLogger(ctx).WithField("operation", op).
		WithField("appeal_id", id.String()).WithField("status", status)
	logger.Debug("Starting database operation: update appeal status")

	result, err := r.db.ExecContext(ctx, updateAppealStatusQuery, id, status)
	if err != nil {
		logger.WithError(err).Error("Database operation failed: update appeal status query")
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logger.WithError(err).Error("Database operation failed: get rows affected")
		return err
	}

	if rowsAffected == 0 {
		logger.Debug("Database operation completed: appeal not found for status update")
		return errors.New("appeal not found")
	}

	logger.Info("Database operation completed successfully: appeal status updated")
	return nil
}

func (r *AppealRepository) UpdateAppealCategory(ctx context.Context, id uuid.UUID, category string) error {
	const op = "AppealRepository.UpdateAppealCategory"

	logger := domains.GetLogger(ctx).WithField("operation", op).
		WithField("appeal_id", id.String()).WithField("category", category)
	logger.Debug("Starting database operation: update appeal category")

	result, err := r.db.ExecContext(ctx, updateAppealCategoryQuery, id, category)
	if err != nil {
		logger.WithError(err).Error("Database operation failed: update appeal category query")
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logger.WithError(err).Error("Database operation failed: get rows affected")
		return err
	}

	if rowsAffected == 0 {
		logger.Debug("Database operation completed: appeal not found for category update")
		return errors.New("appeal not found")
	}

	logger.Info("Database operation completed successfully: appeal category updated")
	return nil
}

func (r *AppealRepository) UpdateAppealTitle(ctx context.Context, id uuid.UUID, title string) error {
	const op = "AppealRepository.UpdateAppealTitle"

	logger := domains.GetLogger(ctx).WithField("operation", op).
		WithField("appeal_id", id.String()).WithField("title", title)
	logger.Debug("Starting database operation: update appeal title")

	result, err := r.db.ExecContext(ctx, updateAppealTitleQuery, id, title)
	if err != nil {
		logger.WithError(err).Error("Database operation failed: update appeal title query")
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logger.WithError(err).Error("Database operation failed: get rows affected")
		return err
	}

	if rowsAffected == 0 {
		logger.Debug("Database operation completed: appeal not found for title update")
		return errors.New("appeal not found")
	}

	logger.Info("Database operation completed successfully: appeal title updated")
	return nil
}

func (r *AppealRepository) CreateAppealMessage(ctx context.Context, message *appealModels.MessageAppeal) error {
	const op = "AppealRepository.CreateAppealMessage"

	logger := domains.GetLogger(ctx).WithField("operation", op).WithField("appeal_id", message.AppealID.String())
	logger.Debug("Starting database operation: create appeal message")

	if message.SenderID != nil {
		// Сообщение от авторизованного пользователя
		err := r.db.QueryRowContext(ctx, createAppealMessageQuery, message.AppealID, message.Text, message.SenderID).
			Scan(&message.ID, &message.CreatedAt, &message.UpdatedAt)

		if err != nil {
			logger.WithError(err).Error("Database operation failed: create appeal message query")
			return err
		}
	} else {
		// Сообщение от анонимного пользователя
		err := r.db.QueryRowContext(ctx, createAnonymousAppealMessageQuery, message.AppealID, message.Text, message.SenderAnonymID).
			Scan(&message.ID, &message.CreatedAt, &message.UpdatedAt)

		if err != nil {
			logger.WithError(err).Error("Database operation failed: create anonymous appeal message query")
			return err
		}
	}

	logger.WithField("message_id", message.ID.String()).Info("Database operation completed successfully: appeal message created")
	return nil
}

func (r *AppealRepository) GetAppealMessages(ctx context.Context, appealID uuid.UUID) ([]*appealModels.MessageAppeal, error) {
	const op = "AppealRepository.GetAppealMessages"

	logger := domains.GetLogger(ctx).WithField("operation", op).WithField("appeal_id", appealID.String())
	logger.Debug("Starting database operation: get appeal messages")

	rows, err := r.db.QueryContext(ctx, getAppealMessagesQuery, appealID)
	if err != nil {
		logger.WithError(err).Error("Database operation failed: get appeal messages query")
		return nil, err
	}
	defer rows.Close()

	var messages []*appealModels.MessageAppeal
	for rows.Next() {
		var message appealModels.MessageAppeal
		err := rows.Scan(&message.ID, &message.AppealID, &message.Text, &message.SenderID, &message.SenderAnonymID, &message.CreatedAt, &message.UpdatedAt)
		if err != nil {
			logger.WithError(err).Error("Database operation failed: scan message row")
			return nil, err
		}
		messages = append(messages, &message)
	}

	if err = rows.Err(); err != nil {
		logger.WithError(err).Error("Database operation failed: rows iteration error")
		return nil, err
	}

	logger.WithField("messages_count", len(messages)).Info("Database operation completed successfully: appeal messages retrieved")
	return messages, nil
}

func (r *AppealRepository) GetUserRole(ctx context.Context, userID uuid.UUID) (string, error) {
	const op = "AppealRepository.GetUserRole"

	logger := domains.GetLogger(ctx).WithField("operation", op).WithField("user_id", userID.String())
	logger.Debug("Starting database operation: get user role")

	var role string
	err := r.db.QueryRowContext(ctx, getUserRoleQuery, userID).Scan(&role)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			logger.Debug("Database operation completed: user role not found")
			return appealModels.RoleUser, nil
		}
		logger.WithError(err).Error("Database operation failed: get user role query")
		return "", err
	}

	logger.WithField("role", role).Info("Database operation completed successfully: user role found")
	return role, nil
}

func (r *AppealRepository) UpdateUserRole(ctx context.Context, userID uuid.UUID, role string) error {
	const op = "AppealRepository.UpdateUserRole"

	logger := domains.GetLogger(ctx).WithField("operation", op).
		WithField("user_id", userID.String()).WithField("role", role)
	logger.Debug("Starting database operation: update user role")

	_, err := r.db.ExecContext(ctx, updateUserRoleQuery, userID, role)
	if err != nil {
		logger.WithError(err).Error("Database operation failed: update user role query")
		return err
	}

	logger.Info("Database operation completed successfully: user role updated")
	return nil
}

func (r *AppealRepository) DeleteUserRole(ctx context.Context, userID uuid.UUID) error {
	const op = "AppealRepository.DeleteUserRole"

	logger := domains.GetLogger(ctx).WithField("operation", op).WithField("user_id", userID.String())
	logger.Debug("Starting database operation: delete user role")

	result, err := r.db.ExecContext(ctx, deleteUserRoleQuery, userID)
	if err != nil {
		logger.WithError(err).Error("Database operation failed: delete user role query")
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logger.WithError(err).Error("Database operation failed: get rows affected")
		return err
	}

	if rowsAffected == 0 {
		logger.Debug("Database operation completed: user role not found for deletion")
		return errors.New("user role not found")
	}

	logger.Info("Database operation completed successfully: user role deleted")
	return nil
}

func (r *AppealRepository) GetAppealsByRole(ctx context.Context, role string, limit, offset int) ([]*appealModels.Appeal, error) {
	const op = "AppealRepository.GetAppealsByRole"

	logger := domains.GetLogger(ctx).WithField("operation", op).
		WithField("role", role).WithField("limit", limit).WithField("offset", offset)
	logger.Debug("Starting database operation: get appeals by role")

	rows, err := r.db.QueryContext(ctx, getAppealsByRoleQuery, limit, offset)
	if err != nil {
		logger.WithError(err).Error("Database operation failed: get appeals by role query")
		return nil, err
	}
	defer rows.Close()

	var appeals []*appealModels.Appeal
	for rows.Next() {
		var appeal appealModels.Appeal
		err := rows.Scan(&appeal.ID, &appeal.UserID, &appeal.AnonymID, &appeal.AnonymEmail, &appeal.AnonymContact,
			&appeal.Title, &appeal.Status, &appeal.Category, &appeal.CreatedAt, &appeal.UpdatedAt)
		if err != nil {
			logger.WithError(err).Error("Database operation failed: scan appeal row")
			return nil, err
		}
		appeals = append(appeals, &appeal)
	}

	if err = rows.Err(); err != nil {
		logger.WithError(err).Error("Database operation failed: rows iteration error")
		return nil, err
	}

	logger.WithField("appeals_count", len(appeals)).Info("Database operation completed successfully: appeals retrieved by role")
	return appeals, nil
}

func (r *AppealRepository) AssignAppealToSupportLevel(ctx context.Context, appealID uuid.UUID, supportLevel string) error {
	const op = "AppealRepository.AssignAppealToSupportLevel"

	logger := domains.GetLogger(ctx).WithField("operation", op).
		WithField("appeal_id", appealID.String()).WithField("support_level", supportLevel)
	logger.Debug("Starting database operation: assign appeal to support level")

	result, err := r.db.ExecContext(ctx, assignAppealToSupportQuery, appealID, supportLevel)
	if err != nil {
		logger.WithError(err).Error("Database operation failed: assign appeal to support query")
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logger.WithError(err).Error("Database operation failed: get rows affected")
		return err
	}

	if rowsAffected == 0 {
		logger.Debug("Database operation completed: appeal not found")
		return errors.New("appeal not found")
	}

	logger.Info("Database operation completed successfully: appeal assigned to support level")
	return nil
}

func (r *AppealRepository) GetAppealsStatsByStatus(ctx context.Context) (map[string]int, error) {
	const op = "AppealRepository.GetAppealsStatsByStatus"

	logger := domains.GetLogger(ctx).WithField("operation", op)
	logger.Debug("Starting database operation: get appeals stats by status")

	rows, err := r.db.QueryContext(ctx, getAppealsStatsByStatusQuery)
	if err != nil {
		logger.WithError(err).Error("Database operation failed: get appeals stats by status query")
		return nil, err
	}
	defer rows.Close()

	stats := make(map[string]int)
	for rows.Next() {
		var status string
		var count int
		err := rows.Scan(&status, &count)
		if err != nil {
			logger.WithError(err).Error("Database operation failed: scan stats row")
			return nil, err
		}
		stats[status] = count
	}

	if err = rows.Err(); err != nil {
		logger.WithError(err).Error("Database operation failed: rows iteration error")
		return nil, err
	}

	logger.WithField("stats_count", len(stats)).Info("Database operation completed successfully: appeals stats by status retrieved")
	return stats, nil
}

func (r *AppealRepository) GetAppealsStatsByCategory(ctx context.Context) (map[string]int, error) {
	const op = "AppealRepository.GetAppealsStatsByCategory"

	logger := domains.GetLogger(ctx).WithField("operation", op)
	logger.Debug("Starting database operation: get appeals stats by category")

	rows, err := r.db.QueryContext(ctx, getAppealsStatsByCategoryQuery)
	if err != nil {
		logger.WithError(err).Error("Database operation failed: get appeals stats by category query")
		return nil, err
	}
	defer rows.Close()

	stats := make(map[string]int)
	for rows.Next() {
		var category string
		var count int
		err := rows.Scan(&category, &count)
		if err != nil {
			logger.WithError(err).Error("Database operation failed: scan stats row")
			return nil, err
		}
		stats[category] = count
	}

	if err = rows.Err(); err != nil {
		logger.WithError(err).Error("Database operation failed: rows iteration error")
		return nil, err
	}

	logger.WithField("stats_count", len(stats)).Info("Database operation completed successfully: appeals stats by category retrieved")
	return stats, nil
}

func (r *AppealRepository) GetAppealsStatsByDateRange(ctx context.Context, startDate, endDate time.Time) ([]*appealModels.AppealStatsByDate, error) {
	const op = "AppealRepository.GetAppealsStatsByDateRange"

	logger := domains.GetLogger(ctx).WithField("operation", op).
		WithField("start_date", startDate.Format("2006-01-02")).
		WithField("end_date", endDate.Format("2006-01-02"))
	logger.Debug("Starting database operation: get appeals stats by date range")

	rows, err := r.db.QueryContext(ctx, getAppealsStatsByDateRangeQuery, startDate, endDate)
	if err != nil {
		logger.WithError(err).Error("Database operation failed: get appeals stats by date range query")
		return nil, err
	}
	defer rows.Close()

	var stats []*appealModels.AppealStatsByDate
	for rows.Next() {
		var stat appealModels.AppealStatsByDate
		err := rows.Scan(&stat.Date, &stat.Count)
		if err != nil {
			logger.WithError(err).Error("Database operation failed: scan stats row")
			return nil, err
		}
		stats = append(stats, &stat)
	}

	if err = rows.Err(); err != nil {
		logger.WithError(err).Error("Database operation failed: rows iteration error")
		return nil, err
	}

	logger.WithField("stats_count", len(stats)).Info("Database operation completed successfully: appeals stats by date range retrieved")
	return stats, nil
}

func (r *AppealRepository) GetAppealsStatsTotal(ctx context.Context) (*appealModels.AppealStatsTotal, error) {
	const op = "AppealRepository.GetAppealsStatsTotal"

	logger := domains.GetLogger(ctx).WithField("operation", op)
	logger.Debug("Starting database operation: get appeals total stats")

	var stats appealModels.AppealStatsTotal
	err := r.db.QueryRowContext(ctx, getAppealsStatsTotalQuery).Scan(
		&stats.TotalAppeals,
		&stats.OpenAppeals,
		&stats.InWorkAppeals,
		&stats.ClosedAppeals,
		&stats.BugReports,
		&stats.FeatureRequests,
		&stats.Claims,
		&stats.Others,
	)

	if err != nil {
		logger.WithError(err).Error("Database operation failed: get appeals total stats query")
		return nil, err
	}

	logger.WithField("total_appeals", stats.TotalAppeals).Info("Database operation completed successfully: appeals total stats retrieved")
	return &stats, nil
}

// GetAppealsByAnonymID возвращает все обращения для анонимного пользователя
func (r *AppealRepository) GetAppealsByAnonymID(ctx context.Context, anonymID uuid.UUID) ([]*appealModels.Appeal, error) {
	const op = "AppealRepository.GetAppealsByAnonymID"

	logger := domains.GetLogger(ctx).WithField("operation", op).WithField("anonym_id", anonymID.String())
	logger.Debug("Starting database operation: get appeals by anonym ID")

	rows, err := r.db.QueryContext(ctx, getAppealByAnonymIDQuery, anonymID)
	if err != nil {
		logger.WithError(err).Error("Database operation failed: get appeals by anonym ID query")
		return nil, err
	}
	defer rows.Close()

	var appeals []*appealModels.Appeal
	for rows.Next() {
		var appeal appealModels.Appeal
		err := rows.Scan(&appeal.ID, &appeal.UserID, &appeal.AnonymID, &appeal.AnonymEmail, &appeal.AnonymContact,
			&appeal.Title, &appeal.Status, &appeal.Category, &appeal.AssignedTo, &appeal.CreatedAt, &appeal.UpdatedAt)
		if err != nil {
			logger.WithError(err).Error("Database operation failed: scan appeal row")
			return nil, err
		}
		appeals = append(appeals, &appeal)
	}

	if err = rows.Err(); err != nil {
		logger.WithError(err).Error("Database operation failed: rows iteration error")
		return nil, err
	}

	logger.WithField("appeals_count", len(appeals)).Info("Database operation completed successfully: appeals retrieved by anonym ID")
	return appeals, nil
}

// CreateAnonymousAppealMessage создает сообщение от анонимного пользователя
func (r *AppealRepository) CreateAnonymousAppealMessage(ctx context.Context, message *appealModels.MessageAppeal) error {
	const op = "AppealRepository.CreateAnonymousAppealMessage"

	logger := domains.GetLogger(ctx).WithField("operation", op).WithField("appeal_id", message.AppealID.String())
	logger.Debug("Starting database operation: create anonymous appeal message")

	err := r.db.QueryRowContext(ctx, createAnonymousAppealMessageQuery, message.AppealID, message.Text, message.SenderAnonymID).
		Scan(&message.ID, &message.CreatedAt, &message.UpdatedAt)

	if err != nil {
		logger.WithError(err).Error("Database operation failed: create anonymous appeal message query")
		return err
	}

	logger.WithField("message_id", message.ID.String()).Info("Database operation completed successfully: anonymous appeal message created")
	return nil
}

// UpdateAppealAssignment обновляет назначение обращения на определенную роль поддержки
func (r *AppealRepository) UpdateAppealAssignment(ctx context.Context, id uuid.UUID, assignedTo string) error {
	const op = "AppealRepository.UpdateAppealAssignment"

	logger := domains.GetLogger(ctx).WithField("operation", op).
		WithField("appeal_id", id.String()).WithField("assigned_to", assignedTo)
	logger.Debug("Starting database operation: update appeal assignment")

	result, err := r.db.ExecContext(ctx, updateAppealAssignmentQuery, id, assignedTo)
	if err != nil {
		logger.WithError(err).Error("Database operation failed: update appeal assignment query")
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logger.WithError(err).Error("Database operation failed: get rows affected")
		return err
	}

	if rowsAffected == 0 {
		logger.Error("Database operation failed: no rows affected")
		return errors.New("appeal not found")
	}

	logger.Info("Database operation completed successfully: appeal assignment updated")
	return nil
}

// GetAppealsByAssignedRole возвращает обращения, назначенные определенной роли
func (r *AppealRepository) GetAppealsByAssignedRole(ctx context.Context, role string, limit, offset int) ([]*appealModels.Appeal, error) {
	const op = "AppealRepository.GetAppealsByAssignedRole"

	logger := domains.GetLogger(ctx).WithField("operation", op).
		WithField("role", role).WithField("limit", limit).WithField("offset", offset)
	logger.Debug("Starting database operation: get appeals by assigned role")

	rows, err := r.db.QueryContext(ctx, getAppealsByAssignedRoleQuery, role, limit, offset)
	if err != nil {
		logger.WithError(err).Error("Database operation failed: get appeals by assigned role query")
		return nil, err
	}
	defer rows.Close()

	var appeals []*appealModels.Appeal
	for rows.Next() {
		var appeal appealModels.Appeal
		err := rows.Scan(&appeal.ID, &appeal.UserID, &appeal.AnonymID, &appeal.AnonymEmail, &appeal.AnonymContact,
			&appeal.Title, &appeal.Status, &appeal.Category, &appeal.AssignedTo, &appeal.CreatedAt, &appeal.UpdatedAt)
		if err != nil {
			logger.WithError(err).Error("Database operation failed: scan appeal row")
			return nil, err
		}
		appeals = append(appeals, &appeal)
	}

	if err = rows.Err(); err != nil {
		logger.WithError(err).Error("Database operation failed: rows iteration error")
		return nil, err
	}

	logger.WithField("appeals_count", len(appeals)).Info("Database operation completed successfully: appeals retrieved by assigned role")
	return appeals, nil
}
