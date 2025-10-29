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

type ColaboradorHandler struct {
	service services.ColaboradorService
}

// Construtor do handler
func NewColaboradorHandler(s services.ColaboradorService) *ColaboradorHandler {
	return &ColaboradorHandler{service: s}
}

// Create cria um novo colaborador
func (h *ColaboradorHandler) Create(c *gin.Context) {
	var dto models.CreateColaboradorDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Requisição inválida: " + err.Error()})
		return
	}

	colab, err := h.service.CreateColaborador(dto.Nome, dto.CPF, dto.RG, dto.DepartamentoID)
	if err != nil {
		switch err {
		case utils.ErrDepartamentoNaoEncontrado:
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		case utils.ErrCPFDuplicado, utils.ErrRGDuplicado:
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao criar colaborador"})
		}
		return
	}

	c.JSON(http.StatusCreated, colab)
}

// GetByID retorna colaborador com nome do gerente do departamento
func (h *ColaboradorHandler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	response, err := h.service.GetColaboradorComGerente(id)
	if err != nil {
		if errors.Is(err, utils.ErrColaboradorNaoEncontrado) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Colaborador não encontrado"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar colaborador"})
		return
	}

	c.JSON(http.StatusOK, response)
}

// Update atualiza um colaborador existente
func (h *ColaboradorHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	var dto models.UpdateColaboradorDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Requisição inválida: " + err.Error()})
		return
	}

	colab, err := h.service.UpdateColaborador(id, dto.Nome, dto.RG, *dto.DepartamentoID)
	if err != nil {
		switch err {
		case utils.ErrColaboradorNaoEncontrado:
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case utils.ErrDepartamentoNaoEncontrado:
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		case utils.ErrRGDuplicado:
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao atualizar colaborador"})
		}
		return
	}

	c.JSON(http.StatusOK, colab)
}

// Delete remove um colaborador (soft delete)
func (h *ColaboradorHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	err = h.service.DeleteColaborador(id)
	if err != nil {
		switch err {
		case utils.ErrColaboradorNaoEncontrado:
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case utils.ErrGerenteNaoPodeSerExcluido:
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao remover colaborador"})
		}
		return
	}

	c.Status(http.StatusNoContent)
}

// List retorna colaboradores paginados com filtros
func (h *ColaboradorHandler) List(c *gin.Context) {
	var dto models.ListColaboradoresDTO

	// Defaults de paginação
	dto.Pagina = 1
	dto.TamanhoPagina = 10

	if err := c.ShouldBindJSON(&dto); err != nil && err.Error() != "EOF" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Requisição inválida: " + err.Error()})
		return
	}

	if dto.Pagina <= 0 {
		dto.Pagina = 1
	}
	if dto.TamanhoPagina <= 0 {
		dto.TamanhoPagina = 10
	}

	colabs, err := h.service.ListColaboradores(dto.Nome, dto.CPF, dto.RG, dto.DepartamentoID, dto.Pagina, dto.TamanhoPagina)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao listar colaboradores"})
		return
	}

	if colabs == nil {
		colabs = []*models.Colaborador{}
	}

	c.JSON(http.StatusOK, colabs)
}
