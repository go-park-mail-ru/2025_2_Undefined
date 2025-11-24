package chats

import (
	"context"

	"github.com/go-park-mail-ru/2025_2_Undefined/internal/models/domains"
	mappers "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/chats-message/mappers"
	dtoUtils "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/dto/utils"
	gen "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/generated/chats"
	chatsInterface "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/interface/chats"
	messageInterface "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/interface/message"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type ChatsGRPCHandler struct {
	gen.UnimplementedChatServiceServer

	chatsUsecase   chatsInterface.ChatsUsecase
	messageUsecase messageInterface.MessageUsecase
}

func NewChatsGRPCHandler(chatsUC chatsInterface.ChatsUsecase, messageUC messageInterface.MessageUsecase) *ChatsGRPCHandler {
	return &ChatsGRPCHandler{
		chatsUsecase:   chatsUC,
		messageUsecase: messageUC,
	}
}

func (h *ChatsGRPCHandler) GetChats(ctx context.Context, in *gen.GetChatsReq) (*gen.GetChatsRes, error) {
	const op = "ChatsGRPCHandler.GetChats"
	logger := domains.GetLogger(ctx).WithField("operation", op)

	userID, err := uuid.Parse(in.GetUserId())
	if err != nil {
		logger.WithError(err).Errorf("error getting userID: %s", in.GetUserId())
		return nil, status.Error(codes.InvalidArgument, "wrong user id format")
	}

	chatsDTO, err := h.chatsUsecase.GetChats(ctx, userID)
	if err != nil {
		logger.WithError(err).Errorf("error getting chats for user %s: %v", in.GetUserId(), err)
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	response := &gen.GetChatsRes{
		Chats: mappers.DTOChatsViewToProto(chatsDTO),
	}

	return response, nil
}

func (h *ChatsGRPCHandler) CreateChat(ctx context.Context, in *gen.CreateChatReq) (*gen.IdRes, error) {
	const op = "ChatsGRPCHandler.CreateChat"
	logger := domains.GetLogger(ctx).WithField("operation", op)

	if err := validateChatCreateDTO(in); err != nil {
		logger.WithError(err).Errorf("validation error")
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	chatDTO, err := mappers.ProtoCreateChatToDTO(in)
	if err != nil {
		logger.WithError(err).Errorf("can't parse chatDTO: %v", chatDTO)
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	chatID, err := h.chatsUsecase.CreateChat(ctx, *chatDTO)
	if err != nil {
		logger.WithError(err).Errorf("error creating chat")
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	err = h.messageUsecase.SubscribeUsersOnChat(ctx, chatID, chatDTO.Members)
	if err != nil {
		logger.WithError(err).Error("Can't subscribe joined users to chat")
	}

	response := &gen.IdRes{
		Id: chatID.String(),
	}

	return response, nil
}

func (h *ChatsGRPCHandler) GetChat(ctx context.Context, in *gen.GetChatReq) (*gen.ChatDetailedInformation, error) {
	const op = "ChatsGRPCHandler.GetChat"
	logger := domains.GetLogger(ctx).WithField("operation", op)

	chatID, err := uuid.Parse(in.GetChatId())
	if err != nil {
		logger.WithError(err).Errorf("error parsing chatId: %s", in.GetChatId())
		return nil, status.Error(codes.InvalidArgument, "wrong chat id format")
	}

	userID, err := uuid.Parse(in.GetUserId())
	if err != nil {
		logger.WithError(err).Errorf("error parsing userId: %s", in.GetUserId())
		return nil, status.Error(codes.InvalidArgument, "wrong user id format")
	}

	chatDTO, err := h.chatsUsecase.GetInformationAboutChat(ctx, userID, chatID)
	if err != nil {
		logger.WithError(err).Errorf("error getting chat %s: %v", in.GetChatId(), err)
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	response := mappers.DTOChatDetailedToProto(chatDTO)

	return response, nil
}

func (h *ChatsGRPCHandler) GetUsersDialog(ctx context.Context, in *gen.GetUsersDialogReq) (*gen.IdRes, error) {
	const op = "ChatsGRPCHandler.GetUsersDialog"
	logger := domains.GetLogger(ctx).WithField("operation", op)

	user1ID, err := uuid.Parse(in.GetUser1Id())
	if err != nil {
		logger.WithError(err).Errorf("error parsing user1_id: %s", in.GetUser1Id())
		return nil, status.Error(codes.InvalidArgument, "wrong user1_id format")
	}

	user2ID, err := uuid.Parse(in.GetUser2Id())
	if err != nil {
		logger.WithError(err).Errorf("error parsing user2_id: %s", in.GetUser2Id())
		return nil, status.Error(codes.InvalidArgument, "wrong user2_id format")
	}

	dialogDTO, err := h.chatsUsecase.GetUsersDialog(ctx, user1ID, user2ID)
	if err != nil {
		logger.WithError(err).Errorf("error getting dialog between users %s and %s: %v", in.GetUser1Id(), in.GetUser2Id(), err)
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	response := &gen.IdRes{
		Id: dialogDTO.ID.String(),
	}

	return response, nil
}

func (h *ChatsGRPCHandler) AddUserToChat(ctx context.Context, in *gen.AddUserToChatReq) (*emptypb.Empty, error) {
	const op = "ChatsGRPCHandler.AddUserToChat"
	logger := domains.GetLogger(ctx).WithField("operation", op)

	userID, err := uuid.Parse(in.GetUserId())
	if err != nil {
		logger.WithError(err).Errorf("error parsing userId: %s", in.GetUserId())
		return nil, status.Error(codes.InvalidArgument, "wrong user id format")
	}

	chatID, err := uuid.Parse(in.GetChatId())
	if err != nil {
		logger.WithError(err).Errorf("error parsing chatId: %s", in.GetChatId())
		return nil, status.Error(codes.InvalidArgument, "wrong chat id format")
	}

	members := in.GetMembers()
	if err := validateAddMembers(members); err != nil {
		logger.Errorf("validation error: %v", err)
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	membersDTO, err := mappers.ProtoAddMembersToDTO(members)
	if err != nil {
		logger.WithError(err).Error("error mapping members to DTO")
		return nil, err
	}

	err = h.chatsUsecase.AddUsersToChat(ctx, chatID, userID, membersDTO)
	if err != nil {
		logger.WithError(err).Errorf("error adding users to chat %s: %v", chatID, err)
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &emptypb.Empty{}, nil
}

func (h *ChatsGRPCHandler) DeleteChat(ctx context.Context, in *gen.GetChatReq) (*emptypb.Empty, error) {
	const op = "ChatsGRPCHandler.DeleteChat"
	logger := domains.GetLogger(ctx).WithField("operation", op)

	chatID, err := uuid.Parse(in.GetChatId())
	if err != nil {
		logger.WithError(err).Errorf("error parsing chatId: %s", in.GetChatId())
		return nil, status.Error(codes.InvalidArgument, "wrong chat id format")
	}

	userID, err := uuid.Parse(in.GetUserId())
	if err != nil {
		logger.WithError(err).Errorf("error parsing userId: %s", in.GetUserId())
		return nil, status.Error(codes.InvalidArgument, "wrong user id format")
	}

	err = h.chatsUsecase.DeleteChat(ctx, userID, chatID)
	if err != nil {
		logger.WithError(err).Errorf("error deleting chat %s: %v", in.GetChatId(), err)
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &emptypb.Empty{}, nil
}

func (h *ChatsGRPCHandler) UpdateChat(ctx context.Context, in *gen.UpdateChatReq) (*emptypb.Empty, error) {
	const op = "ChatsGRPCHandler.UpdateChat"
	logger := domains.GetLogger(ctx).WithField("operation", op)

	chatID, err := uuid.Parse(in.GetChatId())
	if err != nil {
		logger.WithError(err).Errorf("error parsing chatId: %s", in.GetChatId())
		return nil, status.Error(codes.InvalidArgument, "wrong chat id format")
	}

	userID, err := uuid.Parse(in.GetUserId())
	if err != nil {
		logger.WithError(err).Errorf("error parsing userId: %s", in.GetUserId())
		return nil, status.Error(codes.InvalidArgument, "wrong user id format")
	}

	if err := validateChatUpdateReq(in); err != nil {
		logger.Errorf("validation error: %v", err)
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	name := in.GetName()
	description := in.GetDescription()

	err = h.chatsUsecase.UpdateChat(ctx, userID, chatID, name, description)
	if err != nil {
		logger.WithError(err).Errorf("error updating chat %s: %v", in.GetChatId(), err)
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &emptypb.Empty{}, nil
}

func (h *ChatsGRPCHandler) GetChatAvatars(ctx context.Context, in *gen.GetChatAvatarsReq) (*gen.GetChatAvatarsRes, error) {
	const op = "ChatsGRPCHandler.GetChatAvatars"
	logger := domains.GetLogger(ctx).WithField("operation", op)

	userIDStr := in.GetUserId()
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		logger.WithError(err).Errorf("error parsing userId: %s", userIDStr)
		return nil, status.Error(codes.InvalidArgument, "wrong user id format")
	}

	chatIDStrings := in.GetChatIds()
	if len(chatIDStrings) == 0 {
		logger.Debug("No chat IDs provided, returning empty response")
		return &gen.GetChatAvatarsRes{Avatars: make(map[string]string)}, nil
	}

	chatIDs := make([]uuid.UUID, 0, len(chatIDStrings))
	for _, idStr := range chatIDStrings {
		chatID, err := uuid.Parse(idStr)
		if err != nil {
			logger.WithError(err).Errorf("error parsing chatId: %s", idStr)
			return nil, status.Error(codes.InvalidArgument, "wrong chat id format")
		}
		chatIDs = append(chatIDs, chatID)
	}

	avatars, err := h.chatsUsecase.GetChatAvatars(ctx, userID, chatIDs)
	if err != nil {
		logger.WithError(err).Error("error getting chat avatars")
		return nil, status.Error(codes.Internal, err.Error())
	}

	response := &gen.GetChatAvatarsRes{
		Avatars: dtoUtils.PointerMapToStringMap(avatars),
	}

	return response, nil
}

func (h *ChatsGRPCHandler) UploadChatAvatar(ctx context.Context, in *gen.UploadChatAvatarReq) (*gen.UploadChatAvatarRes, error) {
	const op = "ChatsGRPCHandler.UploadChatAvatar"
	logger := domains.GetLogger(ctx).WithField("operation", op)

	chatID, err := uuid.Parse(in.GetChatId())
	if err != nil {
		logger.WithError(err).Errorf("error parsing chatId: %s", in.GetChatId())
		return nil, status.Error(codes.InvalidArgument, "wrong chat id format")
	}

	userID, err := uuid.Parse(in.GetUserId())
	if err != nil {
		logger.WithError(err).Errorf("error parsing userId: %s", in.GetUserId())
		return nil, status.Error(codes.InvalidArgument, "wrong user id format")
	}

	fileData := mappers.ProtoUploadChatAvatarReqToFileData(in)

	avatarURL, err := h.chatsUsecase.UploadChatAvatar(ctx, userID, chatID, fileData)
	if err != nil {
		logger.WithError(err).Error("error uploading chat avatar")
		return nil, status.Error(codes.Internal, err.Error())
	}

	response := &gen.UploadChatAvatarRes{
		AvatarUrl: avatarURL,
	}

	return response, nil
}

func (h *ChatsGRPCHandler) SearchChats(ctx context.Context, in *gen.SearchChatsReq) (*gen.GetChatsRes, error) {
	const op = "ChatsGRPCHandler.SearchChats"
	logger := domains.GetLogger(ctx).WithField("operation", op)

	userID, err := uuid.Parse(in.GetUserId())
	if err != nil {
		logger.WithError(err).Errorf("error parsing userId: %s", in.GetUserId())
		return nil, status.Error(codes.InvalidArgument, "wrong user id format")
	}

	nameQuery := in.GetName()

	chatsDTO, err := h.chatsUsecase.SearchChats(ctx, userID, nameQuery)
	if err != nil {
		logger.WithError(err).Error("Failed to search chats")
		return nil, status.Error(codes.Internal, "can't search chats")
	}

	response := &gen.GetChatsRes{
		Chats: mappers.DTOChatsViewToProto(chatsDTO),
	}

	return response, nil
}
