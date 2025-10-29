package repository

import (
	"ManageEmployeesandDepartments/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DepartamentoRepository interface {
	Create(depto *models.Departamento) error
	FindByID(id uuid.UUID) (*models.Departamento, error)
	FindByIDComGerente(id uuid.UUID) (*models.Departamento, error)
	FindSubDepartamentos(superiorID uuid.UUID) ([]*models.Departamento, error)
	Update(depto *models.Departamento) error
	Delete(id uuid.UUID) error
	CountSubDepartamentos(id uuid.UUID) (int64, error)
	IsGerente(colaboradorID uuid.UUID) (bool, error)
	FindByGerenteID(gerenteID uuid.UUID) ([]*models.Departamento, error)
	IsSubordinado(superiorID, subordinadoID uuid.UUID) (bool, error)
	FindAllSubordinadoIDs(id uuid.UUID) ([]uuid.UUID, error)
	List(nome, gerenteNome *string, superiorID *uuid.UUID, pagina, tamanhoPagina int) ([]*models.Departamento, error)
}

type departamentoRepository struct {
	db *gorm.DB
}

// Construtor que retorna a interface
func NewDepartamentoRepository(db *gorm.DB) DepartamentoRepository {
	return &departamentoRepository{db: db}
}

// Implementação dos métodos
func (r *departamentoRepository) Create(depto *models.Departamento) error {
	return r.db.Create(depto).Error
}

func (r *departamentoRepository) FindByID(id uuid.UUID) (*models.Departamento, error) {
	var depto models.Departamento
	if err := r.db.First(&depto, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &depto, nil
}

func (r *departamentoRepository) FindByIDComGerente(id uuid.UUID) (*models.Departamento, error) {
	var depto models.Departamento
	if err := r.db.Preload("Gerente").First(&depto, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &depto, nil
}

func (r *departamentoRepository) FindSubDepartamentos(superiorID uuid.UUID) ([]*models.Departamento, error) {
	var deptos []*models.Departamento
	if err := r.db.Where("departamento_superior_id = ?", superiorID).Find(&deptos).Error; err != nil {
		return nil, err
	}
	return deptos, nil
}

func (r *departamentoRepository) Update(depto *models.Departamento) error {
	return r.db.Save(depto).Error
}

func (r *departamentoRepository) Delete(id uuid.UUID) error {
	return r.db.Where("id = ?", id).Delete(&models.Departamento{}).Error
}

func (r *departamentoRepository) CountSubDepartamentos(id uuid.UUID) (int64, error) {
	var count int64
	err := r.db.Model(&models.Departamento{}).Where("departamento_superior_id = ?", id).Count(&count).Error
	return count, err
}

func (r *departamentoRepository) IsGerente(colaboradorID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.Model(&models.Departamento{}).Where("gerente_id = ?", colaboradorID).Count(&count).Error
	return count > 0, err
}

func (r *departamentoRepository) FindByGerenteID(gerenteID uuid.UUID) ([]*models.Departamento, error) {
	var deptos []*models.Departamento
	err := r.db.Where("gerente_id = ?", gerenteID).Find(&deptos).Error
	return deptos, err
}

func (r *departamentoRepository) IsSubordinado(superiorID, subordinadoID uuid.UUID) (bool, error) {
	var ids []uuid.UUID
	cte := `
		WITH RECURSIVE sub_deptos AS (
			SELECT id FROM departamentos WHERE id = ?
			UNION ALL
			SELECT d.id FROM departamentos d
			INNER JOIN sub_deptos sd ON d.departamento_superior_id = sd.id
		)
		SELECT id FROM sub_deptos WHERE id != ?;
	`
	err := r.db.Raw(cte, superiorID, superiorID).Scan(&ids).Error
	if err != nil {
		return false, err
	}

	for _, id := range ids {
		if id == subordinadoID {
			return true, nil
		}
	}

	return false, nil
}

func (r *departamentoRepository) FindAllSubordinadoIDs(id uuid.UUID) ([]uuid.UUID, error) {
	var ids []uuid.UUID
	cte := `
		WITH RECURSIVE sub_deptos AS (
			SELECT id FROM departamentos WHERE id = ?
			UNION ALL
			SELECT d.id FROM departamentos d
			INNER JOIN sub_deptos sd ON d.departamento_superior_id = sd.id
		)
		SELECT id FROM sub_deptos;
	`
	err := r.db.Raw(cte, id).Scan(&ids).Error
	return ids, err
}

func (r *departamentoRepository) List(nome, gerenteNome *string, superiorID *uuid.UUID, pagina, tamanhoPagina int) ([]*models.Departamento, error) {
	var deptos []*models.Departamento
	query := r.db.Model(&models.Departamento{}).Preload("Gerente")

	if nome != nil {
		query = query.Where("departamentos.nome ILIKE ?", "%"+*nome+"%")
	}
	if superiorID != nil {
		query = query.Where("departamento_superior_id = ?", *superiorID)
	}
	if gerenteNome != nil {
		query = query.Joins("INNER JOIN colaboradores g ON g.id = departamentos.gerente_id AND g.nome ILIKE ?", "%"+*gerenteNome+"%")
	}

	offset := (pagina - 1) * tamanhoPagina
	err := query.Limit(tamanhoPagina).Offset(offset).Find(&deptos).Error
	return deptos, err
}
