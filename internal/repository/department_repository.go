package repository

import (
	"ManageEmployeesandDepartments/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DepartmentRepository interface {
	Create(dept *models.Department) error
	FindByID(id uuid.UUID) (*models.Department, error)
	FindByIDWithManager(id uuid.UUID) (*models.Department, error)
	FindSubDepartments(parentID uuid.UUID) ([]*models.Department, error)
	Update(dept *models.Department) error
	Delete(id uuid.UUID) error
	CountSubDepartments(id uuid.UUID) (int64, error)
	IsManager(employeeID uuid.UUID) (bool, error)
	FindByManagerID(managerID uuid.UUID) ([]*models.Department, error)
	IsSubordinate(parentID, subordinateID uuid.UUID) (bool, error)
	FindAllSubordinateIDs(id uuid.UUID) ([]uuid.UUID, error)
	List(name, managerName *string, parentID *uuid.UUID, page, pageSize int) ([]*models.Department, error)
}

type departmentRepository struct {
	db *gorm.DB
}

func NewDepartmentRepository(db *gorm.DB) DepartmentRepository {
	return &departmentRepository{db: db}
}

func (r *departmentRepository) Create(dept *models.Department) error {
	return r.db.Create(dept).Error
}

func (r *departmentRepository) FindByID(id uuid.UUID) (*models.Department, error) {
	var dept models.Department
	if err := r.db.First(&dept, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &dept, nil
}

func (r *departmentRepository) FindByIDWithManager(id uuid.UUID) (*models.Department, error) {
	var dept models.Department
	if err := r.db.Preload("Manager").First(&dept, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &dept, nil
}

func (r *departmentRepository) FindSubDepartments(parentID uuid.UUID) ([]*models.Department, error) {
	var departments []*models.Department
	err := r.db.Where("parent_department_id = ?", parentID).Find(&departments).Error
	return departments, err
}

func (r *departmentRepository) Update(dept *models.Department) error {
	return r.db.Save(dept).Error
}

func (r *departmentRepository) Delete(id uuid.UUID) error {
	return r.db.Where("id = ?", id).Delete(&models.Department{}).Error
}

func (r *departmentRepository) CountSubDepartments(id uuid.UUID) (int64, error) {
	var count int64
	err := r.db.Model(&models.Department{}).Where("parent_department_id = ?", id).Count(&count).Error
	return count, err
}

func (r *departmentRepository) IsManager(employeeID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.Model(&models.Department{}).Where("manager_id = ?", employeeID).Count(&count).Error
	return count > 0, err
}

func (r *departmentRepository) FindByManagerID(managerID uuid.UUID) ([]*models.Department, error) {
	var departments []*models.Department
	err := r.db.Where("manager_id = ?", managerID).Find(&departments).Error
	return departments, err
}

func (r *departmentRepository) IsSubordinate(parentID, subordinateID uuid.UUID) (bool, error) {
	// Simplified implementation for now
	return false, nil
}

func (r *departmentRepository) FindAllSubordinateIDs(id uuid.UUID) ([]uuid.UUID, error) {
	// Simplified implementation for now
	return []uuid.UUID{}, nil
}

func (r *departmentRepository) List(name, managerName *string, parentID *uuid.UUID, page, pageSize int) ([]*models.Department, error) {
	var departments []*models.Department
	query := r.db.Model(&models.Department{})

	if name != nil {
		query = query.Where("name ILIKE ?", "%"+*name+"%")
	}
	if parentID != nil {
		query = query.Where("parent_department_id = ?", *parentID)
	}

	offset := (page - 1) * pageSize
	err := query.Limit(pageSize).Offset(offset).Find(&departments).Error
	return departments, err
}
