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

	_ "ManageEmployeesandDepartments/docs" // Import docs for swagger
)

// @title Employee and Department Management API
// @version 1.0
// @description API to manage employees and departments of a company.
// @host localhost:8080
// @BasePath /api/v1
func main() {
	// Load .env locally (not required in Docker, but good for dev)
	if err := godotenv.Load(); err != nil {
		log.Println(".env file not found, using environment variables.")
	}

	// Connect to database
	db, err := db.ConnectDatabase()
	if err != nil {
		log.Fatal("Failed to connect to database: ", err)
	}

	// Repositories
	employeeRepo := repository.NewEmployeeRepository(db)
	deptRepo := repository.NewDepartmentRepository(db)

	// Services
	employeeService := services.NewEmployeeService(deptRepo, employeeRepo)
	deptService := services.NewDepartmentService(deptRepo, employeeRepo)

	// Handlers
	employeeHandler := handlers.NewEmployeeHandler(employeeService)
	deptHandler := handlers.NewDepartmentHandler(deptService)
	managerHandler := handlers.NewManagerHandler(deptService)

	// Initialize Gin Router
	r := gin.Default()

	// Setup Routes
	routes.SetupRoutes(r, employeeHandler, deptHandler, managerHandler)

	// Setup Swagger
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Start Server
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Servidor iniciado na porta :%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal("Falha ao iniciar o servidor: ", err)
	}
}
