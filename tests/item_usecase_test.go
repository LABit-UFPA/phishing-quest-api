package tests

import (
	"testing"

	"phishing-quest/core/usecase"
	"phishing-quest/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// MockItemRepository implementa repository.IItemRepository.
type MockItemRepository struct {
	mock.Mock
}

func (m *MockItemRepository) Create(item *domain.Item) (*domain.Item, error) {
	args := m.Called(item)
	if err := args.Error(0); err != nil {
		return nil, err
	}
	return item, nil
}

func (m *MockItemRepository) Update(item *domain.Item) (*domain.Item, error) {
	args := m.Called(item)
	if args.Get(0) != nil {
		return args.Get(0).(*domain.Item), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockItemRepository) Delete(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockItemRepository) GetByID(id uuid.UUID) (*domain.Item, error) {
	args := m.Called(id)
	if args.Get(0) != nil {
		return args.Get(0).(*domain.Item), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockItemRepository) GetAll() ([]*domain.Item, error) {
	args := m.Called()
	if args.Get(0) != nil {
		return args.Get(0).([]*domain.Item), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockItemRepository) GetByStatus(status domain.ItemStatus) ([]*domain.Item, error) {
	args := m.Called(status)
	if args.Get(0) != nil {
		return args.Get(0).([]*domain.Item), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockItemRepository) GetPublishedByChannel(channel domain.Channel) ([]*domain.Item, error) {
	args := m.Called(channel)
	if args.Get(0) != nil {
		return args.Get(0).([]*domain.Item), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockItemRepository) GetByChannel(channel domain.Channel) ([]*domain.Item, error) {
	args := m.Called(channel)
	if args.Get(0) != nil {
		return args.Get(0).([]*domain.Item), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockItemRepository) CountSeenInSession(userID, sessionID uuid.UUID) (int64, int64, error) {
	args := m.Called(userID, sessionID)
	return args.Get(0).(int64), args.Get(1).(int64), args.Error(2)
}

func (m *MockItemRepository) GetRandomUnseenByCues(userID, sessionID uuid.UUID, cueIDs []uuid.UUID, isMalicious *bool) (*domain.Item, error) {
	args := m.Called(userID, sessionID, cueIDs, isMalicious)
	if args.Get(0) != nil {
		return args.Get(0).(*domain.Item), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockItemRepository) GetRandomUnseen(userID, sessionID uuid.UUID, isMalicious *bool) (*domain.Item, error) {
	args := m.Called(userID, sessionID, isMalicious)
	if args.Get(0) != nil {
		return args.Get(0).(*domain.Item), args.Error(1)
	}
	return nil, args.Error(1)
}

// TestItemUseCase_ListItemsByChannel_SoPublicados garante que a rota
// publica de listagem por canal nunca devolve rascunho (issue #30).
func TestItemUseCase_ListItemsByChannel_SoPublicados(t *testing.T) {
	mockRepo := new(MockItemRepository)
	uc := usecase.NewItemUseCase(mockRepo)

	expected := []*domain.Item{
		{Id: uuid.New(), Channel: domain.ChannelSMS, IsMalicious: true, Status: domain.StatusPublished},
	}
	mockRepo.On("GetPublishedByChannel", domain.ChannelSMS).Return(expected, nil)

	items, err := uc.ListItemsByChannel(domain.ChannelSMS)

	assert.NoError(t, err)
	assert.Len(t, items, 1)
	assert.Equal(t, domain.ChannelSMS, items[0].Channel)
	mockRepo.AssertExpectations(t)
	// A versao sem filtro de status nao pode ser usada pela rota publica.
	mockRepo.AssertNotCalled(t, "GetByChannel", mock.Anything)
}

// TestItemUseCase_ListItems_SoPublicados e o mesmo contrato para a
// listagem sem filtro de canal.
func TestItemUseCase_ListItems_SoPublicados(t *testing.T) {
	mockRepo := new(MockItemRepository)
	uc := usecase.NewItemUseCase(mockRepo)

	mockRepo.On("GetByStatus", domain.StatusPublished).Return([]*domain.Item{
		{Id: uuid.New(), Status: domain.StatusPublished},
	}, nil)

	items, err := uc.ListItems()

	assert.NoError(t, err)
	assert.Len(t, items, 1)
	mockRepo.AssertExpectations(t)
	mockRepo.AssertNotCalled(t, "GetAll")
}

// TestItemUseCase_GetItem_RecusaRascunho e a regressao central do lado
// publico da issue #30: conhecer o id de um rascunho nao pode dar
// acesso ao conteudo antes da revisao humana.
func TestItemUseCase_GetItem_RecusaRascunho(t *testing.T) {
	mockRepo := new(MockItemRepository)
	uc := usecase.NewItemUseCase(mockRepo)

	id := uuid.New()
	mockRepo.On("GetByID", id).Return(&domain.Item{Id: id, Status: domain.StatusDraft}, nil)

	item, err := uc.GetItem(id)

	assert.Nil(t, item)
	assert.ErrorIs(t, err, usecase.ErrItemNotPublished)
}

// TestItemUseCase_GetItem_RecusaApenasRevisado garante que estar
// revisado nao basta: enquanto nao for publicado, nao e servido.
func TestItemUseCase_GetItem_RecusaApenasRevisado(t *testing.T) {
	mockRepo := new(MockItemRepository)
	uc := usecase.NewItemUseCase(mockRepo)

	id := uuid.New()
	reviewer := uuid.New()
	mockRepo.On("GetByID", id).Return(&domain.Item{Id: id, Status: domain.StatusReviewed, ReviewedBy: &reviewer}, nil)

	item, err := uc.GetItem(id)

	assert.Nil(t, item)
	assert.ErrorIs(t, err, usecase.ErrItemNotPublished)
}

func TestItemUseCase_GetItem_DevolvePublicado(t *testing.T) {
	mockRepo := new(MockItemRepository)
	uc := usecase.NewItemUseCase(mockRepo)

	id := uuid.New()
	mockRepo.On("GetByID", id).Return(&domain.Item{Id: id, Status: domain.StatusPublished}, nil)

	item, err := uc.GetItem(id)

	assert.NoError(t, err)
	assert.Equal(t, id, item.Id)
}

func TestItemUseCase_GetItem_PropagaErroDoRepo(t *testing.T) {
	mockRepo := new(MockItemRepository)
	uc := usecase.NewItemUseCase(mockRepo)

	id := uuid.New()
	mockRepo.On("GetByID", id).Return(nil, gorm.ErrRecordNotFound)

	item, err := uc.GetItem(id)

	assert.Nil(t, item)
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
}
