.PHONY: run build test test-verbose test-cover clean deps swagger docker-up docker-down docker-build docker-logs help

## —— 🚀 Aplicação ——————————————————————————————————————————————

run: ## Executa a aplicação localmente
	go run cmd/api/main.go

build: ## Compila o binário da aplicação
	CGO_ENABLED=0 GOOS=linux go build -o bin/server ./cmd/api

clean: ## Remove artefatos de build
	rm -rf bin/

deps: ## Instala e organiza dependências
	go mod tidy

## —— 🧪 Testes ——————————————————————————————————————————————————

test: ## Executa todos os testes
	go test ./...

test-verbose: ## Executa todos os testes com output detalhado
	go test ./... -v

test-cover: ## Executa testes com relatório de cobertura
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html
	@echo "📊 Relatório gerado em coverage.html"

## —— 📖 Documentação ————————————————————————————————————————————

swagger: ## Gera/atualiza a documentação Swagger
	swag init -g cmd/api/main.go --parseDependency --parseInternal
	@echo "✅ Documentação Swagger gerada em docs/"

## —— 🐳 Docker ——————————————————————————————————————————————————

docker-up: ## Sobe os containers (PostgreSQL + API)
	docker compose up --build -d

docker-down: ## Para e remove os containers
	docker compose down

docker-build: ## Builda a imagem Docker da aplicação
	docker compose build

docker-logs: ## Exibe os logs dos containers
	docker compose logs -f

## —— ❓ Ajuda ————————————————————————————————————————————————————

help: ## Exibe este menu de ajuda
	@echo "Comandos disponíveis:"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'
