# Kaizen WPP Scheduler - Backend

## Visão geral

API REST em Go para gerenciamento de agendamentos de envio de mensagens em lote para WhatsApp. Este backend **NÃO envia mensagens** — apenas armazena e gerencia agendamentos via API REST. O envio real é responsabilidade do serviço **messaging-officer**.

## Ecossistema

Este projeto faz parte do ecossistema **"manda-pra-mim"** de automação WhatsApp:
- **kaizen-wpp-scheduler-frontend** → Frontend React que consome esta API (portas 3000/5173)
- **messaging-officer** → API REST que conecta ao WhatsApp via Baileys (porta 3000)
- **kaizen-secretary** → Worker de cronjobs que automatiza rotinas

## Stack

- **Go 1.24.2**
- **Gin** → Framework HTTP
- **GORM** → ORM com PostgreSQL
- **PostgreSQL 16** → Banco de dados
- **UUID** → Identificadores primários (google/uuid)
- **Testify** → Framework de testes e mocks
- **Swaggo** → Geração automática de documentação OpenAPI/Swagger

## Arquitetura

Clean Architecture simplificada com injeção de dependência manual:

```
cmd/api/main.go              # Entry point — inicializa DI e inicia servidor
internal/
  domain/models.go           # Entidades, DTOs, enums e request models
  repository/                # Acesso ao banco (GORM) — interface + implementação
  service/                   # Regras de negócio — interface + implementação
  handler/                   # Controllers HTTP (Gin) + router + testes
    router.go                # Configuração de rotas e middlewares
    schedule_handler.go      # Handlers CRUD de agendamentos
    schedule_handler_test.go # Testes de integração dos handlers
pkg/
  config/config.go           # Carregamento de configuração (.env + env vars)
  database/database.go       # Conexão PostgreSQL + AutoMigrate
  middleware/middleware.go    # CORS configurável + logging de requisições
  response/response.go       # Respostas JSON padronizadas (data/error)
docs/                        # Swagger gerado automaticamente (não editar manualmente)
```

### Fluxo de dependência

```
main.go → Config → Database → Repository → Service → Handler → Router
```

A injeção de dependência é feita manualmente no `main.go`:
```
scheduleRepo := repository.NewScheduleRepository(db)
scheduleService := service.NewScheduleService(scheduleRepo)
scheduleHandler := handler.NewScheduleHandler(scheduleService)
```

## Modelos de domínio

### Schedule (Agendamento)
| Campo | Tipo | Descrição |
|---|---|---|
| `id` | UUID | PK gerada automaticamente via hook BeforeCreate |
| `message` | text | Conteúdo da mensagem |
| `type` | varchar(20) | `once` ou `recurring` |
| `frequency` | varchar(20) | `daily`, `weekly`, `monthly` (nullable, obrigatório se recurring) |
| `scheduledAt` | timestamp | Data/hora programada para envio |
| `status` | varchar(20) | `scheduled` (default), `sent`, `canceled` |
| `contacts` | relação | Lista de Contact (FK, cascade delete) |

### Contact (Contato)
| Campo | Tipo | Descrição |
|---|---|---|
| `id` | UUID | PK |
| `scheduleId` | UUID | FK para Schedule |
| `name` | varchar(255) | Nome do contato |
| `phone` | varchar(20) | Telefone (formato: `5511999999999`) |

## Endpoints da API

Base path: `/api/v1`

| Método | Rota | Handler | Descrição |
|---|---|---|---|
| GET | `/health` | inline | Healthcheck |
| GET | `/swagger/*any` | ginSwagger | Documentação Swagger |
| POST | `/api/v1/schedules` | CreateSchedule | Criar agendamento |
| GET | `/api/v1/schedules` | ListSchedules | Listar (paginado + filtro por status) |
| GET | `/api/v1/schedules/:id` | GetScheduleByID | Buscar por UUID |
| PUT | `/api/v1/schedules/:id` | UpdateSchedule | Atualizar agendamento |
| PATCH | `/api/v1/schedules/:id/cancel` | CancelSchedule | Cancelar agendamento |

## Regras de negócio importantes

- **Agendamento recorrente** (`type=recurring`) **exige** campo `frequency`
- **Frequências válidas**: `daily`, `weekly`, `monthly`
- Agendamentos com status `sent` **não podem** ser atualizados nem cancelados
- Agendamentos com status `canceled` **não podem** ser cancelados novamente
- **Status inicial** é sempre `scheduled`
- Na **atualização de contatos**, os contatos antigos são deletados e os novos inseridos (replace total)
- **Paginação padrão**: page=1, limit=10, máximo=100

## Padrão de respostas HTTP

Todas as respostas seguem o formato padronizado em `pkg/response/`:

**Sucesso:**
```json
{ "data": {...}, "error": null }
```

**Sucesso paginado:**
```json
{ "data": [...], "error": null, "pagination": { "page": 1, "limit": 10, "total": 50, "totalPages": 5 } }
```

**Erro:**
```json
{ "data": null, "error": "mensagem de erro" }
```

## Erros de negócio (service layer)

| Erro | HTTP Status | Cenário |
|---|---|---|
| `ErrScheduleNotFound` | 404 | UUID não existe no banco |
| `ErrScheduleAlreadySent` | 409 | Tentativa de alterar/cancelar agendamento já enviado |
| `ErrFrequencyRequired` | 400 | Tipo recurring sem frequency |
| `ErrInvalidFrequency` | 400 | Frequency diferente de daily/weekly/monthly |
| `ErrAlreadyCanceled` | 409 | Tentativa de cancelar agendamento já cancelado |

## Variáveis de ambiente

| Variável | Default | Descrição |
|---|---|---|
| `SERVER_PORT` | `8080` | Porta do servidor HTTP |
| `DB_HOST` | `localhost` | Host do PostgreSQL |
| `DB_PORT` | `5432` | Porta do PostgreSQL |
| `DB_USER` | `postgres` | Usuário do banco |
| `DB_PASSWORD` | `postgres` | Senha do banco |
| `DB_NAME` | `wpp_scheduler` | Nome do banco |
| `DB_SSLMODE` | `disable` | Modo SSL do PostgreSQL |
| `CORS_ORIGINS` | `*` | Origens permitidas (separadas por vírgula) |

## Testes

- Testes unitários em `internal/service/schedule_service_test.go` (mock do repository)
- Testes de integração em `internal/handler/schedule_handler_test.go` (mock do service)
- Padrão: mocks manuais com `testify/mock`
- Framework: `testify/assert` para asserções
- Executar: `go test ./...` ou `make test`

## Comandos úteis

```bash
make run            # Executa localmente
make test           # Roda testes
make test-cover     # Testes + cobertura
make swagger        # Regenera docs Swagger
make docker-up      # Docker Compose (PostgreSQL + API)
make docker-down    # Para containers
```

## Convenções de código

- Comentários em português nos arquivos de código
- Nomes de packages/funções/variáveis em inglês
- Erros de negócio definidos como `var` no package `service`
- Interfaces definidas no mesmo package da implementação
- DTOs e entidades no package `domain`
- Anotações Swagger nos handlers (godoc comments com `@`)
- Migrations automáticas via GORM AutoMigrate (sem arquivos de migration)
- TimeZone configurado como `America/Sao_Paulo` na DSN
