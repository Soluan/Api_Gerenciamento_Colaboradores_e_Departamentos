package services

import (
	"ManageEmployeesandDepartments/internal/models"
	"ManageEmployeesandDepartments/internal/repository"
	"ManageEmployeesandDepartments/internal/utils"

	"github.com/google/uuid"
)

type ColaboradorService interface {
	CreateColaborador(nome string, cpf string, rg *string, departamentoID uuid.UUID) (*models.Colaborador, error)
	GetColaboradorComGerente(id uuid.UUID) (*ColaboradorComGerenteResponse, error)
	UpdateColaborador(id uuid.UUID, nome *string, rg *string, departamentoID uuid.UUID) (*models.Colaborador, error)
	DeleteColaborador(id uuid.UUID) error
	ListColaboradores(nome *string, cpf *string, rg *string, deptoID *uuid.UUID, pagina, tamanhoPagina int) ([]*models.Colaborador, error)
}

type colaboradorService struct {
	deptoRepo repository.DepartamentoRepository
	colabRepo repository.ColaboradorRepository
}

func NewColaboradorService(deptoRepo repository.DepartamentoRepository, colabRepo repository.ColaboradorRepository) ColaboradorService {
	return &colaboradorService{
		deptoRepo: deptoRepo,
		colabRepo: colabRepo,
	}
}

type ColaboradorComGerenteResponse struct {
	Colaborador *models.Colaborador `json:"colaborador"`
	GerenteNome string              `json:"gerente_nome,omitempty"`
}

// CreateColaborador cria um novo colaborador com validação de CPF/RG e departamento
func (s *colaboradorService) CreateColaborador(nome string, cpf string, rg *string, departamentoID uuid.UUID) (*models.Colaborador, error) {
	// Verifica se departamento existe
	_, err := s.deptoRepo.FindByID(departamentoID)
	if err != nil {
		return nil, utils.ErrDepartamentoNaoEncontrado
	}

	// Cria o colaborador
	colab := &models.Colaborador{
		ID:             uuid.New(),
		Nome:           nome,
		CPF:            cpf,
		RG:             rg,
		DepartamentoID: departamentoID,
	}

	err = s.colabRepo.Create(colab)
	if s.colabRepo.IsCPFDuplicado(err) {
		return nil, utils.ErrCPFDuplicado
	}
	if s.colabRepo.IsRGDuplicado(err) {
		return nil, utils.ErrRGDuplicado
	}
	if err != nil {
		return nil, err
	}

	return colab, nil
}

// GetColaboradorComGerente retorna um colaborador e o nome do gerente do departamento
func (s *colaboradorService) GetColaboradorComGerente(id uuid.UUID) (*ColaboradorComGerenteResponse, error) {
	colab, err := s.colabRepo.FindByID(id)
	if err != nil {
		return nil, err
	}

	depto, err := s.deptoRepo.FindByID(colab.DepartamentoID)
	if err != nil {
		return nil, err
	}

	var gerenteNome string
	if *depto.GerenteID != uuid.Nil {
		gerente, err := s.colabRepo.FindByID(*depto.GerenteID)
		if err == nil {
			gerenteNome = gerente.Nome
		}
	}

	return &ColaboradorComGerenteResponse{
		Colaborador: colab,
		GerenteNome: gerenteNome,
	}, nil
}

// UpdateColaborador atualiza nome, RG e departamento de um colaborador
func (s *colaboradorService) UpdateColaborador(id uuid.UUID, nome *string, rg *string, departamentoID uuid.UUID) (*models.Colaborador, error) {
	colab, err := s.colabRepo.FindByID(id)
	if err != nil {
		return nil, err
	}

	// Valida departamento
	_, err = s.deptoRepo.FindByID(departamentoID)
	if err != nil {
		return nil, utils.ErrDepartamentoNaoEncontrado
	}

	colab.Nome = *nome
	colab.RG = rg
	colab.DepartamentoID = departamentoID

	err = s.colabRepo.Update(colab)
	if s.colabRepo.IsRGDuplicado(err) {
		return nil, utils.ErrRGDuplicado
	}
	if err != nil {
		return nil, err
	}

	return colab, nil
}

// DeleteColaborador remove um colaborador (soft delete)
func (s *colaboradorService) DeleteColaborador(id uuid.UUID) error {
	colab, err := s.colabRepo.FindByID(id)
	if err != nil {
		return err
	}

	// Não permite deletar se for gerente de algum departamento
	isGerente, err := s.deptoRepo.IsGerente(id)
	if err != nil {
		return err
	}
	if isGerente {
		return utils.ErrGerenteNaoPodeSerExcluido
	}

	return s.colabRepo.Delete(colab.ID)
}

// ListColaboradores lista colaboradores com filtros e paginação
func (s *colaboradorService) ListColaboradores(nome, cpf, rg *string, deptoID *uuid.UUID, pagina, tamanhoPagina int) ([]*models.Colaborador, error) {
	return s.colabRepo.List(nome, cpf, rg, deptoID, pagina, tamanhoPagina)
}
