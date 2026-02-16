// Package config gerencia as configurações da aplicação.
// Carrega variáveis de ambiente do arquivo .env e fornece
// uma struct tipada para acesso seguro às configurações.
package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

// Config armazena todas as configurações da aplicação.
type Config struct {
	ServerPort  string
	DBHost      string
	DBPort      string
	DBUser      string
	DBPassword  string
	DBName      string
	DBSSLMode   string
	CORSOrigins string
}

// Load carrega as configurações do arquivo .env e variáveis de ambiente.
// Variáveis de ambiente do sistema têm prioridade sobre o .env.
func Load() (*Config, error) {
	// Carrega .env se existir (ignora erro se não existir, ex: produção)
	_ = godotenv.Load()

	cfg := &Config{
		ServerPort:  getEnv("SERVER_PORT", "8080"),
		DBHost:      getEnv("DB_HOST", "localhost"),
		DBPort:      getEnv("DB_PORT", "5432"),
		DBUser:      getEnv("DB_USER", "postgres"),
		DBPassword:  getEnv("DB_PASSWORD", "postgres"),
		DBName:      getEnv("DB_NAME", "wpp_scheduler"),
		DBSSLMode:   getEnv("DB_SSLMODE", "disable"),
		CORSOrigins: getEnv("CORS_ORIGINS", "*"),
	}

	return cfg, nil
}

// DSN retorna a string de conexão formatada para o PostgreSQL.
func (c *Config) DSN() string {
	return fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=America/Sao_Paulo",
		c.DBHost, c.DBUser, c.DBPassword, c.DBName, c.DBPort, c.DBSSLMode,
	)
}

// getEnv retorna o valor da variável de ambiente ou o fallback fornecido.
func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
