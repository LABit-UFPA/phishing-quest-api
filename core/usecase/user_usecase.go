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

type UserUseCase struct {
	userRepo   repository.IUserRepository
	jwtService service.IJWTService
}

func NewUserUseCase(userRepo repository.IUserRepository, jwtService service.IJWTService) *UserUseCase {
	return &UserUseCase{userRepo: userRepo, jwtService: jwtService}
}

// CreateUser recebe o DTO de cadastro (nao a entidade) porque a senha
// em texto puro nao deve passar pelo dominio: ela e consumida aqui para
// gerar o hash e nao e copiada para o domain.User (issue #60).
func (uc *UserUseCase) CreateUser(userRequest *dto.UserRegisterDTO) (*domain.User, error) {
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
		PasswordHash: hashedPassword,
		Role:         domain.RoleParticipant,
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

	token, err := uc.jwtService.Generate(user.Id, string(user.Role))
	if err != nil {
		return nil, errors.New("erro ao gerar token de autenticacao")
	}

	userResponse := &dto.UserLoginResponseDTO{
		Token:      token,
		Id:         user.Id,
		Username:   user.Username,
		Email:      user.Email,
		Role:       string(user.Role),
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
