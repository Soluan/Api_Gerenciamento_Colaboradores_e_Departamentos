package repository

import (
	"ManageEmployeesandDepartments/internal/models"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Public interface for the employee repository
type EmployeeRepository interface {
	Create(employee *models.Employee) error
	FindByID(id uuid.UUID) (*models.Employee, error)
	FindAll() ([]models.Employee, error)
	Update(employee *models.Employee) error
	Delete(id uuid.UUID) error
	CountByDepartmentID(deptID uuid.UUID) (int64, error)
	FindByDepartmentIDs(deptIDs []uuid.UUID) ([]*models.Employee, error)
	List(name, cpf, rg *string, deptID *uuid.UUID, page, pageSize int) ([]*models.Employee, error)
	IsCPFDuplicated(err error) bool
	IsRGDuplicated(err error) bool
}

type employeeRepository struct {
	db *gorm.DB
}

func NewEmployeeRepository(db *gorm.DB) EmployeeRepository {
	return &employeeRepository{db: db}
}

func (r *employeeRepository) Create(employee *models.Employee) error {
	return r.db.Create(employee).Error
}

func (r *employeeRepository) FindByID(id uuid.UUID) (*models.Employee, error) {
	var employee models.Employee
	if err := r.db.First(&employee, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &employee, nil
}

func (r *employeeRepository) FindAll() ([]models.Employee, error) {
	var employees []models.Employee
	if err := r.db.Find(&employees).Error; err != nil {
		return nil, err
	}
	return employees, nil
}

func (r *employeeRepository) Update(employee *models.Employee) error {
	return r.db.Save(employee).Error
}

func (r *employeeRepository) Delete(id uuid.UUID) error {
	return r.db.Where("id = ?", id).Delete(&models.Employee{}).Error
}

func (r *employeeRepository) CountByDepartmentID(deptID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.Model(&models.Employee{}).Where("department_id = ?", deptID).Count(&count).Error
	return count, err
}

func (r *employeeRepository) FindByDepartmentIDs(deptIDs []uuid.UUID) ([]*models.Employee, error) {
	var employees []*models.Employee
	err := r.db.Where("department_id IN ?", deptIDs).Find(&employees).Error
	return employees, err
}

func (r *employeeRepository) List(name, cpf, rg *string, deptID *uuid.UUID, page, pageSize int) ([]*models.Employee, error) {
	var employees []*models.Employee
	query := r.db.Model(&models.Employee{})

	if name != nil {
		query = query.Where("name ILIKE ?", "%"+*name+"%")
	}
	if cpf != nil {
		query = query.Where("cpf = ?", *cpf)
	}
	if rg != nil {
		query = query.Where("rg = ?", *rg)
	}
	if deptID != nil {
		query = query.Where("department_id = ?", *deptID)
	}

	offset := (page - 1) * pageSize
	err := query.Limit(pageSize).Offset(offset).Find(&employees).Error
	return employees, err
}

func (r *employeeRepository) IsCPFDuplicated(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "uq_cpf") || strings.Contains(err.Error(), "employees_cpf_key")
}

func (r *employeeRepository) IsRGDuplicated(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "uq_rg") || strings.Contains(err.Error(), "employees_rg_key")
}
