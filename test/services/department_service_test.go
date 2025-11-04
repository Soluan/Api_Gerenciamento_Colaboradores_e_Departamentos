package services_test

import (
	"ManageEmployeesandDepartments/internal/models"
	"ManageEmployeesandDepartments/internal/services"
	"ManageEmployeesandDepartments/internal/utils"
	"testing"

	"github.com/google/uuid"
	"go.uber.org/goleak"
	"gorm.io/gorm"
)

func TestDepartmentService_CreateDepartment(t *testing.T) {
	defer goleak.VerifyNone(t)

	managerID := uuid.New()
	parentID := uuid.New()

	testCases := []struct {
		name           string
		departmentName string
		managerID      uuid.UUID
		parentID       *uuid.UUID
		mockSetup      func(*MockDepartmentRepository, *MockEmployeeRepository)
		expectedError  error
	}{
		{
			name:           "success creating department",
			departmentName: "IT",
			managerID:      managerID,
			parentID:       nil,
			mockSetup: func(deptRepo *MockDepartmentRepository, employeeRepo *MockEmployeeRepository) {
				// Manager exists
				manager := &models.Employee{
					ID:   managerID,
					Name: "João Manager",
				}
				employeeRepo.findByIDResult = manager
				employeeRepo.findByIDError = nil
				// Create department works
				deptRepo.createError = nil
			},
			expectedError: nil,
		},
		{
			name:           "error manager not found",
			departmentName: "IT",
			managerID:      managerID,
			parentID:       nil,
			mockSetup: func(deptRepo *MockDepartmentRepository, employeeRepo *MockEmployeeRepository) {
				// Manager doesn't exist
				employeeRepo.findByIDResult = nil
				employeeRepo.findByIDError = gorm.ErrRecordNotFound
			},
			expectedError: utils.ErrManagerNotFound,
		},
		{
			name:           "error parent department not found",
			departmentName: "IT",
			managerID:      managerID,
			parentID:       &parentID,
			mockSetup: func(deptRepo *MockDepartmentRepository, employeeRepo *MockEmployeeRepository) {
				// Manager exists
				manager := &models.Employee{
					ID:   managerID,
					Name: "João Manager",
				}
				employeeRepo.findByIDResult = manager
				employeeRepo.findByIDError = nil
				// Parent department doesn't exist
				deptRepo.findByIDResult = nil
				deptRepo.findByIDError = gorm.ErrRecordNotFound
			},
			expectedError: utils.ErrParentDepartmentNotFound,
		},
		{
			name:           "success creating department with updated manager",
			departmentName: "IT",
			managerID:      managerID,
			parentID:       &parentID,
			mockSetup: func(deptRepo *MockDepartmentRepository, employeeRepo *MockEmployeeRepository) {
				// Manager exists
				manager := &models.Employee{
					ID:           managerID,
					Name:         "João Manager",
					DepartmentID: uuid.New(), // Different department - will be updated
				}
				employeeRepo.findByIDResult = manager
				employeeRepo.findByIDError = nil
				employeeRepo.updateError = nil
				// Parent department exists
				department := &models.Department{
					ID:   parentID,
					Name: "Headquarters",
				}
				deptRepo.findByIDResult = department
				deptRepo.findByIDError = nil
				// Create department works
				deptRepo.createError = nil
			},
			expectedError: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup
			deptRepo := &MockDepartmentRepository{}
			employeeRepo := &MockEmployeeRepository{}
			tc.mockSetup(deptRepo, employeeRepo)

			service := services.NewDepartmentService(deptRepo, employeeRepo)

			// Execute
			result, err := service.CreateDepartment(tc.departmentName, tc.managerID, tc.parentID)

			// Validate
			if tc.expectedError != nil {
				if err != tc.expectedError {
					t.Errorf("Expected error %v, got %v", tc.expectedError, err)
				}
				if result != nil {
					t.Errorf("Expected nil result on error, got %v", result)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}
				if result == nil {
					t.Errorf("Expected result, got nil")
				}
			}
		})
	}
}

func TestDepartmentService_GetDepartmentWithTree(t *testing.T) {
	defer goleak.VerifyNone(t)

	departmentID := uuid.New()

	testCases := []struct {
		name          string
		id            uuid.UUID
		mockSetup     func(*MockDepartmentRepository, *MockEmployeeRepository)
		expectedError error
	}{
		{
			name: "success getting department with tree",
			id:   departmentID,
			mockSetup: func(deptRepo *MockDepartmentRepository, employeeRepo *MockEmployeeRepository) {
				// Department exists
				department := &models.Department{
					ID:   departmentID,
					Name: "IT",
				}
				deptRepo.findByIDWithManagerResult = department
				deptRepo.findByIDWithManagerError = nil

				// Empty sub-departments to avoid infinite recursion
				deptRepo.findSubDepartmentsResult = []*models.Department{}
				deptRepo.findSubDepartmentsError = nil
			},
			expectedError: nil,
		},
		{
			name: "department not found",
			id:   departmentID,
			mockSetup: func(deptRepo *MockDepartmentRepository, employeeRepo *MockEmployeeRepository) {
				deptRepo.findByIDWithManagerResult = nil
				deptRepo.findByIDWithManagerError = gorm.ErrRecordNotFound
			},
			expectedError: gorm.ErrRecordNotFound,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup
			deptRepo := &MockDepartmentRepository{}
			employeeRepo := &MockEmployeeRepository{}
			tc.mockSetup(deptRepo, employeeRepo)

			service := services.NewDepartmentService(deptRepo, employeeRepo)

			// Execute
			result, err := service.GetDepartmentWithTree(tc.id)

			// Validate
			if tc.expectedError != nil {
				if err != tc.expectedError {
					t.Errorf("Expected error %v, got %v", tc.expectedError, err)
				}
				if result != nil {
					t.Errorf("Expected nil result on error, got %v", result)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}
				if result == nil {
					t.Errorf("Expected result, got nil")
				}
			}
		})
	}
}

func TestDepartmentService_DeleteDepartment(t *testing.T) {
	defer goleak.VerifyNone(t)

	departamentoID := uuid.New()

	testCases := []struct {
		name          string
		id            uuid.UUID
		mockSetup     func(*MockDepartmentRepository, *MockEmployeeRepository)
		expectedError error
	}{
		{
			name: "sucesso ao deletar departamento",
			id:   departamentoID,
			mockSetup: func(deptoRepo *MockDepartmentRepository, colabRepo *MockEmployeeRepository) {
				// Departamento existe
				departamento := &models.Department{
					ID:   departamentoID,
					Name: "TI",
				}
				deptoRepo.findByIDResult = departamento
				deptoRepo.findByIDError = nil
				// Não possui colaboradores
				colabRepo.countByDepartmentIDResult = 0
				colabRepo.countByDepartmentIDError = nil
				// Não possui subdepartamentos
				deptoRepo.countSubDepartmentsResult = 0
				deptoRepo.countSubDepartmentsError = nil
				// Delete funciona
				deptoRepo.deleteError = nil
			},
			expectedError: nil,
		},
		{
			name: "erro ao contar colaboradores",
			id:   departamentoID,
			mockSetup: func(deptoRepo *MockDepartmentRepository, colabRepo *MockEmployeeRepository) {
				colabRepo.countByDepartmentIDResult = 0
				colabRepo.countByDepartmentIDError = gorm.ErrRecordNotFound
			},
			expectedError: gorm.ErrRecordNotFound,
		},
		{
			name: "departamento possui colaboradores",
			id:   departamentoID,
			mockSetup: func(deptoRepo *MockDepartmentRepository, colabRepo *MockEmployeeRepository) {
				// Departamento existe
				departamento := &models.Department{
					ID:   departamentoID,
					Name: "TI",
				}
				deptoRepo.findByIDResult = departamento
				deptoRepo.findByIDError = nil
				// Possui colaboradores
				colabRepo.countByDepartmentIDResult = 5
				colabRepo.countByDepartmentIDError = nil
			},
			expectedError: utils.ErrDepartmentHasEmployees,
		},
		{
			name: "departamento possui subdepartamentos",
			id:   departamentoID,
			mockSetup: func(deptoRepo *MockDepartmentRepository, colabRepo *MockEmployeeRepository) {
				// Departamento existe
				departamento := &models.Department{
					ID:   departamentoID,
					Name: "TI",
				}
				deptoRepo.findByIDResult = departamento
				deptoRepo.findByIDError = nil
				// Não possui colaboradores
				colabRepo.countByDepartmentIDResult = 0
				colabRepo.countByDepartmentIDError = nil
				// Possui subdepartamentos
				deptoRepo.countSubDepartmentsResult = 2
				deptoRepo.countSubDepartmentsError = nil
			},
			expectedError: utils.ErrDepartmentHasSubDepartments,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup
			deptoRepo := &MockDepartmentRepository{}
			colabRepo := &MockEmployeeRepository{}
			tc.mockSetup(deptoRepo, colabRepo)

			service := services.NewDepartmentService(deptoRepo, colabRepo)

			// Executar
			err := service.DeleteDepartment(tc.id)

			// Validar
			if tc.expectedError != nil {
				if err != tc.expectedError {
					t.Errorf("Expected error %v, got %v", tc.expectedError, err)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}
			}
		})
	}
}

func TestDepartmentService_GetSubordinateEmployeesRecursively(t *testing.T) {
	defer goleak.VerifyNone(t)

	gerenteID := uuid.New()

	testCases := []struct {
		name          string
		gerenteID     uuid.UUID
		mockSetup     func(*MockDepartmentRepository, *MockEmployeeRepository)
		expectedError error
	}{
		{
			name:      "sucesso ao buscar subordinados",
			gerenteID: gerenteID,
			mockSetup: func(deptoRepo *MockDepartmentRepository, colabRepo *MockEmployeeRepository) {
				// Gerente existe
				gerente := &models.Employee{
					ID:   gerenteID,
					Name: "João Gerente",
				}
				colabRepo.findByIDResult = gerente
				colabRepo.findByIDError = nil

				// Departamentos do gerente
				departamentos := []*models.Department{
					{ID: uuid.New(), Name: "TI"},
				}
				deptoRepo.findByManagerIDResult = departamentos
				deptoRepo.findByManagerIDError = nil

				// IDs subordinados
				subordinadoIDs := []uuid.UUID{uuid.New(), uuid.New()}
				deptoRepo.findAllSubordinateIDsResult = subordinadoIDs
				deptoRepo.findAllSubordinateIDsError = nil

				// Colaboradores subordinados
				colaboradores := []*models.Employee{
					{ID: uuid.New(), Name: "João Silva"},
					{ID: uuid.New(), Name: "Maria Santos"},
				}
				colabRepo.findByDepartmentIDsResult = colaboradores
				colabRepo.findByDepartmentIDsError = nil
			},
			expectedError: nil,
		},
		{
			name:      "gerente não encontrado",
			gerenteID: gerenteID,
			mockSetup: func(deptoRepo *MockDepartmentRepository, colabRepo *MockEmployeeRepository) {
				colabRepo.findByIDResult = nil
				colabRepo.findByIDError = gorm.ErrRecordNotFound
			},
			expectedError: utils.ErrManagerNotFound,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup
			deptoRepo := &MockDepartmentRepository{}
			colabRepo := &MockEmployeeRepository{}
			tc.mockSetup(deptoRepo, colabRepo)

			service := services.NewDepartmentService(deptoRepo, colabRepo)

			// Executar
			result, err := service.GetSubordinateEmployeesRecursively(tc.gerenteID)

			// Validar
			if tc.expectedError != nil {
				if err != tc.expectedError {
					t.Errorf("Expected error %v, got %v", tc.expectedError, err)
				}
				if result != nil {
					t.Errorf("Expected nil result on error, got %v", result)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}
				if result == nil {
					t.Errorf("Expected result, got nil")
				}
			}
		})
	}
}

// Benchmark para teste de performance
func BenchmarkDepartamentoService_CreateDepartment(b *testing.B) {
	gerenteID := uuid.New()

	deptoRepo := &MockDepartmentRepository{
		createError: nil,
	}
	colabRepo := &MockEmployeeRepository{
		findByIDResult: &models.Employee{
			ID:   gerenteID,
			Name: "João Gerente",
		},
		findByIDError: nil,
	}

	service := services.NewDepartmentService(deptoRepo, colabRepo)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.CreateDepartment("TI", gerenteID, nil)
	}
}

func TestDepartmentService_UpdateDepartment(t *testing.T) {
	defer goleak.VerifyNone(t)

	departamentoID := uuid.New()
	gerenteID := uuid.New()
	novoGerenteID := uuid.New()
	superiorID := uuid.New()

	testCases := []struct {
		name           string
		id             uuid.UUID
		departmentName *string
		managerID      *uuid.UUID
		parentID       *uuid.UUID
		mockSetup      func(*MockDepartmentRepository, *MockEmployeeRepository)
		expectedError  error
	}{
		{
			name:           "sucesso ao atualizar departamento",
			id:             departamentoID,
			departmentName: stringPtr("TI Atualizado"),
			managerID:      &novoGerenteID,
			parentID:       &superiorID,
			mockSetup: func(deptoRepo *MockDepartmentRepository, colabRepo *MockEmployeeRepository) {
				// Departamento existe
				departamento := &models.Department{
					ID:        departamentoID,
					Name:      "TI",
					ManagerID: &gerenteID,
				}
				deptoRepo.findByIDResult = departamento
				deptoRepo.findByIDError = nil

				// Novo gerente existe
				novoGerente := &models.Employee{
					ID:           novoGerenteID,
					Name:         "Novo Gerente",
					DepartmentID: departamentoID,
				}
				colabRepo.findByIDResult = novoGerente
				colabRepo.findByIDError = nil
				colabRepo.updateError = nil

				// Verificação de ciclo
				deptoRepo.isSubordinateResult = false
				deptoRepo.isSubordinateError = nil

				// Update funciona
				deptoRepo.updateError = nil
			},
			expectedError: nil,
		},
		{
			name:           "departamento não encontrado",
			id:             departamentoID,
			departmentName: stringPtr("TI"),
			mockSetup: func(deptoRepo *MockDepartmentRepository, colabRepo *MockEmployeeRepository) {
				deptoRepo.findByIDResult = nil
				deptoRepo.findByIDError = gorm.ErrRecordNotFound
			},
			expectedError: gorm.ErrRecordNotFound,
		},
		{
			name:     "ciclo detectado",
			id:       departamentoID,
			parentID: &departamentoID, // Próprio ID como superior
			mockSetup: func(deptoRepo *MockDepartmentRepository, colabRepo *MockEmployeeRepository) {
				departamento := &models.Department{
					ID:   departamentoID,
					Name: "TI",
				}
				deptoRepo.findByIDResult = departamento
				deptoRepo.findByIDError = nil
			},
			expectedError: utils.ErrCycleDetected,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup
			deptoRepo := &MockDepartmentRepository{}
			colabRepo := &MockEmployeeRepository{}
			tc.mockSetup(deptoRepo, colabRepo)

			service := services.NewDepartmentService(deptoRepo, colabRepo)

			// Execute
			result, err := service.UpdateDepartment(tc.id, tc.departmentName, tc.managerID, tc.parentID)

			// Validar
			if tc.expectedError != nil {
				if err != tc.expectedError {
					t.Errorf("Expected error %v, got %v", tc.expectedError, err)
				}
				if result != nil {
					t.Errorf("Expected nil result on error, got %v", result)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}
				if result == nil {
					t.Errorf("Expected result, got nil")
				}
			}
		})
	}
}

func TestDepartmentService_ListDepartments(t *testing.T) {
	defer goleak.VerifyNone(t)

	testCases := []struct {
		name           string
		departmentName *string
		managerName    *string
		parentID       *uuid.UUID
		page           int
		pageSize       int
		mockSetup      func(*MockDepartmentRepository, *MockEmployeeRepository)
		expectedError  error
	}{
		{
			name:           "sucesso ao listar departamentos",
			departmentName: stringPtr("TI"),
			managerName:    stringPtr("João"),
			page:           1,
			pageSize:       10,
			mockSetup: func(deptoRepo *MockDepartmentRepository, colabRepo *MockEmployeeRepository) {
				departamentos := []*models.Department{
					{ID: uuid.New(), Name: "TI"},
					{ID: uuid.New(), Name: "TI Desenvolvimento"},
				}
				deptoRepo.listResult = departamentos
				deptoRepo.listError = nil
			},
			expectedError: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup
			deptoRepo := &MockDepartmentRepository{}
			colabRepo := &MockEmployeeRepository{}
			tc.mockSetup(deptoRepo, colabRepo)

			service := services.NewDepartmentService(deptoRepo, colabRepo)

			// Execute
			result, err := service.ListDepartments(tc.departmentName, tc.managerName, tc.parentID, tc.page, tc.pageSize)

			// Validate
			if tc.expectedError != nil {
				if err != tc.expectedError {
					t.Errorf("Expected error %v, got %v", tc.expectedError, err)
				}
				if result != nil {
					t.Errorf("Expected nil result on error, got %v", result)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}
				if result == nil {
					t.Errorf("Expected result, got nil")
				}
			}
		})
	}
}
