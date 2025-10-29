Projeto API - Gestão de Colaboradores e Departamentos

Este é um projeto de API em Go (Golang) para gerenciar Colaboradores e Departamentos, construído com as especificações da stack solicitada.
Linguagem: Go 1.22+

Framework HTTP: Gin

ORM: GORM

Banco de Dados: PostgreSQL

Migrations: Flyway

Documentação: Swagger (usando swaggo)

Containerização: Docker + docker-compose

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
├── /sql/                           # Migrations Flyway
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