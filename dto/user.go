package dto

import "github.com/google/uuid"

type UserLoginDTO struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UserLoginResponseDTO struct {
	Token      string    `json:"token"`
	Id         uuid.UUID `json:"id"`
	Username   string    `json:"username"`
	Email      string    `json:"email"`
	TotalScore int       `json:"totalScore"`
}

// UserResponseDTO é o formato seguro de resposta para operações que expõem
// um usuário (ex.: cadastro). Nunca inclui senha em texto puro nem hash.
type UserResponseDTO struct {
	Id         uuid.UUID `json:"id"`
	Username   string    `json:"username"`
	Email      string    `json:"email"`
	TotalScore int       `json:"totalScore"`
}
