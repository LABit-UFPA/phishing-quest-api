package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"phishing-quest/adapter/http/handler"
	"phishing-quest/core/service"
	"phishing-quest/core/usecase"
	"phishing-quest/domain"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func newItemReviewHandlerWithMocks() (*handler.ItemReviewHandler, *MockItemRepository) {
	mockRepo := new(MockItemRepository)
	uc := usecase.NewItemReviewUseCase(mockRepo, service.NewTemplateDraftService())
	return handler.NewItemReviewHandler(uc), mockRepo
}

func TestItemReviewHandler_CreateDraft_RetornaRascunho(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h, mockRepo := newItemReviewHandlerWithMocks()
	mockRepo.On("Create", mock.AnythingOfType("*domain.Item")).Return(nil)

	r := gin.New()
	r.POST("/api/v1/admin/items", h.CreateDraft)

	body, _ := json.Marshal(map[string]interface{}{
		"channel":     "sms",
		"isMalicious": true,
		"contentJson": map[string]string{"text": "Sua encomenda esta retida, pague a taxa: bit.ly/xyz"},
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/items", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var created domain.Item
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &created))
	assert.Equal(t, domain.StatusDraft, created.Status)
	assert.Equal(t, domain.ChannelSMS, created.Channel)
	assert.NotEqual(t, uuid.Nil, created.Id)
}

func TestItemReviewHandler_CreateDraft_CanalInvalidoRetorna400(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h, mockRepo := newItemReviewHandlerWithMocks()

	r := gin.New()
	r.POST("/api/v1/admin/items", h.CreateDraft)

	body, _ := json.Marshal(map[string]interface{}{
		"channel":     "carta_pombo",
		"contentJson": map[string]string{},
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/items", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockRepo.AssertNotCalled(t, "Create", mock.Anything)
}

// TestItemReviewHandler_Review_UsaRevisorDoJWT garante que a assinatura
// da revisao vem do usuario autenticado, nunca do corpo da requisicao.
func TestItemReviewHandler_Review_UsaRevisorDoJWT(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h, mockRepo := newItemReviewHandlerWithMocks()

	itemID := uuid.New()
	reviewerDoJWT := uuid.New()
	outroUsuario := uuid.New()
	draft := &domain.Item{Id: itemID, Status: domain.StatusDraft}

	mockRepo.On("GetByID", itemID).Return(draft, nil)
	mockRepo.On("Update", mock.AnythingOfType("*domain.Item")).Return(draft, nil)

	r := gin.New()
	r.POST("/api/v1/admin/items/:id/review", func(c *gin.Context) {
		c.Set("userId", reviewerDoJWT) // simula o AuthRequired
		c.Next()
	}, h.Review)

	// O corpo tenta declarar outro revisor — deve ser ignorado.
	body, _ := json.Marshal(map[string]string{"reviewedBy": outroUsuario.String()})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/items/"+itemID.String()+"/review", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var reviewed domain.Item
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &reviewed))
	assert.Equal(t, domain.StatusReviewed, reviewed.Status)
	assert.NotNil(t, reviewed.ReviewedBy)
	assert.Equal(t, reviewerDoJWT, *reviewed.ReviewedBy)
	assert.NotEqual(t, outroUsuario, *reviewed.ReviewedBy)
}

func TestItemReviewHandler_Review_SemUserIdNoContextoRetorna401(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h, mockRepo := newItemReviewHandlerWithMocks()

	r := gin.New()
	r.POST("/api/v1/admin/items/:id/review", h.Review) // sem middleware de auth

	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/items/"+uuid.New().String()+"/review", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	mockRepo.AssertNotCalled(t, "GetByID", mock.Anything)
}

// TestItemReviewHandler_Publish_RascunhoRetorna409 e o criterio de
// aceite visto pelo HTTP: publicar sem revisao humana e recusado.
func TestItemReviewHandler_Publish_RascunhoRetorna409(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h, mockRepo := newItemReviewHandlerWithMocks()

	itemID := uuid.New()
	mockRepo.On("GetByID", itemID).Return(&domain.Item{Id: itemID, Status: domain.StatusDraft}, nil)

	r := gin.New()
	r.POST("/api/v1/admin/items/:id/publish", h.Publish)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/items/"+itemID.String()+"/publish", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)
	mockRepo.AssertNotCalled(t, "Update", mock.Anything)
}

func TestItemReviewHandler_Publish_ItemRevisadoRetorna200(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h, mockRepo := newItemReviewHandlerWithMocks()

	itemID := uuid.New()
	reviewerID := uuid.New()
	item := &domain.Item{Id: itemID, Status: domain.StatusReviewed, ReviewedBy: &reviewerID}

	mockRepo.On("GetByID", itemID).Return(item, nil)
	mockRepo.On("Update", mock.AnythingOfType("*domain.Item")).Return(item, nil)

	r := gin.New()
	r.POST("/api/v1/admin/items/:id/publish", h.Publish)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/items/"+itemID.String()+"/publish", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var published domain.Item
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &published))
	assert.Equal(t, domain.StatusPublished, published.Status)
	assert.NotNil(t, published.PublishedAt)
}

func TestItemReviewHandler_ListByStatus_DefaultRascunhos(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h, mockRepo := newItemReviewHandlerWithMocks()

	mockRepo.On("GetByStatus", domain.StatusDraft).Return([]*domain.Item{
		{Id: uuid.New(), Status: domain.StatusDraft},
	}, nil)

	r := gin.New()
	r.GET("/api/v1/admin/items", h.ListByStatus)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/items", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockRepo.AssertExpectations(t)
}

func TestItemReviewHandler_ListByStatus_StatusInvalidoRetorna400(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h, mockRepo := newItemReviewHandlerWithMocks()

	r := gin.New()
	r.GET("/api/v1/admin/items", h.ListByStatus)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/items?status=publicado_ontem", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockRepo.AssertNotCalled(t, "GetByStatus", mock.Anything)
}

// TestItemReviewHandler_GenerateDraft_RetornaRascunhoNaoPublicado
// documenta o contrato da geracao automatica no HTTP.
func TestItemReviewHandler_GenerateDraft_RetornaRascunhoNaoPublicado(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h, mockRepo := newItemReviewHandlerWithMocks()
	mockRepo.On("Create", mock.AnythingOfType("*domain.Item")).Return(nil)

	r := gin.New()
	r.POST("/api/v1/admin/items/generate", h.GenerateDraft)

	body, _ := json.Marshal(map[string]interface{}{
		"channel":     "email",
		"isMalicious": true,
		"context":     "atualizacao de cadastro bancario",
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/items/generate", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var created domain.Item
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &created))
	assert.Equal(t, domain.StatusDraft, created.Status)
	assert.Nil(t, created.ReviewedBy)
}
