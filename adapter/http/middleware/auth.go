package middleware

import (
	"net/http"
	"strings"

	"phishing-quest/core/service"

	"github.com/gin-gonic/gin"
)

// Chaves usadas para expor dados do usuario autenticado no gin.Context.
// Handlers/usecases downstream leem via c.Get(ContextUserIDKey) etc.
const (
	ContextUserIDKey = "userId"
	ContextRoleKey   = "role"
)

// AuthRequired valida o header "Authorization: Bearer <token>" usando o
// IJWTService informado. Em caso de token ausente/invalido/expirado,
// responde 401 e interrompe a cadeia de handlers.
func AuthRequired(jwtService service.IJWTService) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "token de autenticacao ausente"})
			return
		}

		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "formato do header Authorization invalido, esperado 'Bearer <token>'"})
			return
		}

		claims, err := jwtService.Parse(parts[1])
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "token invalido ou expirado"})
			return
		}

		c.Set(ContextUserIDKey, claims.UserID)
		c.Set(ContextRoleKey, claims.Role)
		c.Next()
	}
}
