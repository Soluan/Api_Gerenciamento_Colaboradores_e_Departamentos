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

// MockEmployeeRepository simulates the employee repository
type MockEmployeeRepository struct {
	createError               error
	findByIDResult            *models.Employee
	findByIDError             error
	findByIDResults           map[uuid.UUID]*models.Employee
	findByIDErrors            map[uuid.UUID]error
	findAllResult             []models.Employee
	findAllError              error
	updateError               error
	deleteError               error
	countByDepartmentIDResult int64
	countByDepartmentIDError  error
	findByDepartmentIDsResult []*models.Employee
	findByDepartmentIDsError  error
	listResult                []*models.Employee
	listError                 error
	isCPFDuplicatedResult     bool
	isRGDuplicatedResult      bool
}

func (m *MockEmployeeRepository) Create(employee *models.Employee) error {
	return m.createError
}

func (m *MockEmployeeRepository) FindByID(id uuid.UUID) (*models.Employee, error) {
	// If we have a results map, use it first
	if m.findByIDResults != nil {
		if result, exists := m.findByIDResults[id]; exists {
			if m.findByIDErrors != nil {
				return result, m.findByIDErrors[id]
			}
			return result, nil
		}
	}
	return m.findByIDResult, m.findByIDError
}

func (m *MockEmployeeRepository) FindAll() ([]models.Employee, error) {
	return m.findAllResult, m.findAllError
}

func (m *MockEmployeeRepository) Update(employee *models.Employee) error {
	return m.updateError
}

func (m *MockEmployeeRepository) Delete(id uuid.UUID) error {
	return m.deleteError
}

func (m *MockEmployeeRepository) CountByDepartmentID(deptID uuid.UUID) (int64, error) {
	return m.countByDepartmentIDResult, m.countByDepartmentIDError
}

func (m *MockEmployeeRepository) FindByDepartmentIDs(deptIDs []uuid.UUID) ([]*models.Employee, error) {
	return m.findByDepartmentIDsResult, m.findByDepartmentIDsError
}

func (m *MockEmployeeRepository) List(name, cpf, rg *string, deptID *uuid.UUID, page, pageSize int) ([]*models.Employee, error) {
	return m.listResult, m.listError
}

func (m *MockEmployeeRepository) IsCPFDuplicated(err error) bool {
	return m.isCPFDuplicatedResult
}

func (m *MockEmployeeRepository) IsRGDuplicated(err error) bool {
	return m.isRGDuplicatedResult
}

// MockDepartmentRepository implements repository.DepartmentRepository for tests
type MockDepartmentRepository struct {
	findByIDResult              *models.Department
	findByIDError               error
	findByIDWithManagerResult   *models.Department
	findByIDWithManagerError    error
	createError                 error
	updateError                 error
	deleteError                 error
	countSubDepartmentsResult   int64
	countSubDepartmentsError    error
	findSubDepartmentsResult    []*models.Department
	findSubDepartmentsError     error
	findByManagerIDResult       []*models.Department
	findByManagerIDError        error
	findAllSubordinateIDsResult []uuid.UUID
	findAllSubordinateIDsError  error
	listResult                  []*models.Department
	listError                   error
	isSubordinateResult         bool
	isSubordinateError          error
	isManagerResult             bool
	isManagerError              error
}

func (m *MockDepartmentRepository) Create(dept *models.Department) error {
	return m.createError
}

func (m *MockDepartmentRepository) FindByID(id uuid.UUID) (*models.Department, error) {
	return m.findByIDResult, m.findByIDError
}

func (m *MockDepartmentRepository) FindByIDWithManager(id uuid.UUID) (*models.Department, error) {
	return m.findByIDWithManagerResult, m.findByIDWithManagerError
}

func (m *MockDepartmentRepository) FindSubDepartments(parentID uuid.UUID) ([]*models.Department, error) {
	return m.findSubDepartmentsResult, m.findSubDepartmentsError
}

func (m *MockDepartmentRepository) Update(dept *models.Department) error {
	return m.updateError
}

func (m *MockDepartmentRepository) Delete(id uuid.UUID) error {
	return m.deleteError
}

func (m *MockDepartmentRepository) CountSubDepartments(id uuid.UUID) (int64, error) {
	return m.countSubDepartmentsResult, m.countSubDepartmentsError
}

func (m *MockDepartmentRepository) IsManager(employeeID uuid.UUID) (bool, error) {
	return m.isManagerResult, m.isManagerError
}

func (m *MockDepartmentRepository) FindByManagerID(managerID uuid.UUID) ([]*models.Department, error) {
	return m.findByManagerIDResult, m.findByManagerIDError
}

func (m *MockDepartmentRepository) IsSubordinate(parentID, subordinateID uuid.UUID) (bool, error) {
	return m.isSubordinateResult, m.isSubordinateError
}

func (m *MockDepartmentRepository) FindAllSubordinateIDs(id uuid.UUID) ([]uuid.UUID, error) {
	return m.findAllSubordinateIDsResult, m.findAllSubordinateIDsError
}

func (m *MockDepartmentRepository) List(name, managerName *string, parentID *uuid.UUID, page, pageSize int) ([]*models.Department, error) {
	return m.listResult, m.listError
}

// Helper function to create string pointers
func stringPtr(s string) *string {
	return &s
}

func TestEmployeeService_CreateEmployee(t *testing.T) {
	defer goleak.VerifyNone(t)

	testCases := []struct {
		name          string
		employeName   string
		cpf           string
		rg            *string
		departmentID  uuid.UUID
		mockSetup     func(*MockDepartmentRepository, *MockEmployeeRepository)
		expectedError error
	}{
		{
			name:         "success creating employee",
			employeName:  "João Silva",
			cpf:          "12345678901",
			rg:           stringPtr("123456789"),
			departmentID: uuid.New(),
			mockSetup: func(deptRepo *MockDepartmentRepository, employeeRepo *MockEmployeeRepository) {
				// Department exists
				deptRepo.findByIDResult = &models.Department{ID: uuid.New(), Name: "IT"}
				deptRepo.findByIDError = nil
				// Create employee works
				employeeRepo.createError = nil
			},
			expectedError: nil,
		},
		{
			name:         "error department not found",
			employeName:  "João Silva",
			cpf:          "12345678901",
			rg:           stringPtr("123456789"),
			departmentID: uuid.New(),
			mockSetup: func(deptRepo *MockDepartmentRepository, employeeRepo *MockEmployeeRepository) {
				// Department doesn't exist
				deptRepo.findByIDResult = nil
				deptRepo.findByIDError = gorm.ErrRecordNotFound
			},
			expectedError: utils.ErrDepartmentNotFound,
		},
		{
			name:         "error duplicate CPF",
			employeName:  "João Silva",
			cpf:          "12345678901",
			rg:           stringPtr("123456789"),
			departmentID: uuid.New(),
			mockSetup: func(deptRepo *MockDepartmentRepository, employeeRepo *MockEmployeeRepository) {
				// Department exists
				deptRepo.findByIDResult = &models.Department{ID: uuid.New(), Name: "IT"}
				deptRepo.findByIDError = nil
				// Duplicate CPF error
				employeeRepo.createError = gorm.ErrDuplicatedKey
				employeeRepo.isCPFDuplicatedResult = true
			},
			expectedError: utils.ErrCPFDuplicated,
		},
		{
			name:         "error duplicate RG",
			employeName:  "João Silva",
			cpf:          "12345678901",
			rg:           stringPtr("123456789"),
			departmentID: uuid.New(),
			mockSetup: func(deptRepo *MockDepartmentRepository, employeeRepo *MockEmployeeRepository) {
				// Department exists
				deptRepo.findByIDResult = &models.Department{ID: uuid.New(), Name: "IT"}
				deptRepo.findByIDError = nil
				// Duplicate RG error
				employeeRepo.createError = gorm.ErrDuplicatedKey
				employeeRepo.isCPFDuplicatedResult = false
				employeeRepo.isRGDuplicatedResult = true
			},
			expectedError: utils.ErrRGDuplicated,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup
			deptRepo := &MockDepartmentRepository{}
			employeeRepo := &MockEmployeeRepository{}
			tc.mockSetup(deptRepo, employeeRepo)

			service := services.NewEmployeeService(deptRepo, employeeRepo)

			// Execute
			result, err := service.CreateEmployee(tc.employeName, tc.cpf, tc.rg, tc.departmentID)

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

func TestEmployeeService_GetEmployeeWithManager(t *testing.T) {
	defer goleak.VerifyNone(t)

	employeeID := uuid.New()
	departmentID := uuid.New()
	managerID := uuid.New()

	testCases := []struct {
		name          string
		id            uuid.UUID
		mockSetup     func(*MockDepartmentRepository, *MockEmployeeRepository)
		expectedError error
	}{
		{
			name: "success getting employee with manager",
			id:   employeeID,
			mockSetup: func(deptRepo *MockDepartmentRepository, employeeRepo *MockEmployeeRepository) {
				// Employee exists
				employee := &models.Employee{
					ID:           employeeID,
					Name:         "João Silva",
					CPF:          "12345678901",
					DepartmentID: departmentID,
				}
				employeeRepo.findByIDResult = employee
				employeeRepo.findByIDError = nil

				// Department with manager exists
				department := &models.Department{
					ID:        departmentID,
					Name:      "IT",
					ManagerID: &managerID,
				}
				deptRepo.findByIDResult = department
				deptRepo.findByIDError = nil

				// Manager exists
				manager := &models.Employee{
					ID:   managerID,
					Name: "Manager Silva",
				}
				// Configure for when searching for manager
				employeeRepo.findByIDResults = map[uuid.UUID]*models.Employee{
					employeeID: employee,
					managerID:  manager,
				}
				employeeRepo.findByIDErrors = map[uuid.UUID]error{
					employeeID: nil,
					managerID:  nil,
				}
			},
			expectedError: nil,
		},
		{
			name: "employee not found",
			id:   employeeID,
			mockSetup: func(deptRepo *MockDepartmentRepository, employeeRepo *MockEmployeeRepository) {
				employeeRepo.findByIDResult = nil
				employeeRepo.findByIDError = gorm.ErrRecordNotFound
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

			service := services.NewEmployeeService(deptRepo, employeeRepo)

			// Execute
			result, err := service.GetEmployeeWithManager(tc.id)

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

func TestEmployeeService_DeleteEmployee(t *testing.T) {
	defer goleak.VerifyNone(t)

	employeeID := uuid.New()

	testCases := []struct {
		name          string
		id            uuid.UUID
		mockSetup     func(*MockDepartmentRepository, *MockEmployeeRepository)
		expectedError error
	}{
		{
			name: "success deleting employee",
			id:   employeeID,
			mockSetup: func(deptRepo *MockDepartmentRepository, employeeRepo *MockEmployeeRepository) {
				// Employee exists
				employeeRepo.findByIDResult = &models.Employee{ID: employeeID, Name: "João Silva"}
				employeeRepo.findByIDError = nil
				// Is not a manager
				deptRepo.isManagerResult = false
				deptRepo.isManagerError = nil
				// Delete works
				employeeRepo.deleteError = nil
			},
			expectedError: nil,
		},
		{
			name: "employee not found",
			id:   employeeID,
			mockSetup: func(deptRepo *MockDepartmentRepository, employeeRepo *MockEmployeeRepository) {
				employeeRepo.findByIDResult = nil
				employeeRepo.findByIDError = gorm.ErrRecordNotFound
			},
			expectedError: gorm.ErrRecordNotFound,
		},
		{
			name: "manager cannot be deleted",
			id:   employeeID,
			mockSetup: func(deptRepo *MockDepartmentRepository, employeeRepo *MockEmployeeRepository) {
				// Employee exists
				employeeRepo.findByIDResult = &models.Employee{ID: employeeID, Name: "João Silva"}
				employeeRepo.findByIDError = nil
				// Is a manager
				deptRepo.isManagerResult = true
				deptRepo.isManagerError = nil
			},
			expectedError: utils.ErrManagerCannotBeDeleted,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup
			deptRepo := &MockDepartmentRepository{}
			employeeRepo := &MockEmployeeRepository{}
			tc.mockSetup(deptRepo, employeeRepo)

			service := services.NewEmployeeService(deptRepo, employeeRepo)

			// Execute
			err := service.DeleteEmployee(tc.id)

			// Validate
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

// Benchmark for performance testing
func BenchmarkEmployeeService_CreateEmployee(b *testing.B) {
	deptRepo := &MockDepartmentRepository{
		findByIDResult: &models.Department{ID: uuid.New(), Name: "IT"},
		findByIDError:  nil,
	}
	employeeRepo := &MockEmployeeRepository{
		createError: nil,
	}

	service := services.NewEmployeeService(deptRepo, employeeRepo)

	departmentID := uuid.New()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.CreateEmployee("João Silva", "12345678901", stringPtr("123456789"), departmentID)
	}
}
