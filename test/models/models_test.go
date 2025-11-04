package models_test

import (
	"ManageEmployeesandDepartments/internal/models"
	"testing"
	"time"

	"github.com/google/uuid"
	"go.uber.org/goleak"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Discard,
	})
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// Disable foreign key constraints for simpler testing
	db.Exec("PRAGMA foreign_keys = OFF")

	// Auto migrate
	err = db.AutoMigrate(&models.Department{}, &models.Employee{})
	if err != nil {
		t.Fatalf("Failed to migrate tables: %v", err)
	}

	return db
}

func TestColaboradorModel(t *testing.T) {
	defer goleak.VerifyNone(t)

	db := setupTestDB(t)

	// Close connection at the end
	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	// Create a department first
	department := &models.Department{
		Name: "TI",
	}

	err := db.Create(department).Error
	if err != nil {
		t.Fatalf("Failed to create department: %v", err)
	}

	// Test creating a employee
	employee := &models.Employee{
		Name:           "João Silva",
		CPF:            "12345678901",
		DepartmentID: department.ID,
	}

	err = db.Create(employee).Error
	if err != nil {
		t.Fatalf("Failed to create employee: %v", err)
	}

	// Verify the employee was created
	if employee.ID == uuid.Nil {
		t.Error("Expected employee ID to be generated")
	}

	if employee.Name != "João Silva" {
		t.Errorf("Expected nome to be 'João Silva', got %s", employee.Name)
	}

	// Test finding the employee
	var found models.Employee
	err = db.First(&found, employee.ID).Error
	if err != nil {
		t.Fatalf("Failed to find employee: %v", err)
	}

	if found.Name != employee.Name {
		t.Errorf("Expected found employee nome to be %s, got %s", employee.Name, found.Name)
	}
}

func TestDepartamentoModel(t *testing.T) {
	defer goleak.VerifyNone(t)

	db := setupTestDB(t)

	// Close connection at the end
	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	// Test creating a department
	department := &models.Department{
		Name: "Recursos Humanos",
	}

	err := db.Create(department).Error
	if err != nil {
		t.Fatalf("Failed to create department: %v", err)
	}

	// Verify the department was created
	if department.ID == uuid.Nil {
		t.Error("Expected department ID to be generated")
	}

	if department.Name != "Recursos Humanos" {
		t.Errorf("Expected nome to be 'Recursos Humanos', got %s", department.Name)
	}

	// Test finding the department
	var found models.Department
	err = db.First(&found, department.ID).Error
	if err != nil {
		t.Fatalf("Failed to find department: %v", err)
	}

	if found.Name != department.Name {
		t.Errorf("Expected found department nome to be %s, got %s", department.Name, found.Name)
	}
}

func TestColaboradorBeforeCreate(t *testing.T) {
	defer goleak.VerifyNone(t)

	db := setupTestDB(t)

	// Close connection at the end
	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	// Create a department first
	department := &models.Department{
		Name: "Marketing",
	}

	err := db.Create(department).Error
	if err != nil {
		t.Fatalf("Failed to create department: %v", err)
	}

	// Test that BeforeCreate hook generates UUID
	employee := &models.Employee{
		Name:           "Test User",
		CPF:            "12345678901",
		DepartmentID: department.ID,
	}

	// ID should be Nil before creation
	if employee.ID != uuid.Nil {
		t.Error("Expected employee ID to be Nil before creation")
	}

	err = db.Create(employee).Error
	if err != nil {
		t.Fatalf("Failed to create employee: %v", err)
	}

	// ID should be generated after creation
	if employee.ID == uuid.Nil {
		t.Error("Expected employee ID to be generated after creation")
	}
}

func TestDepartamentoBeforeCreate(t *testing.T) {
	defer goleak.VerifyNone(t)

	db := setupTestDB(t)

	// Close connection at the end
	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	// Test that BeforeCreate hook generates UUID
	department := &models.Department{
		Name: "Vendas",
	}

	// ID should be Nil before creation
	if department.ID != uuid.Nil {
		t.Error("Expected department ID to be Nil before creation")
	}

	err := db.Create(department).Error
	if err != nil {
		t.Fatalf("Failed to create department: %v", err)
	}

	// ID should be generated after creation
	if department.ID == uuid.Nil {
		t.Error("Expected department ID to be generated after creation")
	}
}

func TestSoftDelete(t *testing.T) {
	defer goleak.VerifyNone(t)

	db := setupTestDB(t)

	// Close connection at the end
	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	// Create a department and employee
	department := &models.Department{
		Name: "Operações",
	}

	err := db.Create(department).Error
	if err != nil {
		t.Fatalf("Failed to create department: %v", err)
	}

	employee := &models.Employee{
		Name:           "Test Delete",
		CPF:            "99999999999",
		DepartmentID: department.ID,
	}

	err = db.Create(employee).Error
	if err != nil {
		t.Fatalf("Failed to create employee: %v", err)
	}

	// Soft delete the employee
	err = db.Delete(employee).Error
	if err != nil {
		t.Fatalf("Failed to soft delete employee: %v", err)
	}

	// Verify employee is not found with normal query
	var found models.Employee
	err = db.First(&found, employee.ID).Error
	if err == nil {
		t.Error("Expected employee to be soft deleted and not found")
	}

	// Verify employee is found with Unscoped query
	err = db.Unscoped().First(&found, employee.ID).Error
	if err != nil {
		t.Fatalf("Failed to find soft deleted employee with Unscoped: %v", err)
	}

	if found.DeletedAt.Time.IsZero() {
		t.Error("Expected employee to have DeletedAt timestamp")
	}
}

func TestTimestampUpdates(t *testing.T) {
	defer goleak.VerifyNone(t)

	db := setupTestDB(t)

	// Close connection at the end
	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	// Create a department and employee
	department := &models.Department{
		Name: "Financeiro",
	}

	err := db.Create(department).Error
	if err != nil {
		t.Fatalf("Failed to create department: %v", err)
	}

	employee := &models.Employee{
		Name:           "Update Test",
		CPF:            "11111111111",
		DepartmentID: department.ID,
	}

	err = db.Create(employee).Error
	if err != nil {
		t.Fatalf("Failed to create employee: %v", err)
	}

	originalUpdatedAt := employee.UpdatedAt

	// Wait a bit and update
	time.Sleep(10 * time.Millisecond)

	employee.Name = "Updated Name"
	err = db.Save(employee).Error
	if err != nil {
		t.Fatalf("Failed to update employee: %v", err)
	}

	// Verify UpdatedAt was updated
	if !employee.UpdatedAt.After(originalUpdatedAt) {
		t.Error("Expected UpdatedAt to be updated after save")
	}
}
