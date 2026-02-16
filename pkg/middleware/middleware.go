// Package middleware fornece middlewares HTTP para a aplicação.
// Inclui CORS configurável e logging de requisições.
package middleware

import (
	"log"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// CORSMiddleware configura CORS para permitir requisições do frontend React.
// As origens permitidas são configuráveis via variável de ambiente.
func CORSMiddleware(allowedOrigins string) gin.HandlerFunc {
	origins := strings.Split(allowedOrigins, ",")

	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// Verifica se a origem está na lista permitida
		for _, allowed := range origins {
			allowed = strings.TrimSpace(allowed)
			if allowed == "*" || allowed == origin {
				c.Header("Access-Control-Allow-Origin", origin)
				break
			}
		}

		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Max-Age", "86400")

		// Responde preflight requests imediatamente
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// LoggingMiddleware registra informações sobre cada requisição HTTP.
// Inclui método, path, status code e tempo de resposta.
func LoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Processa a requisição
		c.Next()

		// Registra após processamento
		duration := time.Since(start)
		log.Printf(
			"[%s] %s %s | %d | %v",
			c.Request.Method,
			c.Request.URL.Path,
			c.Request.URL.RawQuery,
			c.Writer.Status(),
			duration,
		)
	}
}
