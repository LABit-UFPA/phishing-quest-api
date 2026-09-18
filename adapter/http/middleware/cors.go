package middleware

import (
	"os"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// CORS configura o middleware de CORS a partir da variavel de ambiente
// CORS_ALLOWED_ORIGINS (lista separada por virgula). Sem essa variavel,
// libera "*" — adequado para desenvolvimento local, mas deploys que
// servem um front web (ou o app Flutter web) devem restringir via env.
func CORS() gin.HandlerFunc {
	config := cors.DefaultConfig()
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Type", "Authorization", "ngrok-skip-browser-warning"}
	config.MaxAge = 12 * time.Hour

	if origins := os.Getenv("CORS_ALLOWED_ORIGINS"); origins != "" {
		config.AllowOrigins = strings.Split(origins, ",")
	} else {
		config.AllowAllOrigins = true
	}

	return cors.New(config)
}
