package appeal

import (
	"context"
	"time"

	dtoAppeal "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/dto/appeal"
	"github.com/google/uuid"
)

type AppealUsecase interface {
	GetAppeals(ctx context.Context, userID uuid.UUID) ([]dtoAppeal.AppealDTO, error)
	GetAppealByID(ctx context.Context, userID, appealID uuid.UUID) (*dtoAppeal.AppealDTO, error)
	// GetAppealsForSupport возвращает обращения для техподдержки.
	// Если пользователь — админ, возвращаются все обращения, иначе только обращения для роли поддержки данного пользователя.
	GetAppealsForSupport(ctx context.Context, userID uuid.UUID, limit, offset int) ([]dtoAppeal.AppealDTO, error)
	CreateAppeal(ctx context.Context, appealDTO dtoAppeal.CreateAppealDTO) (uuid.UUID, error)
	EditAppeal(ctx context.Context, userID uuid.UUID, appealDTO dtoAppeal.EditAppealDTO) error
	PostAppealMessage(ctx context.Context, userID uuid.UUID, messageDTO dtoAppeal.AppealCreateMessageDTO) error

	// Методы для анонимных пользователей
	GetAnonymousAppeals(ctx context.Context, anonymID uuid.UUID) ([]dtoAppeal.AnonymousAppealDTO, error)
	CreateAnonymousAppeal(ctx context.Context, appealDTO dtoAppeal.CreateAnonymousAppealDTO) (uuid.UUID, error)
	PostAnonymousMessage(ctx context.Context, messageDTO dtoAppeal.CreateAnonymousMessageDTO) error

	// Админские операции над ролями пользователей
	ChangeUserRole(ctx context.Context, adminID, userID uuid.UUID, role string) error
	DeleteUserRole(ctx context.Context, adminID, userID uuid.UUID) error

	// Получение статистики обращений (только для админа).
	// Если startDate и endDate заданы — возвращается статистика по датам в диапазоне.
	GetAppealsStats(ctx context.Context, adminID uuid.UUID, startDate, endDate *time.Time) (*dtoAppeal.AppealStatsDTO, error)
}
