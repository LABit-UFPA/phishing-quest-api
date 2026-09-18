package handler

import (
	"errors"
	"net/http"
	"phishing-quest/adapter/http/response"
	"phishing-quest/core/usecase"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ItemSelectionHandler struct {
	selectionUseCase *usecase.ItemSelectionUseCase
}

func NewItemSelectionHandler(suc *usecase.ItemSelectionUseCase) *ItemSelectionHandler {
	return &ItemSelectionHandler{selectionUseCase: suc}
}

// NextItem retorna o proximo item da sessao do usuario autenticado
// (GET /api/v1/items/next?sessionId=&mode=). userId vem do JWT (nunca
// do query param, para o cliente nao poder pedir itens em nome de
// outro usuario); sessionId e obrigatorio; mode e opcional
// (default "balanced").
func (ish *ItemSelectionHandler) NextItem(c *gin.Context) {
	userIDValue, exists := c.Get("userId")
	userID, ok := userIDValue.(uuid.UUID)
	if !exists || !ok {
		c.JSON(http.StatusUnauthorized, response.Error("usuario nao autenticado"))
		return
	}

	sessionIDParam := c.Query("sessionId")
	sessionID, err := uuid.Parse(sessionIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error("sessionId invalido ou ausente"))
		return
	}

	mode := usecase.SelectionMode(c.DefaultQuery("mode", string(usecase.ModeBalanced)))

	item, err := ish.selectionUseCase.NextItem(userID, sessionID, mode)
	if err != nil {
		if errors.Is(err, usecase.ErrNoUnseenItems) {
			c.JSON(http.StatusNotFound, response.Error(err.Error()))
			return
		}
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}

	c.JSON(http.StatusOK, item)
}
