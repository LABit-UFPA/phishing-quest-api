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
)

// TestCategoryHandler_ListQuestionsByCategory_LeParametroCorreto e a
// regressao de integracao do bug da issue #13: a rota registra o
// parametro como ":id" (adapter/http/router/category_router.go), e o
// handler chegou a ler "category_id", o que sempre retornava 400.
func TestCategoryHandler_ListQuestionsByCategory_LeParametroCorreto(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockCategoryRepo := new(MockCategoryRepository)
	mockQuestionRepo := new(MockQuestionRepository)
	uc := usecase.NewCategoryUseCase(mockCategoryRepo, mockQuestionRepo)
	h := handler.NewCategoryHandler(uc)

	categoryID := uuid.New()
	mockQuestionRepo.On("GetByCategoryID", categoryID).Return([]*domain.Question{
		{Id: uuid.New(), CategoryId: categoryID, QuestionText: "Isso e phishing?"},
	}, nil)

	r := gin.New()
	r.GET("/api/v1/categories/:id/questions", h.ListQuestionsByCategory)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/categories/"+categoryID.String()+"/questions", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response dto.CategoryQuestionsDTO
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &response))
	assert.Equal(t, categoryID, response.CategoryId)
	assert.Len(t, response.Questions, 1)

	mockQuestionRepo.AssertExpectations(t)
}

func TestCategoryHandler_ListQuestionsByCategory_IDInvalido(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockCategoryRepo := new(MockCategoryRepository)
	mockQuestionRepo := new(MockQuestionRepository)
	uc := usecase.NewCategoryUseCase(mockCategoryRepo, mockQuestionRepo)
	h := handler.NewCategoryHandler(uc)

	r := gin.New()
	r.GET("/api/v1/categories/:id/questions", h.ListQuestionsByCategory)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/categories/not-a-uuid/questions", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockQuestionRepo.AssertNotCalled(t, "GetByCategoryID")
}
