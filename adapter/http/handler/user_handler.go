package handler

import (
	"net/http"
	"phishing-quest/adapter/http/response"
	"phishing-quest/core/usecase"
	"phishing-quest/domain"
	"phishing-quest/dto"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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
		c.JSON(http.StatusBadRequest, response.Error(err.Error()))
		return
	}

	user, err := uh.UserUseCase.CreateUser(userDTO)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}

	c.JSON(http.StatusOK, user.ToDTO())
}

func (uh *UserHandler) Login(c *gin.Context) {
	var userLoginDTO *dto.UserLoginDTO
	if err := c.ShouldBindJSON(&userLoginDTO); err != nil {
		c.JSON(http.StatusBadRequest, response.Error(err.Error()))
		return
	}

	user, err := uh.UserUseCase.Login(userLoginDTO)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}

	c.JSON(http.StatusOK, user)
}

// GetUser busca um usuario pelo ID e responde com o formato seguro de
// resposta (ToDTO), sem senha em texto puro nem hash.
func (uh *UserHandler) GetUser(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error(response.ErrInvalidIDFormat))
		return
	}

	user, err := uh.UserUseCase.GetUser(id)
	if err != nil {
		c.JSON(http.StatusNotFound, response.Error("User not found"))
		return
	}

	c.JSON(http.StatusOK, user.ToDTO())
}
