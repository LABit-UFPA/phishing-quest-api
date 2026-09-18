package tests

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"phishing-quest/adapter/http/handler"
	"phishing-quest/core/usecase"
	"phishing-quest/domain"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestItemSelectionHandler_NextItem_Sucesso(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockItemRepo := new(MockItemRepository)
	uc := usecase.NewItemSelectionUseCase(mockItemRepo)
	h := handler.NewItemSelectionHandler(uc)

	userID := uuid.New()
	sessionID := uuid.New()
	expectedItem := &domain.Item{Id: uuid.New(), Channel: domain.ChannelEmail}

	mockItemRepo.On("CountSeenInSession", userID, sessionID).Return(int64(0), int64(0), nil)
	mockItemRepo.On("GetRandomUnseen", userID, sessionID, (*bool)(nil)).Return(expectedItem, nil)

	r := gin.New()
	r.GET("/api/v1/items/next", func(c *gin.Context) {
		c.Set("userId", userID) // simula o AuthRequired ja tendo rodado
		c.Next()
	}, h.NextItem)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/items/next?sessionId="+sessionID.String(), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var item domain.Item
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &item))
	assert.Equal(t, expectedItem.Id, item.Id)
}

func TestItemSelectionHandler_NextItem_SemSessionId(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockItemRepo := new(MockItemRepository)
	uc := usecase.NewItemSelectionUseCase(mockItemRepo)
	h := handler.NewItemSelectionHandler(uc)

	r := gin.New()
	r.GET("/api/v1/items/next", func(c *gin.Context) {
		c.Set("userId", uuid.New())
		c.Next()
	}, h.NextItem)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/items/next", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockItemRepo.AssertNotCalled(t, "CountSeenInSession")
}

// TestItemSelectionHandler_NextItem_SemUserIdNoContexto garante que o
// handler nunca aceita um userId vindo do cliente (query/body) — so do
// contexto de autenticacao. Sem AuthRequired ter setado userId, deve
// responder 401 em vez de aceitar qualquer coisa.
func TestItemSelectionHandler_NextItem_SemUserIdNoContexto(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockItemRepo := new(MockItemRepository)
	uc := usecase.NewItemSelectionUseCase(mockItemRepo)
	h := handler.NewItemSelectionHandler(uc)

	r := gin.New()
	r.GET("/api/v1/items/next", h.NextItem) // sem middleware simulando auth

	req := httptest.NewRequest(http.MethodGet, "/api/v1/items/next?sessionId="+uuid.New().String(), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	mockItemRepo.AssertNotCalled(t, "CountSeenInSession")
}

func TestItemSelectionHandler_NextItem_SemItensDisponiveisRetorna404(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockItemRepo := new(MockItemRepository)
	uc := usecase.NewItemSelectionUseCase(mockItemRepo)
	h := handler.NewItemSelectionHandler(uc)

	userID := uuid.New()
	sessionID := uuid.New()

	mockItemRepo.On("CountSeenInSession", userID, sessionID).Return(int64(3), int64(3), nil)
	mockItemRepo.On("GetRandomUnseen", userID, sessionID, (*bool)(nil)).Return(nil, assert.AnError)

	r := gin.New()
	r.GET("/api/v1/items/next", func(c *gin.Context) {
		c.Set("userId", userID)
		c.Next()
	}, h.NextItem)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/items/next?sessionId="+sessionID.String(), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}
