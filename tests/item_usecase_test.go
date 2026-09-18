package tests

import (
	"testing"

	"phishing-quest/core/usecase"
	"phishing-quest/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/datatypes"
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

func (m *MockItemRepository) GetByChannel(channel domain.Channel) ([]*domain.Item, error) {
	args := m.Called(channel)
	if args.Get(0) != nil {
		return args.Get(0).([]*domain.Item), args.Error(1)
	}
	return nil, args.Error(1)
}

func TestItemUseCase_CreateItem_DefinePadraoDeLocale(t *testing.T) {
	mockRepo := new(MockItemRepository)
	uc := usecase.NewItemUseCase(mockRepo)

	mockRepo.On("Create", mock.AnythingOfType("*domain.Item")).Return(nil)

	created, err := uc.CreateItem(&domain.Item{
		Channel:     domain.ChannelWhatsApp,
		IsMalicious: true,
		ContentJSON: datatypes.JSON(`{"messages":[]}`),
	})

	assert.NoError(t, err)
	assert.Equal(t, "pt-BR", created.Locale)
	mockRepo.AssertExpectations(t)
}

func TestItemUseCase_CreateItem_RejeitaCanalInvalido(t *testing.T) {
	mockRepo := new(MockItemRepository)
	uc := usecase.NewItemUseCase(mockRepo)

	created, err := uc.CreateItem(&domain.Item{
		Channel:     domain.Channel("carta_pombo"),
		ContentJSON: datatypes.JSON(`{}`),
	})

	assert.Nil(t, created)
	assert.Error(t, err)
	mockRepo.AssertNotCalled(t, "Create", mock.Anything)
}

func TestItemUseCase_ListItemsByChannel(t *testing.T) {
	mockRepo := new(MockItemRepository)
	uc := usecase.NewItemUseCase(mockRepo)

	expected := []*domain.Item{
		{Id: uuid.New(), Channel: domain.ChannelSMS, IsMalicious: true},
	}
	mockRepo.On("GetByChannel", domain.ChannelSMS).Return(expected, nil)

	items, err := uc.ListItemsByChannel(domain.ChannelSMS)

	assert.NoError(t, err)
	assert.Len(t, items, 1)
	assert.Equal(t, domain.ChannelSMS, items[0].Channel)
	mockRepo.AssertExpectations(t)
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
