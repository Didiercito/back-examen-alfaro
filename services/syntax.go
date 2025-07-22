package services

import (
	"regexp"
	"strconv"
	"strings"
	"github.com/didiercito/api-go-examen2/models"
)

func AnalyzeSyntax(code string) models.SyntaxResult {
	lines := strings.Split(code, "\n")
	var errors []string
	isValid := true
	
	braceStack := 0
	parenStack := 0
	bracketStack := 0
	hasMain := false

	for i, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "//") {
			continue
		}

		lineNum := strconv.Itoa(i + 1)

		// Verificar estructura básica de includes
		if strings.HasPrefix(line, "#include") {
			if !isValidInclude(line) {
				errors = append(errors, "Línea "+lineNum+": Include mal formado")
				isValid = false
			}
		}

		// Verificar using namespace
		if strings.HasPrefix(line, "using namespace") {
			if !isValidUsing(line) {
				errors = append(errors, "Línea "+lineNum+": Declaración using namespace incorrecta")
				isValid = false
			}
		}

		// Verificar función main
		if strings.Contains(line, "int main") {
			hasMain = true
			if !isValidMainFunction(line) {
				errors = append(errors, "Línea "+lineNum+": Declaración de main incorrecta")
				isValid = false
			}
		}

		// Verificar declaraciones de variables
		if isVariableDeclaration(line) {
			if !isValidVariableDeclaration(line) {
				errors = append(errors, "Línea "+lineNum+": Declaración de variable incorrecta")
				isValid = false
			}
		}

		// Verificar estructuras de control
		if isControlStructure(line) {
			if !isValidControlStructure(line) {
				errors = append(errors, "Línea "+lineNum+": Estructura de control mal formada")
				isValid = false
			}
		}

		// Verificar statements que deben terminar en punto y coma
		if needsSemicolon(line) && !strings.HasSuffix(line, ";") && !strings.HasSuffix(line, "{") {
			errors = append(errors, "Línea "+lineNum+": Falta punto y coma")
			isValid = false
		}

		// Contar delimitadores
		braceStack += strings.Count(line, "{") - strings.Count(line, "}")
		parenStack += strings.Count(line, "(") - strings.Count(line, ")")
		bracketStack += strings.Count(line, "[") - strings.Count(line, "]")

		// Verificar balance de delimitadores en cada línea para paréntesis
		lineParens := strings.Count(line, "(") - strings.Count(line, ")")
		if lineParens != 0 && !strings.Contains(line, "if") && !strings.Contains(line, "while") && 
		   !strings.Contains(line, "for") && !strings.Contains(line, "main") {
			// Verificar si los paréntesis están balanceados en la línea
			if strings.Count(line, "(") != strings.Count(line, ")") {
				errors = append(errors, "Línea "+lineNum+": Paréntesis desbalanceados")
				isValid = false
			}
		}
	}

	// Verificar balance final de delimitadores
	if braceStack != 0 {
		errors = append(errors, "Error: Llaves desbalanceadas en el código")
		isValid = false
	}
	if parenStack != 0 {
		errors = append(errors, "Error: Paréntesis desbalanceados en el código")
		isValid = false
	}
	if bracketStack != 0 {
		errors = append(errors, "Error: Corchetes desbalanceados en el código")
		isValid = false
	}

	// Verificar si tiene función main
	if !hasMain {
		errors = append(errors, "Error: No se encontró la función main")
		isValid = false
	}

	return models.SyntaxResult{IsValid: isValid, Errors: errors}
}

func isValidInclude(line string) bool {
	// #include <iostream> o #include "archivo.h"
	patterns := []string{
		`^#include\s*<[^>]+>\s*$`,
		`^#include\s*"[^"]+"\s*$`,
	}
	
	for _, pattern := range patterns {
		if matched, _ := regexp.MatchString(pattern, line); matched {
			return true
		}
	}
	return false
}

func isValidUsing(line string) bool {
	// using namespace std;
	pattern := `^using\s+namespace\s+\w+\s*;\s*$`
	matched, _ := regexp.MatchString(pattern, line)
	return matched
}

func isValidMainFunction(line string) bool {
	// int main() o int main(int argc, char* argv[])
	patterns := []string{
		`int\s+main\s*\(\s*\)\s*\{?`,
		`int\s+main\s*\(\s*int\s+\w+\s*,\s*char\s*\*\s*\w+\[\]\s*\)\s*\{?`,
		`int\s+main\s*\(\s*int\s+\w+\s*,\s*char\s*\*\*\s*\w+\s*\)\s*\{?`,
	}
	
	for _, pattern := range patterns {
		if matched, _ := regexp.MatchString(pattern, line); matched {
			return true
		}
	}
	return false
}

func isVariableDeclaration(line string) bool {
	// Tipos básicos de C++
	types := []string{"int", "float", "double", "char", "bool", "string"}
	
	for _, dataType := range types {
		pattern := `^\s*` + dataType + `\s+[a-zA-Z_][a-zA-Z0-9_]*`
		if matched, _ := regexp.MatchString(pattern, line); matched {
			return true
		}
	}
	return false
}

func isValidVariableDeclaration(line string) bool {
	// Patrones más específicos y correctos para C++
	patterns := []string{
		// Declaración simple: int var;
		`^\s*(int|float|double|char|bool|string)\s+[a-zA-Z_][a-zA-Z0-9_]*\s*;\s*$`,
		// Declaración con inicialización: int var = value;
		`^\s*(int|float|double|char|bool|string)\s+[a-zA-Z_][a-zA-Z0-9_]*\s*=\s*[^;]+\s*;\s*$`,
		// Múltiples declaraciones: int a, b;
		`^\s*(int|float|double|char|bool|string)\s+[a-zA-Z_][a-zA-Z0-9_]*(\s*,\s*[a-zA-Z_][a-zA-Z0-9_]*)*\s*;\s*$`,
		// Declaración con inicialización múltiple: int a = 1, b = 2;
		`^\s*(int|float|double|char|bool|string)\s+[a-zA-Z_][a-zA-Z0-9_]*\s*=\s*[^,;]+(\s*,\s*[a-zA-Z_][a-zA-Z0-9_]*\s*=\s*[^,;]+)*\s*;\s*$`,
	}
	
	for _, pattern := range patterns {
		if matched, _ := regexp.MatchString(pattern, line); matched {
			return true
		}
	}
	return false
}

func isControlStructure(line string) bool {
	controlKeywords := []string{"if", "else", "while", "for", "switch", "case", "default"}
	
	for _, keyword := range controlKeywords {
		if strings.Contains(line, keyword) {
			return true
		}
	}
	return false
}

func isValidControlStructure(line string) bool {
	// Verificar estructuras de control básicas
	patterns := []string{
		`^\s*if\s*\(.+\)\s*\{?\s*$`,
		`^\s*else\s*\{?\s*$`,
		`^\s*else\s+if\s*\(.+\)\s*\{?\s*$`,
		`^\s*while\s*\(.+\)\s*\{?\s*$`,
		`^\s*for\s*\(.+\)\s*\{?\s*$`,
		`^\s*switch\s*\(.+\)\s*\{?\s*$`,
		`^\s*case\s+.+:\s*$`,
		`^\s*default\s*:\s*$`,
	}
	
	for _, pattern := range patterns {
		if matched, _ := regexp.MatchString(pattern, line); matched {
			return true
		}
	}
	return false
}

func needsSemicolon(line string) bool {
	// Líneas que NO necesitan punto y coma
	if strings.HasPrefix(line, "#") || 
	   strings.HasSuffix(line, "{") || 
	   strings.HasSuffix(line, "}") ||
	   strings.Contains(line, "if") ||
	   strings.Contains(line, "else") ||
	   strings.Contains(line, "while") ||
	   strings.Contains(line, "for") ||
	   strings.Contains(line, "switch") ||
	   strings.Contains(line, "case") ||
	   strings.Contains(line, "default") ||
	   strings.HasPrefix(line, "using namespace") ||
	   line == "" {
		return false
	}
	
	// Si contiene assignment, cout, cin, return, etc. necesita ;
	statements := []string{"=", "cout", "cin", "return", "++", "--"}
	for _, stmt := range statements {
		if strings.Contains(line, stmt) {
			return true
		}
	}
	
	// Declaraciones de variables
	if isVariableDeclaration(line) {
		return true
	}
	
	return false
}