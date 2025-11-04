package models

import "github.com/google/uuid"

// EmployeeWithManagerResponse is the DTO response for GetByID.
type EmployeeWithManagerResponse struct {
	*Employee
	ManagerName *string `json:"manager_name"`
}

type CreateEmployeeDTO struct {
	Name         string    `json:"name" binding:"required"`
	CPF          string    `json:"cpf" binding:"required"` // Format validation (e.g: 11 digits) can be added
	RG           *string   `json:"rg"`
	DepartmentID uuid.UUID `json:"department_id" binding:"required"`
}

type UpdateEmployeeDTO struct {
	Name         *string    `json:"name"`
	RG           *string    `json:"rg"`
	DepartmentID *uuid.UUID `json:"department_id"`
}

type ListEmployeesDTO struct {
	Name         *string    `json:"name"`
	CPF          *string    `json:"cpf"`
	RG           *string    `json:"rg"`
	DepartmentID *uuid.UUID `json:"department_id"`
	Page         int        `json:"page" binding:"omitempty,gte=1"`
	PageSize     int        `json:"page_size" binding:"omitempty,gte=1"`
}

// Department DTOs
type CreateDepartmentDTO struct {
	Name               string     `json:"name" binding:"required"`
	ManagerID          uuid.UUID  `json:"manager_id" binding:"required"`
	ParentDepartmentID *uuid.UUID `json:"parent_department_id"`
}

// UpdateDepartmentDTO is used to update a department.
type UpdateDepartmentDTO struct {
	Name               *string    `json:"name"`
	ManagerID          *uuid.UUID `json:"manager_id"`
	ParentDepartmentID *uuid.UUID `json:"parent_department_id"`
}

// ListDepartmentsDTO is used for filters and pagination.
type ListDepartmentsDTO struct {
	Name               *string    `json:"name"`
	ManagerName        *string    `json:"manager_name"` // Special filter
	ParentDepartmentID *uuid.UUID `json:"parent_department_id"`
	Page               int        `json:"page" binding:"omitempty,gte=1"`
	PageSize           int        `json:"page_size" binding:"omitempty,gte=1"`
}
