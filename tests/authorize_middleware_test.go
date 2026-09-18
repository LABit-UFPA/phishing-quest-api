package tests

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"phishing-quest/adapter/http/middleware"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupRequireRoleRouter(role string, setRole bool) *gin.Engine {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.GET("/protegido",
		func(c *gin.Context) {
			// Simula o AuthRequired ja tendo rodado e setado a role.
			if setRole {
				c.Set(middleware.ContextRoleKey, role)
			}
			c.Next()
		},
		middleware.RequireRole("researcher", "admin"),
		func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"ok": true})
		},
	)
	return r
}

func TestRequireRole_PermiteRolePermitida(t *testing.T) {
	r := setupRequireRoleRouter("researcher", true)

	req := httptest.NewRequest(http.MethodGet, "/protegido", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRequireRole_BloqueiaRoleNaoPermitida(t *testing.T) {
	r := setupRequireRoleRouter("participant", true)

	req := httptest.NewRequest(http.MethodGet, "/protegido", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

// TestRequireRole_FailClosedSemRoleNoContexto garante que a ausencia
// da role no contexto (ex.: AuthRequired nao rodou antes) bloqueia o
// acesso em vez de liberar por omissao.
func TestRequireRole_FailClosedSemRoleNoContexto(t *testing.T) {
	r := setupRequireRoleRouter("", false)

	req := httptest.NewRequest(http.MethodGet, "/protegido", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestRequireRole_PermiteRoleAdmin(t *testing.T) {
	r := setupRequireRoleRouter("admin", true)

	req := httptest.NewRequest(http.MethodGet, "/protegido", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}
