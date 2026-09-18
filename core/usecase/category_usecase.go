package usecase

import (
	"errors"
	"phishing-quest/adapter/repository"
	"phishing-quest/domain"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CategoryUseCase struct {
	categoryRepo repository.ICategoryRepository
	questionRepo repository.IQuestionRepository
}

func NewCategoryUseCase(categoryRepo repository.ICategoryRepository, questionRepo repository.IQuestionRepository) *CategoryUseCase {
	return &CategoryUseCase{categoryRepo: categoryRepo, questionRepo: questionRepo}
}

func (cuc *CategoryUseCase) CreateCategory(categoryRequest *domain.Category) (*domain.Category, error) {
	existingUser, err := cuc.categoryRepo.GetByCategoryName(categoryRequest.CategoryName)
	if err != nil && !errors.Is(gorm.ErrRecordNotFound, err) {
		return nil, err
	}
	if existingUser != nil {
		return nil, errors.New("categoria já existe")
	}

	category := &domain.Category{
		Id:           uuid.New(),
		CategoryName: categoryRequest.CategoryName,
	}

	err = category.Validate()
	if err != nil {
		return nil, err
	}

	createdCategory, err := cuc.categoryRepo.Create(category)
	if err != nil {
		return nil, err
	}

	return createdCategory, nil
}

func (cuc *CategoryUseCase) ListCategories() ([]*domain.Category, error) {
	categories, err := cuc.categoryRepo.GetAll()
	if err != nil {
		return nil, err
	}

	return categories, err
}

func (cuc *CategoryUseCase) GetQuestionsByCategoryID(categoryID uuid.UUID) ([]*domain.Question, error) {
	questions, err := cuc.questionRepo.GetByCategoryID(categoryID)
	if err != nil {
		return nil, err
	}
	return questions, nil
}
