package handlers_test

import (
	"ManageEmployeesandDepartments/internal/handlers"
	"ManageEmployeesandDepartments/internal/models"
	"ManageEmployeesandDepartments/internal/services"
	"ManageEmployeesandDepartments/internal/utils"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/goleak"
)

// MockEmployeeService simula o serviço de colaborador
type MockEmployeeService struct {
	createResult *models.Employee
	createError  error
	getResult    *services.EmployeeWithManagerResponse
	getError     error
	updateResult *models.Employee
	updateError  error
	deleteError  error
	listResult   []*models.Employee
	listError    error
}

func (m *MockEmployeeService) CreateEmployee(name string, cpf string, rg *string, departmentID uuid.UUID) (*models.Employee, error) {
	return m.createResult, m.createError
}

func (m *MockEmployeeService) GetEmployeeWithManager(id uuid.UUID) (*services.EmployeeWithManagerResponse, error) {
	return m.getResult, m.getError
}

func (m *MockEmployeeService) UpdateEmployee(id uuid.UUID, name *string, rg *string, departmentID uuid.UUID) (*models.Employee, error) {
	return m.updateResult, m.updateError
}

func (m *MockEmployeeService) DeleteEmployee(id uuid.UUID) error {
	return m.deleteError
}

func (m *MockEmployeeService) ListEmployees(name *string, cpf *string, rg *string, deptoID *uuid.UUID, pagina, tamanhoPagina int) ([]*models.Employee, error) {
	return m.listResult, m.listError
}

func setupRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

// Função auxiliar para criar ponteiros de string
func stringPtr(s string) *string {
	return &s
}

func TestColaboradorHandler_Create(t *testing.T) {
	defer goleak.VerifyNone(t)

	testCases := []struct {
		name               string
		requestBody        models.CreateEmployeeDTO
		mockSetup          func(*MockEmployeeService)
		expectedStatus     int
		expectedErrorField string
	}{
		{
			name: "sucesso ao criar colaborador",
			requestBody: models.CreateEmployeeDTO{
				Name:           "João Silva",
				CPF:            "12345678901",
				RG:             stringPtr("123456789"),
				DepartmentID: uuid.New(),
			},
			mockSetup: func(ms *MockEmployeeService) {
				colaborador := &models.Employee{
					ID:   uuid.New(),
					Name: "João Silva",
					CPF:  "12345678901",
					RG:   stringPtr("123456789"),
				}
				ms.createResult = colaborador
				ms.createError = nil
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "erro departamento não encontrado",
			requestBody: models.CreateEmployeeDTO{
				Name:           "João Silva",
				CPF:            "12345678901",
				RG:             stringPtr("123456789"),
				DepartmentID: uuid.New(),
			},
			mockSetup: func(ms *MockEmployeeService) {
				ms.createResult = nil
				ms.createError = utils.ErrDepartmentNotFound
			},
			expectedStatus:     http.StatusUnprocessableEntity,
			expectedErrorField: "error",
		},
		{
			name: "erro CPF duplicado",
			requestBody: models.CreateEmployeeDTO{
				Name:           "João Silva",
				CPF:            "12345678901",
				RG:             stringPtr("123456789"),
				DepartmentID: uuid.New(),
			},
			mockSetup: func(ms *MockEmployeeService) {
				ms.createResult = nil
				ms.createError = utils.ErrCPFDuplicated
			},
			expectedStatus:     http.StatusConflict,
			expectedErrorField: "error",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup
			mockService := &MockEmployeeService{}
			tc.mockSetup(mockService)

			handler := handlers.NewEmployeeHandler(mockService)
			router := setupRouter()
			router.POST("/colaboradores", handler.Create)

			// Preparar request
			bodyBytes, _ := json.Marshal(tc.requestBody)
			req, _ := http.NewRequest("POST", "/colaboradores", bytes.NewBuffer(bodyBytes))
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

func TestColaboradorHandler_GetByID(t *testing.T) {
	defer goleak.VerifyNone(t)

	testCases := []struct {
		name           string
		idParam        string
		mockSetup      func(*MockEmployeeService)
		expectedStatus int
	}{
		{
			name:    "sucesso ao buscar colaborador",
			idParam: uuid.New().String(),
			mockSetup: func(ms *MockEmployeeService) {
				response := &services.EmployeeWithManagerResponse{
					Employee: &models.Employee{
						ID:   uuid.New(),
						Name: "João Silva",
						CPF:  "12345678901",
						RG:   stringPtr("123456789"),
					},
					ManagerName: "Gerente Silva",
				}
				ms.getResult = response
				ms.getError = nil
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "erro ID inválido",
			idParam:        "invalid-uuid",
			mockSetup:      func(ms *MockEmployeeService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:    "colaborador não encontrado",
			idParam: uuid.New().String(),
			mockSetup: func(ms *MockEmployeeService) {
				ms.getResult = nil
				ms.getError = utils.ErrEmployeeNotFound
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup
			mockService := &MockEmployeeService{}
			tc.mockSetup(mockService)

			handler := handlers.NewEmployeeHandler(mockService)
			router := setupRouter()
			router.GET("/colaboradores/:id", handler.GetByID)

			// Executar
			req, _ := http.NewRequest("GET", "/colaboradores/"+tc.idParam, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Validar
			if w.Code != tc.expectedStatus {
				t.Errorf("Expected status %d, got %d", tc.expectedStatus, w.Code)
			}
		})
	}
}

func TestColaboradorHandler_Delete(t *testing.T) {
	defer goleak.VerifyNone(t)

	validID := uuid.New()

	testCases := []struct {
		name           string
		idParam        string
		mockSetup      func(*MockEmployeeService)
		expectedStatus int
	}{
		{
			name:    "sucesso ao deletar colaborador",
			idParam: validID.String(),
			mockSetup: func(ms *MockEmployeeService) {
				ms.deleteError = nil
			},
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "erro ID inválido",
			idParam:        "invalid-uuid",
			mockSetup:      func(ms *MockEmployeeService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:    "colaborador não encontrado",
			idParam: validID.String(),
			mockSetup: func(ms *MockEmployeeService) {
				ms.deleteError = utils.ErrEmployeeNotFound
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup
			mockService := &MockEmployeeService{}
			tc.mockSetup(mockService)

			handler := handlers.NewEmployeeHandler(mockService)
			router := setupRouter()
			router.DELETE("/colaboradores/:id", handler.Delete)

			// Executar
			req, _ := http.NewRequest("DELETE", "/colaboradores/"+tc.idParam, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Validar
			if w.Code != tc.expectedStatus {
				t.Errorf("Expected status %d, got %d", tc.expectedStatus, w.Code)
			}
		})
	}
}

// Benchmark para teste de performance
func BenchmarkColaboradorHandler_Create(b *testing.B) {
	mockService := &MockEmployeeService{
		createResult: &models.Employee{
			ID:   uuid.New(),
			Name: "João Silva",
			CPF:  "12345678901",
			RG:   stringPtr("123456789"),
		},
		createError: nil,
	}

	handler := handlers.NewEmployeeHandler(mockService)
	router := setupRouter()
	router.POST("/colaboradores", handler.Create)

	requestBody := models.CreateEmployeeDTO{
		Name:           "João Silva",
		CPF:            "12345678901",
		RG:             stringPtr("123456789"),
		DepartmentID: uuid.New(),
	}

	bodyBytes, _ := json.Marshal(requestBody)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("POST", "/colaboradores", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}
