package services

import (
	"ManageEmployeesandDepartments/internal/models"
	"ManageEmployeesandDepartments/internal/repository"
	"ManageEmployeesandDepartments/internal/utils"

	"github.com/google/uuid"
)

type DepartmentService interface {
	CreateDepartment(name string, managerID uuid.UUID, parentID *uuid.UUID) (*models.Department, error)
	GetDepartmentWithTree(id uuid.UUID) (*models.Department, error)
	UpdateDepartment(id uuid.UUID, name *string, managerID *uuid.UUID, parentID *uuid.UUID) (*models.Department, error)
	DeleteDepartment(id uuid.UUID) error
	ListDepartments(name, managerName *string, parentID *uuid.UUID, page, pageSize int) ([]*models.Department, error)
	GetSubordinateEmployeesRecursively(managerID uuid.UUID) ([]*models.Employee, error)
}

type departmentService struct {
	deptRepo     repository.DepartmentRepository
	employeeRepo repository.EmployeeRepository
}

func NewDepartmentService(dr repository.DepartmentRepository, cr repository.EmployeeRepository) DepartmentService {
	return &departmentService{deptRepo: dr, employeeRepo: cr}
}

func (s *departmentService) CreateDepartment(name string, managerID uuid.UUID, parentID *uuid.UUID) (*models.Department, error) {
	// Validates Manager
	_, err := s.employeeRepo.FindByID(managerID)
	if err != nil {
		return nil, utils.ErrManagerNotFound
	}

	// Validates Parent Department (if provided)
	if parentID != nil && *parentID != uuid.Nil {
		if _, err := s.deptRepo.FindByID(*parentID); err != nil {
			return nil, utils.ErrParentDepartmentNotFound
		}
	}

	dept := &models.Department{
		Name:               name,
		ManagerID:          &managerID,
		ParentDepartmentID: parentID,
	}

	if err := s.deptRepo.Create(dept); err != nil {
		return nil, err
	}

	return dept, nil
}

func (s *departmentService) GetDepartmentWithTree(id uuid.UUID) (*models.Department, error) {
	dept, err := s.deptRepo.FindByIDWithManager(id)
	if err != nil {
		return nil, err
	}

	// Load sub-departments recursively (simplified for now)
	subDepts, err := s.deptRepo.FindSubDepartments(id)
	if err != nil {
		return nil, err
	}
	dept.SubDepartments = subDepts

	return dept, nil
}

func (s *departmentService) UpdateDepartment(id uuid.UUID, name *string, managerID *uuid.UUID, parentID *uuid.UUID) (*models.Department, error) {
	dept, err := s.deptRepo.FindByID(id)
	if err != nil {
		return nil, utils.ErrDepartmentNotFound
	}

	if name != nil {
		dept.Name = *name
	}
	if managerID != nil {
		dept.ManagerID = managerID
	}
	if parentID != nil {
		dept.ParentDepartmentID = parentID
	}

	if err := s.deptRepo.Update(dept); err != nil {
		return nil, err
	}

	return dept, nil
}

func (s *departmentService) DeleteDepartment(id uuid.UUID) error {
	// Check if department has employees
	count, err := s.employeeRepo.CountByDepartmentID(id)
	if err != nil {
		return err
	}
	if count > 0 {
		return utils.ErrDepartmentHasEmployees
	}

	// Check if department has sub-departments
	subCount, err := s.deptRepo.CountSubDepartments(id)
	if err != nil {
		return err
	}
	if subCount > 0 {
		return utils.ErrDepartmentHasSubDepartments
	}

	return s.deptRepo.Delete(id)
}

func (s *departmentService) ListDepartments(name, managerName *string, parentID *uuid.UUID, page, pageSize int) ([]*models.Department, error) {
	return s.deptRepo.List(name, managerName, parentID, page, pageSize)
}

func (s *departmentService) GetSubordinateEmployeesRecursively(managerID uuid.UUID) ([]*models.Employee, error) {
	// Find departments managed by this manager
	departments, err := s.deptRepo.FindByManagerID(managerID)
	if err != nil {
		return nil, err
	}

	if len(departments) == 0 {
		return nil, utils.ErrManagerNotFound
	}

	// Get department IDs
	var deptIDs []uuid.UUID
	for _, dept := range departments {
		deptIDs = append(deptIDs, dept.ID)
	}

	// Find all employees in these departments
	employees, err := s.employeeRepo.FindByDepartmentIDs(deptIDs)
	if err != nil {
		return nil, err
	}

	return employees, nil
}
