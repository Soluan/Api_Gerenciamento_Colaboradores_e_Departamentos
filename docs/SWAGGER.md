# Documentação da API - Swagger

Este diretório contém os arquivos de documentação Swagger para a API de Gerenciamento de Colaboradores e Departamentos.

## Arquivos

- `docs.go` - Arquivo principal gerado pelo swag que contém a especificação da API em formato Go
- `swagger.json` - Especificação da API em formato JSON
- `swagger.yaml` - Especificação da API em formato YAML

## Como acessar a documentação

Quando a aplicação estiver rodando, você pode acessar a documentação Swagger através da URL:

```
http://localhost:8080/swagger/index.html
```

### Comandos rápidos para acessar o Swagger

#### Via Makefile:
```bash
# Abrir Swagger no navegador (API deve estar rodando)
make swagger

# Iniciar API e abrir Swagger automaticamente
make run-with-swagger

# Apenas abrir o navegador
make docs-open
```

#### Via Script bash:
```bash
# Abrir Swagger no navegador
./swagger.sh

# Verificar se a API está rodando
./swagger.sh status

# Mostrar URLs da API
./swagger.sh url

# Ajuda
./swagger.sh help
```

## Como gerar/atualizar a documentação

### Pré-requisitos

Instale o swag CLI:

```bash
make docs-install
```

ou

```bash
go install github.com/swaggo/swag/cmd/swag@latest
```

### Gerar documentação

Para gerar/atualizar os arquivos de documentação após fazer alterações nos comentários dos handlers:

```bash
make docs
```

ou

```bash
swag init -g cmd/server/main.go
```

## Estrutura da API

A API está organizada em três principais grupos de endpoints:

### Colaboradores (`/api/v1/colaboradores`)
- `POST /colaboradores` - Criar novo colaborador
- `GET /colaboradores/{id}` - Buscar colaborador por ID
- `PUT /colaboradores/{id}` - Atualizar colaborador
- `DELETE /colaboradores/{id}` - Remover colaborador
- `POST /colaboradores/listar` - Listar colaboradores com filtros

### Departamentos (`/api/v1/departamentos`)
- `POST /departamentos` - Criar novo departamento
- `GET /departamentos/{id}` - Buscar departamento por ID (com árvore hierárquica)
- `PUT /departamentos/{id}` - Atualizar departamento
- `DELETE /departamentos/{id}` - Remover departamento
- `POST /departamentos/listar` - Listar departamentos com filtros

### Gerentes (`/api/v1/gerentes`)
- `GET /gerentes/{id}/colaboradores` - Listar colaboradores subordinados recursivamente

## Modelos de Dados

A documentação inclui todos os modelos de dados (DTOs) usados pela API:

- `Employee` - Modelo principal do colaborador
- `Department` - Modelo principal do departamento
- `CreateEmployeeDTO` - DTO para criação de colaborador
- `UpdateEmployeeDTO` - DTO para atualização de colaborador
- `ListEmployeesDTO` - DTO para listagem de colaboradores com filtros
- `CreateDepartmentDTO` - DTO para criação de departamento
- `UpdateDepartmentDTO` - DTO para atualização de departamento
- `ListDepartmentsDTO` - DTO para listagem de departamentos com filtros
- `EmployeeWithManagerResponse` - DTO de resposta com informações do gerente

## Códigos de Resposta

A API usa os seguintes códigos de resposta HTTP:

- `200 OK` - Sucesso em operações de busca/listagem/atualização
- `201 Created` - Sucesso na criação de recursos
- `204 No Content` - Sucesso na remoção de recursos
- `400 Bad Request` - Dados inválidos na requisição
- `404 Not Found` - Recurso não encontrado
- `409 Conflict` - Conflito (ex: CPF/RG duplicado)
- `422 Unprocessable Entity` - Erro de validação de negócio

## Exemplo de Uso

Você pode testar a API diretamente através da interface Swagger ou usar ferramentas como curl, Postman, etc.

Exemplo de criação de colaborador via curl:

```bash
curl -X POST 'http://localhost:8080/api/v1/colaboradores' \
-H 'Content-Type: application/json' \
-d '{
  "name": "João Silva", 
  "cpf": "12345678901",
  "rg": "MG1234567",
  "department_id": "018f3c3e-5c79-7b21-b7e1-d45f80cfa5ab"
}'
```