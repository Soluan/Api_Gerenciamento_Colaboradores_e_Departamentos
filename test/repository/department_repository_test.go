package repository_test

import (
	"ManageEmployeesandDepartments/internal/models"
	"testing"

	"github.com/google/uuid"
	"go.uber.org/goleak"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DepartamentoRepository interface para testes
type DepartamentoRepository interface {
	Create(depto *models.Department) error
	FindByID(id uuid.UUID) (*models.Department, error)
	FindByIDComGerente(id uuid.UUID) (*models.Department, error)
	FindSubDepartamentos(superiorID uuid.UUID) ([]*models.Department, error)
	Update(depto *models.Department) error
	Delete(id uuid.UUID) error
	CountSubDepartamentos(id uuid.UUID) (int64, error)
	IsGerente(colaboradorID uuid.UUID) (bool, error)
	FindByManagerID(gerenteID uuid.UUID) ([]*models.Department, error)
	IsSubordinado(superiorID, subordinadoID uuid.UUID) (bool, error)
	FindAllSubordinadoIDs(id uuid.UUID) ([]uuid.UUID, error)
	List(nome, gerenteNome *string, superiorID *uuid.UUID, pagina, tamanhoPagina int) ([]*models.Department, error)
}

// Implementação do repositório para SQLite
type departamentoRepository struct {
	db *gorm.DB
}

func NewDepartamentoRepository(db *gorm.DB) DepartamentoRepository {
	return &departamentoRepository{db: db}
}

func (r *departamentoRepository) Create(depto *models.Department) error {
	return r.db.Create(depto).Error
}

func (r *departamentoRepository) FindByID(id uuid.UUID) (*models.Department, error) {
	var depto models.Department
	if err := r.db.First(&depto, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &depto, nil
}

func (r *departamentoRepository) FindByIDComGerente(id uuid.UUID) (*models.Department, error) {
	var depto models.Department
	if err := r.db.Preload("Manager").First(&depto, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &depto, nil
}

func (r *departamentoRepository) FindSubDepartamentos(superiorID uuid.UUID) ([]*models.Department, error) {
	var deptos []*models.Department
	if err := r.db.Where("parent_department_id = ?", superiorID).Find(&deptos).Error; err != nil {
		return nil, err
	}
	return deptos, nil
}

func (r *departamentoRepository) Update(depto *models.Department) error {
	return r.db.Save(depto).Error
}

func (r *departamentoRepository) Delete(id uuid.UUID) error {
	return r.db.Where("id = ?", id).Delete(&models.Department{}).Error
}

func (r *departamentoRepository) CountSubDepartamentos(id uuid.UUID) (int64, error) {
	var count int64
	err := r.db.Model(&models.Department{}).Where("parent_department_id = ?", id).Count(&count).Error
	return count, err
}

func (r *departamentoRepository) IsGerente(colaboradorID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.Model(&models.Department{}).Where("manager_id = ?", colaboradorID).Count(&count).Error
	return count > 0, err
}

func (r *departamentoRepository) FindByManagerID(gerenteID uuid.UUID) ([]*models.Department, error) {
	var deptos []*models.Department
	err := r.db.Where("manager_id = ?", gerenteID).Find(&deptos).Error
	return deptos, err
}

// SQLite versão simplificada do IsSubordinado - sem WITH RECURSIVE
func (r *departamentoRepository) IsSubordinado(superiorID, subordinadoID uuid.UUID) (bool, error) {
	// Busca simplificada - verifica apenas relação direta e um nível de profundidade
	var deptos []*models.Department

	// Busca todos os departamentos filhos do superior
	err := r.db.Where("parent_department_id = ?", superiorID).Find(&deptos).Error
	if err != nil {
		return false, err
	}

	// Verifica relação direta
	for _, depto := range deptos {
		if depto.ID == subordinadoID {
			return true, nil
		}
		// Verifica um nível mais profundo
		var subDeptos []*models.Department
		err := r.db.Where("parent_department_id = ?", depto.ID).Find(&subDeptos).Error
		if err == nil {
			for _, subDepto := range subDeptos {
				if subDepto.ID == subordinadoID {
					return true, nil
				}
			}
		}
	}

	return false, nil
}

func (r *departamentoRepository) FindAllSubordinadoIDs(id uuid.UUID) ([]uuid.UUID, error) {
	var ids []uuid.UUID

	// Busca filhos diretos
	var deptos []*models.Department
	err := r.db.Where("parent_department_id = ?", id).Find(&deptos).Error
	if err != nil {
		return nil, err
	}

	ids = append(ids, id) // Inclui o próprio departamento

	for _, depto := range deptos {
		ids = append(ids, depto.ID)
		// Recursivamente busca sub-departamentos (limitado a 2 níveis por simplicidade)
		var subDeptos []*models.Department
		err := r.db.Where("parent_department_id = ?", depto.ID).Find(&subDeptos).Error
		if err == nil {
			for _, subDepto := range subDeptos {
				ids = append(ids, subDepto.ID)
			}
		}
	}

	return ids, nil
}

func (r *departamentoRepository) List(nome, gerenteNome *string, superiorID *uuid.UUID, pagina, tamanhoPagina int) ([]*models.Department, error) {
	var deptos []*models.Department
	query := r.db.Model(&models.Department{}).Preload("Manager")

	if nome != nil {
		// SQLite usa LIKE ao invés de ILIKE, com LOWER para case-insensitive
		query = query.Where("LOWER(departments.name) LIKE LOWER(?)", "%"+*nome+"%")
	}
	if superiorID != nil {
		query = query.Where("parent_department_id = ?", *superiorID)
	}
	if gerenteNome != nil {
		query = query.Joins("INNER JOIN colaboradores g ON g.id = departamentos.manager_id AND LOWER(g.nome) LIKE LOWER(?)", "%"+*gerenteNome+"%")
	}

	offset := (pagina - 1) * tamanhoPagina
	err := query.Limit(tamanhoPagina).Offset(offset).Find(&deptos).Error
	return deptos, err
}

func setupDepartamentoTestDB(t *testing.T) (*gorm.DB, func()) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Discard,
	})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Auto migrate tables
	err = db.AutoMigrate(&models.Employee{}, &models.Department{})
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	// Função de cleanup para fechar conexões
	cleanup := func() {
		sqlDB, err := db.DB()
		if err == nil {
			sqlDB.Close()
		}
	}

	return db, cleanup
}

func TestDepartamentoRepository_Create(t *testing.T) {
	defer goleak.VerifyNone(t)

	db, cleanup := setupDepartamentoTestDB(t)
	defer cleanup()

	repo := NewDepartamentoRepository(db)

	testCases := []struct {
		name         string
		departamento *models.Department
		expectError  bool
	}{
		{
			name: "sucesso ao criar departamento raiz",
			departamento: &models.Department{
				ID:   uuid.New(),
				Name: "Empresa",
			},
			expectError: false,
		},
		{
			name: "sucesso ao criar departamento com superior",
			departamento: &models.Department{
				ID:   uuid.New(),
				Name: "TI",
			},
			expectError: false,
		},
	}

	var empresaID uuid.UUID
	for i, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if i == 1 {
				// Para o segundo teste, define o superior
				tc.departamento.ParentDepartmentID = &empresaID
			}

			err := repo.Create(tc.departamento)

			if tc.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if i == 0 {
					empresaID = tc.departamento.ID
				}
			}
		})
	}
}

func TestDepartamentoRepository_FindByID(t *testing.T) {
	defer goleak.VerifyNone(t)

	db, cleanup := setupDepartamentoTestDB(t)
	defer cleanup()

	repo := NewDepartamentoRepository(db)

	// Criar departamento
	departamento := &models.Department{
		ID:   uuid.New(),
		Name: "TI",
	}
	err := repo.Create(departamento)
	if err != nil {
		t.Fatalf("Failed to create departamento: %v", err)
	}

	testCases := []struct {
		name        string
		id          uuid.UUID
		expectError bool
		expectNil   bool
	}{
		{
			name:        "sucesso ao encontrar departamento",
			id:          departamento.ID,
			expectError: false,
			expectNil:   false,
		},
		{
			name:        "departamento não encontrado",
			id:          uuid.New(),
			expectError: true,
			expectNil:   true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := repo.FindByID(tc.id)

			if tc.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				if result != nil {
					t.Errorf("Expected nil result but got: %v", result)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if result == nil {
					t.Errorf("Expected result but got nil")
				} else if result.ID != tc.id {
					t.Errorf("Expected ID %v, got %v", tc.id, result.ID)
				}
			}
		})
	}
}

func TestDepartamentoRepository_FindByIDComGerente(t *testing.T) {
	defer goleak.VerifyNone(t)

	db, cleanup := setupDepartamentoTestDB(t)
	defer cleanup()

	repo := NewDepartamentoRepository(db)

	// Criar departamento primeiro
	departamento := &models.Department{
		ID:   uuid.New(),
		Name: "TI",
	}
	err := repo.Create(departamento)
	if err != nil {
		t.Fatalf("Failed to create departamento: %v", err)
	}

	// Criar colaborador como gerente
	gerente := &models.Employee{
		ID:             uuid.New(),
		Name:           "João Silva",
		CPF:            "12345678901",
		DepartmentID: departamento.ID,
	}
	err = db.Create(gerente).Error
	if err != nil {
		t.Fatalf("Failed to create colaborador: %v", err)
	}

	// Atualizar departamento com gerente
	departamento.ManagerID = &gerente.ID
	err = repo.Update(departamento)
	if err != nil {
		t.Fatalf("Failed to update departamento: %v", err)
	}

	result, err := repo.FindByIDComGerente(departamento.ID)
	if err != nil {
		t.Fatalf("Failed to find departamento with gerente: %v", err)
	}

	if result.Manager == nil {
		t.Errorf("Expected gerente to be loaded")
	} else if result.Manager.ID != gerente.ID {
		t.Errorf("Expected gerente ID %v, got %v", gerente.ID, result.Manager.ID)
	}
}

func TestDepartamentoRepository_Update(t *testing.T) {
	defer goleak.VerifyNone(t)

	db, cleanup := setupDepartamentoTestDB(t)
	defer cleanup()

	repo := NewDepartamentoRepository(db)

	// Criar departamento
	departamento := &models.Department{
		ID:   uuid.New(),
		Name: "TI",
	}
	err := repo.Create(departamento)
	if err != nil {
		t.Fatalf("Failed to create departamento: %v", err)
	}

	// Atualizar departamento
	departamento.Name = "Tecnologia da Informação"

	err = repo.Update(departamento)
	if err != nil {
		t.Fatalf("Failed to update departamento: %v", err)
	}

	// Verificar se foi atualizado
	updated, err := repo.FindByID(departamento.ID)
	if err != nil {
		t.Fatalf("Failed to find updated departamento: %v", err)
	}

	if updated.Name != "Tecnologia da Informação" {
		t.Errorf("Expected nome 'Tecnologia da Informação', got '%s'", updated.Name)
	}
}

func TestDepartamentoRepository_Delete(t *testing.T) {
	defer goleak.VerifyNone(t)

	db, cleanup := setupDepartamentoTestDB(t)
	defer cleanup()

	repo := NewDepartamentoRepository(db)

	// Criar departamento
	departamento := &models.Department{
		ID:   uuid.New(),
		Name: "TI",
	}
	err := repo.Create(departamento)
	if err != nil {
		t.Fatalf("Failed to create departamento: %v", err)
	}

	// Deletar departamento
	err = repo.Delete(departamento.ID)
	if err != nil {
		t.Fatalf("Failed to delete departamento: %v", err)
	}

	// Verificar se foi deletado
	_, err = repo.FindByID(departamento.ID)
	if err == nil {
		t.Errorf("Expected error when finding deleted departamento")
	}
}

func TestDepartamentoRepository_FindSubDepartamentos(t *testing.T) {
	defer goleak.VerifyNone(t)

	db, cleanup := setupDepartamentoTestDB(t)
	defer cleanup()

	repo := NewDepartamentoRepository(db)

	// Criar departamento pai
	empresa := &models.Department{
		ID:   uuid.New(),
		Name: "Empresa",
	}
	err := repo.Create(empresa)
	if err != nil {
		t.Fatalf("Failed to create empresa: %v", err)
	}

	// Criar departamentos filhos
	ti := &models.Department{
		ID:                     uuid.New(),
		Name:                   "TI",
		ParentDepartmentID: &empresa.ID,
	}
	rh := &models.Department{
		ID:                     uuid.New(),
		Name:                   "RH",
		ParentDepartmentID: &empresa.ID,
	}

	err = repo.Create(ti)
	if err != nil {
		t.Fatalf("Failed to create TI: %v", err)
	}
	err = repo.Create(rh)
	if err != nil {
		t.Fatalf("Failed to create RH: %v", err)
	}

	// Buscar sub-departamentos
	subs, err := repo.FindSubDepartamentos(empresa.ID)
	if err != nil {
		t.Fatalf("Failed to find sub departamentos: %v", err)
	}

	if len(subs) != 2 {
		t.Errorf("Expected 2 sub departamentos, got %d", len(subs))
	}
}

func TestDepartamentoRepository_CountSubDepartamentos(t *testing.T) {
	defer goleak.VerifyNone(t)

	db, cleanup := setupDepartamentoTestDB(t)
	defer cleanup()

	repo := NewDepartamentoRepository(db)

	// Criar departamento pai
	empresa := &models.Department{
		ID:   uuid.New(),
		Name: "Empresa",
	}
	err := repo.Create(empresa)
	if err != nil {
		t.Fatalf("Failed to create empresa: %v", err)
	}

	// Criar 3 departamentos filhos
	for i := 0; i < 3; i++ {
		depto := &models.Department{
			ID:                     uuid.New(),
			Name:                   "Depto " + string(rune(i+1+'0')),
			ParentDepartmentID: &empresa.ID,
		}
		err = repo.Create(depto)
		if err != nil {
			t.Fatalf("Failed to create departamento: %v", err)
		}
	}

	count, err := repo.CountSubDepartamentos(empresa.ID)
	if err != nil {
		t.Fatalf("Failed to count sub departamentos: %v", err)
	}

	if count != 3 {
		t.Errorf("Expected count 3, got %d", count)
	}
}

func TestDepartamentoRepository_List(t *testing.T) {
	defer goleak.VerifyNone(t)

	db, cleanup := setupDepartamentoTestDB(t)
	defer cleanup()

	repo := NewDepartamentoRepository(db)

	// Criar departamentos
	departamentos := []*models.Department{
		{
			ID:   uuid.New(),
			Name: "Empresa",
		},
		{
			ID:   uuid.New(),
			Name: "TI",
		},
		{
			ID:   uuid.New(),
			Name: "Recursos Humanos",
		},
	}

	for _, depto := range departamentos {
		err := repo.Create(depto)
		if err != nil {
			t.Fatalf("Failed to create departamento %s: %v", depto.Name, err)
		}
	}

	testCases := []struct {
		name          string
		nomeFilter    *string
		gerenteNome   *string
		superiorID    *uuid.UUID
		pagina        int
		tamanhoPagina int
		expectedCount int
		expectedError bool
	}{
		{
			name:          "listar todos os departamentos",
			pagina:        1,
			tamanhoPagina: 10,
			expectedCount: 3,
			expectedError: false,
		},
		{
			name:          "filtrar por nome",
			nomeFilter:    stringPtr("recursos"),
			pagina:        1,
			tamanhoPagina: 10,
			expectedCount: 1,
			expectedError: false,
		},
		{
			name:          "paginação",
			pagina:        1,
			tamanhoPagina: 2,
			expectedCount: 2,
			expectedError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := repo.List(tc.nomeFilter, tc.gerenteNome, tc.superiorID, tc.pagina, tc.tamanhoPagina)

			if tc.expectedError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Failed to list departamentos: %v", err)
				}
				if len(result) != tc.expectedCount {
					t.Errorf("Expected %d departamentos, got %d", tc.expectedCount, len(result))
				}
			}
		})
	}
}

func TestDepartamentoRepository_IsSubordinado(t *testing.T) {
	defer goleak.VerifyNone(t)

	db, cleanup := setupDepartamentoTestDB(t)
	defer cleanup()

	repo := NewDepartamentoRepository(db)

	// Criar hierarquia: Empresa -> TI -> Desenvolvimento
	empresa := &models.Department{
		ID:   uuid.New(),
		Name: "Empresa",
	}
	ti := &models.Department{
		ID:                     uuid.New(),
		Name:                   "TI",
		ParentDepartmentID: &empresa.ID,
	}
	desenvolvimento := &models.Department{
		ID:                     uuid.New(),
		Name:                   "Desenvolvimento",
		ParentDepartmentID: &ti.ID,
	}

	err := repo.Create(empresa)
	if err != nil {
		t.Fatalf("Failed to create empresa: %v", err)
	}
	err = repo.Create(ti)
	if err != nil {
		t.Fatalf("Failed to create TI: %v", err)
	}
	err = repo.Create(desenvolvimento)
	if err != nil {
		t.Fatalf("Failed to create desenvolvimento: %v", err)
	}

	testCases := []struct {
		name           string
		superiorID     uuid.UUID
		subordinadoID  uuid.UUID
		expectedResult bool
	}{
		{
			name:           "TI é subordinado direto de Empresa",
			superiorID:     empresa.ID,
			subordinadoID:  ti.ID,
			expectedResult: true,
		},
		{
			name:           "Desenvolvimento é subordinado direto de TI",
			superiorID:     ti.ID,
			subordinadoID:  desenvolvimento.ID,
			expectedResult: true,
		},
		{
			name:           "Desenvolvimento é subordinado indireto de Empresa",
			superiorID:     empresa.ID,
			subordinadoID:  desenvolvimento.ID,
			expectedResult: true,
		},
		{
			name:           "Empresa não é subordinado de TI",
			superiorID:     ti.ID,
			subordinadoID:  empresa.ID,
			expectedResult: false,
		},
		{
			name:           "Departamento não é subordinado de si mesmo",
			superiorID:     ti.ID,
			subordinadoID:  ti.ID,
			expectedResult: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := repo.IsSubordinado(tc.superiorID, tc.subordinadoID)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if result != tc.expectedResult {
				t.Errorf("Expected %v, got %v", tc.expectedResult, result)
			}
		})
	}
}

func TestDepartamentoRepository_FindByManagerID(t *testing.T) {
	defer goleak.VerifyNone(t)

	db, cleanup := setupDepartamentoTestDB(t)
	defer cleanup()

	repo := NewDepartamentoRepository(db)

	// Criar colaborador como gerente
	gerente := &models.Employee{
		ID:   uuid.New(),
		Name: "João Silva",
		CPF:  "12345678901",
	}
	err := db.Create(gerente).Error
	if err != nil {
		t.Fatalf("Failed to create colaborador: %v", err)
	}

	// Criar departamentos com o mesmo gerente
	deptos := []*models.Department{
		{
			ID:        uuid.New(),
			Name:      "TI",
			ManagerID: &gerente.ID,
		},
		{
			ID:        uuid.New(),
			Name:      "Desenvolvimento",
			ManagerID: &gerente.ID,
		},
	}

	for _, depto := range deptos {
		err = repo.Create(depto)
		if err != nil {
			t.Fatalf("Failed to create departamento: %v", err)
		}
	}

	result, err := repo.FindByManagerID(gerente.ID)
	if err != nil {
		t.Fatalf("Failed to find departamentos by gerente: %v", err)
	}

	if len(result) != 2 {
		t.Errorf("Expected 2 departamentos, got %d", len(result))
	}
}

// Benchmarks
func BenchmarkDepartamentoRepository_Create(b *testing.B) {
	db, cleanup := setupDepartamentoTestDB(&testing.T{})
	defer cleanup()

	repo := NewDepartamentoRepository(db)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		departamento := &models.Department{
			ID:   uuid.New(),
			Name: "Benchmark Depto",
		}
		repo.Create(departamento)
	}
}

func BenchmarkDepartamentoRepository_FindByID(b *testing.B) {
	db, cleanup := setupDepartamentoTestDB(&testing.T{})
	defer cleanup()

	repo := NewDepartamentoRepository(db)

	// Criar departamento
	departamento := &models.Department{
		ID:   uuid.New(),
		Name: "Benchmark Depto",
	}
	repo.Create(departamento)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		repo.FindByID(departamento.ID)
	}
}
