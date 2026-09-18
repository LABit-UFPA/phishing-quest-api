package tests

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"phishing-quest/adapter/http/middleware"
	"phishing-quest/core/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func setupAuthTestRouter(t *testing.T) (*gin.Engine, service.IJWTService) {
	gin.SetMode(gin.TestMode)
	t.Setenv("JWT_SECRET", "segredo-de-teste")

	jwtService := service.NewJWTService()
	r := gin.New()
	r.GET("/protegido", middleware.AuthRequired(jwtService), func(c *gin.Context) {
		userID, _ := c.Get(middleware.ContextUserIDKey)
		c.JSON(http.StatusOK, gin.H{"userId": userID})
	})

	return r, jwtService
}

func TestAuthRequired_SemHeader(t *testing.T) {
	r, _ := setupAuthTestRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/protegido", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthRequired_HeaderSemBearer(t *testing.T) {
	r, jwtService := setupAuthTestRouter(t)
	token, _ := jwtService.Generate(uuid.New(), "player")

	req := httptest.NewRequest(http.MethodGet, "/protegido", nil)
	req.Header.Set("Authorization", token) // falta o prefixo "Bearer "
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthRequired_TokenInvalido(t *testing.T) {
	r, _ := setupAuthTestRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/protegido", nil)
	req.Header.Set("Authorization", "Bearer token-invalido")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthRequired_TokenValido(t *testing.T) {
	r, jwtService := setupAuthTestRouter(t)
	token, err := jwtService.Generate(uuid.New(), "player")
	assert.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/protegido", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}
