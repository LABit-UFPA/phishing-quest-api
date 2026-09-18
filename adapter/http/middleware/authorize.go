package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// RequireRole exige que o usuario autenticado (via AuthRequired, que
// deve rodar antes deste middleware na cadeia) tenha uma das roles
// permitidas. Responde 403 caso contrario.
//
// Depende de ContextRoleKey ja estar setado no gin.Context — se
// RequireRole for usado sem AuthRequired antes, a role lida sera vazia
// e o acesso sera negado (fail-closed).
func RequireRole(allowedRoles ...string) gin.HandlerFunc {
	allowed := make(map[string]bool, len(allowedRoles))
	for _, role := range allowedRoles {
		allowed[role] = true
	}

	return func(c *gin.Context) {
		roleValue, exists := c.Get(ContextRoleKey)
		role, ok := roleValue.(string)
		if !exists || !ok || !allowed[role] {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "acesso negado: permissao insuficiente"})
			return
		}
		c.Next()
	}
}
