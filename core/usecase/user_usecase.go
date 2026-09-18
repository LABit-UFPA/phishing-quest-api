package usecase

import (
	"errors"
	"phishing-quest/adapter/repository"
	"phishing-quest/core/service"
	"phishing-quest/domain"
	"phishing-quest/dto"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// defaultUserRole e usada ate a issue #26 introduzir uma coluna de role
// persistida em users. Todo usuario autenticado recebe essa role no token.
const defaultUserRole = "player"

type UserUseCase struct {
	userRepo   repository.IUserRepository
	jwtService service.IJWTService
}

func NewUserUseCase(userRepo repository.IUserRepository, jwtService service.IJWTService) *UserUseCase {
	return &UserUseCase{userRepo: userRepo, jwtService: jwtService}
}

func (uc *UserUseCase) CreateUser(userRequest *domain.User) (*domain.User, error) {
	existingUser, err := uc.userRepo.GetByEmail(userRequest.Email)
	if err != nil && !errors.Is(gorm.ErrRecordNotFound, err) {
		return nil, err
	}
	if existingUser != nil {
		return nil, errors.New("email já está em uso")
	}

	hashedPassword, err := uc.HashPassword(userRequest.Password)
	if err != nil {
		return nil, errors.New("erro ao gerar hash da senha")
	}

	user := &domain.User{
		Id:           uuid.New(),
		Username:     userRequest.Username,
		Email:        userRequest.Email,
		Password:     userRequest.Password,
		PasswordHash: hashedPassword,
		TotalScore:   0,
		CreatedAt:    time.Now(),
	}

	err = user.Validate()
	if err != nil {
		return nil, err
	}

	createdUser, err := uc.userRepo.Create(user)
	if err != nil {
		return nil, err
	}

	return createdUser, nil
}

func (uc *UserUseCase) Login(userRequest *dto.UserLoginDTO) (*dto.UserLoginResponseDTO, error) {
	user, err := uc.userRepo.GetByEmail(userRequest.Email)
	if err != nil {
		return nil, errors.New("usuário não encontrado")
	}

	if !uc.CheckPasswordHash(userRequest.Password, user.PasswordHash) {
		return nil, errors.New("senha incorreta")
	}

	token, err := uc.jwtService.Generate(user.Id, defaultUserRole)
	if err != nil {
		return nil, errors.New("erro ao gerar token de autenticacao")
	}

	userResponse := &dto.UserLoginResponseDTO{
		Token:      token,
		Id:         user.Id,
		Username:   user.Username,
		Email:      user.Email,
		TotalScore: user.TotalScore,
	}

	return userResponse, nil
}

func (uc *UserUseCase) GetUser(id uuid.UUID) (*domain.User, error) {
	return uc.userRepo.GetByID(id)
}

func (uc *UserUseCase) UpdatePassword(user *domain.User, newPasswordHash string) {
	user.PasswordHash = newPasswordHash
	user.UpdatedAt = time.Now()
}

func (uc *UserUseCase) AddScore(user *domain.User, score int) {
	user.TotalScore += score
	user.UpdatedAt = time.Now()
}

func (uc *UserUseCase) HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func (uc *UserUseCase) CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
