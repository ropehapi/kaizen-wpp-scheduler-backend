# Build stage - compila o binário Go
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copia arquivos de dependência primeiro (melhor cache de Docker layers)
COPY go.mod go.sum ./
RUN go mod download

# Copia o restante do código e compila
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/server ./cmd/api

# Run stage - imagem mínima para produção
FROM alpine:3.19

WORKDIR /app

# Instala certificados CA para conexões HTTPS
RUN apk --no-cache add ca-certificates tzdata

# Copia o binário compilado
COPY --from=builder /app/server .

# Copia o .env (pode ser sobrescrito via volume/env vars)
COPY .env .

EXPOSE 8080

CMD ["./server"]
