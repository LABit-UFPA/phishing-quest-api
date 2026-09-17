package handler

import (
	"net/http"
	"phishing-quest/core/usecase"
	"phishing-quest/domain"
	"phishing-quest/dto"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	UserUseCase *usecase.UserUseCase
}

func NewUserHandler(uuc *usecase.UserUseCase) *UserHandler {
	return &UserHandler{UserUseCase: uuc}
}

func (uh *UserHandler) CreateUser(c *gin.Context) {
	var userDTO *domain.User
	if err := c.ShouldBindJSON(&userDTO); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := uh.UserUseCase.CreateUser(userDTO)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, user.ToDTO())
}

func (uh *UserHandler) Login(c *gin.Context) {
	var userLoginDTO *dto.UserLoginDTO
	if err := c.ShouldBindJSON(&userLoginDTO); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := uh.UserUseCase.Login(userLoginDTO)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, user)
}

// GetUser lida com a obtenção de um usuário pelo ID
func (uh *UserHandler) GetUser(c *gin.Context) {
	id := c.Param("id")

	c.JSON(http.StatusOK, id)
}

func (uh *UserHandler) GetTeste(c *gin.Context) {
	c.JSON(http.StatusOK, "olhaaaaaa")
}
