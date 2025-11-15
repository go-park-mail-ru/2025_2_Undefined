package appeal

import (
	"context"
	"time"

	appealModels "github.com/go-park-mail-ru/2025_2_Undefined/internal/models/appeal"
	"github.com/google/uuid"
)

type AppealRepository interface {
	CreateAppeal(ctx context.Context, appeal *appealModels.Appeal) error
	GetAppealByID(ctx context.Context, id uuid.UUID) (*appealModels.Appeal, error)
	GetAppealsByUserID(ctx context.Context, userID uuid.UUID) ([]*appealModels.Appeal, error)
	GetAppealsWithFilters(ctx context.Context, status, category string, limit, offset int) ([]*appealModels.Appeal, error)
	UpdateAppealStatus(ctx context.Context, id uuid.UUID, status string) error
	UpdateAppealCategory(ctx context.Context, id uuid.UUID, category string) error
	UpdateAppealTitle(ctx context.Context, id uuid.UUID, title string) error

	// Работа с сообщениями обращений
	CreateAppealMessage(ctx context.Context, message *appealModels.MessageAppeal) error
	GetAppealMessages(ctx context.Context, appealID uuid.UUID) ([]*appealModels.MessageAppeal, error)

	// Статистика обращений
	GetAppealsStatsByStatus(ctx context.Context) (map[string]int, error)
	GetAppealsStatsByCategory(ctx context.Context) (map[string]int, error)
	GetAppealsStatsByDateRange(ctx context.Context, startDate, endDate time.Time) ([]*appealModels.AppealStatsByDate, error)
	GetAppealsStatsTotal(ctx context.Context) (*appealModels.AppealStatsTotal, error)

	// Работа с ролями пользователей для техподдержки
	GetUserRole(ctx context.Context, userID uuid.UUID) (string, error)
	UpdateUserRole(ctx context.Context, userID uuid.UUID, role string) error
	DeleteUserRole(ctx context.Context, userID uuid.UUID) error

	// Получение обращений для конкретной линии поддержки
	GetAppealsByRole(ctx context.Context, role string, limit, offset int) ([]*appealModels.Appeal, error)

	AssignAppealToSupportLevel(ctx context.Context, appealID uuid.UUID, supportLevel string) error

	// Методы для анонимных пользователей
	GetAppealsByAnonymID(ctx context.Context, anonymID uuid.UUID) ([]*appealModels.Appeal, error)
	CreateAnonymousAppealMessage(ctx context.Context, message *appealModels.MessageAppeal) error
}
