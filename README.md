# WPP Scheduler - Backend

Backend em Go para gerenciamento de agendamentos de envio de mensagens em lote para WhatsApp.

> ⚠️ Este backend **NÃO envia mensagens**. Ele apenas armazena e gerencia agendamentos via API REST.

## 🏗️ Arquitetura

O projeto segue uma **Clean Architecture simplificada** com separação clara de responsabilidades:

```
cmd/api/              # Ponto de entrada da aplicação
docs/                 # Documentação OpenAPI gerada (Swagger)
internal/
  domain/             # Modelos de domínio e DTOs
  repository/         # Acesso ao banco de dados (GORM)
  service/            # Regras de negócio
  handler/            # Controllers HTTP e rotas
pkg/
  config/             # Configuração da aplicação
  database/           # Conexão com PostgreSQL
  middleware/          # CORS e logging
  response/           # Respostas padronizadas
Dockerfile
docker-compose.yml
.env
```

## 🚀 Como Rodar

### Pré-requisitos

- Docker e Docker Compose

### Com Docker Compose (recomendado)

```bash
# Sobe PostgreSQL + API
docker compose up --build

# A API estará disponível em http://localhost:8080
```

### Sem Docker (desenvolvimento local)

```bash
# 1. Tenha Go 1.22+ e PostgreSQL rodando

# 2. Copie e configure o .env
cp .env.example .env

# 3. Instale dependências e execute
make deps
make run
```

### Makefile

O projeto inclui um `Makefile` com os principais comandos. Execute `make help` para ver todos:

```bash
make run            # Executa a aplicação localmente
make build          # Compila o binário
make test           # Executa todos os testes
make test-verbose   # Testes com output detalhado
make test-cover     # Testes com relatório de cobertura
make swagger        # Gera/atualiza documentação Swagger
make docker-up      # Sobe containers (PostgreSQL + API)
make docker-down    # Para os containers
make docker-logs    # Exibe logs dos containers
make deps           # Instala dependências
make clean          # Remove artefatos de build
```

## 📖 Documentação da API (Swagger)

A documentação interativa da API está disponível via Swagger UI. Com a aplicação rodando, acesse:

```
http://localhost:8080/swagger/index.html
```

A especificação OpenAPI é gerada automaticamente a partir das anotações no código-fonte usando [swaggo/swag](https://github.com/swaggo/swag).

### Regenerar a documentação

Após alterar as anotações nos handlers, regenere os arquivos de documentação:

```bash
# Instale o swag CLI (se ainda não tiver)
go install github.com/swaggo/swag/cmd/swag@latest

# Gere a documentação
swag init -g cmd/api/main.go --parseDependency --parseInternal
```

## 📡 Endpoints

| Método | Endpoint                       | Descrição              |
|--------|--------------------------------|------------------------|
| GET    | `/health`                      | Healthcheck            |
| GET    | `/swagger/*`                   | Documentação Swagger   |
| POST   | `/api/v1/schedules`            | Criar agendamento      |
| GET    | `/api/v1/schedules`            | Listar agendamentos    |
| GET    | `/api/v1/schedules/:id`        | Buscar por ID          |
| PUT    | `/api/v1/schedules/:id`        | Atualizar agendamento  |
| PATCH  | `/api/v1/schedules/:id/cancel` | Cancelar agendamento   |

### Exemplo: Criar agendamento

```bash
curl -X POST http://localhost:8080/api/v1/schedules \
  -H "Content-Type: application/json" \
  -d '{
    "message": "Olá! Lembrete da reunião.",
    "type": "once",
    "scheduledAt": "2026-02-20T15:00:00Z",
    "contacts": [
      {"name": "João", "phone": "+5511999999999"},
      {"name": "Maria", "phone": "+5511888888888"}
    ]
  }'
```

### Listar com filtros

```bash
# Filtrar por status
curl http://localhost:8080/api/v1/schedules?status=scheduled

# Com paginação
curl http://localhost:8080/api/v1/schedules?page=1&limit=20
```

## 🧪 Testes

```bash
go test ./... -v
```

## 📦 Stack

- **Go 1.24**
- **Gin** - Framework HTTP
- **GORM** - ORM
- **PostgreSQL** - Banco de dados
- **UUID** - Identificadores primários
- **Testify** - Testes e mocks
- **Swaggo** - Documentação OpenAPI/Swagger