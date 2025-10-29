package utils

import (
	"errors"
	"fmt"
	"net/http"
)

var (
	ErrColaboradorNaoEncontrado           = errors.New("colaborador nao encontrado")
	ErrDepartamentoSuperiorNaoEncontrado  = errors.New("departamento superior não encontrado")
	ErrCicloDetectado                     = errors.New("ciclo hierárquico detectado")
	ErrDepartamentoPossuiColaboradores    = errors.New("departamento possui colaboradores vinculados")
	ErrDepartamentoPossuiSubDepartamentos = errors.New("departamento possui sub-departamentos vinculados")
	ErrGerenteNaoPertenceAoDepto          = errors.New("o gerente deve pertencer ao departamento que irá gerenciar")
	ErrNaoEncontrado                      = errors.New("recurso não encontrado")
	ErrInvalido                           = errors.New("dados fornecidos são inválidos")
	ErrDepartamentoNaoEncontrado          = errors.New("departamento não encontrado")
	ErrCPFDuplicado                       = errors.New("CPF já cadastrado")
	ErrRGDuplicado                        = errors.New("RG já cadastrado")
	ErrGerenteNaoEncontrado               = errors.New("gerente não encontrado")
	ErrGerenteNaoPodeSerExcluido          = errors.New("colaborador é gerente e não pode ser removido")
)

// CustomError representa uma estrutura de erro padronizada para a API (Resposta HTTP).
type CustomError struct {
	Code    int    `json:"-"` // Código HTTP (não serializado no JSON)
	Message string `json:"message"`
	Details string `json:"details,omitempty"` // Detalhes técnicos (o erro original .Error())
}

func (e *CustomError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("Erro Customizado (HTTP %d): %s - Detalhes: %s", e.Code, e.Message, e.Details)
	}
	return fmt.Sprintf("Erro Customizado (HTTP %d): %s", e.Code, e.Message)
}

func NewCustomError(code int, message string, details string) *CustomError {
	return &CustomError{
		Code:    code,
		Message: message,
		Details: details,
	}
}

// MapErrorToCustom converte um erro de regra de negócio (padrão Go error) para um CustomError
func MapErrorToCustom(err error) *CustomError {
	switch {
	case errors.Is(err, ErrDepartamentoSuperiorNaoEncontrado),
		errors.Is(err, ErrNaoEncontrado):
		return NewCustomError(http.StatusNotFound, "Recurso não encontrado.", err.Error())
	}

	switch {
	case errors.Is(err, ErrCicloDetectado),
		errors.Is(err, ErrDepartamentoPossuiColaboradores),
		errors.Is(err, ErrDepartamentoPossuiSubDepartamentos),
		errors.Is(err, ErrGerenteNaoPertenceAoDepto),
		errors.Is(err, ErrInvalido):
		return NewCustomError(http.StatusBadRequest, "Falha na regra de negócio ou dados inválidos.", err.Error())
	}

	var customErr *CustomError
	if errors.As(err, &customErr) {
		return customErr
	}

	return NewCustomError(http.StatusInternalServerError, "Erro interno do servidor.", err.Error())
}
