package handlers

import (
	"ManageEmployeesandDepartments/internal/models"
	"ManageEmployeesandDepartments/internal/services"
	"ManageEmployeesandDepartments/internal/utils"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// GerenteHandler lida com as requisições HTTP para Gerentes.
type GerenteHandler struct {
	deptoService services.DepartmentService
}

// NewGerenteHandler cria um novo handler de gerente.
func NewGerenteHandler(ds services.DepartmentService) *GerenteHandler {
	return &GerenteHandler{deptoService: ds}
}

// GetSubordinados @Summary Lista colaboradores subordinados a um gerente
// @Description Retorna todos os colaboradores dos departamentos subordinados ao gerente, recursivamente
// @Tags Gerentes
// @Produce json
// @Param id path string true "ID do Gerente (UUID do Colaborador)"
// @Success 200 {array} models.Colaborador
// @Failure 400 {object} map[string]string "ID inválido"
// @Failure 404 {object} map[string]string "Gerente não encontrado"
// @Router /gerentes/{id}/colaboradores [get]
func (h *GerenteHandler) GetSubordinados(c *gin.Context) {
	gerenteID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID de gerente inválido"})
		return
	}

	colaboradores, err := h.deptoService.GetSubordinateEmployeesRecursively(gerenteID)
	if err != nil {
		if errors.Is(err, utils.ErrEmployeeNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar colaboradores subordinados"})
		return
	}

	if colaboradores == nil {
		colaboradores = []*models.Employee{} // Retorna lista vazia
	}

	c.JSON(http.StatusOK, colaboradores)
}
