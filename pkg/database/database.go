// Package database gerencia a conexão com o PostgreSQL via GORM.
// Também executa migrations automáticas dos modelos de domínio.
package database

import (
	"fmt"
	"log"

	"github.com/ropehapi/kaizen-wpp-scheduler-backend/internal/domain"
	"github.com/ropehapi/kaizen-wpp-scheduler-backend/pkg/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Connect estabelece conexão com o PostgreSQL e executa migrations.
// Retorna a instância do GORM pronta para uso.
func Connect(cfg *config.Config) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(cfg.DSN()), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("falha ao conectar com o banco de dados: %w", err)
	}

	// Executa migrations automáticas dos modelos de domínio.
	// GORM cria/atualiza tabelas baseado nas structs fornecidas.
	if err := db.AutoMigrate(&domain.Schedule{}, &domain.Contact{}); err != nil {
		return nil, fmt.Errorf("falha ao executar migrations: %w", err)
	}

	log.Println("✅ Conexão com banco de dados estabelecida e migrations executadas")
	return db, nil
}
