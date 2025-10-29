// (Esta é uma implementação de exemplo, não use em produção sem testar)
# Estágio 1: Build da aplicação
FROM golang:1.22-alpine AS builder

# Instala git e ferramentas de build
RUN apk add --no-cache git build-base

WORKDIR /app

# Copia go.mod e go.sum e baixa as dependências
COPY go.mod go.sum ./
RUN go mod download

# Copia todo o código-fonte
COPY . .

# Instala o Swag para gerar a documentação
RUN go install github.com/swaggo/swag/cmd/swag@latest

# Gera a documentação do Swagger
# O -g aponta para o entrypoint principal
RUN swag init -g cmd/server/main.go

# Builda o executável final
# CGO_ENABLED=0 e -ldflags "-s -w" criam um binário estático e menor
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/main ./cmd/server/main.go

# Estágio 2: Imagem final
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copia o binário buildado do estágio anterior
COPY --from=builder /app/main .

# Copia os docs do swagger gerados
COPY --from=builder /app/docs/ ./docs/

# Expõe a porta que a aplicação usa
EXPOSE 8080

# Comando para rodar a aplicação
CMD ["./main"]
