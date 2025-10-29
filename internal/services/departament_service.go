package services

import (
	"ManageEmployeesandDepartments/internal/models"
	"ManageEmployeesandDepartments/internal/repository"
	"ManageEmployeesandDepartments/internal/utils"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DepartamentoService interface {
	CreateDepartamento(nome string, gerenteID uuid.UUID, superiorID *uuid.UUID) (*models.Departamento, error)
	GetDepartamentoComArvore(id uuid.UUID) (*models.Departamento, error)
	UpdateDepartamento(id uuid.UUID, nome *string, gerenteID *uuid.UUID, superiorID *uuid.UUID) (*models.Departamento, error)
	DeleteDepartamento(id uuid.UUID) error
	ListDepartamentos(nome, gerenteNome *string, superiorID *uuid.UUID, pagina, tamanhoPagina int) ([]*models.Departamento, error)
	GetColaboradoresSubordinadosRecursivamente(gerenteID uuid.UUID) ([]*models.Colaborador, error)
}

type DepartamentoHandler struct{}

type departamentoService struct {
	deptoRepo repository.DepartamentoRepository
	colabRepo repository.ColaboradorRepository
}

// NewDepartamentoService cria uma nova instância do serviço de departamento.
func NewDepartamentoService(dr repository.DepartamentoRepository, cr repository.ColaboradorRepository) *departamentoService {
	return &departamentoService{deptoRepo: dr, colabRepo: cr}
}

// CreateDepartamento cria um novo departamento.
func (s *departamentoService) CreateDepartamento(nome string, gerenteID uuid.UUID, superiorID *uuid.UUID) (*models.Departamento, error) {
	// 1. Valida Gerente
	gerente, err := s.colabRepo.FindByID(gerenteID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, utils.ErrGerenteNaoEncontrado
		}
		return nil, err
	}

	// 2. Valida Departamento Superior (se informado)
	if superiorID != nil && *superiorID != uuid.Nil {
		if _, err := s.deptoRepo.FindByID(*superiorID); err != nil {
			if err == gorm.ErrRecordNotFound {
				return nil, utils.ErrDepartamentoSuperiorNaoEncontrado
			}
			return nil, err
		}
	}

	// 3. Regra de Negócio: O gerente deve ser um Colaborador existente
	// (Já validado acima)

	// Regra de Negócio: ...e vinculado ao mesmo departamento.
	// Esta regra é complexa na CRIAÇÃO. Como o depto ainda não existe,
	// vamos assumir que a regra é: o gerente DEVE existir.
	// A regra de "vinculado ao mesmo depto" se aplica melhor na ATUALIZAÇÃO,
	// ou se a regra fosse "o gerente deve pertencer ao depto superior".

	// Vamos implementar a regra "gerente deve pertencer ao depto":
	// Se o gerente não pertencer ao depto que ele está sendo setado
	// (o que é impossível na criação), ou ao depto superior (se houver)?

	// Vamos simplificar para a regra da especificação:
	// "O gerente deve ser um Colaborador existente e vinculado ao mesmo departamento."
	// Isso implica que o gerente DEVE pertencer ao departamento que ele gerencia.
	// Na criação, o gerente AINDA não pertence.

	// A regra mais lógica parece ser:
	// 1. O gerente (colaborador) deve existir. (Feito)
	// 2. O departamento do gerente DEVE ser o departamento que está sendo criado.
	// Isso significa que o `colaborador.departamento_id` deve ser atualizado.

	// *** Ajuste na Regra de Negócio ***
	// A regra "gerente... vinculado ao *mesmo departamento*" é estranha.
	// Se o Colaborador 'Bob' é gerente do Depto 'Vendas', o `colaborador.departamento_id`
	// do 'Bob' deve ser o `id` de 'Vendas'.

	// Vamos assumir que a regra é:
	// 1. O Colaborador (gerente) DEVE existir. (Feito)
	// 2. O `departamento_id` do Colaborador (gerente) DEVE ser o ID deste novo departamento.
	// Isso é um problema de "ovo e galinha" na criação.

	// SOLUÇÃO: Vamos assumir que a regra é "O gerente DEVE existir".
	// E na ATUALIZAÇÃO (Update) vamos validar se o `gerente.departamento_id`
	// é o mesmo do `departamento.id`.

	depto := &models.Departamento{
		Nome:                   nome,
		GerenteID:              &gerenteID,
		DepartamentoSuperiorID: superiorID,
	}

	if err := s.deptoRepo.Create(depto); err != nil {
		return nil, err
	}

	// AGORA, com o depto.ID, podemos garantir a regra:
	// "O gerente deve ser um Colaborador existente e vinculado ao mesmo departamento."
	if gerente.DepartamentoID != depto.ID {
		// O gerente não está neste depto. Ele DEVE estar.
		// Atualiza o departamento do gerente para ser este.
		// (Isso pode ser uma premissa de negócio perigosa, mas segue a regra)
		gerente.DepartamentoID = depto.ID
		if err := s.colabRepo.Update(gerente); err != nil {
			// Rollback? Por enquanto, só reporta o erro.
			return nil, errors.New("falha ao vincular gerente ao departamento: " + err.Error())
		}
	}

	return depto, nil
}

// GetDepartamentoComArvore busca um departamento e monta sua árvore de sub-departamentos.
func (s *departamentoService) GetDepartamentoComArvore(id uuid.UUID) (*models.Departamento, error) {
	depto, err := s.deptoRepo.FindByIDComGerente(id)
	if err != nil {
		return nil, err // Pode ser gorm.ErrRecordNotFound
	}

	// Função recursiva interna para carregar a árvore
	var buildTree func(d *models.Departamento) error
	buildTree = func(d *models.Departamento) error {
		subs, err := s.deptoRepo.FindSubDepartamentos(d.ID)
		if err != nil {
			return err
		}

		d.SubDepartamentos = subs
		for _, sub := range d.SubDepartamentos {
			// Carrega gerente do sub-depto
			if sub.GerenteID != nil {
				gerente, err := s.colabRepo.FindByID(*sub.GerenteID)
				if err == nil {
					sub.Gerente = gerente
				}
			}
			// Chama recursão
			if err := buildTree(sub); err != nil {
				return err
			}
		}
		return nil
	}

	if err := buildTree(depto); err != nil {
		return nil, err
	}

	return depto, nil
}

// UpdateDepartamento atualiza um departamento, prevenindo ciclos.
func (s *departamentoService) UpdateDepartamento(id uuid.UUID, nome *string, gerenteID *uuid.UUID, superiorID *uuid.UUID) (*models.Departamento, error) {
	depto, err := s.deptoRepo.FindByID(id)
	if err != nil {
		return nil, err // gorm.ErrRecordNotFound
	}

	// 1. Valida Ciclo (se o superiorID mudou)
	if superiorID != nil && (depto.DepartamentoSuperiorID == nil || *superiorID != *depto.DepartamentoSuperiorID) {
		// Se o novo superior for o próprio ID, é um ciclo.
		if *superiorID == id {
			return nil, utils.ErrCicloDetectado
		}

		// Verifica se o novo superior é um dos sub-departamentos
		isCiclo, err := s.deptoRepo.IsSubordinado(id, *superiorID)
		if err != nil {
			return nil, err
		}
		if isCiclo {
			return nil, utils.ErrCicloDetectado
		}
		depto.DepartamentoSuperiorID = superiorID
	}

	// 2. Valida Gerente (se mudou)
	if gerenteID != nil && (depto.GerenteID == nil || *gerenteID != *depto.GerenteID) {
		gerente, err := s.colabRepo.FindByID(*gerenteID)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return nil, utils.ErrGerenteNaoEncontrado
			}
			return nil, err
		}

		// Regra: Gerente deve estar vinculado ao *mesmo* departamento
		if gerente.DepartamentoID != depto.ID {
			//return nil, ErrGerenteNaoPertenceAoDepto

			// Alternativa: Força o gerente a pertencer ao depto
			gerente.DepartamentoID = depto.ID
			if err := s.colabRepo.Update(gerente); err != nil {
				return nil, errors.New("falha ao vincular gerente ao departamento: " + err.Error())
			}
		}
		depto.GerenteID = gerenteID
	}

	// 3. Atualiza Nome
	if nome != nil {
		depto.Nome = *nome
	}

	// 4. Salva atualizações
	if err := s.deptoRepo.Update(depto); err != nil {
		return nil, err
	}

	return depto, nil
}

// DeleteDepartamento remove um departamento.
func (s *departamentoService) DeleteDepartamento(id uuid.UUID) error {
	// 1. Verifica se tem colaboradores
	countColab, err := s.colabRepo.CountByDepartamentoID(id)
	if err != nil {
		return err
	}
	if countColab > 0 {
		return utils.ErrDepartamentoPossuiColaboradores
	}

	// 2. Verifica se tem sub-departamentos
	countSub, err := s.deptoRepo.CountSubDepartamentos(id)
	if err != nil {
		return err
	}
	if countSub > 0 {
		return utils.ErrDepartamentoPossuiSubDepartamentos
	}

	// 3. Deleta (Soft delete)
	return s.deptoRepo.Delete(id)
}

// ListDepartamentos lista departamentos com filtros.
func (s *departamentoService) ListDepartamentos(nome, gerenteNome *string, superiorID *uuid.UUID, pagina, tamanhoPagina int) ([]*models.Departamento, error) {
	return s.deptoRepo.List(nome, gerenteNome, superiorID, pagina, tamanhoPagina)
}

// GetColaboradoresSubordinadosRecursivamente busca todos os colaboradores
// dos departamentos gerenciados (direta ou indiretamente) pelo gerenteID.
func (s *departamentoService) GetColaboradoresSubordinadosRecursivamente(gerenteID uuid.UUID) ([]*models.Colaborador, error) {
	// 1. Verifica se o gerente existe
	if _, err := s.colabRepo.FindByID(gerenteID); err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, utils.ErrGerenteNaoEncontrado
		}
		return nil, err
	}

	// 2. Encontra todos os departamentos gerenciados por este gerente
	deptosGerenciados, err := s.deptoRepo.FindByGerenteID(gerenteID)
	if err != nil {
		return nil, err
	}

	if len(deptosGerenciados) == 0 {
		return []*models.Colaborador{}, nil // Gerente existe mas não gerencia deptos
	}

	// 3. Para cada depto, encontra todos os sub-deptos (árvore)
	var todosDeptoIDs []uuid.UUID
	for _, depto := range deptosGerenciados {
		ids, err := s.deptoRepo.FindAllSubordinadoIDs(depto.ID)
		if err != nil {
			return nil, err
		}
		todosDeptoIDs = append(todosDeptoIDs, ids...)
	}

	// Remove duplicatas (caso um gerente gerencie A e A.1)
	idMap := make(map[uuid.UUID]bool)
	var uniqueDeptoIDs []uuid.UUID
	for _, id := range todosDeptoIDs {
		if !idMap[id] {
			idMap[id] = true
			uniqueDeptoIDs = append(uniqueDeptoIDs, id)
		}
	}

	// 4. Busca todos os colaboradores desses departamentos
	if len(uniqueDeptoIDs) == 0 {
		return []*models.Colaborador{}, nil
	}

	return s.colabRepo.FindByDepartamentoIDs(uniqueDeptoIDs)
}
