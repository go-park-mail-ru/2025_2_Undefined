package models

import (
	"time"

	"github.com/google/uuid"
)

const (
	StatusOpen       = "open"
	StatusInProgress = "in_work"
	StatusClosed     = "closed"
)

const (
	CategoryBugReport     = "bug"
	CategoryFeatureReport = "feature"
	CategoryClaimReport   = "claim"
	CategoryOtherReport   = "other"
)

const (
	RoleUser      = "user"
	RoleSupportV1 = "support-l1"
	RoleAdmin     = "admin"
	RoleDeveloper = "developer"
	RoleSupportV2 = "support-l2"
)

type Appeal struct {
	ID            uuid.UUID
	UserID        *uuid.UUID // Nullable для анонимных пользователей
	AnonymID      *uuid.UUID // ID анонимного пользователя
	AnonymEmail   *string    // Email для обратной связи
	AnonymContact *string    // Дополнительная контактная информация
	Title         string
	Status        string
	Category      string
	AppealRole    string
	AssignedTo    string // Роль (support-l1, support-l2, developer, admin)
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type MessageAppeal struct {
	ID             uuid.UUID
	AppealID       uuid.UUID
	Text           string
	SenderID       *uuid.UUID // Nullable для анонимных отправителей
	SenderAnonymID *uuid.UUID // ID анонимного отправителя
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type AppealRole struct {
	ID     uuid.UUID
	UserID uuid.UUID
	Role   string
}

type AppealStatsByDate struct {
	Date  time.Time
	Count int
}

type AppealStatsTotal struct {
	TotalAppeals    int
	OpenAppeals     int
	InWorkAppeals   int
	ClosedAppeals   int
	BugReports      int
	FeatureRequests int
	Claims          int
	Others          int
}
