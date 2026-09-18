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

// MockCueRepository implementa repository.ICueRepository.
type MockCueRepository struct {
	mock.Mock
}

func (m *MockCueRepository) Create(cue *domain.Cue) (*domain.Cue, error) {
	args := m.Called(cue)
	if err := args.Error(0); err != nil {
		return nil, err
	}
	return cue, nil
}

func (m *MockCueRepository) Update(cue *domain.Cue) (*domain.Cue, error) {
	args := m.Called(cue)
	if args.Get(0) != nil {
		return args.Get(0).(*domain.Cue), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockCueRepository) Delete(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockCueRepository) GetByID(id uuid.UUID) (*domain.Cue, error) {
	args := m.Called(id)
	if args.Get(0) != nil {
		return args.Get(0).(*domain.Cue), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockCueRepository) GetAll() ([]*domain.Cue, error) {
	args := m.Called()
	if args.Get(0) != nil {
		return args.Get(0).([]*domain.Cue), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockCueRepository) GetByCode(code string) (*domain.Cue, error) {
	args := m.Called(code)
	if args.Get(0) != nil {
		return args.Get(0).(*domain.Cue), args.Error(1)
	}
	return nil, args.Error(1)
}

// MockItemCueRepository implementa repository.IItemCueRepository.
type MockItemCueRepository struct {
	mock.Mock
}

func (m *MockItemCueRepository) Create(itemCue *domain.ItemCue) (*domain.ItemCue, error) {
	args := m.Called(itemCue)
	if err := args.Error(0); err != nil {
		return nil, err
	}
	return itemCue, nil
}

func (m *MockItemCueRepository) Update(itemCue *domain.ItemCue) (*domain.ItemCue, error) {
	args := m.Called(itemCue)
	if args.Get(0) != nil {
		return args.Get(0).(*domain.ItemCue), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockItemCueRepository) Delete(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockItemCueRepository) GetByID(id uuid.UUID) (*domain.ItemCue, error) {
	args := m.Called(id)
	if args.Get(0) != nil {
		return args.Get(0).(*domain.ItemCue), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockItemCueRepository) GetAll() ([]*domain.ItemCue, error) {
	args := m.Called()
	if args.Get(0) != nil {
		return args.Get(0).([]*domain.ItemCue), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockItemCueRepository) GetByItemID(itemID uuid.UUID) ([]*domain.ItemCue, error) {
	args := m.Called(itemID)
	if args.Get(0) != nil {
		return args.Get(0).([]*domain.ItemCue), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockItemCueRepository) GetByCueID(cueID uuid.UUID) ([]*domain.ItemCue, error) {
	args := m.Called(cueID)
	if args.Get(0) != nil {
		return args.Get(0).([]*domain.ItemCue), args.Error(1)
	}
	return nil, args.Error(1)
}

func TestCueUseCase_ListCues(t *testing.T) {
	mockCueRepo := new(MockCueRepository)
	mockItemCueRepo := new(MockItemCueRepository)
	mockItemRepo := new(MockItemRepository)
	uc := usecase.NewCueUseCase(mockCueRepo, mockItemCueRepo, mockItemRepo)

	expected := []*domain.Cue{
		{Id: uuid.New(), Code: "urgency", LabelPt: "Urgência", Category: "psychological"},
	}
	mockCueRepo.On("GetAll").Return(expected, nil)

	cues, err := uc.ListCues()

	assert.NoError(t, err)
	assert.Len(t, cues, 1)
	mockCueRepo.AssertExpectations(t)
}

// TestCueUseCase_AssociateItemCue_ValidaExistenciaDoItem e a regressao
// que garante que a associacao nunca e criada para um item inexistente
// — a FK do banco tambem protege isso, mas o usecase devolve um erro
// mais claro antes de tentar o INSERT.
func TestCueUseCase_AssociateItemCue_ValidaExistenciaDoItem(t *testing.T) {
	mockCueRepo := new(MockCueRepository)
	mockItemCueRepo := new(MockItemCueRepository)
	mockItemRepo := new(MockItemRepository)
	uc := usecase.NewCueUseCase(mockCueRepo, mockItemCueRepo, mockItemRepo)

	itemID := uuid.New()
	cueID := uuid.New()
	mockItemRepo.On("GetByID", itemID).Return(nil, gorm.ErrRecordNotFound)

	created, err := uc.AssociateItemCue(&domain.ItemCue{ItemId: itemID, CueId: cueID})

	assert.Nil(t, created)
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
	mockCueRepo.AssertNotCalled(t, "GetByID", mock.Anything)
	mockItemCueRepo.AssertNotCalled(t, "Create", mock.Anything)
}

func TestCueUseCase_AssociateItemCue_ValidaExistenciaDaCue(t *testing.T) {
	mockCueRepo := new(MockCueRepository)
	mockItemCueRepo := new(MockItemCueRepository)
	mockItemRepo := new(MockItemRepository)
	uc := usecase.NewCueUseCase(mockCueRepo, mockItemCueRepo, mockItemRepo)

	itemID := uuid.New()
	cueID := uuid.New()
	mockItemRepo.On("GetByID", itemID).Return(&domain.Item{Id: itemID}, nil)
	mockCueRepo.On("GetByID", cueID).Return(nil, gorm.ErrRecordNotFound)

	created, err := uc.AssociateItemCue(&domain.ItemCue{ItemId: itemID, CueId: cueID})

	assert.Nil(t, created)
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
	mockItemCueRepo.AssertNotCalled(t, "Create", mock.Anything)
}

func TestCueUseCase_AssociateItemCue_Sucesso(t *testing.T) {
	mockCueRepo := new(MockCueRepository)
	mockItemCueRepo := new(MockItemCueRepository)
	mockItemRepo := new(MockItemRepository)
	uc := usecase.NewCueUseCase(mockCueRepo, mockItemCueRepo, mockItemRepo)

	itemID := uuid.New()
	cueID := uuid.New()
	spanStart := 10
	spanEnd := 20

	mockItemRepo.On("GetByID", itemID).Return(&domain.Item{Id: itemID}, nil)
	mockCueRepo.On("GetByID", cueID).Return(&domain.Cue{Id: cueID}, nil)
	mockItemCueRepo.On("Create", mock.AnythingOfType("*domain.ItemCue")).Return(nil)

	created, err := uc.AssociateItemCue(&domain.ItemCue{
		ItemId:    itemID,
		CueId:     cueID,
		SpanStart: &spanStart,
		SpanEnd:   &spanEnd,
	})

	assert.NoError(t, err)
	assert.Equal(t, itemID, created.ItemId)
	assert.Equal(t, cueID, created.CueId)
	assert.NotEqual(t, uuid.Nil, created.Id)
	mockItemCueRepo.AssertExpectations(t)
}

func TestCueUseCase_GetCuesByItemID(t *testing.T) {
	mockCueRepo := new(MockCueRepository)
	mockItemCueRepo := new(MockItemCueRepository)
	mockItemRepo := new(MockItemRepository)
	uc := usecase.NewCueUseCase(mockCueRepo, mockItemCueRepo, mockItemRepo)

	itemID := uuid.New()
	expected := []*domain.ItemCue{{Id: uuid.New(), ItemId: itemID, CueId: uuid.New()}}
	mockItemCueRepo.On("GetByItemID", itemID).Return(expected, nil)

	itemCues, err := uc.GetCuesByItemID(itemID)

	assert.NoError(t, err)
	assert.Len(t, itemCues, 1)
	mockItemCueRepo.AssertExpectations(t)
}
