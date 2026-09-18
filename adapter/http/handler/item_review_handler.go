package handler

import (
	"errors"
	"net/http"
	"phishing-quest/adapter/http/response"
	"phishing-quest/core/service"
	"phishing-quest/core/usecase"
	"phishing-quest/domain"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ItemReviewHandler struct {
	reviewUseCase *usecase.ItemReviewUseCase
}

func NewItemReviewHandler(iruc *usecase.ItemReviewUseCase) *ItemReviewHandler {
	return &ItemReviewHandler{reviewUseCase: iruc}
}

// generateDraftRequest e declarado aqui (e nao no pacote dto) para
// nao criar dependencia dto -> domain/service no sentido inverso ao
// usado no resto do projeto, seguindo o mesmo caminho adotado na
// ingestao de telemetria (#29).
type generateDraftRequest struct {
	Channel     string `json:"channel" binding:"required"`
	IsMalicious bool   `json:"isMalicious"`
	Locale      string `json:"locale"`
	Context     string `json:"context"`
}

// CreateDraft cria um item em rascunho (POST /api/v1/admin/items).
// Protegido por role no router. O item nasce sempre como draft,
// independente do que o corpo enviar.
func (irh *ItemReviewHandler) CreateDraft(c *gin.Context) {
	var itemDTO *domain.Item
	if err := c.ShouldBindJSON(&itemDTO); err != nil {
		c.JSON(http.StatusBadRequest, response.Error(err.Error()))
		return
	}

	created, err := irh.reviewUseCase.CreateDraft(itemDTO)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error(err.Error()))
		return
	}

	c.JSON(http.StatusCreated, created)
}

// GenerateDraft gera um rascunho pelo servico de geracao configurado
// (POST /api/v1/admin/items/generate). O resultado entra como rascunho
// e ainda precisa de revisao humana para ser publicado.
func (irh *ItemReviewHandler) GenerateDraft(c *gin.Context) {
	var req generateDraftRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error(err.Error()))
		return
	}

	created, err := irh.reviewUseCase.GenerateDraft(service.DraftSpec{
		Channel:     domain.Channel(req.Channel),
		IsMalicious: req.IsMalicious,
		Locale:      req.Locale,
		Context:     req.Context,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error(err.Error()))
		return
	}

	c.JSON(http.StatusCreated, created)
}

// ListByStatus lista itens por estado do pipeline
// (GET /api/v1/admin/items?status=draft). Sem o parametro, lista os
// rascunhos — a fila de trabalho de quem revisa.
func (irh *ItemReviewHandler) ListByStatus(c *gin.Context) {
	status := domain.ItemStatus(c.DefaultQuery("status", string(domain.StatusDraft)))

	switch status {
	case domain.StatusDraft, domain.StatusReviewed, domain.StatusPublished:
	default:
		c.JSON(http.StatusBadRequest, response.Error("status invalido: use draft, reviewed ou published"))
		return
	}

	items, err := irh.reviewUseCase.ListByStatus(status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}

	c.JSON(http.StatusOK, items)
}

// GetItem devolve um item em qualquer estado, para inspecao pelo
// revisor (GET /api/v1/admin/items/:id).
func (irh *ItemReviewHandler) GetItem(c *gin.Context) {
	itemID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error(response.ErrInvalidIDFormat))
		return
	}

	item, err := irh.reviewUseCase.GetItem(itemID)
	if err != nil {
		c.JSON(http.StatusNotFound, response.Error("Item not found"))
		return
	}

	c.JSON(http.StatusOK, item)
}

// Review registra a revisao humana do item
// (POST /api/v1/admin/items/:id/review). O revisor e SEMPRE o usuario
// autenticado (do JWT), nunca um id enviado no corpo — do contrario a
// assinatura da revisao nao valeria nada.
func (irh *ItemReviewHandler) Review(c *gin.Context) {
	itemID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error(response.ErrInvalidIDFormat))
		return
	}

	reviewerIDValue, exists := c.Get("userId")
	reviewerID, ok := reviewerIDValue.(uuid.UUID)
	if !exists || !ok {
		c.JSON(http.StatusUnauthorized, response.Error("usuario nao autenticado"))
		return
	}

	item, err := irh.reviewUseCase.MarkReviewed(itemID, reviewerID)
	if err != nil {
		irh.respondTransitionError(c, err)
		return
	}

	c.JSON(http.StatusOK, item)
}

// Publish libera o item para o jogo
// (POST /api/v1/admin/items/:id/publish). Recusa item que nao passou
// pela revisao humana.
func (irh *ItemReviewHandler) Publish(c *gin.Context) {
	itemID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error(response.ErrInvalidIDFormat))
		return
	}

	item, err := irh.reviewUseCase.Publish(itemID)
	if err != nil {
		irh.respondTransitionError(c, err)
		return
	}

	c.JSON(http.StatusOK, item)
}

// respondTransitionError separa violacao de regra de fluxo (409, o
// cliente pediu algo que o pipeline nao permite) de item inexistente
// (404) e falha real (500).
func (irh *ItemReviewHandler) respondTransitionError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrInvalidItemTransition), errors.Is(err, domain.ErrReviewerRequired):
		c.JSON(http.StatusConflict, response.Error(err.Error()))
	case errors.Is(err, gorm.ErrRecordNotFound):
		c.JSON(http.StatusNotFound, response.Error("Item not found"))
	default:
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
	}
}
