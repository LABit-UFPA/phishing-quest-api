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

// MockCategoryRepository implementa repository.ICategoryRepository.
type MockCategoryRepository struct {
	mock.Mock
}

func (m *MockCategoryRepository) Create(category *domain.Category) (*domain.Category, error) {
	args := m.Called(category)
	if err := args.Error(0); err != nil {
		return nil, err
	}
	return category, nil
}

func (m *MockCategoryRepository) Update(category *domain.Category) (*domain.Category, error) {
	args := m.Called(category)
	if args.Get(0) != nil {
		return args.Get(0).(*domain.Category), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockCategoryRepository) Delete(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockCategoryRepository) GetByID(id uuid.UUID) (*domain.Category, error) {
	args := m.Called(id)
	if args.Get(0) != nil {
		return args.Get(0).(*domain.Category), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockCategoryRepository) GetAll() ([]*domain.Category, error) {
	args := m.Called()
	if args.Get(0) != nil {
		return args.Get(0).([]*domain.Category), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockCategoryRepository) GetByCategoryName(categoryName string) (*domain.Category, error) {
	args := m.Called(categoryName)
	if args.Get(0) != nil {
		return args.Get(0).(*domain.Category), args.Error(1)
	}
	return nil, args.Error(1)
}

// MockQuestionRepository implementa repository.IQuestionRepository.
type MockQuestionRepository struct {
	mock.Mock
}

func (m *MockQuestionRepository) Create(question *domain.Question) (*domain.Question, error) {
	args := m.Called(question)
	if err := args.Error(0); err != nil {
		return nil, err
	}
	return question, nil
}

func (m *MockQuestionRepository) Update(question *domain.Question) (*domain.Question, error) {
	args := m.Called(question)
	if args.Get(0) != nil {
		return args.Get(0).(*domain.Question), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockQuestionRepository) Delete(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockQuestionRepository) GetByID(id uuid.UUID) (*domain.Question, error) {
	args := m.Called(id)
	if args.Get(0) != nil {
		return args.Get(0).(*domain.Question), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockQuestionRepository) GetAll() ([]*domain.Question, error) {
	args := m.Called()
	if args.Get(0) != nil {
		return args.Get(0).([]*domain.Question), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockQuestionRepository) GetByCategoryID(categoryID uuid.UUID) ([]*domain.Question, error) {
	args := m.Called(categoryID)
	if args.Get(0) != nil {
		return args.Get(0).([]*domain.Question), args.Error(1)
	}
	return nil, args.Error(1)
}

// TestCategoryUseCase_GetQuestionsByCategoryID_NaoDeveSerNilPointer e a
// regressao do bug da issue #13: NewCategoryUseCase criava o usecase sem
// injetar questionRepo, entao GetQuestionsByCategoryID sempre causava
// nil pointer dereference. Este teste falha (panic) se questionRepo nao
// estiver de fato setado no construtor.
func TestCategoryUseCase_GetQuestionsByCategoryID_NaoDeveSerNilPointer(t *testing.T) {
	mockCategoryRepo := new(MockCategoryRepository)
	mockQuestionRepo := new(MockQuestionRepository)
	uc := usecase.NewCategoryUseCase(mockCategoryRepo, mockQuestionRepo)

	categoryID := uuid.New()
	expectedQuestions := []*domain.Question{
		{Id: uuid.New(), CategoryId: categoryID, QuestionText: "Isso e phishing?"},
	}

	mockQuestionRepo.On("GetByCategoryID", categoryID).Return(expectedQuestions, nil)

	questions, err := uc.GetQuestionsByCategoryID(categoryID)

	assert.NoError(t, err)
	assert.Len(t, questions, 1)
	assert.Equal(t, expectedQuestions[0].QuestionText, questions[0].QuestionText)
	mockQuestionRepo.AssertExpectations(t)
}

func TestCategoryUseCase_GetQuestionsByCategoryID_PropagaErroDoRepo(t *testing.T) {
	mockCategoryRepo := new(MockCategoryRepository)
	mockQuestionRepo := new(MockQuestionRepository)
	uc := usecase.NewCategoryUseCase(mockCategoryRepo, mockQuestionRepo)

	categoryID := uuid.New()
	mockQuestionRepo.On("GetByCategoryID", categoryID).Return(nil, gorm.ErrRecordNotFound)

	questions, err := uc.GetQuestionsByCategoryID(categoryID)

	assert.Nil(t, questions)
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
	mockQuestionRepo.AssertExpectations(t)
}

func TestCategoryUseCase_CreateCategory_FalhaQuandoJaExiste(t *testing.T) {
	mockCategoryRepo := new(MockCategoryRepository)
	mockQuestionRepo := new(MockQuestionRepository)
	uc := usecase.NewCategoryUseCase(mockCategoryRepo, mockQuestionRepo)

	existing := &domain.Category{Id: uuid.New(), CategoryName: "Phishing Corporativo"}
	mockCategoryRepo.On("GetByCategoryName", existing.CategoryName).Return(existing, nil)

	created, err := uc.CreateCategory(&domain.Category{CategoryName: existing.CategoryName})

	assert.Nil(t, created)
	assert.EqualError(t, err, "categoria já existe")
	mockCategoryRepo.AssertExpectations(t)
}
