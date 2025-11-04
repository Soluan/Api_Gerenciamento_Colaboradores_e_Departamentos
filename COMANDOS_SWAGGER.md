# Guia Rápido - Comandos Swagger

Este documento contém exemplos práticos de como usar os comandos criados para acessar a documentação Swagger.

## 🚀 Comandos Mais Usados

### 1. Acesso Rápido ao Swagger
```bash
# Comando mais simples - abre o Swagger no navegador
make swagger
```

### 2. Iniciar API + Swagger
```bash
# Inicia a API e abre o Swagger automaticamente
make run-with-swagger
```

### 3. Verificar Status da API
```bash
# Verifica se a API está rodando
./swagger.sh status
```

## 📋 Todos os Comandos Disponíveis

### Via Makefile:
```bash
make docs           # Gerar documentação Swagger
make docs-install   # Instalar ferramenta swag
make docs-open      # Abrir Swagger no navegador
make swagger        # Atalho para abrir Swagger
make run-with-swagger # Iniciar API + abrir Swagger
```

### Via Script swagger.sh:
```bash
./swagger.sh        # Abrir Swagger (comando padrão)
./swagger.sh open   # Abrir Swagger no navegador
./swagger.sh status # Verificar se API está rodando
./swagger.sh url    # Mostrar URLs da API
./swagger.sh help   # Mostrar ajuda
```

## 🔧 Fluxo de Trabalho Típico

### Primeira vez usando:
```bash
# 1. Instalar dependências
make deps

# 2. Instalar swag CLI
make docs-install

# 3. Gerar documentação
make docs

# 4. Iniciar API e abrir Swagger
make run-with-swagger
```

### Desenvolvimento diário:
```bash
# Iniciar e acessar rapidamente
make run-with-swagger

# Ou, se a API já estiver rodando:
make swagger
```

### Após fazer mudanças nos handlers:
```bash
# Regenerar documentação
make docs

# Recarregar página no navegador para ver mudanças
```

## 🌐 URLs Importantes

- **API Base:** http://localhost:8080
- **Swagger UI:** http://localhost:8080/swagger/index.html
- **Health Check:** http://localhost:8080/health (se implementado)

## 🐛 Solução de Problemas

### Problema: "API não está rodando"
```bash
# Verificar status
./swagger.sh status

# Iniciar API
make run
# OU
go run cmd/server/main.go
```

### Problema: "Comando não encontrado"
```bash
# Verificar se está no diretório correto
pwd

# Tornar script executável
chmod +x swagger.sh
```

### Problema: "swag command not found"
```bash
# Instalar swag CLI
make docs-install
```

## 💡 Dicas

1. **Use `make swagger`** - é o comando mais rápido para acessar o Swagger
2. **Use `make run-with-swagger`** - quando quiser iniciar tudo de uma vez
3. **Use `./swagger.sh status`** - para verificar se a API está rodando antes de abrir o navegador
4. **Mantenha a documentação atualizada** - execute `make docs` após mudanças nos handlers