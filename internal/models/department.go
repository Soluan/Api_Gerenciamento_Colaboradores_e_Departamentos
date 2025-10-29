package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Departamento representa o modelo de dados de um departamento.
type Departamento struct {
	ID   uuid.UUID `gorm:"type:uuid;primary_key;" json:"id"`
	Nome string    `gorm:"not null" json:"nome"`

	GerenteID *uuid.UUID   `json:"gerente_id"` // Ponteiro para aceitar NULL
	Gerente   *Colaborador `gorm:"foreignKey:GerenteID" json:"gerente,omitempty"`

	DepartamentoSuperiorID *uuid.UUID    `json:"departamento_superior_id"`                   // Ponteiro para aceitar NULL
	DepartamentoSuperior   *Departamento `gorm:"foreignKey:DepartamentoSuperiorID" json:"-"` // Evita recursão

	// Usado para carregar a árvore hierárquica
	SubDepartamentos []*Departamento `gorm:"foreignKey:DepartamentoSuperiorID" json:"sub_departamentos,omitempty"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
