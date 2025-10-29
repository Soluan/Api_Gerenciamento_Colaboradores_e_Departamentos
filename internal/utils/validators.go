package utils

// Este arquivo conteria validadores customizados, como o de CPF.
// Por brevidade, vamos apenas definir a estrutura.

import (
	"strconv"
)

// IsCPFValido implementa o algoritmo de validação de CPF.
// (Esta é uma implementação de exemplo, não use em produção sem testar)
func IsCPFValido(cpf string) bool {
	// Remove pontuação
	cpf = removeNaoDigitos(cpf)

	if len(cpf) != 11 {
		return false
	}

	// Verifica se todos são iguais (ex: 111.111.111-11)
	todosIguais := true
	for i := 1; i < len(cpf); i++ {
		if cpf[i] != cpf[0] {
			todosIguais = false
			break
		}
	}
	if todosIguais {
		return false
	}

	// Calcula o primeiro dígito verificador
	soma := 0
	for i := 0; i < 9; i++ {
		digito, _ := strconv.Atoi(string(cpf[i]))
		soma += digito * (10 - i)
	}
	resto := soma % 11
	dv1 := 0
	if resto >= 2 {
		dv1 = 11 - resto
	}

	dv1Real, _ := strconv.Atoi(string(cpf[9]))
	if dv1 != dv1Real {
		return false
	}

	// Calcula o segundo dígito verificador
	soma = 0
	for i := 0; i < 10; i++ {
		digito, _ := strconv.Atoi(string(cpf[i]))
		soma += digito * (11 - i)
	}
	resto = soma % 11
	dv2 := 0
	if resto >= 2 {
		dv2 = 11 - resto
	}

	dv2Real, _ := strconv.Atoi(string(cpf[10]))
	if dv2 != dv2Real {
		return false
	}

	return true
}

func removeNaoDigitos(s string) string {
	var result string
	for _, r := range s {
		if r >= '0' && r <= '9' {
			result += string(r)
		}
	}
	return result
}
