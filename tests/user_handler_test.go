package tests

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"phishing-quest/adapter/http/handler"
	"phishing-quest/core/usecase"
	"phishing-quest/domain"
	"phishing-quest/dto"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// TestUserHandler_GetUser_RetornaUsuarioSemSenha e a regressao da issue
// #16: GET /api/v1/users/:id era um stub que so ecoava o parametro
// (c.JSON(200, id)), sem consultar o banco. Agora deve buscar de
// verdade e responder com o DTO seguro (sem senha/hash).
func TestUserHandler_GetUser_RetornaUsuarioSemSenha(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(MockUserRepository)
	mockJWT := new(MockJWTService)
	uc := usecase.NewUserUseCase(mockRepo, mockJWT)
	h := handler.NewUserHandler(uc)

	userID := uuid.New()
	existingUser := &domain.User{
		Id:           userID,
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: "$2a$10$hash-nao-deve-aparecer",
		TotalScore:   15,
	}
	mockRepo.On("GetByID", userID).Return(existingUser, nil)

	r := gin.New()
	r.GET("/api/v1/users/:id", h.GetUser)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/"+userID.String(), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NotContains(t, w.Body.String(), "hash-nao-deve-aparecer")
	assert.NotContains(t, w.Body.String(), "passwordHash")

	var response dto.UserResponseDTO
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &response))
	assert.Equal(t, userID, response.Id)
	assert.Equal(t, existingUser.Username, response.Username)
	assert.Equal(t, existingUser.TotalScore, response.TotalScore)

	mockRepo.AssertExpectations(t)
}

func TestUserHandler_GetUser_IDInvalido(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(MockUserRepository)
	mockJWT := new(MockJWTService)
	uc := usecase.NewUserUseCase(mockRepo, mockJWT)
	h := handler.NewUserHandler(uc)

	r := gin.New()
	r.GET("/api/v1/users/:id", h.GetUser)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/not-a-uuid", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockRepo.AssertNotCalled(t, "GetByID", mock.Anything)
}

func TestUserHandler_GetUser_NaoEncontrado(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(MockUserRepository)
	mockJWT := new(MockJWTService)
	uc := usecase.NewUserUseCase(mockRepo, mockJWT)
	h := handler.NewUserHandler(uc)

	userID := uuid.New()
	mockRepo.On("GetByID", userID).Return(nil, gorm.ErrRecordNotFound)

	r := gin.New()
	r.GET("/api/v1/users/:id", h.GetUser)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/"+userID.String(), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockRepo.AssertExpectations(t)
}

func TestUserUseCase_GetUser_PropagaErroDoRepo(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockJWT := new(MockJWTService)
	uc := usecase.NewUserUseCase(mockRepo, mockJWT)

	userID := uuid.New()
	mockRepo.On("GetByID", userID).Return(nil, gorm.ErrRecordNotFound)

	user, err := uc.GetUser(userID)

	assert.Nil(t, user)
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
	mockRepo.AssertExpectations(t)
}
