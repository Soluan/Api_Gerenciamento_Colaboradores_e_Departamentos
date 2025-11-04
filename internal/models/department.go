package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Department represents the data model of a department.
type Department struct {
	ID   uuid.UUID `gorm:"type:uuid;primary_key;" json:"id"`
	Name string    `gorm:"not null" json:"name"`

	ManagerID *uuid.UUID `json:"manager_id"` // Pointer to accept NULL
	Manager   *Employee  `gorm:"foreignKey:ManagerID" json:"manager,omitempty"`

	ParentDepartmentID *uuid.UUID  `json:"parent_department_id"`                   // Pointer to accept NULL
	ParentDepartment   *Department `gorm:"foreignKey:ParentDepartmentID" json:"-"` // Avoids recursion

	// Used to load the hierarchical tree
	SubDepartments []*Department `gorm:"foreignKey:ParentDepartmentID" json:"sub_departments,omitempty"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// BeforeCreate is a GORM hook to generate UUID v7 before creating.
func (d *Department) BeforeCreate(tx *gorm.DB) (err error) {
	if d.ID == uuid.Nil {
		d.ID, err = uuid.NewV7()
	}
	return err
}

// TableName specifies the table name for this model
func (Department) TableName() string {
	return "departments"
}
