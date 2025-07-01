package services

import (
	"regexp"
	"strconv"
	"strings"
	"github.com/didiercito/api-go-examen2/models"
)

type Variable struct {
	Name         string
	DeclaredType string
	Value        string
	ActualType   string
	Line         int
}

func AnalyzeSemantic(code string) models.SemanticResult {
	lines := strings.Split(code, "\n")
	variables := 0
	functions := 0
	var errors []string
	var declaredVars []Variable
	var usedVars []string

	for lineNum, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		
		if strings.HasPrefix(line, "def ") {
			functions++
		}
		
		if strings.Contains(line, "=") && !strings.Contains(line, "==") && !strings.Contains(line, "!=") {
			variables++
			
			if hasExplicitType(line) {
				variable := parseTypedAssignment(line, lineNum+1)
				if variable.Name != "" {
					declaredVars = append(declaredVars, variable)
					
					if !isTypeCompatible(variable.DeclaredType, variable.ActualType) {
						errors = append(errors, 
							"Línea "+strconv.Itoa(lineNum+1)+": Error de tipo - Variable '"+variable.Name+
							"' declarada como "+variable.DeclaredType+" pero asignada valor "+variable.ActualType)
					}
				}
			} else {
				variable := parseNormalAssignment(line, lineNum+1)
				if variable.Name != "" {
					declaredVars = append(declaredVars, variable)
				}
			}
		}
		
		varsInLine := extractVariablesFromLine(line)
		usedVars = append(usedVars, varsInLine...)
	}
	
	for _, usedVar := range usedVars {
		if !isVariableDeclared(usedVar, declaredVars) && !isPythonBuiltin(usedVar) {
			errors = append(errors, "Variable '"+usedVar+"' usada pero no declarada")
		}
	}

	return models.SemanticResult{
		Variables: variables,
		Functions: functions,
		Errors:    errors,
	}
}

func hasExplicitType(line string) bool {
	pattern := "^\\s*\\w+\\s+(int|string|float|bool)\\s*="
	matched, _ := regexp.MatchString(pattern, line)
	return matched
}

func parseTypedAssignment(line string, lineNum int) Variable {
	pattern := "^\\s*(\\w+)\\s+(int|string|float|bool)\\s*=\\s*(.+)$"
	re := regexp.MustCompile(pattern)
	matches := re.FindStringSubmatch(line)
	
	if len(matches) == 4 {
		name := strings.TrimSpace(matches[1])
		declaredType := strings.TrimSpace(matches[2])
		value := strings.TrimSpace(matches[3])
		actualType := inferType(value)
		
		return Variable{
			Name:         name,
			DeclaredType: declaredType,
			Value:        value,
			ActualType:   actualType,
			Line:         lineNum,
		}
	}
	
	return Variable{}
}

func parseNormalAssignment(line string, lineNum int) Variable {
	parts := strings.Split(line, "=")
	if len(parts) >= 2 {
		name := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		actualType := inferType(value)
		
		return Variable{
			Name:         name,
			DeclaredType: actualType,
			Value:        value,
			ActualType:   actualType,
			Line:         lineNum,
		}
	}
	
	return Variable{}
}

func inferType(value string) string {
	value = strings.TrimSpace(value)
	
	if (strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"")) ||
	   (strings.HasPrefix(value, "'") && strings.HasSuffix(value, "'")) {
		return "string"
	}
	
	matched, _ := regexp.MatchString("^\\d+$", value)
	if matched {
		return "int"
	}
	
	matched, _ = regexp.MatchString("^\\d+\\.\\d+$", value)
	if matched {
		return "float"
	}
	
	if value == "True" || value == "False" {
		return "bool"
	}
	
	return "identifier"
}

func isTypeCompatible(declared, actual string) bool {
	if declared == actual {
		return true
	}
	
	switch declared {
	case "int":
		return actual == "int"
	case "string":
		return actual == "string"
	case "float":
		return actual == "float" || actual == "int"
	case "bool":
		return actual == "bool"
	}
	
	return false
}

func extractVariablesFromLine(line string) []string {
	var variables []string
	
	re := regexp.MustCompile("\\b[a-zA-Z_][a-zA-Z0-9_]*\\b")
	matches := re.FindAllString(line, -1)
	
	for _, match := range matches {
		if !isPythonKeyword(match) && !isPythonBuiltin(match) {
			variables = append(variables, match)
		}
	}
	
	return variables
}

func isVariableDeclared(varName string, declaredVars []Variable) bool {
	for _, v := range declaredVars {
		if v.Name == varName {
			return true
		}
	}
	return false
}

func isPythonKeyword(word string) bool {
	keywords := []string{
		"def", "if", "else", "elif", "while", "for", "in", "return", 
		"print", "True", "False", "None", "and", "or", "not", "import", 
		"from", "as", "class", "try", "except", "finally", "with", 
		"lambda", "pass", "break", "continue", "global", "nonlocal",
	}
	
	for _, keyword := range keywords {
		if word == keyword {
			return true
		}
	}
	return false
}

func isPythonBuiltin(word string) bool {
	builtins := []string{
		"print", "len", "str", "int", "float", "list", "dict", "tuple", 
		"set", "range", "enumerate", "zip", "map", "filter", "sum", 
		"min", "max", "abs", "round", "type", "isinstance", "lower", "upper",
	}
	
	for _, builtin := range builtins {
		if word == builtin {
			return true
		}
	}
	return false
}