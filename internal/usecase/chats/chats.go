package usecase

import (
	"context"

	modelsChats "github.com/go-park-mail-ru/2025_2_Undefined/internal/models/chats"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/models/domains"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/models/errs"
	modelsMessage "github.com/go-park-mail-ru/2025_2_Undefined/internal/models/message"
	dtoChats "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/dto/chats"
	dtoMessage "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/dto/message"
	dtoUtils "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/dto/utils"
	interfaceChatsRepository "github.com/go-park-mail-ru/2025_2_Undefined/internal/usecase/interface/chats"
	interfaceMessageRepository "github.com/go-park-mail-ru/2025_2_Undefined/internal/usecase/interface/message"
	interfaceFileStorage "github.com/go-park-mail-ru/2025_2_Undefined/internal/usecase/interface/storage"
	interfaceUserRepository "github.com/go-park-mail-ru/2025_2_Undefined/internal/usecase/interface/user"
	"github.com/google/uuid"

	"github.com/go-park-mail-ru/2025_2_Undefined/internal/repository/minio"
)

type ChatsUsecase struct {
	chatsRepo   interfaceChatsRepository.ChatsRepository
	messageRepo interfaceMessageRepository.MessageRepository
	usersRepo   interfaceUserRepository.UserRepository
	fileStorage interfaceFileStorage.FileStorage
}

func NewChatsUsecase(chatsRepo interfaceChatsRepository.ChatsRepository, usersRepo interfaceUserRepository.UserRepository, messageRepo interfaceMessageRepository.MessageRepository, fileStorage interfaceFileStorage.FileStorage) *ChatsUsecase {
	return &ChatsUsecase{
		chatsRepo:   chatsRepo,
		messageRepo: messageRepo,
		usersRepo:   usersRepo,
		fileStorage: fileStorage,
	}
}

func (uc *ChatsUsecase) GetChats(ctx context.Context, userId uuid.UUID) ([]dtoChats.ChatViewInformationDTO, error) {
	const op = "ChatsUsecase.GetChats"

	logger := domains.GetLogger(ctx).WithField("operation", op)

	chats, err := uc.chatsRepo.GetChats(ctx, userId)
	if err != nil {
		return nil, err
	}

	lastMessages, err := uc.messageRepo.GetLastMessagesOfChats(ctx, userId)
	if err != nil {
		return nil, err
	}

	// Создаем мапу для быстрого поиска последних сообщений по chat_id
	messageMap := make(map[uuid.UUID]modelsMessage.Message, len(lastMessages))
	for _, msg := range lastMessages {
		messageMap[msg.ChatID] = msg
	}

	result := make([]dtoChats.ChatViewInformationDTO, 0, len(chats))
	for _, chat := range chats {
		chatName := chat.Name

		// Для диалогов определяем имя собеседника
		if chat.Type == modelsChats.ChatTypeDialog {
			users, err := uc.chatsRepo.GetUsersOfChat(ctx, chat.ID)
			if err != nil {
				logger.Warningf("could not get users for dialog %s: %v", chat.ID, err)
			} else {
				// Ищем собеседника (не текущего пользователя)
				for _, user := range users {
					if user.UserID != userId {
						chatName = user.UserName
						break
					}
				}
			}
		}

		chatDTO := dtoChats.ChatViewInformationDTO{
			ID:   chat.ID,
			Name: chatName,
			Type: chat.Type,
		}

		if lastMsg, exists := messageMap[chat.ID]; exists {
			chatDTO.LastMessage = dtoMessage.MessageDTO{
				ID:         lastMsg.ID,
				SenderID:   lastMsg.UserID,
				SenderName: lastMsg.UserName,
				Text:       lastMsg.Text,
				CreatedAt:  lastMsg.CreatedAt,
				UpdatedAt:  lastMsg.UpdatedAt,
				ChatID:     lastMsg.ChatID,
				Type:       lastMsg.Type,
			}
		}

		result = append(result, chatDTO)
	}

	return result, nil
}

func (uc *ChatsUsecase) GetInformationAboutChat(ctx context.Context, userID, chatID uuid.UUID) (*dtoChats.ChatDetailedInformationDTO, error) {
	const op = "ChatsUsecase.GetInformationAboutChat"

	logger := domains.GetLogger(ctx).WithField("operation", op)
	defer logger.Debugf("Succesfull get information about chat: %s", chatID)

	chat, err := uc.chatsRepo.GetChat(ctx, chatID)
	if err != nil {
		return nil, err
	}

	messages, err := uc.messageRepo.GetMessagesOfChat(ctx, chatID, 0, 40)
	if err != nil {
		return nil, err
	}

	users, err := uc.chatsRepo.GetUsersOfChat(ctx, chatID)
	if err != nil {
		return nil, err
	}

	userInfo, err := uc.chatsRepo.GetUserInfo(ctx, userID, chatID)
	if err != nil {
		return nil, err
	}

	// Проверяем, что пользователь есть в участниках чата. Иначе - возвращаем ошибку
	isUserInChat := false
	for _, user := range users {
		if user.UserID == userID {
			isUserInChat = true
			break
		}
	}
	if !isUserInChat {
		return nil, errs.ErrNotFound
	}

	messagesDTO := make([]dtoMessage.MessageDTO, len(messages))
	for i, message := range messages {
		messagesDTO[i] = dtoMessage.MessageDTO{
			ID:         message.ID,
			SenderID:   message.UserID,
			SenderName: message.UserName,
			Text:       message.Text,
			CreatedAt:  message.CreatedAt,
			UpdatedAt:  message.UpdatedAt,
			ChatID:     message.ChatID,
			Type:       message.Type,
		}
	}

	usersDTO := make([]dtoChats.UserInfoChatDTO, len(users))
	for i, user := range users {
		usersDTO[i] = dtoChats.UserInfoChatDTO{
			UserId:   user.UserID,
			UserName: user.UserName,
			Role:     user.Role,
		}
	}

	var isAdmin, canChat, isMember, isPrivate bool = false, false, false, false
	switch userInfo.Role {
	case modelsChats.RoleAdmin:
		isAdmin = true
		fallthrough
	case modelsChats.RoleMember:
		isMember = true
		canChat = true
	}

	if chat.Type == modelsChats.ChatTypeDialog {
		isPrivate = true
	}

	// Определяем название чата
	chatName := chat.Name
	if chat.Type == modelsChats.ChatTypeDialog {
		// Для диалогов название - это имя собеседника
		for _, user := range usersDTO {
			if user.UserId != userID {
				chatName = user.UserName
				break
			}
		}
	}

	result := &dtoChats.ChatDetailedInformationDTO{
		ID:          chat.ID,
		Name:        chatName,
		IsAdmin:     isAdmin,
		CanChat:     canChat,
		IsMember:    isMember,
		IsPrivate:   isPrivate,
		Type:        chat.Type,
		Members:     usersDTO,
		Messages:    messagesDTO,
		Description: chat.Description,
	}

	return result, nil
}

func (uc *ChatsUsecase) GetUsersDialog(ctx context.Context, user1ID, user2ID uuid.UUID) (*dtoUtils.IdDTO, error) {
	idDialog, err := uc.chatsRepo.GetUsersDialog(ctx, user1ID, user2ID)
	if err != nil {
		return nil, err
	}

	resultDTO := &dtoUtils.IdDTO{
		ID: idDialog,
	}

	return resultDTO, nil
}

func (uc *ChatsUsecase) CreateChat(ctx context.Context, chatDTO dtoChats.ChatCreateInformationDTO) (uuid.UUID, error) {
	chat := modelsChats.Chat{
		ID:          uuid.New(),
		Name:        chatDTO.Name,
		Type:        chatDTO.Type,
		Description: "",
	}

	usersIds := make([]uuid.UUID, len(chatDTO.Members))
	for i, memberDTO := range chatDTO.Members {
		usersIds[i] = memberDTO.UserId
	}

	usersNames, err := uc.usersRepo.GetUsersNames(ctx, usersIds)
	if err != nil {
		return uuid.Nil, err
	}

	usersInfo := make([]modelsChats.UserInfo, len(chatDTO.Members))
	for i, memberDTO := range chatDTO.Members {
		usersInfo[i] = modelsChats.UserInfo{
			UserID: memberDTO.UserId,
			ChatID: chat.ID,
			Role:   memberDTO.Role,
		}
	}

	err = uc.chatsRepo.CreateChat(ctx, chat, usersInfo, usersNames)
	if err != nil {
		return uuid.Nil, err
	}
	return chat.ID, nil
}

func (uc *ChatsUsecase) AddUsersToChat(ctx context.Context, chatID, userID uuid.UUID, users []dtoChats.AddChatMemberDTO) error {
	ok, err := uc.chatsRepo.CheckUserHasRole(ctx, userID, chatID, modelsChats.RoleAdmin)
	if err != nil {
		return err
	}

	if !ok {
		return errs.ErrNoRights
	}

	usersInfo := make([]modelsChats.UserInfo, len(users))
	for i, user := range users {
		usersInfo[i] = modelsChats.UserInfo{
			UserID: user.UserId,
			ChatID: chatID,
			Role:   user.Role,
		}
	}

	err = uc.chatsRepo.InsertUsersToChat(ctx, chatID, usersInfo)
	if err != nil {
		return err
	}
	return nil
}

func (uc *ChatsUsecase) DeleteChat(ctx context.Context, userId, chatId uuid.UUID) error {
	return uc.chatsRepo.DeleteChat(ctx, userId, chatId)
}

func (uc *ChatsUsecase) UpdateChat(ctx context.Context, userId, chatId uuid.UUID, name, description string) error {
	return uc.chatsRepo.UpdateChat(ctx, userId, chatId, name, description)
}

func (uc *ChatsUsecase) GetChatAvatars(ctx context.Context, userId uuid.UUID, chatIDs []uuid.UUID) (map[string]*string, error) {
	const op = "ChatsUsecase.GetChatAvatars"

	logger := domains.GetLogger(ctx).WithField("operation", op).WithField("chat_ids_count", len(chatIDs))
	logger.Debug("Starting usecase operation: get chat avatars")

	// Инициализируем карту для всех запрошенных ID со значением nil
	avatars := make(map[string]*string, len(chatIDs))
	for _, chatID := range chatIDs {
		avatars[chatID.String()] = nil
	}

	avatarsIDs, err := uc.chatsRepo.GetChatAvatars(ctx, userId, chatIDs)
	if err != nil {
		logger.WithError(err).Error("Failed to get chat avatars from repository")
		return nil, err
	}

	logger.WithField("avatars_count", len(avatarsIDs)).Debug("Got avatar IDs from repository")
	for chatID, attachmentID := range avatarsIDs {
		logger.WithField("chat_id", chatID).WithField("attachment_id", attachmentID.String()).Debug("Getting URL from file storage")
		url, err := uc.fileStorage.GetOne(ctx, &attachmentID)
		if err != nil {
			logger.WithError(err).WithField("chat_id", chatID).WithField("attachment_id", attachmentID.String()).Error("Failed to get URL from file storage")
			avatars[chatID] = nil
		} else {
			logger.WithField("chat_id", chatID).WithField("url", url).Debug("Got URL from file storage")
			u := url
			avatars[chatID] = &u
		}
	}

	logger.WithField("avatars_count", len(avatars)).Info("Usecase operation completed successfully: chat avatars retrieved")
	return avatars, nil
}

func (uc *ChatsUsecase) UploadChatAvatar(ctx context.Context, userID, chatID uuid.UUID, fileData minio.FileData) (string, error) {
	const op = "ChatsUsecase.UploadChatAvatar"
	logger := domains.GetLogger(ctx).WithField("operation", op)

	isAdmin, err := uc.chatsRepo.CheckUserHasRole(ctx, userID, chatID, modelsChats.RoleAdmin)
	if err != nil {
		logger.WithError(err).Error("Failed to check user role")
		return "", err
	}

	if !isAdmin {
		logger.Error("User is not admin")
		return "", errs.ErrNoRights
	}

	// Генерируем UUID для файла в MinIO
	attachmentID := uuid.New()

	// Сохраняем файл в MinIO
	avatarURL, err := uc.fileStorage.CreateOne(ctx, fileData, attachmentID)
	if err != nil {
		logger.WithError(err).Error("Failed to save avatar file")
		return "", err
	}

	// Сохраняем attachment_id в БД
	err = uc.chatsRepo.UpdateChatAvatar(ctx, chatID, attachmentID, int64(len(fileData.Data)))
	if err != nil {
		logger.WithError(err).Error("Failed to update chat avatar in repository")
		return "", err
	}

	logger.Info("Chat avatar uploaded successfully")
	return avatarURL, nil
}

func (uc *ChatsUsecase) SearchChats(ctx context.Context, userID uuid.UUID, name string) ([]dtoChats.ChatViewInformationDTO, error) {
	const op = "ChatsUsecase.SearchChats"

	logger := domains.GetLogger(ctx).WithField("operation", op)

	chats, err := uc.chatsRepo.SearchChats(ctx, userID, name)
	if err != nil {
		logger.WithError(err).Errorf("Failed to search chats for user %s with name query '%s'", userID, name)
		return nil, err
	}

	if len(chats) == 0 {
		return []dtoChats.ChatViewInformationDTO{}, nil
	}

	chatsIDs := make([]uuid.UUID, len(chats))
	for i, chat := range chats {
		chatsIDs[i] = chat.ID
	}

	lastMessages, err := uc.messageRepo.GetLastMessagesOfChatsByIDs(ctx, chatsIDs)
	if err != nil {
		logger.WithError(err).Errorf("Failed to get last messages for searched chats for user %s", userID)
		return nil, err
	}

	result := make([]dtoChats.ChatViewInformationDTO, len(chats))
	for i, chat := range chats {
		lastMessage := lastMessages[chat.ID]
		result[i] = dtoChats.ChatViewInformationDTO{
			ID:   chat.ID,
			Name: chat.Name,
			Type: chat.Type,
			LastMessage: dtoMessage.MessageDTO{
				ID:         lastMessage.ID,
				SenderID:   lastMessage.UserID,
				SenderName: lastMessage.UserName,
				Text:       lastMessage.Text,
				CreatedAt:  lastMessage.CreatedAt,
				UpdatedAt:  lastMessage.UpdatedAt,
				ChatID:     lastMessage.ChatID,
				Type:       lastMessage.Type,
			},
		}
	}

	return result, nil
}
