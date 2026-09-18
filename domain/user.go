package domain

import (
	"phishing-quest/dto"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

// Role enumera os niveis de acesso do usuario. participant e o
// default para qualquer cadastro novo; researcher e admin sao
// atribuidos manualmente (nao ha auto-promocao via API).
type Role string

const (
	RoleParticipant Role = "participant"
	RoleResearcher  Role = "researcher"
	RoleAdmin       Role = "admin"
)

type User struct {
	Id           uuid.UUID `json:"id" gorm:"primaryKey"`
	Username     string    `json:"username" validate:"required,min=1,max=255"`
	Email        string    `json:"email" validate:"required,email,max=255"`
	Password     string    `json:"password" validate:"required" gorm:"-"`
	PasswordHash string    `json:"-" validate:"required,min=1,max=255"`
	Role         Role      `json:"role,omitempty" gorm:"default:participant"`
	TotalScore   int       `json:"totalScore" validate:"gte=0"`
	CreatedAt    time.Time `json:"-"`
	UpdatedAt    time.Time `json:"-"`
}

func (u *User) TableName() string {
	return "phishing_quest.users"
}

func (u *User) Validate() error {
	validate := validator.New()
	return validate.Struct(u)
}

// ToDTO converte para o formato seguro de resposta, sem senha em texto
// puro nem hash. Deve ser usado em qualquer endpoint que devolve um User.
func (u *User) ToDTO() *dto.UserResponseDTO {
	return &dto.UserResponseDTO{
		Id:         u.Id,
		Username:   u.Username,
		Email:      u.Email,
		Role:       string(u.Role),
		TotalScore: u.TotalScore,
	}
}
