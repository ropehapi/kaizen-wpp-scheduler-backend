// Package main é o ponto de entrada da aplicação.
// Responsável por inicializar todas as dependências e iniciar o servidor HTTP.
package main

import (
	"fmt"
	"log"

	_ "github.com/ropehapi/kaizen-wpp-scheduler-backend/docs"
	"github.com/ropehapi/kaizen-wpp-scheduler-backend/internal/handler"
	"github.com/ropehapi/kaizen-wpp-scheduler-backend/internal/repository"
	"github.com/ropehapi/kaizen-wpp-scheduler-backend/internal/service"
	"github.com/ropehapi/kaizen-wpp-scheduler-backend/pkg/config"
	"github.com/ropehapi/kaizen-wpp-scheduler-backend/pkg/database"
)

// @title WPP Scheduler API
// @version 1.0
// @description API para gerenciamento de agendamentos de envio de mensagens em lote para WhatsApp.
// @description Este backend apenas armazena e gerencia agendamentos via API REST — não envia mensagens.

// @contact.name Suporte WPP Scheduler
// @contact.url https://github.com/ropehapi/kaizen-wpp-scheduler-backend

// @host localhost:8080
// @BasePath /api/v1
// @schemes http

func main() {
	// Carrega configurações do .env e variáveis de ambiente
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("❌ Erro ao carregar configurações: %v", err)
	}

	// Conecta ao banco de dados e executa migrations
	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatalf("❌ Erro ao conectar ao banco de dados: %v", err)
	}

	// Inicializa as camadas seguindo o padrão de injeção de dependência:
	// Repository → Service → Handler
	scheduleRepo := repository.NewScheduleRepository(db)
	scheduleService := service.NewScheduleService(scheduleRepo)
	scheduleHandler := handler.NewScheduleHandler(scheduleService)

	// Configura o router com todas as rotas e middlewares
	router := handler.SetupRouter(scheduleHandler, cfg.CORSOrigins)

	// Inicia o servidor HTTP
	addr := fmt.Sprintf(":%s", cfg.ServerPort)
	log.Printf("🚀 Servidor iniciado na porta %s", cfg.ServerPort)
	if err := router.Run(addr); err != nil {
		log.Fatalf("❌ Erro ao iniciar servidor: %v", err)
	}
}
