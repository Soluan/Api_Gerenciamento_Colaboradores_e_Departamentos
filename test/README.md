# Documentação dos Testes

Este documento descreve a estrutura e execução dos testes unitários do projeto API de Gerenciamento de Colaboradores e Departamentos.

## Estrutura dos Testes

## Estrutura dos Testes

```
test/
├── utils/                          ✅ FUNCIONANDO
│   ├── validators_test.go           # Testes de validação de CPF
│   └── custom_error_test.go         # Testes de erros customizados
├── models/                          ⚠️ IMPLEMENTADO (problemas de migração SQLite)
│   └── models_test.go               # Testes dos modelos GORM
├── handlers/                        ✅ FUNCIONANDO
│   ├── collaborator_handler_test.go # Testes do handler de colaboradores
│   ├── departament_handler_test.go  # Testes do handler de departamentos
│   └── gerente_handler_test.go      # Testes do handler de gerentes
├── services/                        🔄 ESTRUTURA CRIADA (sem mocks funcionais)
├── mocks/                          🔄 ESTRUTURA CRIADA
└── README.md                       ✅ COMPLETO
```

## Tecnologias Utilizadas

### Framework de Testes
- **Go standard library testing**: Framework nativo do Go
- **go.uber.org/goleak**: Detecção de vazamentos de goroutines
- **go.uber.org/mock**: Geração e uso de mocks

### Banco de Dados para Testes
- **SQLite in-memory**: Para testes dos modelos e repositórios
- **gorm.io/driver/sqlite**: Driver SQLite para GORM

## Executando os Testes

### Usando Makefile (Recomendado)

```bash
# Executar todos os testes
make test

# Executar testes específicos
make test-utils      # Testes dos utilitários
make test-models     # Testes dos modelos
make test-services   # Testes dos serviços
make test-handlers   # Testes dos handlers

# Executar com verbose
make test-verbose

# Executar com coverage
make test-coverage

# Executar benchmarks
make bench
```

### Usando go test diretamente

```bash
# Todos os testes
go test ./test/...

# Testes específicos com verbose
go test ./test/utils/... -v
go test ./test/models/... -v

# Com coverage
go test ./test/... -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html

# Benchmarks
go test ./test/... -bench=.

# Detectar race conditions
go test ./test/... -race
```

## Tipos de Testes Implementados

### 1. Testes Unitários (utils/)

#### validators_test.go
- **TestIsCPFValido**: Valida diferentes cenários de CPF (válidos e inválidos)
- **TestRemoveNaoDigitos**: Testa remoção de caracteres não numéricos
- **BenchmarkIsCPFValido**: Benchmark da validação de CPF
- **TestIsCPFValidoPanicRecovery**: Teste de robustez contra panic

#### custom_error_test.go
- **TestCustomError_Error**: Teste do método Error() do tipo CustomError
- **TestNewCustomError**: Teste da criação de erros customizados
- **TestMapErrorToCustom**: Teste do mapeamento de erros para erros customizados
- **TestPredefinedErrors**: Validação de todos os erros predefinidos
- **TestErrorsIs**: Teste da função errors.Is com erros customizados

### 2. Testes de Modelos (models/)

#### models_test.go
- **TestColaboradorModel**: Teste de criação e validação de colaboradores
- **TestColaboradorBeforeCreate**: Teste do hook BeforeCreate (geração de UUID v7)
- **TestDepartamentoModel**: Teste de criação e validação de departamentos
- **TestDepartamentoHierarchy**: Teste de hierarquia entre departamentos
- **TestSoftDelete**: Teste de soft delete do GORM
- **TestModelValidations**: Teste de validações e constraints únicos
- **TestTimestampUpdates**: Teste de atualização automática de timestamps

### 3. Testes de Serviços (services/)

#### collaborator_service_test.go
- **TestColaboradorService_CreateColaborador**: Teste de criação de colaboradores
- **TestColaboradorService_GetColaboradorComGerente**: Teste de busca com gerente
- **TestColaboradorService_UpdateColaborador**: Teste de atualização
- **TestColaboradorService_DeleteColaborador**: Teste de deleção
- **TestColaboradorService_ListColaboradores**: Teste de listagem com filtros

### 4. Testes de Handlers (handlers/)

Testes de integração para endpoints HTTP:
- Teste de requisições e respostas HTTP
- Validação de códigos de status
- Teste de serialização/deserialização JSON
- Validação de erros de negócio

## Padrões de Teste

### Nomenclatura
- Arquivos de teste terminam com `_test.go`
- Funções de teste começam com `Test`
- Benchmarks começam com `Benchmark`

### Estrutura dos Testes
```go
func TestFunctionName(t *testing.T) {
    tests := []struct {
        name     string
        input    InputType
        expected ExpectedType
        error    bool
    }{
        {
            name:     "Valid case",
            input:    validInput,
            expected: expectedOutput,
            error:    false,
        },
        // mais casos...
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // implementação do teste
        })
    }
}
```

### Uso de Mocks
```go
func TestWithMocks(t *testing.T) {
    ctrl := gomock.NewController(t)
    defer ctrl.Finish()

    mockRepo := mocks.NewMockRepository(ctrl)
    mockRepo.EXPECT().
        Method(gomock.Any()).
        Return(expectedValue, nil).
        Times(1)

    // usar o mock no teste
}
```

## Coverage Reports

Os relatórios de coverage são gerados em HTML:

```bash
make test-coverage
# Gera coverage.html
```

O arquivo `coverage.html` pode ser aberto no navegador para visualizar:
- Linhas cobertas (verde)
- Linhas não cobertas (vermelho)
- Percentual de coverage por arquivo/função

## Mocks

### Geração Automática
```bash
make generate-mocks
```

### Regeneração Manual
Se necessário regenerar mocks específicos:
```bash
mockgen -source=internal/repository/collaborator_repository.go \
        -destination=test/mocks/collaborator_repository_mock.go \
        -package=mocks
```

## CI/CD

Para integração contínua, use:
```bash
make ci
```

Isso executará:
1. Download de dependências
2. Verificação de formato e lint
3. Execução de todos os testes
4. Geração de relatório de coverage

## Boas Práticas

### 1. Isolamento
- Cada teste deve ser independente
- Use banco de dados em memória para testes
- Limpe recursos após cada teste

### 2. Nomenclatura Clara
- Nomes descritivos para casos de teste
- Use table-driven tests para múltiplos cenários

### 3. Coverage
- Mantenha coverage acima de 80%
- Teste casos de erro além dos casos de sucesso
- Teste edge cases

### 4. Performance
- Use benchmarks para código crítico
- Monitore vazamentos de memória com goleak
- Teste concorrência com flag -race

### 5. Manutenibilidade
- Mantenha testes simples e legíveis
- Use helpers para setup/teardown comum
- Atualize testes junto com mudanças de código

## Debugging

### Executar teste específico
```bash
go test -run TestSpecificFunction ./test/utils/
```

### Debug com verbose
```bash
go test -v -run TestSpecificFunction ./test/utils/
```

### Apenas testes rápidos
```bash
go test -short ./test/...
```

## Dependências de Teste

As seguintes dependências são necessárias apenas para testes:

```go
// go.mod
require (
    go.uber.org/goleak v1.3.0
    go.uber.org/mock v0.6.0
    gorm.io/driver/sqlite v1.6.0
    github.com/DATA-DOG/go-sqlmock v1.5.2
)
```

Para instalar:
```bash
make deps
```