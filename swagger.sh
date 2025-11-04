#!/bin/bash

# Script para facilitar o acesso à documentação Swagger
# Uso: ./swagger.sh [comando]

SWAGGER_URL="http://localhost:8080/swagger/index.html"
API_URL="http://localhost:8080"

show_help() {
    echo "Uso: $0 [comando]"
    echo ""
    echo "Comandos disponíveis:"
    echo "  open      - Abre a documentação Swagger no navegador"
    echo "  status    - Verifica se a API está rodando"
    echo "  url       - Mostra a URL do Swagger"
    echo "  help      - Mostra esta ajuda"
    echo ""
    echo "Se nenhum comando for especificado, tentará abrir o Swagger no navegador."
}

check_api_status() {
    echo "Verificando se a API está rodando..."
    if curl -s --max-time 5 "$API_URL" > /dev/null 2>&1; then
        echo "✅ API está rodando em $API_URL"
        return 0
    else
        echo "❌ API não está rodando ou não está acessível em $API_URL"
        echo "💡 Execute 'make run' ou 'go run cmd/server/main.go' para iniciar a API"
        return 1
    fi
}

open_swagger() {
    echo "Abrindo documentação Swagger..."
    echo "URL: $SWAGGER_URL"
    
    if ! check_api_status; then
        echo ""
        echo "⚠️  A API precisa estar rodando para acessar o Swagger."
        echo "Execute um dos comandos abaixo para iniciar a API:"
        echo "  make run"
        echo "  go run cmd/server/main.go"
        echo "  make run-with-swagger  (inicia a API e abre o Swagger automaticamente)"
        return 1
    fi
    
    echo ""
    echo "🚀 Abrindo Swagger no navegador..."
    
    # Detecta o sistema operacional e abre o navegador adequado
    if command -v xdg-open > /dev/null 2>&1; then
        # Linux
        xdg-open "$SWAGGER_URL"
    elif command -v open > /dev/null 2>&1; then
        # macOS
        open "$SWAGGER_URL"
    elif command -v start > /dev/null 2>&1; then
        # Windows
        start "$SWAGGER_URL"
    else
        echo "❌ Não foi possível detectar como abrir o navegador automaticamente."
        echo "📋 Copie e cole esta URL no seu navegador:"
        echo "   $SWAGGER_URL"
    fi
}

show_url() {
    echo "URLs da API:"
    echo "  API Base: $API_URL"
    echo "  Swagger:  $SWAGGER_URL"
}

# Comando principal
case "${1:-open}" in
    open)
        open_swagger
        ;;
    status)
        check_api_status
        ;;
    url)
        show_url
        ;;
    help|--help|-h)
        show_help
        ;;
    *)
        echo "❌ Comando desconhecido: $1"
        echo ""
        show_help
        exit 1
        ;;
esac