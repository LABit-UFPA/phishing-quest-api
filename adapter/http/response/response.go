// Package response centraliza a construcao de respostas de erro da API.
//
// O formato ja usado em todos os handlers e {"error": "<mensagem>"}.
// Este pacote nao muda esse formato — apenas evita repetir
// c.JSON(status, gin.H{"error": ...}) em cada handler e da um lugar
// unico para evoluir o formato no futuro (ex.: incluir um codigo de
// erro estavel, sem precisar tocar em cada handler).
package response

import "github.com/gin-gonic/gin"

// Error monta o corpo padrao de erro: {"error": "<mensagem>"}.
func Error(message string) gin.H {
	return gin.H{"error": message}
}

// Mensagens reutilizadas por múltiplos handlers para o mesmo tipo de
// falha, evitando strings literais divergentes (ex.: "Invalid ID
// format" vs "Invalid ID" apareciam ambas antes desta mudanca).
const (
	ErrInvalidIDFormat = "Invalid ID format"
)
