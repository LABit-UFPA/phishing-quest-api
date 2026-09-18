package handler

import (
	"errors"
	"net/http"
	"phishing-quest/adapter/http/response"
	"phishing-quest/core/usecase"
	"phishing-quest/domain"
	"phishing-quest/dto"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AttemptHandler struct {
	attemptUseCase *usecase.AttemptUseCase
}

func NewAttemptHandler(auc *usecase.AttemptUseCase) *AttemptHandler {
	return &AttemptHandler{attemptUseCase: auc}
}

// CreateAttempt registra uma tentativa completa (POST /api/v1/attempts).
// Recusa usuarios sem consentimento registrado com 403.
func (ah *AttemptHandler) CreateAttempt(c *gin.Context) {
	var attemptDTO *domain.Attempt
	if err := c.ShouldBindJSON(&attemptDTO); err != nil {
		c.JSON(http.StatusBadRequest, response.Error(err.Error()))
		return
	}

	createdAttempt, err := ah.attemptUseCase.RegisterAttempt(attemptDTO)
	if err != nil {
		if errors.Is(err, usecase.ErrConsentRequired) {
			c.JSON(http.StatusForbidden, response.Error(err.Error()))
			return
		}
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}

	c.JSON(http.StatusCreated, createdAttempt)
}

// ListAttemptsByUser retorna o historico de tentativas de um usuario.
func (ah *AttemptHandler) ListAttemptsByUser(c *gin.Context) {
	idParam := c.Param("id")
	userID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error(response.ErrInvalidIDFormat))
		return
	}

	attempts, err := ah.attemptUseCase.ListAttemptsByUser(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}

	c.JSON(http.StatusOK, attempts)
}

// RegisterConsent registra o consentimento do usuario para participar
// da coleta de dados de pesquisa (POST /api/v1/auth/consent).
// Idempotente: consentir de novo com o mesmo userId retorna sucesso
// sem erro, em vez de expor a violacao de PK ao cliente.
func (ah *AttemptHandler) RegisterConsent(c *gin.Context) {
	var consentDTO dto.ConsentRequestDTO
	if err := c.ShouldBindJSON(&consentDTO); err != nil {
		c.JSON(http.StatusBadRequest, response.Error(err.Error()))
		return
	}

	participant, alreadyConsented, err := ah.attemptUseCase.RegisterConsent(consentDTO.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}

	if alreadyConsented {
		c.JSON(http.StatusOK, gin.H{"userId": participant.UserId, "alreadyConsented": true})
		return
	}

	c.JSON(http.StatusCreated, participant)
}
