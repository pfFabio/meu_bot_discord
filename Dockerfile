# Build stage
FROM golang:1.23-alpine AS builder
WORKDIR /app

# Cachear dependências separadamente do código-fonte
COPY go.mod go.sum ./
RUN go mod download

# Compilar binário estático otimizado
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /bot ./cmd/bot

# Runtime stage — imagem mínima sem compilador nem ferramentas
FROM alpine:3.20
RUN apk add --no-cache ca-certificates
COPY --from=builder /bot /bot

ENTRYPOINT ["/bot"]
