package handler

import (
	"net/http"
	"phishing-quest/adapter/http/response"
	"phishing-quest/core/usecase"
	"phishing-quest/domain"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ItemHandler struct {
	itemUseCase *usecase.ItemUseCase
}

func NewItemHandler(iuc *usecase.ItemUseCase) *ItemHandler {
	return &ItemHandler{itemUseCase: iuc}
}

func (ih *ItemHandler) CreateItem(c *gin.Context) {
	var itemDTO *domain.Item
	if err := c.ShouldBindJSON(&itemDTO); err != nil {
		c.JSON(http.StatusBadRequest, response.Error(err.Error()))
		return
	}

	createdItem, err := ih.itemUseCase.CreateItem(itemDTO)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}

	c.JSON(http.StatusOK, createdItem)
}

func (ih *ItemHandler) GetItem(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error(response.ErrInvalidIDFormat))
		return
	}

	item, err := ih.itemUseCase.GetItem(id)
	if err != nil {
		c.JSON(http.StatusNotFound, response.Error("Item not found"))
		return
	}

	c.JSON(http.StatusOK, item)
}

// ListItems lista todos os itens, opcionalmente filtrados por canal
// via query param (?channel=email). A selecao balanceada/aleatoria
// para o fluxo de jogo fica a cargo de GET /items/next.
func (ih *ItemHandler) ListItems(c *gin.Context) {
	channelParam := c.Query("channel")
	if channelParam != "" {
		items, err := ih.itemUseCase.ListItemsByChannel(domain.Channel(channelParam))
		if err != nil {
			c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
			return
		}
		c.JSON(http.StatusOK, items)
		return
	}

	items, err := ih.itemUseCase.ListItems()
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}

	c.JSON(http.StatusOK, items)
}
