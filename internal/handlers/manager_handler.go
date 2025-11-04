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

// ManagerHandler handles HTTP requests for Managers.
type ManagerHandler struct {
	deptService services.DepartmentService
}

// NewManagerHandler creates a new manager handler.
func NewManagerHandler(ds services.DepartmentService) *ManagerHandler {
	return &ManagerHandler{deptService: ds}
}

// GetSubordinates lists employees subordinated to a manager
// @Summary List employees subordinated to a manager
// @Description Returns all employees from departments subordinated to the manager, recursively
// @Tags Gerentes
// @Produce json
// @Param id path string true "Manager ID (Employee UUID)"
// @Success 200 {array} models.Employee
// @Failure 400 {object} map[string]string "Invalid ID"
// @Failure 404 {object} map[string]string "Manager not found"
// @Router /gerentes/{id}/colaboradores [get]
func (h *ManagerHandler) GetSubordinates(c *gin.Context) {
	managerID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid manager ID"})
		return
	}

	employees, err := h.deptService.GetSubordinateEmployeesRecursively(managerID)
	if err != nil {
		if errors.Is(err, utils.ErrManagerNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching subordinate employees"})
		return
	}

	if employees == nil {
		employees = []*models.Employee{} // Returns empty list
	}

	c.JSON(http.StatusOK, employees)
}
