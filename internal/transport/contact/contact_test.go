package transport

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2025_2_Undefined/internal/models/errs"
	ContactDTO "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/dto/contact"
	UserDTO "github.com/go-park-mail-ru/2025_2_Undefined/internal/transport/dto/user"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockContactUsecase struct {
	mock.Mock
}

func (m *MockContactUsecase) CreateContact(ctx context.Context, req *ContactDTO.PostContactDTO, userID uuid.UUID) error {
	args := m.Called(ctx, req, userID)
	return args.Error(0)
}

func (m *MockContactUsecase) GetContacts(ctx context.Context, userID uuid.UUID) ([]*ContactDTO.GetContactsDTO, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*ContactDTO.GetContactsDTO), args.Error(1)
}

type MockSessionUtils struct {
	mock.Mock
}

func (m *MockSessionUtils) GetUserIDFromSession(r *http.Request) (uuid.UUID, error) {
	args := m.Called(r)
	return args.Get(0).(uuid.UUID), args.Error(1)
}

func TestContactHandler_CreateContact_Success(t *testing.T) {
	mockUsecase := new(MockContactUsecase)
	mockSessionUtils := new(MockSessionUtils)

	handler := New(mockUsecase, mockSessionUtils)

	userID := uuid.New()
	contactUserID := uuid.New()
	req := ContactDTO.PostContactDTO{
		ContactUserID: contactUserID,
	}

	mockSessionUtils.On("GetUserIDFromSession", mock.Anything).Return(userID, nil)
	mockUsecase.On("CreateContact", mock.Anything, &req, userID).Return(nil)

	body, _ := json.Marshal(req)
	request := httptest.NewRequest(http.MethodPost, "/contacts", bytes.NewBuffer(body))
	request.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()

	handler.CreateContact(recorder, request)

	assert.Equal(t, http.StatusCreated, recorder.Code)
	mockSessionUtils.AssertExpectations(t)
	mockUsecase.AssertExpectations(t)
}

func TestContactHandler_CreateContact_Unauthorized(t *testing.T) {
	mockUsecase := new(MockContactUsecase)
	mockSessionUtils := new(MockSessionUtils)

	handler := New(mockUsecase, mockSessionUtils)

	mockSessionUtils.On("GetUserIDFromSession", mock.Anything).Return(uuid.Nil, errors.New("unauthorized"))

	request := httptest.NewRequest(http.MethodPost, "/contacts", nil)
	recorder := httptest.NewRecorder()

	handler.CreateContact(recorder, request)

	assert.Equal(t, http.StatusUnauthorized, recorder.Code)
	mockSessionUtils.AssertExpectations(t)
}

func TestContactHandler_CreateContact_InvalidJSON(t *testing.T) {
	mockUsecase := new(MockContactUsecase)
	mockSessionUtils := new(MockSessionUtils)

	handler := New(mockUsecase, mockSessionUtils)

	userID := uuid.New()
	mockSessionUtils.On("GetUserIDFromSession", mock.Anything).Return(userID, nil)

	request := httptest.NewRequest(http.MethodPost, "/contacts", bytes.NewBuffer([]byte("invalid json")))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	handler.CreateContact(recorder, request)

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
	mockSessionUtils.AssertExpectations(t)
}

func TestContactHandler_CreateContact_SelfContact(t *testing.T) {
	mockUsecase := new(MockContactUsecase)
	mockSessionUtils := new(MockSessionUtils)

	handler := New(mockUsecase, mockSessionUtils)

	userID := uuid.New()
	req := ContactDTO.PostContactDTO{
		ContactUserID: userID,
	}

	mockSessionUtils.On("GetUserIDFromSession", mock.Anything).Return(userID, nil)

	body, _ := json.Marshal(req)
	request := httptest.NewRequest(http.MethodPost, "/contacts", bytes.NewBuffer(body))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	handler.CreateContact(recorder, request)

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
	mockSessionUtils.AssertExpectations(t)
}

func TestContactHandler_CreateContact_DuplicateContact(t *testing.T) {
	mockUsecase := new(MockContactUsecase)
	mockSessionUtils := new(MockSessionUtils)

	handler := New(mockUsecase, mockSessionUtils)

	userID := uuid.New()
	contactUserID := uuid.New()
	req := ContactDTO.PostContactDTO{
		ContactUserID: contactUserID,
	}

	mockSessionUtils.On("GetUserIDFromSession", mock.Anything).Return(userID, nil)
	mockUsecase.On("CreateContact", mock.Anything, &req, userID).Return(errs.ErrIsDuplicateKey)

	body, _ := json.Marshal(req)
	request := httptest.NewRequest(http.MethodPost, "/contacts", bytes.NewBuffer(body))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	handler.CreateContact(recorder, request)

	assert.Equal(t, http.StatusConflict, recorder.Code)
	mockSessionUtils.AssertExpectations(t)
	mockUsecase.AssertExpectations(t)
}

func TestContactHandler_CreateContact_UserNotFound(t *testing.T) {
	mockUsecase := new(MockContactUsecase)
	mockSessionUtils := new(MockSessionUtils)

	handler := New(mockUsecase, mockSessionUtils)

	userID := uuid.New()
	contactUserID := uuid.New()
	req := ContactDTO.PostContactDTO{
		ContactUserID: contactUserID,
	}

	mockSessionUtils.On("GetUserIDFromSession", mock.Anything).Return(userID, nil)
	mockUsecase.On("CreateContact", mock.Anything, &req, userID).Return(errs.ErrUserNotFound)

	body, _ := json.Marshal(req)
	request := httptest.NewRequest(http.MethodPost, "/contacts", bytes.NewBuffer(body))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	handler.CreateContact(recorder, request)

	assert.Equal(t, http.StatusNotFound, recorder.Code)
	mockSessionUtils.AssertExpectations(t)
	mockUsecase.AssertExpectations(t)
}

func TestContactHandler_CreateContact_InternalError(t *testing.T) {
	mockUsecase := new(MockContactUsecase)
	mockSessionUtils := new(MockSessionUtils)

	handler := New(mockUsecase, mockSessionUtils)

	userID := uuid.New()
	contactUserID := uuid.New()
	req := ContactDTO.PostContactDTO{
		ContactUserID: contactUserID,
	}

	mockSessionUtils.On("GetUserIDFromSession", mock.Anything).Return(userID, nil)
	mockUsecase.On("CreateContact", mock.Anything, &req, userID).Return(errors.New("internal error"))

	body, _ := json.Marshal(req)
	request := httptest.NewRequest(http.MethodPost, "/contacts", bytes.NewBuffer(body))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	handler.CreateContact(recorder, request)

	assert.Equal(t, http.StatusInternalServerError, recorder.Code)
	mockSessionUtils.AssertExpectations(t)
	mockUsecase.AssertExpectations(t)
}

func TestContactHandler_GetContacts_Success(t *testing.T) {
	mockUsecase := new(MockContactUsecase)
	mockSessionUtils := new(MockSessionUtils)

	handler := New(mockUsecase, mockSessionUtils)

	userID := uuid.New()
	contactUserID := uuid.New()

	contacts := []*ContactDTO.GetContactsDTO{
		{
			UserID: userID,
			ContactUser: &UserDTO.User{
				ID:          contactUserID,
				Name:        "Contact User",
				PhoneNumber: "+79998887777",
				Username:    "contact_user",
				AccountType: "user",
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	mockSessionUtils.On("GetUserIDFromSession", mock.Anything).Return(userID, nil)
	mockUsecase.On("GetContacts", mock.Anything, userID).Return(contacts, nil)

	request := httptest.NewRequest(http.MethodGet, "/contacts", nil)
	recorder := httptest.NewRecorder()

	handler.GetContacts(recorder, request)

	assert.Equal(t, http.StatusOK, recorder.Code)

	var response []*ContactDTO.GetContactsDTO
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Len(t, response, 1)
	assert.Equal(t, userID, response[0].UserID)
	assert.Equal(t, contactUserID, response[0].ContactUser.ID)

	mockSessionUtils.AssertExpectations(t)
	mockUsecase.AssertExpectations(t)
}

func TestContactHandler_GetContacts_Unauthorized(t *testing.T) {
	mockUsecase := new(MockContactUsecase)
	mockSessionUtils := new(MockSessionUtils)

	handler := New(mockUsecase, mockSessionUtils)

	mockSessionUtils.On("GetUserIDFromSession", mock.Anything).Return(uuid.Nil, errors.New("unauthorized"))

	request := httptest.NewRequest(http.MethodGet, "/contacts", nil)
	recorder := httptest.NewRecorder()

	handler.GetContacts(recorder, request)

	assert.Equal(t, http.StatusUnauthorized, recorder.Code)
	mockSessionUtils.AssertExpectations(t)
}

func TestContactHandler_GetContacts_EmptyList(t *testing.T) {
	mockUsecase := new(MockContactUsecase)
	mockSessionUtils := new(MockSessionUtils)

	handler := New(mockUsecase, mockSessionUtils)

	userID := uuid.New()

	mockSessionUtils.On("GetUserIDFromSession", mock.Anything).Return(userID, nil)
	mockUsecase.On("GetContacts", mock.Anything, userID).Return([]*ContactDTO.GetContactsDTO{}, nil)

	request := httptest.NewRequest(http.MethodGet, "/contacts", nil)
	recorder := httptest.NewRecorder()

	handler.GetContacts(recorder, request)

	assert.Equal(t, http.StatusOK, recorder.Code)

	var response []*ContactDTO.GetContactsDTO
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Empty(t, response)

	mockSessionUtils.AssertExpectations(t)
	mockUsecase.AssertExpectations(t)
}

func TestContactHandler_GetContacts_InternalError(t *testing.T) {
	mockUsecase := new(MockContactUsecase)
	mockSessionUtils := new(MockSessionUtils)

	handler := New(mockUsecase, mockSessionUtils)

	userID := uuid.New()

	mockSessionUtils.On("GetUserIDFromSession", mock.Anything).Return(userID, nil)
	mockUsecase.On("GetContacts", mock.Anything, userID).Return(nil, errors.New("internal error"))

	request := httptest.NewRequest(http.MethodGet, "/contacts", nil)
	recorder := httptest.NewRecorder()

	handler.GetContacts(recorder, request)

	assert.Equal(t, http.StatusInternalServerError, recorder.Code)
	mockSessionUtils.AssertExpectations(t)
	mockUsecase.AssertExpectations(t)
}
