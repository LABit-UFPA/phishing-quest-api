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
)

// TestGameHandler_SubmitAnswer_AceitaECamelCase e a regressao de
// integracao da issue #15: POST /api/v1/game/answer era o unico
// endpoint da API em snake_case (user_id, question_id, answer_id,
// is_correct), divergindo do resto (categoryName, questionText,
// isCorrect...). Este teste envia e valida o corpo em camelCase.
func TestGameHandler_SubmitAnswer_AceitaECamelCase(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockAnswerRepo := new(MockAnswerRepository)
	mockUserRepo := new(MockUserRepository)
	mockUserScoreRepo := new(MockUserScoreRepository)
	mockUserAnswerRepo := new(MockUserAnswerRepository)

	uc := usecase.NewGameUseCase(mockAnswerRepo, mockUserRepo, mockUserScoreRepo, mockUserAnswerRepo)
	h := handler.NewGameHandler(uc)

	userID := uuid.New()
	questionID := uuid.New()
	answerID := uuid.New()

	answer := &domain.Answer{Id: answerID, QuestionId: questionID, IsCorrect: true}
	mockAnswerRepo.On("GetByID", answerID).Return(answer, nil)
	mockUserAnswerRepo.On("Create", mock.AnythingOfType("*domain.UserAnswer")).Return(nil)
	mockUserScoreRepo.On("IncrementScore", userID, 10).Return(nil)
	mockUserScoreRepo.On("GetUserScore", userID).Return(&domain.UserScore{UserId: userID, Score: 40}, nil)

	r := gin.New()
	r.POST("/api/v1/game/answer", h.SubmitAnswer)

	// Corpo da requisicao em camelCase — se o DTO ainda esperasse
	// snake_case (user_id, question_id, answer_id), o bind falharia
	// silenciosamente (campos ficariam com uuid.Nil) e o teste
	// quebraria no mock de GetByID (chamado com uuid.Nil, sem
	// expectativa configurada).
	body := map[string]string{
		"userId":     userID.String(),
		"questionId": questionID.String(),
		"answerId":   answerID.String(),
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/game/answer", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// A resposta tambem deve estar em camelCase.
	var response map[string]interface{}
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &response))
	assert.Contains(t, response, "isCorrect")
	assert.Contains(t, response, "totalScore")
	assert.NotContains(t, response, "is_correct")
	assert.NotContains(t, response, "total_score")
	assert.Equal(t, true, response["isCorrect"])
	assert.Equal(t, float64(40), response["totalScore"])

	mockAnswerRepo.AssertExpectations(t)
	mockUserScoreRepo.AssertExpectations(t)
}

func TestGameHandler_SubmitAnswer_RejeitaCorpoSnakeCase(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockAnswerRepo := new(MockAnswerRepository)
	mockUserRepo := new(MockUserRepository)
	mockUserScoreRepo := new(MockUserScoreRepository)
	mockUserAnswerRepo := new(MockUserAnswerRepository)

	uc := usecase.NewGameUseCase(mockAnswerRepo, mockUserRepo, mockUserScoreRepo, mockUserAnswerRepo)
	h := handler.NewGameHandler(uc)

	r := gin.New()
	r.POST("/api/v1/game/answer", h.SubmitAnswer)

	// Corpo no formato antigo (snake_case) deve falhar o binding, ja
	// que os campos tem "binding:required" e nao seriam preenchidos.
	body := map[string]string{
		"user_id":     uuid.New().String(),
		"question_id": uuid.New().String(),
		"answer_id":   uuid.New().String(),
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/game/answer", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockAnswerRepo.AssertNotCalled(t, "GetByID", mock.Anything)
}
