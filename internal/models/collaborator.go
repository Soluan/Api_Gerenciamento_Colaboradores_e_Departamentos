package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Employee represents the data model of an employee.
type Employee struct {
	ID   uuid.UUID `gorm:"type:uuid;primary_key;" json:"id"`
	Name string    `gorm:"not null" json:"name"`
	CPF  string    `gorm:"not null;uniqueIndex" json:"cpf"`
	RG   *string   `gorm:"uniqueIndex" json:"rg"` // Pointer to accept NULL

	DepartmentID uuid.UUID  `gorm:"not null" json:"department_id"`
	Department   Department `gorm:"foreignKey:DepartmentID" json:"-"` // Avoids recursion in JSON

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// BeforeCreate is a GORM hook to generate UUID v7 before creating.
func (c *Employee) BeforeCreate(tx *gorm.DB) (err error) {
	if c.ID == uuid.Nil {
		c.ID, err = uuid.NewV7()
	}
	return err
}

// TableName specifies the table name for this model
func (Employee) TableName() string {
	return "employees"
}
