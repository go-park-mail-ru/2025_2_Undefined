package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-park-mail-ru/2025_2_Undefined/internal/handlers/dto"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/handlers/jwt"
	utils "github.com/go-park-mail-ru/2025_2_Undefined/internal/handlers/utils/response"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/models/domains"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/models/errs"
	"github.com/google/uuid"
)

type ChatsServiceInterface interface {
	GetChats(userId uuid.UUID) ([]dto.ChatViewInformationDTO, error)
	CreateChat(chatDTO dto.ChatCreateInformationDTO) (uuid.UUID, error)
	GetInformationAboutChat(userId, chatId uuid.UUID) (*dto.ChatDetailedInformationDTO, error)
}

type ChatsHandler struct {
	chatService ChatsServiceInterface
}

func NewChatsHandler(chatService ChatsServiceInterface) *ChatsHandler {
	return &ChatsHandler{
		chatService: chatService,
	}
}

func (h *ChatsHandler) GetChats(w http.ResponseWriter, r *http.Request) {
	// ! Получаем id пользователя из JWT токена
	jwtCookie, err := r.Cookie(domains.TokenCookieName)
	if err != nil {
		utils.SendError(w, http.StatusUnauthorized, `{"error": "JWT token required"}`)
		return
	}
	jwttoken := jwt.NewTokenator()
	claims, err := jwttoken.ParseJWT(jwtCookie.Value)
	if err != nil {
		utils.SendError(w, http.StatusUnauthorized, `{"error": "`+err.Error()+`"}`)
		return
	}
	userUUID, err := uuid.Parse(claims.UserID)
	if err != nil {
		utils.SendError(w, http.StatusUnauthorized, `{"error": "`+err.Error()+`"}`)
		return
	}

	chats, err := h.chatService.GetChats(userUUID)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, err.Error())
		return
	}
	utils.SendJSONResponse(w, http.StatusOK, chats)
}

func (h *ChatsHandler) PostChats(w http.ResponseWriter, r *http.Request) {

	chatDTO := &dto.ChatCreateInformationDTO{}

	if err := json.NewDecoder(r.Body).Decode(chatDTO); err != nil {
		utils.SendError(w, http.StatusBadRequest, err.Error())
		return
	}

	idOfCreatedChat, err := h.chatService.CreateChat(*chatDTO)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.SendJSONResponse(w, http.StatusCreated, dto.IdDTO{ID: idOfCreatedChat})
}

func (h *ChatsHandler) GetInformationAboutChat(w http.ResponseWriter, r *http.Request) {
	// Получаем id чата из пути
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 2 {
		utils.SendError(w, http.StatusBadRequest, errs.ErrBadRequest.Error())
		return
	}
	idStr := parts[len(parts)-1]
	chatUUID, err := uuid.Parse(idStr)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, errs.ErrBadRequest.Error())
		return
	}

	// ! Получаем id пользователя из JWT токена
	jwtCookie, err := r.Cookie(domains.TokenCookieName)
	if err != nil {
		utils.SendError(w, http.StatusUnauthorized, `{"error": "JWT token required"}`)
		return
	}
	jwttoken := jwt.NewTokenator()
	claims, err := jwttoken.ParseJWT(jwtCookie.Value)
	if err != nil {
		utils.SendError(w, http.StatusUnauthorized, `{"error": "`+err.Error()+`"}`)
		return
	}
	userUUID, err := uuid.Parse(claims.UserID)
	if err != nil {
		utils.SendError(w, http.StatusUnauthorized, `{"error": "`+err.Error()+`"}`)
		return
	}

	informationDTO, err := h.chatService.GetInformationAboutChat(userUUID, chatUUID)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.SendJSONResponse(w, http.StatusOK, informationDTO)
}
