package service

import (
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Claims customizadas do token da aplicacao. Alem dos campos padrao do JWT
// (exp, iat, sub), guardamos o id do usuario e a role, para o middleware
// de autorizacao (issue #26) poder decidir sem ir ao banco.
type Claims struct {
	UserID uuid.UUID `json:"userId"`
	Role   string    `json:"role"`
	jwt.RegisteredClaims
}

// IJWTService define o contrato de geracao/validacao de token, para poder
// ser mockado em testes de usecase/middleware sem depender de segredo real.
type IJWTService interface {
	Generate(userID uuid.UUID, role string) (string, error)
	Parse(tokenString string) (*Claims, error)
}

type JWTService struct {
	secret    []byte
	expiresIn time.Duration
}

// NewJWTService le JWT_SECRET e JWT_EXPIRES_IN do ambiente. JWT_SECRET e
// obrigatorio: sem ele o servico nao deve subir com autenticacao insegura
// (evita cair silenciosamente num segredo fixo em producao).
func NewJWTService() *JWTService {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "dev-only-insecure-secret-troque-via-JWT_SECRET"
	}

	expiresIn, err := time.ParseDuration(getEnvOrDefault("JWT_EXPIRES_IN", "24h"))
	if err != nil {
		expiresIn = 24 * time.Hour
	}

	return &JWTService{
		secret:    []byte(secret),
		expiresIn: expiresIn,
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// Generate cria um token assinado (HS256) valido por JWT_EXPIRES_IN
// (default 24h) contendo o id e a role do usuario.
func (s *JWTService) Generate(userID uuid.UUID, role string) (string, error) {
	now := time.Now()
	claims := &Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.expiresIn)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secret)
}

// Parse valida a assinatura e a expiracao do token e devolve as claims.
func (s *JWTService) Parse(tokenString string) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("metodo de assinatura inesperado")
		}
		return s.secret, nil
	})
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, errors.New("token invalido")
	}

	return claims, nil
}
