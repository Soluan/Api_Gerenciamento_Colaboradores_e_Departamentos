package handlers_test

import (
	"ManageEmployeesandDepartments/internal/handlers"
	"ManageEmployeesandDepartments/internal/models"
	"ManageEmployeesandDepartments/internal/utils"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"go.uber.org/goleak"
	"gorm.io/gorm"
)

// MockDepartmentService simulates the department service
type MockDepartmentService struct {
	createResult                  *models.Department
	createError                   error
	getResult                     *models.Department
	getError                      error
	updateResult                  *models.Department
	updateError                   error
	deleteError                   error
	listResult                    []*models.Department
	listError                     error
	getSubordinateEmployeesResult []*models.Employee
	getSubordinateEmployeesError  error
}

func (m *MockDepartmentService) CreateDepartment(name string, managerID uuid.UUID, parentID *uuid.UUID) (*models.Department, error) {
	return m.createResult, m.createError
}

func (m *MockDepartmentService) GetDepartmentWithTree(id uuid.UUID) (*models.Department, error) {
	return m.getResult, m.getError
}

func (m *MockDepartmentService) UpdateDepartment(id uuid.UUID, name *string, managerID *uuid.UUID, parentID *uuid.UUID) (*models.Department, error) {
	return m.updateResult, m.updateError
}

func (m *MockDepartmentService) DeleteDepartment(id uuid.UUID) error {
	return m.deleteError
}

func (m *MockDepartmentService) ListDepartments(name, managerName *string, parentID *uuid.UUID, page, pageSize int) ([]*models.Department, error) {
	return m.listResult, m.listError
}

func (m *MockDepartmentService) GetSubordinateEmployeesRecursively(managerID uuid.UUID) ([]*models.Employee, error) {
	return m.getSubordinateEmployeesResult, m.getSubordinateEmployeesError
}

func TestDepartamentoHandler_Create(t *testing.T) {
	defer goleak.VerifyNone(t)

	testCases := []struct {
		name               string
		requestBody        models.CreateDepartmentDTO
		mockSetup          func(*MockDepartmentService)
		expectedStatus     int
		expectedErrorField string
	}{
		{
			name: "sucesso ao criar departamento",
			requestBody: models.CreateDepartmentDTO{
				Name:               "TI",
				ManagerID:          uuid.New(),
				ParentDepartmentID: nil,
			},
			mockSetup: func(ms *MockDepartmentService) {
				departamento := &models.Department{
					ID:   uuid.New(),
					Name: "TI",
				}
				ms.createResult = departamento
				ms.createError = nil
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "erro gerente não encontrado",
			requestBody: models.CreateDepartmentDTO{
				Name:               "TI",
				ManagerID:          uuid.New(),
				ParentDepartmentID: nil,
			},
			mockSetup: func(ms *MockDepartmentService) {
				ms.createResult = nil
				ms.createError = utils.ErrManagerNotFound
			},
			expectedStatus:     http.StatusUnprocessableEntity,
			expectedErrorField: "error",
		},
		{
			name: "erro departamento superior não encontrado",
			requestBody: models.CreateDepartmentDTO{
				Name:               "TI",
				ManagerID:          uuid.New(),
				ParentDepartmentID: func() *uuid.UUID { id := uuid.New(); return &id }(),
			},
			mockSetup: func(ms *MockDepartmentService) {
				ms.createResult = nil
				ms.createError = utils.ErrParentDepartmentNotFound
			},
			expectedStatus:     http.StatusUnprocessableEntity,
			expectedErrorField: "error",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup
			mockService := &MockDepartmentService{}
			tc.mockSetup(mockService)

			handler := handlers.NewDepartmentHandler(mockService)
			router := setupRouter()
			router.POST("/departamentos", handler.Create)

			// Preparar request
			bodyBytes, _ := json.Marshal(tc.requestBody)
			req, _ := http.NewRequest("POST", "/departamentos", bytes.NewBuffer(bodyBytes))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()

			// Executar
			router.ServeHTTP(w, req)

			// Validar
			if w.Code != tc.expectedStatus {
				t.Errorf("Expected status %d, got %d", tc.expectedStatus, w.Code)
			}

			if tc.expectedErrorField != "" {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				if err != nil {
					t.Errorf("Failed to unmarshal response: %v", err)
				}
				if _, exists := response[tc.expectedErrorField]; !exists {
					t.Errorf("Expected error field %s not found in response", tc.expectedErrorField)
				}
			}
		})
	}
}

func TestDepartamentoHandler_GetByID(t *testing.T) {
	defer goleak.VerifyNone(t)

	testCases := []struct {
		name           string
		idParam        string
		mockSetup      func(*MockDepartmentService)
		expectedStatus int
	}{
		{
			name:    "sucesso ao buscar departamento",
			idParam: uuid.New().String(),
			mockSetup: func(ms *MockDepartmentService) {
				departamento := &models.Department{
					ID:   uuid.New(),
					Name: "TI",
				}
				ms.getResult = departamento
				ms.getError = nil
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "erro ID inválido",
			idParam:        "invalid-uuid",
			mockSetup:      func(ms *MockDepartmentService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:    "departamento não encontrado",
			idParam: uuid.New().String(),
			mockSetup: func(ms *MockDepartmentService) {
				ms.getResult = nil
				ms.getError = gorm.ErrRecordNotFound
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup
			mockService := &MockDepartmentService{}
			tc.mockSetup(mockService)

			handler := handlers.NewDepartmentHandler(mockService)
			router := setupRouter()
			router.GET("/departamentos/:id", handler.GetByID)

			// Executar
			req, _ := http.NewRequest("GET", "/departamentos/"+tc.idParam, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Validar
			if w.Code != tc.expectedStatus {
				t.Errorf("Expected status %d, got %d", tc.expectedStatus, w.Code)
			}
		})
	}
}

func TestDepartamentoHandler_Update(t *testing.T) {
	defer goleak.VerifyNone(t)

	validID := uuid.New()
	gerenteID := uuid.New()

	testCases := []struct {
		name           string
		idParam        string
		requestBody    models.UpdateDepartmentDTO
		mockSetup      func(*MockDepartmentService)
		expectedStatus int
	}{
		{
			name:    "sucesso ao atualizar departamento",
			idParam: validID.String(),
			requestBody: models.UpdateDepartmentDTO{
				Name:      stringPtr("TI Atualizado"),
				ManagerID: &gerenteID,
			},
			mockSetup: func(ms *MockDepartmentService) {
				departamento := &models.Department{
					ID:   validID,
					Name: "TI Atualizado",
				}
				ms.updateResult = departamento
				ms.updateError = nil
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "erro ID inválido",
			idParam:        "invalid-uuid",
			requestBody:    models.UpdateDepartmentDTO{},
			mockSetup:      func(ms *MockDepartmentService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:    "departamento não encontrado",
			idParam: validID.String(),
			requestBody: models.UpdateDepartmentDTO{
				Name: stringPtr("TI"),
			},
			mockSetup: func(ms *MockDepartmentService) {
				ms.updateResult = nil
				ms.updateError = gorm.ErrRecordNotFound
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:    "ciclo detectado",
			idParam: validID.String(),
			requestBody: models.UpdateDepartmentDTO{
				Name:               stringPtr("TI"),
				ParentDepartmentID: &validID, // Ciclo - se referencia
			},
			mockSetup: func(ms *MockDepartmentService) {
				ms.updateResult = nil
				ms.updateError = utils.ErrCycleDetected
			},
			expectedStatus: http.StatusUnprocessableEntity,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup
			mockService := &MockDepartmentService{}
			tc.mockSetup(mockService)

			handler := handlers.NewDepartmentHandler(mockService)
			router := setupRouter()
			router.PUT("/departamentos/:id", handler.Update)

			// Preparar request
			bodyBytes, _ := json.Marshal(tc.requestBody)
			req, _ := http.NewRequest("PUT", "/departamentos/"+tc.idParam, bytes.NewBuffer(bodyBytes))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()

			// Executar
			router.ServeHTTP(w, req)

			// Validar
			if w.Code != tc.expectedStatus {
				t.Errorf("Expected status %d, got %d", tc.expectedStatus, w.Code)
			}
		})
	}
}

func TestDepartamentoHandler_Delete(t *testing.T) {
	defer goleak.VerifyNone(t)

	validID := uuid.New()

	testCases := []struct {
		name           string
		idParam        string
		mockSetup      func(*MockDepartmentService)
		expectedStatus int
	}{
		{
			name:    "sucesso ao deletar departamento",
			idParam: validID.String(),
			mockSetup: func(ms *MockDepartmentService) {
				ms.deleteError = nil
			},
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "erro ID inválido",
			idParam:        "invalid-uuid",
			mockSetup:      func(ms *MockDepartmentService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:    "departamento não encontrado",
			idParam: validID.String(),
			mockSetup: func(ms *MockDepartmentService) {
				ms.deleteError = gorm.ErrRecordNotFound
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:    "departamento possui colaboradores",
			idParam: validID.String(),
			mockSetup: func(ms *MockDepartmentService) {
				ms.deleteError = utils.ErrDepartmentHasEmployees
			},
			expectedStatus: http.StatusUnprocessableEntity,
		},
		{
			name:    "departamento possui subdepartamentos",
			idParam: validID.String(),
			mockSetup: func(ms *MockDepartmentService) {
				ms.deleteError = utils.ErrDepartmentHasSubDepartments
			},
			expectedStatus: http.StatusUnprocessableEntity,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup
			mockService := &MockDepartmentService{}
			tc.mockSetup(mockService)

			handler := handlers.NewDepartmentHandler(mockService)
			router := setupRouter()
			router.DELETE("/departamentos/:id", handler.Delete)

			// Executar
			req, _ := http.NewRequest("DELETE", "/departamentos/"+tc.idParam, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Validar
			if w.Code != tc.expectedStatus {
				t.Errorf("Expected status %d, got %d", tc.expectedStatus, w.Code)
			}
		})
	}
}

func TestDepartamentoHandler_List(t *testing.T) {
	defer goleak.VerifyNone(t)

	testCases := []struct {
		name           string
		requestBody    models.ListDepartmentsDTO
		mockSetup      func(*MockDepartmentService)
		expectedStatus int
	}{
		{
			name: "sucesso ao listar departamentos",
			requestBody: models.ListDepartmentsDTO{
				Page:     1,
				PageSize: 10,
			},
			mockSetup: func(ms *MockDepartmentService) {
				departamentos := []*models.Department{
					{ID: uuid.New(), Name: "TI"},
					{ID: uuid.New(), Name: "RH"},
				}
				ms.listResult = departamentos
				ms.listError = nil
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:        "sucesso com lista vazia",
			requestBody: models.ListDepartmentsDTO{},
			mockSetup: func(ms *MockDepartmentService) {
				ms.listResult = nil
				ms.listError = nil
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup
			mockService := &MockDepartmentService{}
			tc.mockSetup(mockService)

			handler := handlers.NewDepartmentHandler(mockService)
			router := setupRouter()
			router.POST("/departamentos/listar", handler.List)

			// Preparar request
			bodyBytes, _ := json.Marshal(tc.requestBody)
			req, _ := http.NewRequest("POST", "/departamentos/listar", bytes.NewBuffer(bodyBytes))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()

			// Executar
			router.ServeHTTP(w, req)

			// Validar
			if w.Code != tc.expectedStatus {
				t.Errorf("Expected status %d, got %d", tc.expectedStatus, w.Code)
			}

			if tc.expectedStatus == http.StatusOK {
				var response []models.Department
				err := json.Unmarshal(w.Body.Bytes(), &response)
				if err != nil {
					t.Errorf("Failed to unmarshal response: %v", err)
				}
			}
		})
	}
}

// Benchmark para teste de performance
func BenchmarkDepartamentoHandler_Create(b *testing.B) {
	mockService := &MockDepartmentService{
		createResult: &models.Department{
			ID:   uuid.New(),
			Name: "TI",
		},
		createError: nil,
	}

	handler := handlers.NewDepartmentHandler(mockService)
	router := setupRouter()
	router.POST("/departamentos", handler.Create)

	requestBody := models.CreateDepartmentDTO{
		Name:               "TI",
		ManagerID:          uuid.New(),
		ParentDepartmentID: nil,
	}

	bodyBytes, _ := json.Marshal(requestBody)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("POST", "/departamentos", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}
