package utils_test

import (
	"ManageEmployeesandDepartments/internal/utils"
	"errors"
	"net/http"
	"testing"
)

func TestCustomError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *utils.CustomError
		expected string
	}{
		{
			name: "Error with details",
			err: &utils.CustomError{
				Code:    http.StatusBadRequest,
				Message: "Bad request",
				Details: "Invalid input data",
			},
			expected: "Custom Error (HTTP 400): Bad request - Details: Invalid input data",
		},
		{
			name: "Error without details",
			err: &utils.CustomError{
				Code:    http.StatusNotFound,
				Message: "Resource not found",
				Details: "",
			},
			expected: "Custom Error (HTTP 404): Resource not found",
		},
		{
			name: "Internal server error",
			err: &utils.CustomError{
				Code:    http.StatusInternalServerError,
				Message: "Internal error",
				Details: "Database connection failed",
			},
			expected: "Custom Error (HTTP 500): Internal error - Details: Database connection failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.err.Error()
			if result != tt.expected {
				t.Errorf("CustomError.Error() = %q, expected %q", result, tt.expected)
			}
		})
	}
}

func TestNewCustomError(t *testing.T) {
	tests := []struct {
		name     string
		code     int
		message  string
		details  string
		expected *utils.CustomError
	}{
		{
			name:    "Create custom error with details",
			code:    http.StatusBadRequest,
			message: "Validation failed",
			details: "CPF already exists",
			expected: &utils.CustomError{
				Code:    http.StatusBadRequest,
				Message: "Validation failed",
				Details: "CPF already exists",
			},
		},
		{
			name:    "Create custom error without details",
			code:    http.StatusNotFound,
			message: "User not found",
			details: "",
			expected: &utils.CustomError{
				Code:    http.StatusNotFound,
				Message: "User not found",
				Details: "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.NewCustomError(tt.code, tt.message, tt.details)

			if result.Code != tt.expected.Code {
				t.Errorf("NewCustomError().Code = %d, expected %d", result.Code, tt.expected.Code)
			}
			if result.Message != tt.expected.Message {
				t.Errorf("NewCustomError().Message = %q, expected %q", result.Message, tt.expected.Message)
			}
			if result.Details != tt.expected.Details {
				t.Errorf("NewCustomError().Details = %q, expected %q", result.Details, tt.expected.Details)
			}
		})
	}
}

func TestMapErrorToCustom(t *testing.T) {
	tests := []struct {
		name         string
		inputError   error
		expectedCode int
		expectedMsg  string
	}{
		{
			name:         "ErrParentDepartmentNotFound",
			inputError:   utils.ErrParentDepartmentNotFound,
			expectedCode: http.StatusNotFound,
			expectedMsg:  "Resource not found.",
		},
		{
			name:         "ErrNotFound",
			inputError:   utils.ErrNotFound,
			expectedCode: http.StatusNotFound,
			expectedMsg:  "Resource not found.",
		},
		{
			name:         "ErrCycleDetected",
			inputError:   utils.ErrCycleDetected,
			expectedCode: http.StatusBadRequest,
			expectedMsg:  "Business rule failure or invalid data.",
		},
		{
			name:         "ErrDepartmentHasEmployees",
			inputError:   utils.ErrDepartmentHasEmployees,
			expectedCode: http.StatusBadRequest,
			expectedMsg:  "Business rule failure or invalid data.",
		},
		{
			name:         "ErrDepartmentHasSubDepartments",
			inputError:   utils.ErrDepartmentHasSubDepartments,
			expectedCode: http.StatusBadRequest,
			expectedMsg:  "Business rule failure or invalid data.",
		},
		{
			name:         "ErrManagerNotBelongToDepartment",
			inputError:   utils.ErrManagerNotBelongToDepartment,
			expectedCode: http.StatusBadRequest,
			expectedMsg:  "Business rule failure or invalid data.",
		},
		{
			name:         "ErrInvalid",
			inputError:   utils.ErrInvalid,
			expectedCode: http.StatusBadRequest,
			expectedMsg:  "Business rule failure or invalid data.",
		},
		{
			name:         "Unknown error",
			inputError:   errors.New("unknown error"),
			expectedCode: http.StatusInternalServerError,
			expectedMsg:  "Internal server error.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.MapErrorToCustom(tt.inputError)

			if result.Code != tt.expectedCode {
				t.Errorf("MapErrorToCustom().Code = %d, expected %d", result.Code, tt.expectedCode)
			}
			if result.Message != tt.expectedMsg {
				t.Errorf("MapErrorToCustom().Message = %q, expected %q", result.Message, tt.expectedMsg)
			}
			if result.Details != tt.inputError.Error() {
				t.Errorf("MapErrorToCustom().Details = %q, expected %q", result.Details, tt.inputError.Error())
			}
		})
	}
}

func TestPredefinedErrors(t *testing.T) {
	tests := []struct {
		name string
		err  error
		msg  string
	}{
		{
			name: "ErrEmployeeNotFound",
			err:  utils.ErrEmployeeNotFound,
			msg:  "employee not found",
		},
		{
			name: "ErrParentDepartmentNotFound",
			err:  utils.ErrParentDepartmentNotFound,
			msg:  "parent department not found",
		},
		{
			name: "ErrCycleDetected",
			err:  utils.ErrCycleDetected,
			msg:  "hierarchical cycle detected",
		},
		{
			name: "ErrDepartmentHasEmployees",
			err:  utils.ErrDepartmentHasEmployees,
			msg:  "department has linked employees",
		},
		{
			name: "ErrDepartmentHasSubDepartments",
			err:  utils.ErrDepartmentHasSubDepartments,
			msg:  "department has linked sub-departments",
		},
		{
			name: "ErrManagerNotBelongToDepartment",
			err:  utils.ErrManagerNotBelongToDepartment,
			msg:  "the manager must belong to the department they will manage",
		},
		{
			name: "ErrNotFound",
			err:  utils.ErrNotFound,
			msg:  "resource not found",
		},
		{
			name: "ErrInvalid",
			err:  utils.ErrInvalid,
			msg:  "provided data is invalid",
		},
		{
			name: "ErrDepartmentNotFound",
			err:  utils.ErrDepartmentNotFound,
			msg:  "department not found",
		},
		{
			name: "ErrCPFDuplicated",
			err:  utils.ErrCPFDuplicated,
			msg:  "CPF already registered",
		},
		{
			name: "ErrRGDuplicated",
			err:  utils.ErrRGDuplicated,
			msg:  "RG already registered",
		},
		{
			name: "ErrManagerNotFound",
			err:  utils.ErrManagerNotFound,
			msg:  "manager not found",
		},
		{
			name: "ErrManagerCannotBeDeleted",
			err:  utils.ErrManagerCannotBeDeleted,
			msg:  "employee is a manager and cannot be removed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Error() != tt.msg {
				t.Errorf("Error message = %q, expected %q", tt.err.Error(), tt.msg)
			}
		})
	}
}

// Benchmark para medir performance do mapeamento de erros
func BenchmarkMapErrorToCustom(b *testing.B) {
	err := utils.ErrParentDepartmentNotFound

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		utils.MapErrorToCustom(err)
	}
}

// Test para verificar se errors.Is funciona corretamente com nossos erros personalizados
func TestErrorsIs(t *testing.T) {
	tests := []struct {
		name   string
		target error
		err    error
		match  bool
	}{
		{
			name:   "ErrParentDepartmentNotFound matches",
			target: utils.ErrParentDepartmentNotFound,
			err:    utils.ErrParentDepartmentNotFound,
			match:  true,
		},
		{
			name:   "Different errors don't match",
			target: utils.ErrParentDepartmentNotFound,
			err:    utils.ErrCycleDetected,
			match:  false,
		},
		{
			name:   "Wrapped error matches",
			target: utils.ErrParentDepartmentNotFound,
			err:    errors.New("wrapper: " + utils.ErrParentDepartmentNotFound.Error()),
			match:  false, // errors.New não envelopa, então não vai dar match
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := errors.Is(tt.err, tt.target)
			if result != tt.match {
				t.Errorf("errors.Is() = %v, expected %v", result, tt.match)
			}
		})
	}
}
