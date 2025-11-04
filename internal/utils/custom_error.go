package utils

import (
	"errors"
	"fmt"
	"net/http"
)

var (
	ErrEmployeeNotFound             = errors.New("employee not found")
	ErrParentDepartmentNotFound     = errors.New("parent department not found")
	ErrCycleDetected                = errors.New("hierarchical cycle detected")
	ErrDepartmentHasEmployees       = errors.New("department has linked employees")
	ErrDepartmentHasSubDepartments  = errors.New("department has linked sub-departments")
	ErrManagerNotBelongToDepartment = errors.New("the manager must belong to the department they will manage")
	ErrNotFound                     = errors.New("resource not found")
	ErrInvalid                      = errors.New("provided data is invalid")
	ErrDepartmentNotFound           = errors.New("department not found")
	ErrCPFDuplicated                = errors.New("CPF already registered")
	ErrRGDuplicated                 = errors.New("RG already registered")
	ErrManagerNotFound              = errors.New("manager not found")
	ErrManagerCannotBeDeleted       = errors.New("employee is a manager and cannot be removed")
)

// CustomError represents a standardized error structure for the API (HTTP Response).
type CustomError struct {
	Code    int    `json:"-"` // HTTP Code (not serialized in JSON)
	Message string `json:"message"`
	Details string `json:"details,omitempty"` // Technical details (the original error .Error())
}

func (e *CustomError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("Custom Error (HTTP %d): %s - Details: %s", e.Code, e.Message, e.Details)
	}
	return fmt.Sprintf("Custom Error (HTTP %d): %s", e.Code, e.Message)
}

func NewCustomError(code int, message string, details string) *CustomError {
	return &CustomError{
		Code:    code,
		Message: message,
		Details: details,
	}
}

// MapErrorToCustom converts a business rule error (standard Go error) to a CustomError
func MapErrorToCustom(err error) *CustomError {
	switch {
	case errors.Is(err, ErrParentDepartmentNotFound),
		errors.Is(err, ErrNotFound):
		return NewCustomError(http.StatusNotFound, "Resource not found.", err.Error())
	}

	switch {
	case errors.Is(err, ErrCycleDetected),
		errors.Is(err, ErrDepartmentHasEmployees),
		errors.Is(err, ErrDepartmentHasSubDepartments),
		errors.Is(err, ErrManagerNotBelongToDepartment),
		errors.Is(err, ErrInvalid):
		return NewCustomError(http.StatusBadRequest, "Business rule failure or invalid data.", err.Error())
	}

	var customErr *CustomError
	if errors.As(err, &customErr) {
		return customErr
	}

	return NewCustomError(http.StatusInternalServerError, "Internal server error.", err.Error())
}
