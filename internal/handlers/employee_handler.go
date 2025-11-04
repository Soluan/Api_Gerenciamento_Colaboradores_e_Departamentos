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

type EmployeeHandler struct {
	service services.EmployeeService
}

// Constructor for the handler
func NewEmployeeHandler(s services.EmployeeService) *EmployeeHandler {
	return &EmployeeHandler{service: s}
}

// Create creates a new employee
// @Summary Create a new employee
// @Description Creates a new employee with the provided data
// @Tags Colaboradores
// @Accept json
// @Produce json
// @Param employee body models.CreateEmployeeDTO true "Employee data"
// @Success 201 {object} models.Employee
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 409 {object} map[string]string "CPF or RG already exists"
// @Failure 422 {object} map[string]string "Department not found"
// @Router /colaboradores [post]
func (h *EmployeeHandler) Create(c *gin.Context) {
	var dto models.CreateEmployeeDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	employee, err := h.service.CreateEmployee(dto.Name, dto.CPF, dto.RG, dto.DepartmentID)
	if err != nil {
		switch err {
		case utils.ErrDepartmentNotFound:
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		case utils.ErrCPFDuplicated, utils.ErrRGDuplicated:
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating employee"})
		}
		return
	}

	c.JSON(http.StatusCreated, employee)
}

// GetByID returns employee with manager name from department
// @Summary Get employee by ID
// @Description Returns an employee by ID with manager information
// @Tags Colaboradores
// @Produce json
// @Param id path string true "Employee ID (UUID)"
// @Success 200 {object} models.EmployeeWithManagerResponse
// @Failure 400 {object} map[string]string "Invalid ID"
// @Failure 404 {object} map[string]string "Employee not found"
// @Router /colaboradores/{id} [get]
func (h *EmployeeHandler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	response, err := h.service.GetEmployeeWithManager(id)
	if err != nil {
		if errors.Is(err, utils.ErrEmployeeNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Employee not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching employee"})
		return
	}

	c.JSON(http.StatusOK, response)
}

// Update updates an existing employee
// @Summary Update an employee
// @Description Updates an existing employee with the provided data
// @Tags Colaboradores
// @Accept json
// @Produce json
// @Param id path string true "Employee ID (UUID)"
// @Param employee body models.UpdateEmployeeDTO true "Updated employee data"
// @Success 200 {object} models.Employee
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 404 {object} map[string]string "Employee not found"
// @Failure 409 {object} map[string]string "RG already exists"
// @Failure 422 {object} map[string]string "Department not found"
// @Router /colaboradores/{id} [put]
func (h *EmployeeHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var dto models.UpdateEmployeeDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	employee, err := h.service.UpdateEmployee(id, dto.Name, dto.RG, *dto.DepartmentID)
	if err != nil {
		switch err {
		case utils.ErrEmployeeNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case utils.ErrDepartmentNotFound:
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		case utils.ErrRGDuplicated:
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating employee"})
		}
		return
	}

	c.JSON(http.StatusOK, employee)
}

// Delete removes an employee (soft delete)
// @Summary Delete an employee
// @Description Removes an employee (soft delete)
// @Tags Colaboradores
// @Param id path string true "Employee ID (UUID)"
// @Success 204
// @Failure 400 {object} map[string]string "Invalid ID"
// @Failure 404 {object} map[string]string "Employee not found"
// @Failure 422 {object} map[string]string "Manager cannot be deleted"
// @Router /colaboradores/{id} [delete]
func (h *EmployeeHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	err = h.service.DeleteEmployee(id)
	if err != nil {
		switch err {
		case utils.ErrEmployeeNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case utils.ErrManagerCannotBeDeleted:
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error removing employee"})
		}
		return
	}

	c.Status(http.StatusNoContent)
}

// List returns paginated employees with filters
// @Summary List employees with filters
// @Description Returns a paginated list of employees based on filters
// @Tags Colaboradores
// @Accept json
// @Produce json
// @Param filters body models.ListEmployeesDTO false "Filters and pagination"
// @Success 200 {array} models.Employee
// @Failure 400 {object} map[string]string "Invalid request"
// @Router /colaboradores/listar [post]
func (h *EmployeeHandler) List(c *gin.Context) {
	var dto models.ListEmployeesDTO

	// Pagination defaults
	dto.Page = 1
	dto.PageSize = 10

	if err := c.ShouldBindJSON(&dto); err != nil && err.Error() != "EOF" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	if dto.Page <= 0 {
		dto.Page = 1
	}
	if dto.PageSize <= 0 {
		dto.PageSize = 10
	}

	employees, err := h.service.ListEmployees(dto.Name, dto.CPF, dto.RG, dto.DepartmentID, dto.Page, dto.PageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error listing employees"})
		return
	}

	if employees == nil {
		employees = []*models.Employee{}
	}

	c.JSON(http.StatusOK, employees)
}
