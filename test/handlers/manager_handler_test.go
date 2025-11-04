package handlers_test

import (
	"ManageEmployeesandDepartments/internal/handlers"
	"ManageEmployeesandDepartments/internal/models"
	"ManageEmployeesandDepartments/internal/utils"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"go.uber.org/goleak"
)

func TestGerenteHandler_GetSubordinates(t *testing.T) {
	defer goleak.VerifyNone(t)

	testCases := []struct {
		name           string
		idParam        string
		mockSetup      func(*MockDepartmentService)
		expectedStatus int
	}{
		{
			name:    "sucesso ao buscar subordinados",
			idParam: uuid.New().String(),
			mockSetup: func(ms *MockDepartmentService) {
				colaboradores := []*models.Employee{
					{ID: uuid.New(), Name: "João Silva"},
					{ID: uuid.New(), Name: "Maria Santos"},
				}
				ms.getSubordinateEmployeesResult = colaboradores
				ms.getSubordinateEmployeesError = nil
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
			name:    "gerente não encontrado",
			idParam: uuid.New().String(),
			mockSetup: func(ms *MockDepartmentService) {
				ms.getSubordinateEmployeesResult = nil
				ms.getSubordinateEmployeesError = utils.ErrManagerNotFound
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:    "sucesso com lista vazia",
			idParam: uuid.New().String(),
			mockSetup: func(ms *MockDepartmentService) {
				ms.getSubordinateEmployeesResult = nil
				ms.getSubordinateEmployeesError = nil
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup
			mockService := &MockDepartmentService{}
			tc.mockSetup(mockService)

			handler := handlers.NewManagerHandler(mockService)
			router := setupRouter()
			router.GET("/gerentes/:id/colaboradores", handler.GetSubordinates)

			// Executar
			req, _ := http.NewRequest("GET", "/gerentes/"+tc.idParam+"/colaboradores", nil)
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
func BenchmarkGerenteHandler_GetSubordinates(b *testing.B) {
	mockService := &MockDepartmentService{
		getSubordinateEmployeesResult: []*models.Employee{
			{ID: uuid.New(), Name: "João Silva"},
			{ID: uuid.New(), Name: "Maria Santos"},
		},
		getSubordinateEmployeesError: nil,
	}

	handler := handlers.NewManagerHandler(mockService)
	router := setupRouter()
	router.GET("/gerentes/:id/colaboradores", handler.GetSubordinates)

	validID := uuid.New().String()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("GET", "/gerentes/"+validID+"/colaboradores", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}
