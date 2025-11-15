package dto

import "github.com/google/uuid"

type CreateAppealDTO struct {
	UserID   uuid.UUID
	Title    string `json:"title"`
	Category string `json:"category"`
}

type EditAppealDTO struct {
	ID       uuid.UUID `json:"id" swaggertype:"string" format:"uuid"`
	Title    string    `json:"title"`
	Status   string    `json:"status"`
	Category string    `json:"category"`
}

type AppealDTO struct {
	ID         uuid.UUID          `json:"id" swaggertype:"string" format:"uuid"`
	Title      string             `json:"title"`
	Status     string             `json:"status"`
	Category   string             `json:"category"`
	UserID     *uuid.UUID         `json:"user_id" swaggertype:"string" format:"uuid"`
	AssignedTo string             `json:"assigned_to"`
	Messages   []AppealMessageDTO `json:"messages"`
}

type AppealMessageDTO struct {
	AppealID uuid.UUID  `json:"appeal_id" swaggertype:"string" format:"uuid"`
	Text     string     `json:"text"`
	SenderID *uuid.UUID `json:"sender_id" swaggertype:"string" format:"uuid"`
	AnonymID *uuid.UUID `json:"anonym_id" swaggertype:"string" format:"uuid"`
	IsUser   bool       `json:"is_user"`
}

type AppealCreateMessageDTO struct {
	AppealID uuid.UUID `json:"appeal_id" swaggertype:"string" format:"uuid"`
	Text     string    `json:"text"`
}

// DTO для создания анонимного обращения
type CreateAnonymousAppealDTO struct {
	AnonymID uuid.UUID `json:"anonym_id" swaggertype:"string" format:"uuid"`
	Title    string    `json:"title"`
	Category string    `json:"category"`
	Email    string    `json:"email"`
	Contact  string    `json:"contact"`
}

// DTO для создания анонимного сообщения
type CreateAnonymousMessageDTO struct {
	AppealID uuid.UUID `json:"appeal_id" swaggertype:"string" format:"uuid"`
	AnonymID uuid.UUID `json:"anonym_id" swaggertype:"string" format:"uuid"`
	Text     string    `json:"text"`
}

// DTO для изменения роли пользователя
type ChangeRoleDTO struct {
	Role string `json:"role"`
}

// DTO для получения анонимного обращения
type AnonymousAppealDTO struct {
	ID         uuid.UUID          `json:"id" swaggertype:"string" format:"uuid"`
	Title      string             `json:"title"`
	Status     string             `json:"status"`
	Category   string             `json:"category"`
	AnonymID   uuid.UUID          `json:"anonym_id" swaggertype:"string" format:"uuid"`
	Email      string             `json:"email"`
	AssignedTo string             `json:"assigned_to"`
	Messages   []AppealMessageDTO `json:"messages"`
}

// DTOs для статистики
type AppealStatsByDateDTO struct {
	Date  string `json:"date"`
	Count int    `json:"count"`
}

type AppealStatsTotalDTO struct {
	TotalAppeals    int `json:"total_appeals"`
	OpenAppeals     int `json:"open_appeals"`
	InWorkAppeals   int `json:"in_work_appeals"`
	ClosedAppeals   int `json:"closed_appeals"`
	BugReports      int `json:"bug_reports"`
	FeatureRequests int `json:"feature_requests"`
	Claims          int `json:"claims"`
	Others          int `json:"others"`
}

type AppealStatsDTO struct {
	ByStatus    map[string]int         `json:"by_status"`
	ByCategory  map[string]int         `json:"by_category"`
	ByDateRange []AppealStatsByDateDTO `json:"by_date_range,omitempty"`
	Total       AppealStatsTotalDTO    `json:"total"`
}
