package services

import (
	"strconv"
	"strings"
	"github.com/didiercito/api-go-examen2/models"
)

func AnalyzeSyntax(code string) models.SyntaxResult {
	lines := strings.Split(code, "\n")
	var errors []string
	isValid := true

	for i, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		lineNum := strconv.Itoa(i + 1)

		// Verificar estructura básica
		if strings.HasPrefix(line, "def ") && !strings.HasSuffix(line, ":") {
			errors = append(errors, "Línea "+lineNum+": Falta ':' en definición de función")
			isValid = false
		}
		
		if strings.HasPrefix(line, "if ") && !strings.HasSuffix(line, ":") {
			errors = append(errors, "Línea "+lineNum+": Falta ':' en estructura if")
			isValid = false
		}

		// Verificar paréntesis balanceados
		openCount := strings.Count(line, "(")
		closeCount := strings.Count(line, ")")
		if openCount != closeCount {
			errors = append(errors, "Línea "+lineNum+": Paréntesis no balanceados")
			isValid = false
		}
	}

	return models.SyntaxResult{IsValid: isValid, Errors: errors}
}