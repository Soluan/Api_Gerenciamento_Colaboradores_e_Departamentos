package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Colaborador representa o modelo de dados de um colaborador.
type Colaborador struct {
	ID   uuid.UUID `gorm:"type:uuid;primary_key;" json:"id"`
	Nome string    `gorm:"not null" json:"nome"`
	CPF  string    `gorm:"not null;uniqueIndex" json:"cpf"`
	RG   *string   `gorm:"uniqueIndex" json:"rg"` // Ponteiro para aceitar NULL

	DepartamentoID uuid.UUID    `gorm:"not null" json:"departamento_id"`
	Departamento   Departamento `gorm:"foreignKey:DepartamentoID" json:"-"` // Evita recursão no JSON

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// BeforeCreate é um hook do GORM para gerar UUID v7 antes de criar.
func (c *Colaborador) BeforeCreate(tx *gorm.DB) (err error) {
	if c.ID == uuid.Nil {
		c.ID, err = uuid.NewV7()
	}
	return err
}
