package models

import "github.com/google/uuid"

// ColaboradorComGerenteResponse é o DTO de resposta para GetByID.
type ColaboradorComGerenteResponse struct {
	*Colaborador
	GerenteNome *string `json:"gerente_nome"`
}

type CreateColaboradorDTO struct {
	Nome           string    `json:"nome" binding:"required"`
	CPF            string    `json:"cpf" binding:"required"` // Validação de formato (ex: 11 dígitos) pode ser adicionada
	RG             *string   `json:"rg"`
	DepartamentoID uuid.UUID `json:"departamento_id" binding:"required"`
}

type UpdateColaboradorDTO struct {
	Nome           *string    `json:"nome"`
	RG             *string    `json:"rg"`
	DepartamentoID *uuid.UUID `json:"departamento_id"`
}

type ListColaboradoresDTO struct {
	Nome           *string    `json:"nome"`
	CPF            *string    `json:"cpf"`
	RG             *string    `json:"rg"`
	DepartamentoID *uuid.UUID `json:"departamento_id"`
	Pagina         int        `json:"pagina" binding:"omitempty,gte=1"`
	TamanhoPagina  int        `json:"tamanho_pagina" binding:"omitempty,gte=1"`
}

// Departament
type CreateDepartamentoDTO struct {
	Nome                   string     `json:"nome" binding:"required"`
	GerenteID              uuid.UUID  `json:"gerente_id" binding:"required"`
	DepartamentoSuperiorID *uuid.UUID `json:"departamento_superior_id"`
}

// UpdateDepartamentoDTO é usado para atualizar um departamento.
type UpdateDepartamentoDTO struct {
	Nome                   *string    `json:"nome"`
	GerenteID              *uuid.UUID `json:"gerente_id"`
	DepartamentoSuperiorID *uuid.UUID `json:"departamento_superior_id"`
}

// ListDepartamentosDTO é usado para filtros e paginação.
type ListDepartamentosDTO struct {
	Nome                   *string    `json:"nome"`
	GerenteNome            *string    `json:"gerente_nome"` // Filtro especial
	DepartamentoSuperiorID *uuid.UUID `json:"departamento_superior_id"`
	Pagina                 int        `json:"pagina" binding:"omitempty,gte=1"`
	TamanhoPagina          int        `json:"tamanho_pagina" binding:"omitempty,gte=1"`
}
