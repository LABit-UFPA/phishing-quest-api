package tests

import (
	"testing"
	"time"

	"phishing-quest/core/usecase"
	"phishing-quest/domain"
	"phishing-quest/dto"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// MockUserRepository implementa repository.IUserRepository (IRepository[domain.User] + GetByEmail)
// para permitir testar UserUseCase sem depender de um banco real.
type MockUserRepository struct {
	mock.Mock
}

// Create espelha o comportamento do Repository[T] real: devolve a própria
// entidade recebida (mesmo ponteiro) quando não há erro configurado.
func (m *MockUserRepository) Create(user *domain.User) (*domain.User, error) {
	args := m.Called(user)
	if err := args.Error(0); err != nil {
		return nil, err
	}
	return user, nil
}

func (m *MockUserRepository) Update(user *domain.User) (*domain.User, error) {
	args := m.Called(user)
	if args.Get(0) != nil {
		return args.Get(0).(*domain.User), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockUserRepository) Delete(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockUserRepository) GetByID(id uuid.UUID) (*domain.User, error) {
	args := m.Called(id)
	if args.Get(0) != nil {
		return args.Get(0).(*domain.User), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockUserRepository) GetAll() ([]*domain.User, error) {
	args := m.Called()
	if args.Get(0) != nil {
		return args.Get(0).([]*domain.User), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockUserRepository) GetByEmail(email string) (*domain.User, error) {
	args := m.Called(email)
	if args.Get(0) != nil {
		return args.Get(0).(*domain.User), args.Error(1)
	}
	return nil, args.Error(1)
}

func TestUserUseCase_CreateUser(t *testing.T) {
	t.Run("user creation success", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		uc := usecase.NewUserUseCase(mockRepo)

		userRequest := &domain.User{
			Username: "testuser",
			Email:    "test@example.com",
			Password: "password123",
		}

		mockRepo.On("GetByEmail", userRequest.Email).Return(nil, gorm.ErrRecordNotFound)
		mockRepo.On("Create", mock.AnythingOfType("*domain.User")).Return(nil)

		user, err := uc.CreateUser(userRequest)

		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, userRequest.Username, user.Username)
		assert.Equal(t, userRequest.Email, user.Email)
		mockRepo.AssertExpectations(t)
	})

	t.Run("user creation fails when email already exists", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		uc := usecase.NewUserUseCase(mockRepo)

		existingUser := &domain.User{
			Id:        uuid.New(),
			Username:  "existinguser",
			Email:     "test@example.com",
			CreatedAt: time.Now(),
		}

		mockRepo.On("GetByEmail", existingUser.Email).Return(existingUser, nil)

		userRequest := &domain.User{
			Username: "newuser",
			Email:    "test@example.com",
			Password: "password123",
		}

		user, err := uc.CreateUser(userRequest)

		assert.Nil(t, user)
		assert.EqualError(t, err, "email já está em uso")
		mockRepo.AssertExpectations(t)
	})
}

func TestUserUseCase_Login(t *testing.T) {
	t.Run("login success", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		uc := usecase.NewUserUseCase(mockRepo)

		hash, err := uc.HashPassword("password123")
		assert.NoError(t, err)

		existingUser := &domain.User{
			Id:           uuid.New(),
			Username:     "testuser",
			Email:        "test@example.com",
			PasswordHash: hash,
			TotalScore:   10,
		}

		mockRepo.On("GetByEmail", existingUser.Email).Return(existingUser, nil)

		loginRequest := &dto.UserLoginDTO{
			Email:    existingUser.Email,
			Password: "password123",
		}

		response, err := uc.Login(loginRequest)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, existingUser.Id, response.Id)
		assert.Equal(t, existingUser.Username, response.Username)
		assert.Equal(t, existingUser.TotalScore, response.TotalScore)
		mockRepo.AssertExpectations(t)
	})

	t.Run("login fails with wrong password", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		uc := usecase.NewUserUseCase(mockRepo)

		hash, err := uc.HashPassword("password123")
		assert.NoError(t, err)

		existingUser := &domain.User{
			Id:           uuid.New(),
			Username:     "testuser",
			Email:        "test@example.com",
			PasswordHash: hash,
		}

		mockRepo.On("GetByEmail", existingUser.Email).Return(existingUser, nil)

		loginRequest := &dto.UserLoginDTO{
			Email:    existingUser.Email,
			Password: "senha-errada",
		}

		response, err := uc.Login(loginRequest)

		assert.Nil(t, response)
		assert.EqualError(t, err, "senha incorreta")
		mockRepo.AssertExpectations(t)
	})

	t.Run("login fails when user does not exist", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		uc := usecase.NewUserUseCase(mockRepo)

		mockRepo.On("GetByEmail", "missing@example.com").Return(nil, gorm.ErrRecordNotFound)

		loginRequest := &dto.UserLoginDTO{
			Email:    "missing@example.com",
			Password: "password123",
		}

		response, err := uc.Login(loginRequest)

		assert.Nil(t, response)
		assert.EqualError(t, err, "usuário não encontrado")
		mockRepo.AssertExpectations(t)
	})
}
