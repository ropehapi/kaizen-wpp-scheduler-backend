# WPP Scheduler - Backend

Backend em Go para gerenciamento de agendamentos de envio de mensagens em lote para WhatsApp.

> ⚠️ Este backend **NÃO envia mensagens**. Ele apenas armazena e gerencia agendamentos via API REST.

## 🏗️ Arquitetura

O projeto segue uma **Clean Architecture simplificada** com separação clara de responsabilidades:

```
├── cmd/api/              # Ponto de entrada da aplicação
├── internal/
│   ├── domain/           # Modelos de domínio e DTOs
│   ├── repository/       # Acesso ao banco de dados (GORM)
│   ├── service/          # Regras de negócio
│   └── handler/          # Controllers HTTP e rotas
├── pkg/
│   ├── config/           # Configuração da aplicação
│   ├── database/         # Conexão com PostgreSQL
│   ├── middleware/        # CORS e logging
│   └── response/         # Respostas padronizadas
├── Dockerfile
├── docker-compose.yml
└── .env
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

# 3. Instale dependências
go mod tidy

# 4. Execute
go run cmd/api/main.go
```

## 📡 Endpoints

| Método | Endpoint                       | Descrição              |
|--------|--------------------------------|------------------------|
| GET    | `/health`                      | Healthcheck            |
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
