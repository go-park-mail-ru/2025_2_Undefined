package transport

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-park-mail-ru/2025_2_Undefined/internal/models/domains"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/models/errs"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/dto"
	"github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/jwt"
	utils "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/utils/response"
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

// GetChats получает список всех чатов пользователя
// @Summary      Получить список чатов
// @Description  Возвращает список всех чатов текущего пользователя с информацией о последнем сообщении
// @Tags         chats
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Success      200  {array}   dto.ChatViewInformationDTO  "Список чатов"
// @Failure      400  {object}  dto.ErrorDTO                "Некорректный запрос"
// @Failure      401  {object}  dto.ErrorDTO                "Неавторизованный доступ"
// @Router       /chats [get]
func (h *ChatsHandler) GetChats(w http.ResponseWriter, r *http.Request) {
	// ! Получаем id пользователя из JWT токена
	jwtCookie, err := r.Cookie(domains.TokenCookieName)
	if err != nil {
		utils.SendError(w, http.StatusUnauthorized, `{"error": "JWT token требуется"}`)
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

// PostChats создает новый чат
// @Summary      Создать новый чат
// @Description  Создает новый чат с указанными участниками и настройками
// @Tags         chats
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        chat  body      dto.ChatCreateInformationDTO  true  "Данные для создания чата"
// @Success      201   {object}  dto.IdDTO                     "ID созданного чата"
// @Failure      400   {object}  dto.ErrorDTO                  "Некорректный запрос"
// @Failure      401   {object}  dto.ErrorDTO                  "Неавторизованный доступ"
// @Router       /chats [post]
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

// GetInformationAboutChat получает детальную информацию о чате
// @Summary      Получить информацию о чате
// @Description  Возвращает детальную информацию о конкретном чате, включая сообщения и участников
// @Tags         chats
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        chatId  path      string  true  "ID чата"  format(uuid)
// @Success      200     {object}  dto.ChatDetailedInformationDTO  "Детальная информация о чате"
// @Failure      400     {object}  dto.ErrorDTO                    "Некорректный запрос"
// @Failure      401     {object}  dto.ErrorDTO                    "Неавторизованный доступ"
// @Router       /chats/{chatId} [get]
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
		utils.SendError(w, http.StatusUnauthorized, errs.ErrJWTIsRequired.Error())
		return
	}
	jwttoken := jwt.NewTokenator()
	claims, err := jwttoken.ParseJWT(jwtCookie.Value)
	if err != nil {
		utils.SendError(w, http.StatusUnauthorized, err.Error())
		return
	}
	userUUID, err := uuid.Parse(claims.UserID)
	if err != nil {
		utils.SendError(w, http.StatusUnauthorized, err.Error())
		return
	}

	informationDTO, err := h.chatService.GetInformationAboutChat(userUUID, chatUUID)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.SendJSONResponse(w, http.StatusOK, informationDTO)
}
