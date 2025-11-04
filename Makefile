# Makefile for API Gerenciamento Colaboradores e Departamentos

# Variables
GOTEST=go test
GOMOD=go mod
GOFMT=go fmt
GOVET=go vet
PROJECT_PATH=ManageEmployeesandDepartments

# Default target
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  test         - Run all tests"
	@echo "  test-utils   - Run utils tests"
	@echo "  test-models  - Run models tests"
	@echo "  test-services - Run services tests"
	@echo "  test-handlers - Run handlers tests"
	@echo "  test-repository - Run repository tests"
	@echo "  test-coverage - Run tests with coverage"
	@echo "  test-verbose - Run tests with verbose output"
	@echo "  clean        - Clean test cache"
	@echo "  fmt          - Format code"
	@echo "  vet          - Run go vet"
	@echo "  deps         - Download dependencies"
	@echo "  generate-mocks - Generate mocks using mockgen"
	@echo "  docs         - Generate Swagger documentation"
	@echo "  docs-install - Install swag tool for docs generation"
	@echo "  docs-open    - Open Swagger documentation in browser"
	@echo "  swagger      - Quick access to Swagger UI (opens browser)"
	@echo "  run-with-swagger - Start app and open Swagger automatically"
	@echo "  run          - Run the application"
	@echo "  build        - Build the application"

# Test targets
.PHONY: test
test:
	$(GOTEST) ./test/...

.PHONY: test-utils
test-utils:
	$(GOTEST) ./test/utils/... -v

.PHONY: test-models
test-models:
	$(GOTEST) ./test/models/... -v

.PHONY: test-services
test-services:
	$(GOTEST) ./test/services/... -v

.PHONY: test-handlers
test-handlers:
	$(GOTEST) ./test/handlers/... -v

.PHONY: test-repository
test-repository:
	$(GOTEST) ./test/repository/... -v

.PHONY: test-all
test-all:
	$(GOTEST) ./test/... -v

.PHONY: test-coverage
test-coverage:
	$(GOTEST) ./test/... -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

.PHONY: test-verbose
test-verbose:
	$(GOTEST) ./test/... -v

.PHONY: test-race
test-race:
	$(GOTEST) ./test/... -race

.PHONY: test-short
test-short:
	$(GOTEST) ./test/... -short

# Benchmark targets
.PHONY: bench
bench:
	$(GOTEST) ./test/... -bench=.

.PHONY: bench-utils
bench-utils:
	$(GOTEST) ./test/utils/... -bench=.

# Code quality targets
.PHONY: fmt
fmt:
	$(GOFMT) ./...

.PHONY: vet
vet:
	$(GOVET) ./...

.PHONY: lint
lint: fmt vet

# Dependency management
.PHONY: deps
deps:
	$(GOMOD) download
	$(GOMOD) tidy

# Mock generation
.PHONY: generate-mocks
generate-mocks:
	@echo "Generating mocks..."
	@mkdir -p test/mocks
	mockgen -source=internal/repository/collaborator_repository.go -destination=test/mocks/collaborator_repository_mock.go -package=mocks
	mockgen -source=internal/repository/departament_repository.go -destination=test/mocks/departament_repository_mock.go -package=mocks
	mockgen -source=internal/services/collaborator_service.go -destination=test/mocks/collaborator_service_mock.go -package=mocks
	mockgen -source=internal/services/departament_service.go -destination=test/mocks/departament_service_mock.go -package=mocks
	@echo "Mocks generated successfully!"

# Cleanup targets
.PHONY: clean
clean:
	go clean -testcache
	rm -f coverage.out coverage.html

.PHONY: clean-all
clean-all: clean
	rm -rf test/mocks/*

# CI/CD targets
.PHONY: ci
ci: deps lint test-coverage

# Development targets
.PHONY: dev-setup
dev-setup: deps generate-mocks

# Database targets (for integration tests)
.PHONY: test-db-up
test-db-up:
	docker-compose -f docker-compose.test.yml up -d db

.PHONY: test-db-down
test-db-down:
	docker-compose -f docker-compose.test.yml down

# Run application
.PHONY: run
run:
	go run cmd/server/main.go

# Build application
.PHONY: build
build:
	go build -o bin/server cmd/server/main.go

# Docker targets
.PHONY: docker-test
docker-test:
	docker-compose -f docker-compose.test.yml up --build --abort-on-container-exit test

# Install tools
.PHONY: install-tools
install-tools:
	go install go.uber.org/mock/mockgen@latest
	go install github.com/swaggo/swag/cmd/swag@latest

# Swagger documentation
.PHONY: docs
docs:
	swag init -g cmd/server/main.go

.PHONY: docs-install
docs-install:
	go install github.com/swaggo/swag/cmd/swag@latest

.PHONY: docs-open
docs-open:
	@echo "Opening Swagger documentation in browser..."
	@echo "Make sure the application is running on http://localhost:8080"
	@if command -v xdg-open > /dev/null; then \
		xdg-open http://localhost:8080/swagger/index.html; \
	elif command -v open > /dev/null; then \
		open http://localhost:8080/swagger/index.html; \
	elif command -v start > /dev/null; then \
		start http://localhost:8080/swagger/index.html; \
	else \
		echo "Please open http://localhost:8080/swagger/index.html in your browser"; \
	fi

.PHONY: swagger
swagger: docs-open

.PHONY: run-with-swagger
run-with-swagger:
	@echo "Starting application and opening Swagger documentation..."
	@echo "Application will start on http://localhost:8080"
	@echo "Swagger docs will be available at http://localhost:8080/swagger/index.html"
	@go run cmd/server/main.go &
	@sleep 3
	@$(MAKE) docs-open

# All-in-one targets
.PHONY: all
all: deps lint test

.PHONY: setup
setup: install-tools dev-setup