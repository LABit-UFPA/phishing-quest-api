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
	"gorm.io/datatypes"
)

func TestItemHandler_ListItems_FiltraPorCanal(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(MockItemRepository)
	uc := usecase.NewItemUseCase(mockRepo)
	h := handler.NewItemHandler(uc)

	r := gin.New()
	r.GET("/api/v1/items", h.ListItems)

	expected := []*domain.Item{
		{Id: uuid.New(), Channel: domain.ChannelPixQR, IsMalicious: true, ContentJSON: datatypes.JSON(`{}`), Status: domain.StatusPublished},
	}
	mockRepo.On("GetPublishedByChannel", domain.ChannelPixQR).Return(expected, nil)

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

// TestItemHandler_GetItem_RascunhoRetorna404 garante, no nivel HTTP,
// que a rota publica nao expoe item pendente de revisao (issue #30).
func TestItemHandler_GetItem_RascunhoRetorna404(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(MockItemRepository)
	uc := usecase.NewItemUseCase(mockRepo)
	h := handler.NewItemHandler(uc)

	r := gin.New()
	r.GET("/api/v1/items/:id", h.GetItem)

	id := uuid.New()
	mockRepo.On("GetByID", id).Return(&domain.Item{Id: id, Status: domain.StatusDraft}, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/items/"+id.String(), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
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
