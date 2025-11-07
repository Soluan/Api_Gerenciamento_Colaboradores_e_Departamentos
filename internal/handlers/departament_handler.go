package handlers

import (
	"ManageEmployeesandDepartments/internal/models"
	"ManageEmployeesandDepartments/internal/services"
	"ManageEmployeesandDepartments/internal/utils"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// DepartamentoHandler lida com as requisições HTTP para Departamentos.
type DepartamentoHandler struct {
	service services.DepartamentoService
}

// NewDepartamentoHandler cria um novo handler de departamento.
func NewDepartamentoHandler(s services.DepartamentoService) *DepartamentoHandler {
	return &DepartamentoHandler{service: s}
}

// Create @Summary Cria um novo departamento
// @Description Cria um novo departamento (valida gerente_id)
// @Tags Departamentos
// @Accept json
// @Produce json
// @Param departamento body CreateDepartamentoDTO true "Dados do Departamento"
// @Success 201 {object} models.Departamento
// @Failure 400 {object} map[string]string "Requisição inválida"
// @Failure 422 {object} map[string]string "Erro de validação (Gerente/Depto Superior inválido)"
// @Router /departamentos [post]
func (h *DepartamentoHandler) Create(c *gin.Context) {
	var dto models.CreateDepartmentDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Requisição inválida: " + err.Error()})
		return
	}

	depto, err := h.service.CreateDepartamento(dto.Name, dto.ManagerID, dto.ParentDepartmentID)
	if err != nil {
		if err == utils.ErrManagerNotFound || err == utils.ErrDepartmentNotFound || err == utils.ErrManagerNotBelongToDepartment {
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao criar departamento"})
		return
	}

	c.JSON(http.StatusCreated, depto)
}

// GetByID @Summary Retorna um departamento por ID com árvore hierárquica
// @Description Retorna departamento, gerente e a árvore hierárquica completa dos subdepartamentos
// @Tags Departamentos
// @Produce json
// @Param id path string true "ID do Departamento (UUID)"
// @Success 200 {object} models.Departamento "Departamento com SubDepartamentos preenchidos"
// @Failure 404 {object} map[string]string "Departamento não encontrado"
// @Router /departamentos/{id} [get]
func (h *DepartamentoHandler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	// O serviço deve carregar a árvore completa
	depto, err := h.service.GetDepartamentoComArvore(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Departamento não encontrado"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar departamento"})
		return
	}

	c.JSON(http.StatusOK, depto)
}

// Update @Summary Atualiza um departamento
// @Description Atualiza dados de um departamento (impede ciclos)
// @Tags Departamentos
// @Accept json
// @Produce json
// @Param id path string true "ID do Departamento (UUID)"
// @Param departamento body UpdateDepartamentoDTO true "Dados para atualizar"
// @Success 200 {object} models.Departamento
// @Failure 400 {object} map[string]string "Requisição inválida"
// @Failure 404 {object} map[string]string "Departamento não encontrado"
// @Failure 422 {object} map[string]string "Erro de validação (Gerente/Depto Superior inválido ou Ciclo detectado)"
// @Router /departamentos/{id} [put]
func (h *DepartamentoHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	var dto models.UpdateDepartmentDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Requisição inválida: " + err.Error()})
		return
	}

	depto, err := h.service.UpdateDepartamento(id, dto.Name, dto.ManagerID, dto.ParentDepartmentID)
	if err != nil {
		switch err {
		case gorm.ErrRecordNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "Departamento não encontrado"})
		case utils.ErrCycleDetected, utils.ErrManagerNotFound, utils.ErrDepartmentHasSubDepartments, utils.ErrManagerNotBelongToDepartment:
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao atualizar departamento"})
		}
		return
	}

	c.JSON(http.StatusOK, depto)
}

// Delete @Summary Remove um departamento
// @Description Remove um departamento (soft delete)
// @Tags Departamentos
// @Param id path string true "ID do Departamento (UUID)"
// @Success 204 "Sem conteúdo"
// @Failure 404 {object} map[string]string "Departamento não encontrado"
// @Failure 422 {object} map[string]string "Não é possível remover depto com colaboradores ou sub-deptos"
// @Router /departamentos/{id} [delete]
func (h *DepartamentoHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	err = h.service.DeleteDepartamento(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Departamento não encontrado"})
			return
		}
		if errors.Is(err, utils.ErrEmployeeNotFound) || errors.Is(err, utils.ErrDepartmentHasSubDepartments) {
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao remover departamento"})
		return
	}

	c.Status(http.StatusNoContent)
}

// List @Summary Lista departamentos com filtros
// @Description Retorna uma lista paginada de departamentos com base nos filtros
// @Tags Departamentos
// @Accept json
// @Produce json
// @Param filtros body ListDepartamentosDTO false "Filtros e Paginação"
// @Success 200 {array} models.Departamento
// @Failure 400 {object} map[string]string "Requisição inválida"
// @Router /departamentos/listar [post]
func (h *DepartamentoHandler) List(c *gin.Context) {
	var dto models.ListDepartmentsDTO

	// Defaults
	dto.Page = 1
	dto.PageSize = 10

	if err := c.ShouldBindJSON(&dto); err != nil {
		if err.Error() != "EOF" { // Permite body vazio
			c.JSON(http.StatusBadRequest, gin.H{"error": "Requisição inválida: " + err.Error()})
			return
		}
	}

	if dto.Page <= 0 {
		dto.Page = 1
	}
	if dto.PageSize <= 0 {
		dto.PageSize = 10
	}

	deptos, err := h.service.ListDepartamentos(dto.Name, dto.ManagerName, dto.ParentDepartmentID, dto.Page, dto.PageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao listar departamentos"})
		return
	}

	if deptos == nil {
		deptos = []*models.Department{}
	}

	c.JSON(http.StatusOK, deptos)
}
