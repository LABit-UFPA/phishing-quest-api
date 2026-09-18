package dto

import "github.com/google/uuid"

// UserRegisterDTO e o formato de ENTRADA do cadastro. A senha em texto
// puro existe so aqui, no limite da API, e nunca entra na entidade de
// dominio (issue #60) — o usecase converte direto para hash.
type UserRegisterDTO struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

type UserLoginDTO struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UserLoginResponseDTO struct {
	Token      string    `json:"token"`
	Id         uuid.UUID `json:"id"`
	Username   string    `json:"username"`
	Email      string    `json:"email"`
	Role       string    `json:"role"`
	TotalScore int       `json:"totalScore"`
}

// UserResponseDTO é o formato seguro de resposta para operações que expõem
// um usuário (ex.: cadastro). Nunca inclui senha em texto puro nem hash.
type UserResponseDTO struct {
	Id         uuid.UUID `json:"id"`
	Username   string    `json:"username"`
	Email      string    `json:"email"`
	Role       string    `json:"role"`
	TotalScore int       `json:"totalScore"`
}
