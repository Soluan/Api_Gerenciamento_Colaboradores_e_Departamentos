Projeto API - Gestão de Colaboradores e Departamentos

Este é um projeto de API em Go (Golang) para gerenciar Colaboradores e Departamentos, construído com as especificações da stack solicitada.
Linguagem: Go 1.22+

Framework HTTP: Gin

ORM: GORM

Banco de Dados: PostgreSQL

Migrations: Flyway

Documentação: Swagger (usando swaggo)

Containerização: Docker + docker-compose

## Documentação da API

A documentação completa da API está disponível através do Swagger. Com a aplicação rodando, acesse:

**http://localhost:8080/swagger/index.html**

### Gerando/Atualizando a Documentação

Para gerar ou atualizar os arquivos de documentação Swagger após fazer alterações nos comentários dos handlers:

```bash
make docs
```

ou

```bash
swag init -g cmd/server/main.go
```

Para instalar o swag CLI:

```bash
make docs-install
```

Estrutura do Projeto

Aqui está a estrutura de diretórios recomendada para este projeto:

/projeto-gestao/
├── /cmd/
│   └── /server/
│       └── main.go                 # Entrypoint da aplicação
├── /internal/
│   ├── /config/                    # Carregamento de config (env)
│   │   └── config.go
│   ├── /database/                  # Conexão com DB (GORM)
│   │   └── database.go
│   ├── /handlers/                  # Handlers Gin (controllers)
│   │   ├── colaborador_handler.go
│   │   ├── departamento_handler.go
│   │   ├── gerente_handler.go
│   │   └── routes.go               # Definição de todas as rotas
│   ├── /models/                    # Modelos de domínio (structs GORM)
│   │   └── models.go
│   ├── /repository/                # Camada de acesso a dados (padrão Repository)
│   │   ├── colaborador_repo.go
│   │   └── departamento_repo.go
│   ├── /services/                  # Lógica de negócio (validações, ciclos, etc)
│   │   ├── colaborador_service.go
│   │   └── departamento_service.go
│   └── /validators/                # Validadores customizados (CPF, RG)
│       └── validators.go
├── /docs/                          # Arquivos gerados pelo Swagger
│   ├── docs.go
│   ├── swagger.json
│   └── swagger.yaml
├── /migrate/                           # Migrations Flyway
│   └── V1__init_schema.sql
├── .env.example                    # Exemplo de variáveis de ambiente
├── .gitignore
├── docker-compose.yml
├── Dockerfile
├── go.mod
└── go.sum


Como Executar

Clone o Repositório (hipotético):

git clone [https://github.com/seu-usuario/projeto-gestao.git](https://github.com/seu-usuario/projeto-gestao.git)
cd projeto-gestao


Crie o arquivo .env:
Copie o .env.example para um novo arquivo .env e preencha as variáveis de ambiente:

cp .env.example .env


Edite o .env com suas senhas.

Suba os Containers:
Este comando irá construir a imagem da aplicação Go, baixar a imagem do PostgreSQL e do Flyway, e iniciar os containers. O Flyway executará as migrations antes que a aplicação Go inicie.

docker-compose up --build


Acesse a Aplicação:

API: http://localhost:8080

Swagger Docs: http://localhost:8080/swagger/index.html

## Testando a API

### Via Swagger UI
1. Acesse http://localhost:8080/swagger/index.html
2. Explore os endpoints disponíveis organizados por tags (Colaboradores, Departamentos, Gerentes)
3. Clique em "Try it out" para testar diretamente na interface
4. Preencha os parâmetros necessários e execute as requisições

### Comandos Rápidos

#### Para acessar o Swagger rapidamente:
```bash
# Abrir Swagger no navegador (mais rápido)
make swagger

# Ou usar o script
./swagger.sh

# Iniciar API e abrir Swagger automaticamente
make run-with-swagger
```

#### Para verificar status:
```bash
# Verificar se API está rodando
./swagger.sh status
```

### Via cURL
Você também pode testar usando cURL. Veja os exemplos abaixo.

Documentação da API (Swagger)

Este projeto usa swaggo/swag para gerar a documentação do Swagger a partir de comentários no código (especialmente nos handlers).

Para (re)gerar a documentação após fazer alterações nos comentários:

Instale o swag (se ainda não o fez):

go install [github.com/swaggo/swag/cmd/swag@latest](https://github.com/swaggo/swag/cmd/swag@latest)


Execute o swag init na raiz do projeto, especificando o diretório do main.go:

swag init -g cmd/server/main.go


Isso atualizará os arquivos no diretório /docs.

Lógica de Negócio Implementada

Validação de CPF/RG: A unicidade é garantida pelo banco. A validação de formato (ex: CPF válido) deve ser implementada na camada de serviço/validação.

Prevenção de Ciclos: Implementada na camada de serviço (DepartamentoService) antes de qualquer atualização no DepartamentoSuperiorID. A lógica verifica se o novo "superior" não é, na verdade, um "subordinado" do departamento que está sendo movido.

Busca de Árvore Hierárquica: O endpoint GET /api/v1/departamentos/:id retorna a árvore completa usando uma função recursiva (ou CTEs no repositório) para carregar os SubDepartamentos.

Respostas de Erro: Os handlers estão configurados para retornar os códigos HTTP corretos (404, 409, 422) conforme especificado.

Exemplos de Requisições (cURL)
Aqui estão alguns exemplos de como interagir com a API usando o curl.

(Assumindo que a API está rodando em http://localhost:8080)

Colaboradores
1. Criar um novo colaborador

Bash

curl -X POST 'http://localhost:8080/api/v1/colaboradores' \
-H 'Content-Type: application/json' \
-d '{
"nome": "Ana Silva",
"cpf": "12345678909",
"rg": "MG1234567",
"departamento_id": "018f3c3e-5c79-7b21-b7e1-d45f80cfa5ab"
}'
2. Buscar um colaborador por ID

Bash

curl -X GET 'http://localhost:8080/api/v1/colaboradores/SEU_UUID_AQUI'
3. Listar colaboradores (com filtro e paginação)

Bash

curl -X POST 'http://localhost:8080/api/v1/colaboradores/listar' \
-H 'Content-Type: application/json' \
-d '{
"filtros": {
"nome": "Ana",
"departamento_id": "018f3c3e-5c79-7b21-b7e1-d45f80cfa5ab"
},
"pagina": 1,
"limite": 10
}'
4. Atualizar um colaborador

Bash

curl -X PUT 'http://localhost:8080/api/v1/colaboradores/SEU_UUID_AQUI' \
-H 'Content-Type: application/json' \
-d '{
"nome": "Ana Silva Souza",
"rg": "MG7654321"
}'
5. Deletar um colaborador

Bash

curl -X DELETE 'http://localhost:8080/api/v1/colaboradores/SEU_UUID_AQUI'
Departamentos
1. Criar um novo departamento

(Nota: O gerente_id e o departamento_superior_id devem existir previamente)

Bash

curl -X POST 'http://localhost:8080/api/v1/departamentos' \
-H 'Content-Type: application/json' \
-d '{
"nome": "Engenharia de Software",
"gerente_id": "018f3c3e-5c79-7b21-b7e1-d45f80cfa5ab",
"departamento_superior_id": "018f3c3e-0000-0000-0000-d45f80cfa5ab"
}'
2. Buscar um departamento por ID (com árvore hierárquica)

Bash

curl -X GET 'http://localhost:8080/api/v1/departamentos/SEU_UUID_AQUI'
3. Listar departamentos (com filtro e paginação)

Bash

curl -X POST 'http://localhost:8080/api/v1/departamentos/listar' \
-H 'Content-Type: application/json' \
-d '{
"filtros": {
"nome": "Engenharia",
"gerente_nome": "Ana Silva"
},
"pagina": 1,
"limite": 10
}'
Gerentes
1. Listar colaboradores subordinados (recursivamente)

Bash

curl -X GET 'http://localhost:8080/api/v1/gerentes/UUID_DO_GERENTE_AQUI/colaboradores'