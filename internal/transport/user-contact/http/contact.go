package transport

import (
	"encoding/json"
	"net/http"
	"time"

	ContactDTO "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/dto/contact"
	gen "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/generated/user"
	contextUtils "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/utils/context"
	grpcUtils "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/utils/grpc"
	utils "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/utils/response"
)

// CreateContact создает новый контакт через gRPC
// @Summary      Добавление контакта
// @Description  Добавляет нового пользователя в список контактов текущего пользователя
// @Tags         contacts
// @Accept       json
// @Produce      json
// @Param X-CSRF-Token header string true "CSRF Token"
// @Param        contact  body  dto.PostContactDTO  true  "Данные контакта для добавления"
// @Success      201   "Контакт успешно добавлен"
// @Failure      400   {object}  dto.ErrorDTO  			  "Неверные данные запроса"
// @Failure      401   {object}  dto.ErrorDTO 			  "Неавторизованный доступ"
// @Failure      404   {object}  dto.ErrorDTO 			  "Пользователь не найден"
// @Failure      500   {object}  dto.ErrorDTO          	  "Внутренняя ошибка сервера"
// @Security     ApiKeyAuth
// @Router       /contacts [post]
func (h *UserGRPCProxyHandler) CreateContact(w http.ResponseWriter, r *http.Request) {
	const op = "UserGRPCProxyHandler.CreateContact"

	userID, err := contextUtils.GetUserIDFromContext(r)
	if err != nil {
		utils.SendError(r.Context(), op, w, http.StatusUnauthorized, err.Error())
		return
	}

	var req ContactDTO.PostContactDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.SendError(r.Context(), op, w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	if userID == req.ContactUserID {
		utils.SendError(r.Context(), op, w, http.StatusBadRequest, "Cannot add yourself as contact")
		return
	}

	_, err = h.userClient.CreateContact(r.Context(), &gen.CreateContactReq{
		UserId:        userID.String(),
		ContactUserId: req.ContactUserID.String(),
	})
	if err != nil {
		grpcUtils.HandleGRPCError(r.Context(), op, w, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// GetContacts получает список контактов через gRPC
// @Summary      Получение списка контактов
// @Description  Возвращает список всех контактов текущего пользователя с полной информацией о пользователях
// @Tags         contacts
// @Accept       json
// @Produce      json
// @Success      200   {array}   dto.GetContactsDTO      "Список контактов успешно получен"
// @Failure      401   {object}  dto.ErrorDTO   		 "Неавторизованный доступ"
// @Failure      500   {object}  dto.ErrorDTO 			 "Внутренняя ошибка сервера"
// @Security     ApiKeyAuth
// @Router       /contacts [get]
func (h *UserGRPCProxyHandler) GetContacts(w http.ResponseWriter, r *http.Request) {
	const op = "UserGRPCProxyHandler.GetContacts"

	userID, err := contextUtils.GetUserIDFromContext(r)
	if err != nil {
		utils.SendError(r.Context(), op, w, http.StatusUnauthorized, err.Error())
		return
	}

	res, err := h.userClient.GetContacts(r.Context(), &gen.GetContactsReq{
		UserId: userID.String(),
	})
	if err != nil {
		grpcUtils.HandleGRPCError(r.Context(), op, w, err)
		return
	}

	contacts := make([]*ContactDTO.GetContactsDTO, 0)
	for _, c := range res.Contacts {
		createdAt, _ := time.Parse(time.RFC3339, c.CreatedAt)
		updatedAt, _ := time.Parse(time.RFC3339, c.UpdatedAt)

		contacts = append(contacts, &ContactDTO.GetContactsDTO{
			UserID:      userID,
			ContactUser: mapProtoUserToDTO(convertContactToUser(c)),
			CreatedAt:   createdAt,
			UpdatedAt:   updatedAt,
		})
	}

	utils.SendJSONResponse(r.Context(), op, w, http.StatusOK, contacts)
}

// SearchContacts выполняет поиск контактов через gRPC
// @Summary      Поиск контактов
// @Description  Возвращает список контактов текущего пользователя, отфильтрованных по поисковому запросу (имя или username)
// @Tags         contacts
// @Accept       json
// @Produce      json
// @Param        query  query     string  true  "Поисковый запрос"
// @Success      200   {array}   dto.GetContactsDTO      "Список найденных контактов"
// @Failure      400   {object}  dto.ErrorDTO   		 "Неверный запрос"
// @Failure      401   {object}  dto.ErrorDTO   		 "Неавторизованный доступ"
// @Failure      500   {object}  dto.ErrorDTO 			 "Внутренняя ошибка сервера"
// @Security     ApiKeyAuth
// @Router       /contacts/search [get]
func (h *UserGRPCProxyHandler) SearchContacts(w http.ResponseWriter, r *http.Request) {
	const op = "UserGRPCProxyHandler.SearchContacts"

	userID, err := contextUtils.GetUserIDFromContext(r)
	if err != nil {
		utils.SendError(r.Context(), op, w, http.StatusUnauthorized, err.Error())
		return
	}

	query := r.URL.Query().Get("query")
	if query == "" {
		utils.SendError(r.Context(), op, w, http.StatusBadRequest, "query parameter is required")
		return
	}

	res, err := h.userClient.SearchContacts(r.Context(), &gen.SearchContactsReq{
		UserId: userID.String(),
		Query:  query,
	})
	if err != nil {
		grpcUtils.HandleGRPCError(r.Context(), op, w, err)
		return
	}

	contacts := make([]*ContactDTO.GetContactsDTO, 0)
	for _, c := range res.Contacts {
		createdAt, _ := time.Parse(time.RFC3339, c.CreatedAt)
		updatedAt, _ := time.Parse(time.RFC3339, c.UpdatedAt)

		contacts = append(contacts, &ContactDTO.GetContactsDTO{
			UserID:      userID,
			ContactUser: mapProtoUserToDTO(convertContactToUser(c)),
			CreatedAt:   createdAt,
			UpdatedAt:   updatedAt,
		})
	}

	utils.SendJSONResponse(r.Context(), op, w, http.StatusOK, contacts)
}

func convertContactToUser(contact *gen.Contact) *gen.User {
	return &gen.User{
		Id:          contact.Id,
		PhoneNumber: contact.PhoneNumber,
		Name:        contact.Name,
		Username:    contact.Username,
		Bio:         contact.Bio,
		AvatarUrl:   contact.AvatarUrl,
		AccountType: contact.AccountType,
		CreatedAt:   contact.CreatedAt,
		UpdatedAt:   contact.UpdatedAt,
	}
}
