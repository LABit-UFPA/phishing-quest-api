package handler

import (
	"net/http"
	"phishing-quest/adapter/http/response"
	"phishing-quest/core/usecase"
	"phishing-quest/domain"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type CueHandler struct {
	cueUseCase *usecase.CueUseCase
}

func NewCueHandler(cuc *usecase.CueUseCase) *CueHandler {
	return &CueHandler{cueUseCase: cuc}
}

// ListCues retorna a taxonomia completa de pistas.
func (ch *CueHandler) ListCues(c *gin.Context) {
	cues, err := ch.cueUseCase.ListCues()
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}

	c.JSON(http.StatusOK, cues)
}

func (ch *CueHandler) CreateCue(c *gin.Context) {
	var cueDTO *domain.Cue
	if err := c.ShouldBindJSON(&cueDTO); err != nil {
		c.JSON(http.StatusBadRequest, response.Error(err.Error()))
		return
	}

	createdCue, err := ch.cueUseCase.CreateCue(cueDTO)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}

	c.JSON(http.StatusOK, createdCue)
}

// AssociateItemCue liga um item a uma pista presente nele.
func (ch *CueHandler) AssociateItemCue(c *gin.Context) {
	var itemCueDTO *domain.ItemCue
	if err := c.ShouldBindJSON(&itemCueDTO); err != nil {
		c.JSON(http.StatusBadRequest, response.Error(err.Error()))
		return
	}

	created, err := ch.cueUseCase.AssociateItemCue(itemCueDTO)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}

	c.JSON(http.StatusOK, created)
}

// GetCuesByItem retorna as pistas associadas a um item especifico.
func (ch *CueHandler) GetCuesByItem(c *gin.Context) {
	idParam := c.Param("id")
	itemID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error(response.ErrInvalidIDFormat))
		return
	}

	itemCues, err := ch.cueUseCase.GetCuesByItemID(itemID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}

	c.JSON(http.StatusOK, itemCues)
}
