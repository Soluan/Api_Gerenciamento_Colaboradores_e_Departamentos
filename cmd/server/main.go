package main

import (
	"ManageEmployeesandDepartments/internal/db"
	"ManageEmployeesandDepartments/internal/handlers"
	"ManageEmployeesandDepartments/internal/repository"
	"ManageEmployeesandDepartments/internal/routes"
	"ManageEmployeesandDepartments/internal/services"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title API de Gestão de Colaboradores e Departamentos
// @version 1.0
// @description API para gerenciar colaboradores e departamentos de uma empresa.
// @host localhost:8080
// @BasePath /api/v1
func main() {
	// Carrega .env localmente (não necessário no Docker, mas bom para dev)
	if err := godotenv.Load(); err != nil {
		log.Println("Arquivo .env não encontrado, usando variáveis de ambiente.")
	}

	// Conecta ao banco de dados
	db, err := db.ConnectDatabase()
	if err != nil {
		log.Fatal("Falha ao conectar ao banco de dados: ", err)
	}

	// Repositórios
	colabRepo := repository.NewColaboradorRepository(db)
	deptoRepo := repository.NewDepartamentoRepository(db)

	// Serviços
	colabService := services.NewColaboradorService(deptoRepo, colabRepo)
	deptoService := services.NewDepartamentoService(deptoRepo, colabRepo)

	// Handlers
	colabHandler := handlers.NewColaboradorHandler(colabService)
	deptoHandler := handlers.NewDepartamentoHandler(deptoService)
	gerenteHandler := handlers.NewGerenteHandler(deptoService)

	// Inicializa o Roteador Gin
	r := gin.Default()

	// Configura as Rotas
	routes.SetupRoutes(r, colabHandler, deptoHandler, gerenteHandler)

	// Configura o Swagger
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Inicia o Servidor
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Servidor iniciado na porta :%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal("Falha ao iniciar o servidor: ", err)
	}
}
