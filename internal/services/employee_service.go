package services

import (
	"ManageEmployeesandDepartments/internal/models"
	"ManageEmployeesandDepartments/internal/repository"
	"ManageEmployeesandDepartments/internal/utils"

	"github.com/google/uuid"
)

type EmployeeService interface {
	CreateEmployee(name string, cpf string, rg *string, departmentID uuid.UUID) (*models.Employee, error)
	GetEmployeeWithManager(id uuid.UUID) (*EmployeeWithManagerResponse, error)
	UpdateEmployee(id uuid.UUID, name *string, rg *string, departmentID uuid.UUID) (*models.Employee, error)
	DeleteEmployee(id uuid.UUID) error
	ListEmployees(name *string, cpf *string, rg *string, deptID *uuid.UUID, page, pageSize int) ([]*models.Employee, error)
}

type employeeService struct {
	deptRepo     repository.DepartmentRepository
	employeeRepo repository.EmployeeRepository
}

func NewEmployeeService(deptRepo repository.DepartmentRepository, employeeRepo repository.EmployeeRepository) EmployeeService {
	return &employeeService{
		deptRepo:     deptRepo,
		employeeRepo: employeeRepo,
	}
}

type EmployeeWithManagerResponse struct {
	Employee    *models.Employee `json:"employee"`
	ManagerName string           `json:"manager_name,omitempty"`
}

// CreateEmployee creates a new employee with CPF/RG and department validation
func (s *employeeService) CreateEmployee(name string, cpf string, rg *string, departmentID uuid.UUID) (*models.Employee, error) {
	// Checks if department exists
	_, err := s.deptRepo.FindByID(departmentID)
	if err != nil {
		return nil, utils.ErrDepartmentNotFound
	}

	// Creates the employee
	employee := &models.Employee{
		ID:           uuid.New(),
		Name:         name,
		CPF:          cpf,
		RG:           rg,
		DepartmentID: departmentID,
	}

	err = s.employeeRepo.Create(employee)
	if s.employeeRepo.IsCPFDuplicated(err) {
		return nil, utils.ErrCPFDuplicated
	}
	if s.employeeRepo.IsRGDuplicated(err) {
		return nil, utils.ErrRGDuplicated
	}
	if err != nil {
		return nil, err
	}

	return employee, nil
}

// GetEmployeeWithManager returns an employee and the manager name from the department
func (s *employeeService) GetEmployeeWithManager(id uuid.UUID) (*EmployeeWithManagerResponse, error) {
	employee, err := s.employeeRepo.FindByID(id)
	if err != nil {
		return nil, err
	}

	dept, err := s.deptRepo.FindByID(employee.DepartmentID)
	if err != nil {
		return nil, err
	}

	var managerName string
	if *dept.ManagerID != uuid.Nil {
		manager, err := s.employeeRepo.FindByID(*dept.ManagerID)
		if err == nil {
			managerName = manager.Name
		}
	}

	return &EmployeeWithManagerResponse{
		Employee:    employee,
		ManagerName: managerName,
	}, nil
}

// UpdateEmployee updates name, RG and department of an employee
func (s *employeeService) UpdateEmployee(id uuid.UUID, name *string, rg *string, departmentID uuid.UUID) (*models.Employee, error) {
	employee, err := s.employeeRepo.FindByID(id)
	if err != nil {
		return nil, err
	}

	// Validates department
	_, err = s.deptRepo.FindByID(departmentID)
	if err != nil {
		return nil, utils.ErrDepartmentNotFound
	}

	employee.Name = *name
	employee.RG = rg
	employee.DepartmentID = departmentID

	err = s.employeeRepo.Update(employee)
	if s.employeeRepo.IsRGDuplicated(err) {
		return nil, utils.ErrRGDuplicated
	}
	if err != nil {
		return nil, err
	}

	return employee, nil
}

// DeleteEmployee removes an employee (soft delete)
func (s *employeeService) DeleteEmployee(id uuid.UUID) error {
	employee, err := s.employeeRepo.FindByID(id)
	if err != nil {
		return err
	}

	// Does not allow deletion if they are a manager of any department
	isManager, err := s.deptRepo.IsManager(id)
	if err != nil {
		return err
	}
	if isManager {
		return utils.ErrManagerCannotBeDeleted
	}

	return s.employeeRepo.Delete(employee.ID)
}

// ListEmployees lists employees with filters and pagination
func (s *employeeService) ListEmployees(name, cpf, rg *string, deptID *uuid.UUID, page, pageSize int) ([]*models.Employee, error) {
	return s.employeeRepo.List(name, cpf, rg, deptID, page, pageSize)
}
