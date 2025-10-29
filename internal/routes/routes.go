package routes

import (
	"ManageEmployeesandDepartments/internal/handlers"

	"github.com/gin-gonic/gin"
)

// SetupRoutes configura todos os endpoints da API no roteador Gin.
func SetupRoutes(
	r *gin.Engine,
	colabHandler *handlers.ColaboradorHandler,
	deptoHandler *handlers.DepartamentoHandler,
	gerenteHandler *handlers.GerenteHandler,
) {
	v1 := r.Group("/api/v1")
	{
		// Rotas de Colaboradores
		colab := v1.Group("/colaboradores")
		{
			colab.POST("", colabHandler.Create)
			colab.GET("/:id", colabHandler.GetByID)
			colab.PUT("/:id", colabHandler.Update)
			colab.DELETE("/:id", colabHandler.Delete)
			colab.POST("/listar", colabHandler.List)
		}

		// Rotas de Departamentos
		depto := v1.Group("/departamentos")
		{
			depto.POST("", deptoHandler.Create)
			depto.GET("/:id", deptoHandler.GetByID)
			depto.PUT("/:id", deptoHandler.Update)
			depto.DELETE("/:id", deptoHandler.Delete)
			depto.POST("/listar", deptoHandler.List)
		}

		// Rotas de Gerentes
		gerentes := v1.Group("/gerentes")
		{
			gerentes.GET("/:id/colaboradores", gerenteHandler.GetSubordinados)
		}
	}
}
