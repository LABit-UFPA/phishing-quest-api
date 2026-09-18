package tests

import (
	"bytes"
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
	"github.com/stretchr/testify/mock"
	"gorm.io/datatypes"
)

// TestItemHandler_CreateAndGet_FluxoCompleto e um teste de integracao
// (via httptest, sem banco real) do modulo de items introduzido pela
// issue #20 — generaliza phishing_emails (hoje so mockado no front)
// para suportar multiplos canais.
func TestItemHandler_CreateAndGet_FluxoCompleto(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(MockItemRepository)
	uc := usecase.NewItemUseCase(mockRepo)
	h := handler.NewItemHandler(uc)

	r := gin.New()
	r.POST("/api/v1/items", h.CreateItem)
	r.GET("/api/v1/items/:id", h.GetItem)

	mockRepo.On("Create", mock.AnythingOfType("*domain.Item")).Return(nil)

	body := map[string]interface{}{
		"channel":     "sms",
		"isMalicious": true,
		"contentJson": map[string]string{"text": "Sua encomenda esta retida, pague a taxa: bit.ly/xyz"},
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/items", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var created domain.Item
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &created))
	assert.Equal(t, domain.ChannelSMS, created.Channel)
	assert.True(t, created.IsMalicious)
	assert.NotEqual(t, uuid.Nil, created.Id)

	mockRepo.AssertExpectations(t)
}

func TestItemHandler_CreateItem_CanalInvalidoRetornaErro(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(MockItemRepository)
	uc := usecase.NewItemUseCase(mockRepo)
	h := handler.NewItemHandler(uc)

	r := gin.New()
	r.POST("/api/v1/items", h.CreateItem)

	body := map[string]interface{}{
		"channel":     "carta_pombo",
		"contentJson": map[string]string{},
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/items", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.NotEqual(t, http.StatusOK, w.Code)
	mockRepo.AssertNotCalled(t, "Create", mock.Anything)
}

func TestItemHandler_ListItems_FiltraPorCanal(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(MockItemRepository)
	uc := usecase.NewItemUseCase(mockRepo)
	h := handler.NewItemHandler(uc)

	r := gin.New()
	r.GET("/api/v1/items", h.ListItems)

	expected := []*domain.Item{
		{Id: uuid.New(), Channel: domain.ChannelPixQR, IsMalicious: true, ContentJSON: datatypes.JSON(`{}`)},
	}
	mockRepo.On("GetByChannel", domain.ChannelPixQR).Return(expected, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/items?channel=pix_qr", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var items []domain.Item
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &items))
	assert.Len(t, items, 1)
	assert.Equal(t, domain.ChannelPixQR, items[0].Channel)

	mockRepo.AssertExpectations(t)
	mockRepo.AssertNotCalled(t, "GetAll")
}

func TestItemHandler_GetItem_IDInvalido(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(MockItemRepository)
	uc := usecase.NewItemUseCase(mockRepo)
	h := handler.NewItemHandler(uc)

	r := gin.New()
	r.GET("/api/v1/items/:id", h.GetItem)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/items/not-a-uuid", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
