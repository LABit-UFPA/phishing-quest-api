package usecase

import (
	"errors"
	"phishing-quest/adapter/repository"
	"phishing-quest/domain"

	"github.com/google/uuid"
)

// ErrItemNotPublished e retornado quando se tenta acessar por uma rota
// publica um item que ainda nao passou pela revisao humana.
var ErrItemNotPublished = errors.New("item nao publicado")

// ItemUseCase atende as rotas PUBLICAS de leitura de itens. Todas
// filtram por status publicado: o pipeline de curadoria (issue #30)
// vive no ItemReviewUseCase, e rascunho nao vaza para participante.
type ItemUseCase struct {
	itemRepo repository.IItemRepository
}

func NewItemUseCase(itemRepo repository.IItemRepository) *ItemUseCase {
	return &ItemUseCase{itemRepo: itemRepo}
}

// GetItem devolve o item somente se ele estiver publicado. Sem esse
// filtro, um rascunho ainda em revisao poderia ser lido por qualquer
// um que descobrisse o id.
func (iuc *ItemUseCase) GetItem(id uuid.UUID) (*domain.Item, error) {
	item, err := iuc.itemRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if item.Status != domain.StatusPublished {
		return nil, ErrItemNotPublished
	}
	return item, nil
}

func (iuc *ItemUseCase) ListItems() ([]*domain.Item, error) {
	return iuc.itemRepo.GetByStatus(domain.StatusPublished)
}

func (iuc *ItemUseCase) ListItemsByChannel(channel domain.Channel) ([]*domain.Item, error) {
	return iuc.itemRepo.GetPublishedByChannel(channel)
}
