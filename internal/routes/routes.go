package routes

import (
	"ManageEmployeesandDepartments/internal/handlers"

	"github.com/gin-gonic/gin"
)

// SetupRoutes configures all API endpoints in the Gin router.
func SetupRoutes(
	r *gin.Engine,
	employeeHandler *handlers.EmployeeHandler,
	deptHandler *handlers.DepartamentoHandler,
	managerHandler *handlers.ManagerHandler,
) {
	v1 := r.Group("/api/v1")
	{
		// Rotas de Colaboradores
		colab := v1.Group("/colaboradores")
		{
			colab.POST("", employeeHandler.Create)
			colab.GET("/:id", employeeHandler.GetByID)
			colab.PUT("/:id", employeeHandler.Update)
			colab.DELETE("/:id", employeeHandler.Delete)
			colab.POST("/listar", employeeHandler.List)
		}

		// Rotas de Departamentos
		depto := v1.Group("/departamentos")
		{
			depto.POST("", deptHandler.Create)
			depto.GET("/:id", deptHandler.GetByID)
			depto.PUT("/:id", deptHandler.Update)
			depto.DELETE("/:id", deptHandler.Delete)
			depto.POST("/listar", deptHandler.List)
		}

		// Rotas de Gerentes
		gerentes := v1.Group("/gerentes")
		{
			gerentes.GET("/:id/colaboradores", managerHandler.GetSubordinates)
		}
	}
}
