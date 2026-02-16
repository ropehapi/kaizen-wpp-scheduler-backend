// Package handler - router.go configura todas as rotas da aplicação.
// Centraliza o registro de endpoints e aplica middlewares.
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ropehapi/kaizen-wpp-scheduler-backend/pkg/middleware"
	"github.com/ropehapi/kaizen-wpp-scheduler-backend/pkg/response"
)

// SetupRouter configura o router do Gin com todas as rotas e middlewares.
// Recebe as dependências (handlers) por injeção.
func SetupRouter(scheduleHandler *ScheduleHandler, corsOrigins string) *gin.Engine {
	router := gin.New()

	// Middlewares globais
	router.Use(gin.Recovery())               // Recupera de panics
	router.Use(middleware.LoggingMiddleware()) // Log de requisições
	router.Use(middleware.CORSMiddleware(corsOrigins)) // CORS configurável

	// Healthcheck endpoint
	router.GET("/health", func(c *gin.Context) {
		response.Success(c, http.StatusOK, gin.H{
			"status": "ok",
		})
	})

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		schedules := v1.Group("/schedules")
		{
			schedules.POST("", scheduleHandler.CreateSchedule)
			schedules.GET("", scheduleHandler.ListSchedules)
			schedules.GET("/:id", scheduleHandler.GetScheduleByID)
			schedules.PUT("/:id", scheduleHandler.UpdateSchedule)
			schedules.PATCH("/:id/cancel", scheduleHandler.CancelSchedule)
		}
	}

	return router
}
