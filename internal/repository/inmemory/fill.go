package inmemory

import (
	"time"

	chatsModels "github.com/go-park-mail-ru/2025_2_Undefined/internal/models/chats"
	userModels "github.com/go-park-mail-ru/2025_2_Undefined/internal/models/user"
	"github.com/google/uuid"
)

// FillWithFakeData заполняет репозитории фейковыми данными
func FillWithFakeData(userRepo *UserRepo, chatsRepo *ChatsRepo) {
	users := createFakeUsers()

	for _, user := range users {
		userRepo.Create(&user)
	}

	createFakeChats(chatsRepo, users)
}

// пароли везде - admin
// createFakeUsers создает список фейковых пользователей
func createFakeUsers() []userModels.User {
	now := time.Now()

	users := []userModels.User{
		{
			ID:           uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"),
			Email:        "alice.johnson@example.com",
			PhoneNumber:  "+79001234567",
			PasswordHash: "$2a$10$fZAsBF3Itv8a2LMkfK0GLuJ/ADve/bY4RWQViOmoKFTXTCrU7MwrK",
			Name:         "Алиса Джонсон",
			Username:     "alice_j",
			Bio:          "Люблю программировать и читать книги",
			AccountType:  userModels.UserAccount,
			CreatedAt:    now.AddDate(0, -6, 0),
			UpdatedAt:    now.AddDate(0, -1, 0),
		},
		{
			ID:           uuid.MustParse("550e8400-e29b-41d4-a716-446655440001"),
			Email:        "bob.smith@example.com",
			PhoneNumber:  "+79002345678",
			PasswordHash: "$2a$10$fZAsBF3Itv8a2LMkfK0GLuJ/ADve/bY4RWQViOmoKFTXTCrU7MwrK",
			Name:         "Боб Смит",
			Username:     "bob_smith",
			Bio:          "Фотограф и путешественник",
			AccountType:  userModels.PremiumAccount,
			CreatedAt:    now.AddDate(0, -4, 0),
			UpdatedAt:    now.AddDate(0, 0, -5),
		},
		{
			ID:           uuid.MustParse("550e8400-e29b-41d4-a716-446655440002"),
			Email:        "carol.white@example.com",
			PhoneNumber:  "+79003456789",
			PasswordHash: "$2a$10$fZAsBF3Itv8a2LMkfK0GLuJ/ADve/bY4RWQViOmoKFTXTCrU7MwrK",
			Name:         "Кэрол Уайт",
			Username:     "carol_w",
			Bio:          "Дизайнер и художник",
			AccountType:  userModels.VerifiedAccount,
			CreatedAt:    now.AddDate(0, -8, 0),
			UpdatedAt:    now.AddDate(0, 0, -10),
		},
		{
			ID:           uuid.MustParse("550e8400-e29b-41d4-a716-446655440003"),
			Email:        "david.brown@example.com",
			PhoneNumber:  "+79004567890",
			PasswordHash: "$2a$10$fZAsBF3Itv8a2LMkfK0GLuJ/ADve/bY4RWQViOmoKFTXTCrU7MwrK",
			Name:         "Дэвид Браун",
			Username:     "david_b",
			Bio:          "Музыкант и преподаватель",
			AccountType:  userModels.UserAccount,
			CreatedAt:    now.AddDate(0, -2, 0),
			UpdatedAt:    now.AddDate(0, 0, -2),
		},
		{
			ID:           uuid.MustParse("550e8400-e29b-41d4-a716-446655440004"),
			Email:        "eva.green@example.com",
			PhoneNumber:  "+79005678901",
			PasswordHash: "$2a$10$fZAsBF3Itv8a2LMkfK0GLuJ/ADve/bY4RWQViOmoKFTXTCrU7MwrK",
			Name:         "Ева Грин",
			Username:     "eva_green",
			Bio:          "Спортсменка и блогер",
			AccountType:  userModels.PremiumAccount,
			CreatedAt:    now.AddDate(0, -3, 0),
			UpdatedAt:    now.AddDate(0, 0, -7),
		},
	}

	return users
}

// createFakeChats создает фейковые чаты и сообщения
func createFakeChats(chatsRepo *ChatsRepo, users []userModels.User) {
	now := time.Now()

	// Чат 1: Групповой чат "Команда разработчиков"
	chat1ID := uuid.MustParse("660e8400-e29b-41d4-a716-446655440000")
	chat1 := chatsModels.Chat{
		ID:          chat1ID,
		Type:        chatsModels.ChatGroup,
		Name:        "Команда разработчиков",
		Description: "Обсуждение проектов и задач",
	}

	chat1Users := []chatsModels.UserInfo{
		{UserID: users[0].ID, ChatID: chat1ID, Role: chatsModels.RoleAdmin},
		{UserID: users[1].ID, ChatID: chat1ID, Role: chatsModels.RoleMember},
		{UserID: users[2].ID, ChatID: chat1ID, Role: chatsModels.RoleMember},
	}

	chatsRepo.CreateChat(chat1, chat1Users)

	// Добавляем сообщения в первый чат
	chat1Messages := []chatsModels.Message{
		{
			ID:        uuid.New(),
			ChatID:    chat1ID,
			UserID:    users[0].ID,
			Text:      "Привет команда! Как дела с проектом?",
			CreatedAt: now.Add(-2 * time.Hour),
			Type:      chatsModels.UserMessage,
		},
		{
			ID:        uuid.New(),
			ChatID:    chat1ID,
			UserID:    users[1].ID,
			Text:      "Всё идёт по плану, сегодня завершу API",
			CreatedAt: now.Add(-1 * time.Hour),
			Type:      chatsModels.UserMessage,
		},
		{
			ID:        uuid.New(),
			ChatID:    chat1ID,
			UserID:    users[2].ID,
			Text:      "Дизайн готов, отправлю файлы сегодня",
			CreatedAt: now.Add(-30 * time.Minute),
			Type:      chatsModels.UserMessage,
		},
	}

	addMessagesToChat(chatsRepo, chat1Messages)

	// Чат 2: Диалог между двумя пользователями
	chat2ID := uuid.MustParse("660e8400-e29b-41d4-a716-446655440001")
	chat2 := chatsModels.Chat{
		ID:          chat2ID,
		Type:        chatsModels.ChatDialog,
		Name:        "Диалог: Алиса и Дэвид",
		Description: "Личный диалог",
	}

	chat2Users := []chatsModels.UserInfo{
		{UserID: users[0].ID, ChatID: chat2ID, Role: chatsModels.RoleMember},
		{UserID: users[3].ID, ChatID: chat2ID, Role: chatsModels.RoleMember},
	}

	chatsRepo.CreateChat(chat2, chat2Users)

	// Добавляем сообщения во второй чат
	chat2Messages := []chatsModels.Message{
		{
			ID:        uuid.New(),
			ChatID:    chat2ID,
			UserID:    users[0].ID,
			Text:      "Привет! Как дела с музыкой?",
			CreatedAt: now.Add(-3 * time.Hour),
			Type:      chatsModels.UserMessage,
		},
		{
			ID:        uuid.New(),
			ChatID:    chat2ID,
			UserID:    users[3].ID,
			Text:      "Отлично! Записал новую песню, хочешь послушать?",
			CreatedAt: now.Add(-2*time.Hour + 30*time.Minute),
			Type:      chatsModels.UserMessage,
		},
		{
			ID:        uuid.New(),
			ChatID:    chat2ID,
			UserID:    users[0].ID,
			Text:      "Конечно! Отправляй ссылку",
			CreatedAt: now.Add(-15 * time.Minute),
			Type:      chatsModels.UserMessage,
		},
	}

	addMessagesToChat(chatsRepo, chat2Messages)

	// Чат 3: Канал с объявлениями
	chat3ID := uuid.MustParse("660e8400-e29b-41d4-a716-446655440002")
	chat3 := chatsModels.Chat{
		ID:          chat3ID,
		Type:        chatsModels.ChatChannel,
		Name:        "Объявления компании",
		Description: "Важные новости и обновления",
	}

	chat3Users := []chatsModels.UserInfo{
		{UserID: users[2].ID, ChatID: chat3ID, Role: chatsModels.RoleAdmin},
		{UserID: users[0].ID, ChatID: chat3ID, Role: chatsModels.RoleViewer},
		{UserID: users[1].ID, ChatID: chat3ID, Role: chatsModels.RoleViewer},
		{UserID: users[4].ID, ChatID: chat3ID, Role: chatsModels.RoleViewer},
	}

	chatsRepo.CreateChat(chat3, chat3Users)

	// Добавляем сообщения в третий чат
	chat3Messages := []chatsModels.Message{
		{
			ID:        uuid.New(),
			ChatID:    chat3ID,
			UserID:    users[2].ID,
			Text:      "🎉 Поздравляем с успешным запуском нового проекта!",
			CreatedAt: now.Add(-24 * time.Hour),
			Type:      chatsModels.UserMessage,
		},
		{
			ID:        uuid.New(),
			ChatID:    chat3ID,
			UserID:    users[2].ID,
			Text:      "Завтра в 14:00 будет общее собрание команды",
			CreatedAt: now.Add(-4 * time.Hour),
			Type:      chatsModels.UserMessage,
		},
	}

	addMessagesToChat(chatsRepo, chat3Messages)

	// Чат 4: Ещё один групповой чат
	chat4ID := uuid.MustParse("660e8400-e29b-41d4-a716-446655440003")
	chat4 := chatsModels.Chat{
		ID:          chat4ID,
		Type:        chatsModels.ChatGroup,
		Name:        "Хобби и увлечения",
		Description: "Обсуждаем свободное время",
	}

	chat4Users := []chatsModels.UserInfo{
		{UserID: users[1].ID, ChatID: chat4ID, Role: chatsModels.RoleAdmin},
		{UserID: users[3].ID, ChatID: chat4ID, Role: chatsModels.RoleMember},
		{UserID: users[4].ID, ChatID: chat4ID, Role: chatsModels.RoleMember},
	}

	chatsRepo.CreateChat(chat4, chat4Users)

	// Добавляем сообщения в четвёртый чат
	chat4Messages := []chatsModels.Message{
		{
			ID:        uuid.New(),
			ChatID:    chat4ID,
			UserID:    users[1].ID,
			Text:      "Кто-нибудь увлекается фотографией? Хочу обсудить новые техники",
			CreatedAt: now.Add(-6 * time.Hour),
			Type:      chatsModels.UserMessage,
		},
		{
			ID:        uuid.New(),
			ChatID:    chat4ID,
			UserID:    users[4].ID,
			Text:      "Я! Недавно купила новый объектив, очень довольна результатом",
			CreatedAt: now.Add(-5*time.Hour + 15*time.Minute),
			Type:      chatsModels.UserMessage,
		},
		{
			ID:        uuid.New(),
			ChatID:    chat4ID,
			UserID:    users[3].ID,
			Text:      "А я больше в музыке разбираюсь, но интересно послушать",
			CreatedAt: now.Add(-45 * time.Minute),
			Type:      chatsModels.UserMessage,
		},
	}

	addMessagesToChat(chatsRepo, chat4Messages)
}

// addMessagesToChat добавляет сообщения в чат напрямую (обходя CreateChat)
func addMessagesToChat(chatsRepo *ChatsRepo, messages []chatsModels.Message) {
	chatsRepo.mutexChatMessages.Lock()
	defer chatsRepo.mutexChatMessages.Unlock()

	for _, message := range messages {
		chatsRepo.chatMessages[message.ChatID] = append(chatsRepo.chatMessages[message.ChatID], message)
	}
}
