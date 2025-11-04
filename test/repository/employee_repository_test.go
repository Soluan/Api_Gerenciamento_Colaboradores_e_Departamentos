package repository_test

import (
	"ManageEmployeesandDepartments/internal/models"
	"strings"
	"testing"

	"github.com/google/uuid"
	"go.uber.org/goleak"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// ColaboradorRepository interface para testes
type ColaboradorRepository interface {
	Create(colab *models.Employee) error
	FindByID(id uuid.UUID) (*models.Employee, error)
	FindAll() ([]models.Employee, error)
	Update(colab *models.Employee) error
	Delete(id uuid.UUID) error
	CountByDepartamentoID(deptoID uuid.UUID) (int64, error)
	FindByDepartamentoIDs(deptoIDs []uuid.UUID) ([]*models.Employee, error)
	List(nome, cpf, rg *string, deptoID *uuid.UUID, pagina, tamanhoPagina int) ([]*models.Employee, error)
	IsCPFDuplicado(err error) bool
	IsRGDuplicado(err error) bool
}

// Implementação do repositório para SQLite
type employeeRepository struct {
	db *gorm.DB
}

func NewColaboradorRepository(db *gorm.DB) ColaboradorRepository {
	return &employeeRepository{db: db}
}

func (r *employeeRepository) Create(colab *models.Employee) error {
	return r.db.Create(colab).Error
}

func (r *employeeRepository) FindByID(id uuid.UUID) (*models.Employee, error) {
	var colab models.Employee
	if err := r.db.First(&colab, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &colab, nil
}

func (r *employeeRepository) FindAll() ([]models.Employee, error) {
	var colabs []models.Employee
	if err := r.db.Find(&colabs).Error; err != nil {
		return nil, err
	}
	return colabs, nil
}

func (r *employeeRepository) Update(colab *models.Employee) error {
	return r.db.Save(colab).Error
}

func (r *employeeRepository) Delete(id uuid.UUID) error {
	return r.db.Where("id = ?", id).Delete(&models.Employee{}).Error
}

func (r *employeeRepository) CountByDepartamentoID(deptoID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.Model(&models.Employee{}).Where("department_id = ?", deptoID).Count(&count).Error
	return count, err
}

func (r *employeeRepository) FindByDepartamentoIDs(deptoIDs []uuid.UUID) ([]*models.Employee, error) {
	var colabs []*models.Employee
	err := r.db.Where("department_id IN ?", deptoIDs).Find(&colabs).Error
	return colabs, err
}

func (r *employeeRepository) List(nome, cpf, rg *string, deptoID *uuid.UUID, pagina, tamanhoPagina int) ([]*models.Employee, error) {
	var colabs []*models.Employee
	query := r.db.Model(&models.Employee{})

	if nome != nil {
		// SQLite usa LIKE ao invés de ILIKE, com LOWER para case-insensitive
		query = query.Where("LOWER(name) LIKE LOWER(?)", "%"+*nome+"%")
	}
	if cpf != nil {
		query = query.Where("cpf = ?", *cpf)
	}
	if rg != nil {
		query = query.Where("rg = ?", *rg)
	}
	if deptoID != nil {
		query = query.Where("department_id = ?", *deptoID)
	}

	offset := (pagina - 1) * tamanhoPagina
	err := query.Limit(tamanhoPagina).Offset(offset).Find(&colabs).Error
	return colabs, err
}

func (r *employeeRepository) IsCPFDuplicado(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "uq_cpf") || strings.Contains(err.Error(), "employeees_cpf_key") ||
		strings.Contains(err.Error(), "UNIQUE constraint failed")
}

func (r *employeeRepository) IsRGDuplicado(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "uq_rg") || strings.Contains(err.Error(), "employeees_rg_key") ||
		strings.Contains(err.Error(), "UNIQUE constraint failed")
}

func setupTestDB(t *testing.T) (*gorm.DB, func()) {
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

func TestColaboradorRepository_Create(t *testing.T) {
	defer goleak.VerifyNone(t)

	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewColaboradorRepository(db)

	// Criar departamento primeiro
	departamento := &models.Department{
		ID:   uuid.New(),
		Name: "TI",
	}
	err := db.Create(departamento).Error
	if err != nil {
		t.Fatalf("Failed to create department: %v", err)
	}

	testCases := []struct {
		name          string
		employee   *models.Employee
		expectedError bool
	}{
		{
			name: "sucesso ao criar employee",
			employee: &models.Employee{
				ID:             uuid.New(),
				Name:           "João Silva",
				CPF:            "12345678901",
				RG:             stringPtr("MG1234567"),
				DepartmentID: departamento.ID,
			},
			expectedError: false,
		},
		{
			name: "sucesso ao criar employee sem RG",
			employee: &models.Employee{
				ID:             uuid.New(),
				Name:           "Maria Santos",
				CPF:            "98765432109",
				RG:             nil,
				DepartmentID: departamento.ID,
			},
			expectedError: false,
		},
		{
			name: "erro CPF duplicado",
			employee: &models.Employee{
				ID:             uuid.New(),
				Name:           "Pedro Oliveira",
				CPF:            "12345678901", // CPF já usado
				RG:             stringPtr("SP9876543"),
				DepartmentID: departamento.ID,
			},
			expectedError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := repo.Create(tc.employee)

			if tc.expectedError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}

				// Verificar se foi criado no banco
				var found models.Employee
				err = db.First(&found, "id = ?", tc.employee.ID).Error
				if err != nil {
					t.Errorf("Failed to find created employee: %v", err)
				}

				if found.Name != tc.employee.Name {
					t.Errorf("Expected name %s, got %s", tc.employee.Name, found.Name)
				}
			}
		})
	}
}

func TestColaboradorRepository_FindByID(t *testing.T) {
	defer goleak.VerifyNone(t)

	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewColaboradorRepository(db)

	// Criar departamento primeiro
	departamento := &models.Department{
		ID:   uuid.New(),
		Name: "TI",
	}
	err := db.Create(departamento).Error
	if err != nil {
		t.Fatalf("Failed to create department: %v", err)
	}

	// Criar employee para teste
	employee := &models.Employee{
		ID:             uuid.New(),
		Name:           "João Silva",
		CPF:            "12345678901",
		RG:             stringPtr("MG1234567"),
		DepartmentID: departamento.ID,
	}
	err = db.Create(employee).Error
	if err != nil {
		t.Fatalf("Failed to create employee: %v", err)
	}

	testCases := []struct {
		name          string
		id            uuid.UUID
		expectedError bool
		expectedName  string
	}{
		{
			name:          "sucesso ao encontrar employee",
			id:            employee.ID,
			expectedError: false,
			expectedName:  "João Silva",
		},
		{
			name:          "employee não encontrado",
			id:            uuid.New(),
			expectedError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := repo.FindByID(tc.id)

			if tc.expectedError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				if result != nil {
					t.Error("Expected nil result on error")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}
				if result == nil {
					t.Error("Expected result, got nil")
				} else if result.Name != tc.expectedName {
					t.Errorf("Expected name %s, got %s", tc.expectedName, result.Name)
				}
			}
		})
	}
}

func TestColaboradorRepository_Update(t *testing.T) {
	defer goleak.VerifyNone(t)

	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewColaboradorRepository(db)

	// Criar departamento primeiro
	departamento := &models.Department{
		ID:   uuid.New(),
		Name: "TI",
	}
	err := db.Create(departamento).Error
	if err != nil {
		t.Fatalf("Failed to create department: %v", err)
	}

	// Criar employee para teste
	employee := &models.Employee{
		ID:             uuid.New(),
		Name:           "João Silva",
		CPF:            "12345678901",
		RG:             stringPtr("MG1234567"),
		DepartmentID: departamento.ID,
	}
	err = db.Create(employee).Error
	if err != nil {
		t.Fatalf("Failed to create employee: %v", err)
	}

	// Atualizar nome
	employee.Name = "João Silva Santos"
	err = repo.Update(employee)
	if err != nil {
		t.Errorf("Failed to update employee: %v", err)
	}

	// Verificar se foi atualizado
	var updated models.Employee
	err = db.First(&updated, "id = ?", employee.ID).Error
	if err != nil {
		t.Fatalf("Failed to find updated employee: %v", err)
	}

	if updated.Name != "João Silva Santos" {
		t.Errorf("Expected updated name 'João Silva Santos', got %s", updated.Name)
	}
}

func TestColaboradorRepository_Delete(t *testing.T) {
	defer goleak.VerifyNone(t)

	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewColaboradorRepository(db)

	// Criar departamento primeiro
	departamento := &models.Department{
		ID:   uuid.New(),
		Name: "TI",
	}
	err := db.Create(departamento).Error
	if err != nil {
		t.Fatalf("Failed to create department: %v", err)
	}

	// Criar employee para teste
	employee := &models.Employee{
		ID:             uuid.New(),
		Name:           "João Silva",
		CPF:            "12345678901",
		RG:             stringPtr("MG1234567"),
		DepartmentID: departamento.ID,
	}
	err = db.Create(employee).Error
	if err != nil {
		t.Fatalf("Failed to create employee: %v", err)
	}

	// Deletar employee (soft delete)
	err = repo.Delete(employee.ID)
	if err != nil {
		t.Errorf("Failed to delete employee: %v", err)
	}

	// Verificar se foi soft deleted
	var deleted models.Employee
	err = db.Unscoped().First(&deleted, "id = ?", employee.ID).Error
	if err != nil {
		t.Fatalf("Failed to find deleted employee: %v", err)
	}

	if deleted.DeletedAt.Time.IsZero() {
		t.Error("Expected employee to be soft deleted, but DeletedAt is zero")
	}

	// Verificar que não aparece em busca normal
	err = db.First(&deleted, "id = ?", employee.ID).Error
	if err == nil {
		t.Error("Expected employee to not be found after soft delete")
	}
}

func TestColaboradorRepository_List(t *testing.T) {
	defer goleak.VerifyNone(t)

	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewColaboradorRepository(db)

	// Criar departamento primeiro
	departamento := &models.Department{
		ID:   uuid.New(),
		Name: "TI",
	}
	err := db.Create(departamento).Error
	if err != nil {
		t.Fatalf("Failed to create department: %v", err)
	}

	// Criar employeees para teste
	employeees := []*models.Employee{
		{
			ID:             uuid.New(),
			Name:           "João Silva",
			CPF:            "12345678901",
			RG:             stringPtr("MG1234567"),
			DepartmentID: departamento.ID,
		},
		{
			ID:             uuid.New(),
			Name:           "Maria Santos",
			CPF:            "98765432109",
			RG:             stringPtr("SP9876543"),
			DepartmentID: departamento.ID,
		},
	}

	for _, colab := range employeees {
		err = db.Create(colab).Error
		if err != nil {
			t.Fatalf("Failed to create employee: %v", err)
		}
	}

	testCases := []struct {
		name          string
		nomeFilter    *string
		cpfFilter     *string
		rgFilter      *string
		deptoIDFilter *uuid.UUID
		pagina        int
		tamanhoPagina int
		expectedCount int
	}{
		{
			name:          "listar todos os employeees",
			pagina:        1,
			tamanhoPagina: 10,
			expectedCount: 2,
		},
		{
			name:          "filtrar por nome",
			nomeFilter:    stringPtr("João"),
			pagina:        1,
			tamanhoPagina: 10,
			expectedCount: 1,
		},
		{
			name:          "filtrar por departamento",
			deptoIDFilter: &departamento.ID,
			pagina:        1,
			tamanhoPagina: 10,
			expectedCount: 2,
		},
		{
			name:          "paginação",
			pagina:        1,
			tamanhoPagina: 1,
			expectedCount: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := repo.List(tc.nomeFilter, tc.cpfFilter, tc.rgFilter, tc.deptoIDFilter, tc.pagina, tc.tamanhoPagina)
			if err != nil {
				t.Errorf("Failed to list employeees: %v", err)
			}

			if len(result) != tc.expectedCount {
				t.Errorf("Expected %d employeees, got %d", tc.expectedCount, len(result))
			}
		})
	}
}

func TestColaboradorRepository_CountByDepartamentoID(t *testing.T) {
	defer goleak.VerifyNone(t)

	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewColaboradorRepository(db)

	// Criar departamento primeiro
	departamento := &models.Department{
		ID:   uuid.New(),
		Name: "TI",
	}
	err := db.Create(departamento).Error
	if err != nil {
		t.Fatalf("Failed to create department: %v", err)
	}

	// Criar employeees
	for i := 0; i < 3; i++ {
		employee := &models.Employee{
			ID:             uuid.New(),
			Name:           "Colaborador " + string(rune(i+1)),
			CPF:            "1234567890" + string(rune(i+1)),
			DepartmentID: departamento.ID,
		}
		err = db.Create(employee).Error
		if err != nil {
			t.Fatalf("Failed to create employee: %v", err)
		}
	}

	count, err := repo.CountByDepartamentoID(departamento.ID)
	if err != nil {
		t.Errorf("Failed to count employeees: %v", err)
	}

	if count != 3 {
		t.Errorf("Expected 3 employeees, got %d", count)
	}

	// Testar departamento inexistente
	count, err = repo.CountByDepartamentoID(uuid.New())
	if err != nil {
		t.Errorf("Failed to count employeees for non-existent department: %v", err)
	}

	if count != 0 {
		t.Errorf("Expected 0 employeees for non-existent department, got %d", count)
	}
}

// Benchmark para teste de performance
func BenchmarkColaboradorRepository_Create(b *testing.B) {
	db, cleanup := setupTestDB(&testing.T{})
	defer cleanup()

	repo := NewColaboradorRepository(db)

	// Criar departamento primeiro
	departamento := &models.Department{
		ID:   uuid.New(),
		Name: "TI",
	}
	db.Create(departamento)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		employee := &models.Employee{
			ID:             uuid.New(),
			Name:           "Test User",
			CPF:            "12345678901",
			DepartmentID: departamento.ID,
		}
		repo.Create(employee)
	}
}

func BenchmarkColaboradorRepository_FindByID(b *testing.B) {
	db, cleanup := setupTestDB(&testing.T{})
	defer cleanup()

	repo := NewColaboradorRepository(db)

	// Criar departamento primeiro
	departamento := &models.Department{
		ID:   uuid.New(),
		Name: "TI",
	}
	db.Create(departamento)

	// Criar employee para busca
	employee := &models.Employee{
		ID:             uuid.New(),
		Name:           "Test User",
		CPF:            "12345678901",
		DepartmentID: departamento.ID,
	}
	db.Create(employee)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		repo.FindByID(employee.ID)
	}
}

// Função auxiliar
func stringPtr(s string) *string {
	return &s
}
