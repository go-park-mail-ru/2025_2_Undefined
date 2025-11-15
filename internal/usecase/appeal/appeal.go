package usecase

import (
	"context"
	"time"

	appealModels "github.com/go-park-mail-ru/2025_2_Undefined/internal/models/appeal"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/models/errs"
	dtoAppeal "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/dto/appeal"
	interfaceAppealRepository "github.com/go-park-mail-ru/2025_2_Undefined/internal/usecase/interface/appeal"
	"github.com/google/uuid"
)

type AppealUsecase struct {
	repo interfaceAppealRepository.AppealRepository
}

func NewAppealUsecase(repo interfaceAppealRepository.AppealRepository) *AppealUsecase {
	return &AppealUsecase{repo: repo}
}

// buildAppealsDTO формирует []dtoAppeal.AppealDTO из списка моделей Appeal,
// подтягивая сообщения для каждого обращения и конвертируя их через buildMessagesDTO.
func (uc *AppealUsecase) buildAppealsDTO(ctx context.Context, appeals []*appealModels.Appeal) ([]dtoAppeal.AppealDTO, error) {
	result := make([]dtoAppeal.AppealDTO, 0, len(appeals))
	for _, a := range appeals {
		msgs, err := uc.repo.GetAppealMessages(ctx, a.ID)
		if err != nil {
			return nil, err
		}

		messagesDTO := make([]dtoAppeal.AppealMessageDTO, 0, len(msgs))
		for _, m := range msgs {
			var isUser bool
			// Сравниваем значения UUID, а не указатели — учитываем и авторизованных и анонимных отправителей
			if a.UserID != nil && m.SenderID != nil && *a.UserID == *m.SenderID {
				isUser = true
			} else if a.AnonymID != nil && m.SenderAnonymID != nil && *a.AnonymID == *m.SenderAnonymID {
				isUser = true
			} else {
				isUser = false
			}

			messagesDTO = append(messagesDTO, dtoAppeal.AppealMessageDTO{
				AppealID: m.AppealID,
				Text:     m.Text,
				SenderID: m.SenderID,
				AnonymID: m.SenderAnonymID,
				IsUser:   isUser,
			})
		}

		result = append(result, dtoAppeal.AppealDTO{
			ID:         a.ID,
			Title:      a.Title,
			Status:     a.Status,
			Category:   a.Category,
			UserID:     a.UserID,
			AssignedTo: a.AssignedTo,
			Messages:   messagesDTO,
		})
	}

	return result, nil
}

// GetAppeals возвращает обращения пользователя вместе с их сообщениями
func (uc *AppealUsecase) GetAppeals(ctx context.Context, userID uuid.UUID) ([]dtoAppeal.AppealDTO, error) {
	appeals, err := uc.repo.GetAppealsByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return uc.buildAppealsDTO(ctx, appeals)
}

func (uc *AppealUsecase) GetAppealByID(ctx context.Context, userID, appealID uuid.UUID) (*dtoAppeal.AppealDTO, error) {
	appeal, err := uc.repo.GetAppealByID(ctx, appealID)
	if err != nil {
		return nil, err
	}

	role, err := uc.repo.GetUserRole(ctx, userID)
	if err != nil {
		return nil, err
	}

	if appeal.UserID == nil || *appeal.UserID != userID && role == appealModels.RoleUser {
		return nil, errs.ErrNoRights
	}

	msgs, err := uc.repo.GetAppealMessages(ctx, appeal.ID)
	if err != nil {
		return nil, err
	}

	messagesDTO := make([]dtoAppeal.AppealMessageDTO, 0, len(msgs))
	for _, m := range msgs {
		messagesDTO = append(messagesDTO, dtoAppeal.AppealMessageDTO{
			AppealID: m.AppealID,
			Text:     m.Text,
			SenderID: m.SenderID,
			AnonymID: m.SenderAnonymID,
		})
	}

	return &dtoAppeal.AppealDTO{
		ID:         appeal.ID,
		Title:      appeal.Title,
		Status:     appeal.Status,
		Category:   appeal.Category,
		UserID:     appeal.UserID,
		AssignedTo: appeal.AssignedTo,
		Messages:   messagesDTO,
	}, nil
}

// CreateAppeal создает новое обращение и возвращает его ID
func (uc *AppealUsecase) CreateAppeal(ctx context.Context, appealDTO dtoAppeal.CreateAppealDTO) (uuid.UUID, error) {
	now := time.Now()
	appeal := &appealModels.Appeal{
		ID:         uuid.New(),
		UserID:     &appealDTO.UserID,
		Title:      appealDTO.Title,
		Status:     appealModels.StatusOpen,
		Category:   appealDTO.Category,
		AssignedTo: appealModels.RoleSupportV1, // По умолчанию назначается support-l1
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	if err := uc.repo.CreateAppeal(ctx, appeal); err != nil {
		return uuid.Nil, err
	}

	return appeal.ID, nil
}

// EditAppeal обновляет статус и/или категорию обращения
func (uc *AppealUsecase) EditAppeal(ctx context.Context, userID uuid.UUID, appealDTO dtoAppeal.EditAppealDTO) error {
	role, err := uc.repo.GetUserRole(ctx, userID)
	if err != nil {
		return err
	}

	if role == appealModels.RoleUser && appealDTO.Status != "" {
		return errs.ErrNoRights
	}

	if appealDTO.Status != "" {
		if err := uc.repo.UpdateAppealStatus(ctx, appealDTO.ID, appealDTO.Status); err != nil {
			return err
		}
	}

	if appealDTO.Category != "" {
		if err := uc.repo.UpdateAppealCategory(ctx, appealDTO.ID, appealDTO.Category); err != nil {
			return err
		}
	}

	if appealDTO.Title != "" {
		if err := uc.repo.UpdateAppealTitle(ctx, appealDTO.ID, appealDTO.Title); err != nil {
			return err
		}
	}

	return nil
}

// PostAppealMessage добавляет сообщение к обращению
func (uc *AppealUsecase) PostAppealMessage(ctx context.Context, userID uuid.UUID, messageDTO dtoAppeal.AppealCreateMessageDTO) error {
	now := time.Now()
	msg := &appealModels.MessageAppeal{
		ID:        uuid.New(),
		AppealID:  messageDTO.AppealID,
		Text:      messageDTO.Text,
		CreatedAt: now,
		UpdatedAt: now,
		SenderID:  &userID,
	}

	return uc.repo.CreateAppealMessage(ctx, msg)
}

// GetAnonymousAppeals возвращает анонимные обращения по anonym_id
func (uc *AppealUsecase) GetAnonymousAppeals(ctx context.Context, anonymID uuid.UUID) ([]dtoAppeal.AnonymousAppealDTO, error) {
	appeals, err := uc.repo.GetAppealsByAnonymID(ctx, anonymID)
	if err != nil {
		return nil, err
	}

	// Используем buildAppealsDTO для получения данных с сообщениями
	appealsDTO, err := uc.buildAppealsDTO(ctx, appeals)
	if err != nil {
		return nil, err
	}

	// Конвертируем AppealDTO в AnonymousAppealDTO
	result := make([]dtoAppeal.AnonymousAppealDTO, 0, len(appealsDTO))
	for i, appealDTO := range appealsDTO {
		var email string
		if appeals[i].AnonymEmail != nil {
			email = *appeals[i].AnonymEmail
		}

		result = append(result, dtoAppeal.AnonymousAppealDTO{
			ID:         appealDTO.ID,
			Title:      appealDTO.Title,
			Status:     appealDTO.Status,
			Category:   appealDTO.Category,
			AnonymID:   *appeals[i].AnonymID,
			Email:      email,
			AssignedTo: appealDTO.AssignedTo,
			Messages:   appealDTO.Messages,
		})
	}

	return result, nil
}

// CreateAnonymousAppeal создает анонимное обращение
func (uc *AppealUsecase) CreateAnonymousAppeal(ctx context.Context, appealDTO dtoAppeal.CreateAnonymousAppealDTO) (uuid.UUID, error) {
	now := time.Now()

	var email *string
	if appealDTO.Email != "" {
		email = &appealDTO.Email
	}

	var contact *string
	if appealDTO.Contact != "" {
		contact = &appealDTO.Contact
	}

	appeal := &appealModels.Appeal{
		ID:            uuid.New(),
		AnonymID:      &appealDTO.AnonymID,
		AnonymEmail:   email,
		AnonymContact: contact,
		Title:         appealDTO.Title,
		Status:        appealModels.StatusOpen,
		Category:      appealDTO.Category,
		AssignedTo:    appealModels.RoleSupportV1, // По умолчанию назначается support-l1
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	if err := uc.repo.CreateAppeal(ctx, appeal); err != nil {
		return uuid.Nil, err
	}

	return appeal.ID, nil
}

// PostAnonymousMessage отправляет сообщение от анонимного пользователя
func (uc *AppealUsecase) PostAnonymousMessage(ctx context.Context, messageDTO dtoAppeal.CreateAnonymousMessageDTO) error {
	// Проверяем, что обращение существует и принадлежит данному анонимному пользователю
	appeal, err := uc.repo.GetAppealByID(ctx, messageDTO.AppealID)
	if err != nil {
		return errs.ErrNotFound
	}

	// Проверяем, что обращение принадлежит этому анонимному пользователю
	if appeal.AnonymID == nil || *appeal.AnonymID != messageDTO.AnonymID {
		return errs.ErrNoRights
	}

	now := time.Now()
	msg := &appealModels.MessageAppeal{
		ID:             uuid.New(),
		AppealID:       messageDTO.AppealID,
		Text:           messageDTO.Text,
		SenderAnonymID: &messageDTO.AnonymID,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	return uc.repo.CreateAppealMessage(ctx, msg)
}

// GetAppealsForSupport возвращает обращения для техподдержки.
// Если пользователь — админ, возвращаются все обращения, иначе только обращения для роли поддержки данного пользователя.
func (uc *AppealUsecase) GetAppealsForSupport(ctx context.Context, userID uuid.UUID, limit, offset int) ([]dtoAppeal.AppealDTO, error) {
	role, err := uc.repo.GetUserRole(ctx, userID)
	if err != nil {
		return nil, err
	}

	var appeals []*appealModels.Appeal
	// Блокируем доступ пользователям с ролью по умолчанию (user) — ручка доступна только для support/admin/developer
	if role == appealModels.RoleUser {
		return nil, errs.ErrNoRights
	}

	if role == appealModels.RoleAdmin {
		// admin получает все обращения
		appeals, err = uc.repo.GetAppealsWithFilters(ctx, "", "", limit, offset)
		if err != nil {
			return nil, err
		}
	} else {
		// остальные получают обращения только для своей роли
		appeals, err = uc.repo.GetAppealsByRole(ctx, role, limit, offset)
		if err != nil {
			return nil, err
		}
	}

	return uc.buildAppealsDTO(ctx, appeals)
}

// ChangeUserRole позволяет админу изменить роль другого пользователя
func (uc *AppealUsecase) ChangeUserRole(ctx context.Context, adminID, userID uuid.UUID, role string) error {
	callerRole, err := uc.repo.GetUserRole(ctx, adminID)
	if err != nil {
		return err
	}

	if callerRole != appealModels.RoleAdmin {
		return errs.ErrNoRights
	}

	if role == appealModels.RoleUser {
		return uc.repo.DeleteUserRole(ctx, userID)
	}

	switch role {
	case appealModels.RoleAdmin, appealModels.RoleDeveloper, appealModels.RoleSupportV1, appealModels.RoleSupportV2:
		return uc.repo.UpdateUserRole(ctx, userID, role)
	default:
		return errs.ErrBadRequest
	}
}

// DeleteUserRole позволяет админу удалить роль пользователя (вернуть к роли по умолчанию)
func (uc *AppealUsecase) DeleteUserRole(ctx context.Context, adminID, userID uuid.UUID) error {
	callerRole, err := uc.repo.GetUserRole(ctx, adminID)
	if err != nil {
		return err
	}

	if callerRole != appealModels.RoleAdmin {
		return errs.ErrNoRights
	}

	return uc.repo.DeleteUserRole(ctx, userID)
}

// GetAppealsStats возвращает агрегированную статистику по обращениям.
// Операция доступна только админу. Если startDate и endDate заданы — включает breakdown по датам в диапазоне.
func (uc *AppealUsecase) GetAppealsStats(ctx context.Context, adminID uuid.UUID, startDate, endDate *time.Time) (*dtoAppeal.AppealStatsDTO, error) {
	callerRole, err := uc.repo.GetUserRole(ctx, adminID)
	if err != nil {
		return nil, err
	}

	if callerRole != appealModels.RoleAdmin {
		return nil, errs.ErrNoRights
	}

	byStatus, err := uc.repo.GetAppealsStatsByStatus(ctx)
	if err != nil {
		return nil, err
	}

	byCategory, err := uc.repo.GetAppealsStatsByCategory(ctx)
	if err != nil {
		return nil, err
	}

	var byDateDTO []dtoAppeal.AppealStatsByDateDTO
	if startDate != nil && endDate != nil {
		byDate, err := uc.repo.GetAppealsStatsByDateRange(ctx, *startDate, *endDate)
		if err != nil {
			return nil, err
		}

		for _, d := range byDate {
			byDateDTO = append(byDateDTO, dtoAppeal.AppealStatsByDateDTO{
				Date:  d.Date.Format("2006-01-02"),
				Count: d.Count,
			})
		}
	}

	total, err := uc.repo.GetAppealsStatsTotal(ctx)
	if err != nil {
		return nil, err
	}

	stats := &dtoAppeal.AppealStatsDTO{
		ByStatus:    byStatus,
		ByCategory:  byCategory,
		ByDateRange: byDateDTO,
		Total: dtoAppeal.AppealStatsTotalDTO{
			TotalAppeals:    total.TotalAppeals,
			OpenAppeals:     total.OpenAppeals,
			InWorkAppeals:   total.InWorkAppeals,
			ClosedAppeals:   total.ClosedAppeals,
			BugReports:      total.BugReports,
			FeatureRequests: total.FeatureRequests,
			Claims:          total.Claims,
			Others:          total.Others,
		},
	}

	return stats, nil
}
