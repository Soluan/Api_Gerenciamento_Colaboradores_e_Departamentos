package repository

import (
	"ManageEmployeesandDepartments/internal/models"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Interface pública para o repositório de colaboradores
type ColaboradorRepository interface {
	Create(colab *models.Colaborador) error
	FindByID(id uuid.UUID) (*models.Colaborador, error)
	FindAll() ([]models.Colaborador, error)
	Update(colab *models.Colaborador) error
	Delete(id uuid.UUID) error
	CountByDepartamentoID(deptoID uuid.UUID) (int64, error)
	FindByDepartamentoIDs(deptoIDs []uuid.UUID) ([]*models.Colaborador, error)
	List(nome, cpf, rg *string, deptoID *uuid.UUID, pagina, tamanhoPagina int) ([]*models.Colaborador, error)
	IsCPFDuplicado(err error) bool
	IsRGDuplicado(err error) bool
}

type colaboradorRepository struct {
	db *gorm.DB
}

func NewColaboradorRepository(db *gorm.DB) ColaboradorRepository {
	return &colaboradorRepository{db: db}
}

func (r *colaboradorRepository) Create(colab *models.Colaborador) error {
	return r.db.Create(colab).Error
}

func (r *colaboradorRepository) FindByID(id uuid.UUID) (*models.Colaborador, error) {
	var colab models.Colaborador
	if err := r.db.First(&colab, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &colab, nil
}

func (r *colaboradorRepository) FindAll() ([]models.Colaborador, error) {
	var colabs []models.Colaborador
	if err := r.db.Find(&colabs).Error; err != nil {
		return nil, err
	}
	return colabs, nil
}

func (r *colaboradorRepository) Update(colab *models.Colaborador) error {
	return r.db.Save(colab).Error
}

func (r *colaboradorRepository) Delete(id uuid.UUID) error {
	return r.db.Where("id = ?", id).Delete(&models.Colaborador{}).Error
}

func (r *colaboradorRepository) CountByDepartamentoID(deptoID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.Model(&models.Colaborador{}).Where("departamento_id = ?", deptoID).Count(&count).Error
	return count, err
}

func (r *colaboradorRepository) FindByDepartamentoIDs(deptoIDs []uuid.UUID) ([]*models.Colaborador, error) {
	var colabs []*models.Colaborador
	err := r.db.Where("departamento_id IN ?", deptoIDs).Find(&colabs).Error
	return colabs, err
}

func (r *colaboradorRepository) List(nome, cpf, rg *string, deptoID *uuid.UUID, pagina, tamanhoPagina int) ([]*models.Colaborador, error) {
	var colabs []*models.Colaborador
	query := r.db.Model(&models.Colaborador{})

	if nome != nil {
		query = query.Where("nome ILIKE ?", "%"+*nome+"%")
	}
	if cpf != nil {
		query = query.Where("cpf = ?", *cpf)
	}
	if rg != nil {
		query = query.Where("rg = ?", *rg)
	}
	if deptoID != nil {
		query = query.Where("departamento_id = ?", *deptoID)
	}

	offset := (pagina - 1) * tamanhoPagina
	err := query.Limit(tamanhoPagina).Offset(offset).Find(&colabs).Error
	return colabs, err
}

func (r *colaboradorRepository) IsCPFDuplicado(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "uq_cpf") || strings.Contains(err.Error(), "colaboradores_cpf_key")
}

func (r *colaboradorRepository) IsRGDuplicado(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "uq_rg") || strings.Contains(err.Error(), "colaboradores_rg_key")
}
