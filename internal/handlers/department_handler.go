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

// DepartmentHandler handles HTTP requests for Departments.
type DepartmentHandler struct {
	service services.DepartmentService
}

// NewDepartmentHandler creates a new department handler.
func NewDepartmentHandler(s services.DepartmentService) *DepartmentHandler {
	return &DepartmentHandler{service: s}
}

// Create creates a new department
// @Summary Create a new department
// @Description Creates a new department with the provided data
// @Tags Departamentos
// @Accept json
// @Produce json
// @Param department body models.CreateDepartmentDTO true "Department data"
// @Success 201 {object} models.Department
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 422 {object} map[string]string "Validation error (Invalid Manager/Parent Department)"
// @Router /departamentos [post]
func (h *DepartmentHandler) Create(c *gin.Context) {
	var dto models.CreateDepartmentDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	dept, err := h.service.CreateDepartment(dto.Name, dto.ManagerID, dto.ParentDepartmentID)
	if err != nil {
		if err == utils.ErrManagerNotFound || err == utils.ErrParentDepartmentNotFound || err == utils.ErrManagerNotBelongToDepartment {
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating department"})
		return
	}

	c.JSON(http.StatusCreated, dept)
}

// GetByID returns a department by ID with hierarchical tree
// @Summary Get department by ID with hierarchical tree
// @Description Returns department with manager and complete hierarchical tree of sub-departments
// @Tags Departamentos
// @Produce json
// @Param id path string true "Department ID (UUID)"
// @Success 200 {object} models.Department "Department with SubDepartments filled"
// @Failure 400 {object} map[string]string "Invalid ID"
// @Failure 404 {object} map[string]string "Department not found"
// @Router /departamentos/{id} [get]
func (h *DepartmentHandler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	// The service should load the complete tree
	dept, err := h.service.GetDepartmentWithTree(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Department not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching department"})
		return
	}

	c.JSON(http.StatusOK, dept)
}

// Update updates a department
// @Summary Update a department
// @Description Updates department data and prevents cycles
// @Tags Departamentos
// @Accept json
// @Produce json
// @Param id path string true "Department ID (UUID)"
// @Param department body models.UpdateDepartmentDTO true "Data to update"
// @Success 200 {object} models.Department
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 404 {object} map[string]string "Department not found"
// @Failure 422 {object} map[string]string "Validation error (Invalid Manager/Parent Department or Cycle detected)"
// @Router /departamentos/{id} [put]
func (h *DepartmentHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var dto models.UpdateDepartmentDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	dept, err := h.service.UpdateDepartment(id, dto.Name, dto.ManagerID, dto.ParentDepartmentID)
	if err != nil {
		switch err {
		case gorm.ErrRecordNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "Department not found"})
		case utils.ErrCycleDetected, utils.ErrManagerNotFound, utils.ErrParentDepartmentNotFound, utils.ErrManagerNotBelongToDepartment:
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating department"})
		}
		return
	}

	c.JSON(http.StatusOK, dept)
}

// Delete removes a department
// @Summary Delete a department
// @Description Removes a department (soft delete)
// @Tags Departamentos
// @Param id path string true "Department ID (UUID)"
// @Success 204
// @Failure 400 {object} map[string]string "Invalid ID"
// @Failure 404 {object} map[string]string "Department not found"
// @Failure 422 {object} map[string]string "Cannot remove department with employees or sub-departments"
// @Router /departamentos/{id} [delete]
func (h *DepartmentHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	err = h.service.DeleteDepartment(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Department not found"})
			return
		}
		if errors.Is(err, utils.ErrDepartmentHasEmployees) || errors.Is(err, utils.ErrDepartmentHasSubDepartments) {
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error removing department"})
		return
	}

	c.Status(http.StatusNoContent)
}

// List lists departments with filters
// @Summary List departments with filters
// @Description Returns a paginated list of departments based on filters
// @Tags Departamentos
// @Accept json
// @Produce json
// @Param filters body models.ListDepartmentsDTO false "Filters and pagination"
// @Success 200 {array} models.Department
// @Failure 400 {object} map[string]string "Invalid request"
// @Router /departamentos/listar [post]
func (h *DepartmentHandler) List(c *gin.Context) {
	var dto models.ListDepartmentsDTO

	// Defaults
	dto.Page = 1
	dto.PageSize = 10

	if err := c.ShouldBindJSON(&dto); err != nil {
		if err.Error() != "EOF" { // Allows empty body
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
			return
		}
	}

	if dto.Page <= 0 {
		dto.Page = 1
	}
	if dto.PageSize <= 0 {
		dto.PageSize = 10
	}

	departments, err := h.service.ListDepartments(dto.Name, dto.ManagerName, dto.ParentDepartmentID, dto.Page, dto.PageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error listing departments"})
		return
	}

	if departments == nil {
		departments = []*models.Department{} // Returns empty list
	}

	c.JSON(http.StatusOK, departments)
}
