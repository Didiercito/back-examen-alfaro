package services

import (
	"regexp"
	"strconv"
	"strings"
	"github.com/didiercito/api-go-examen2/models"
)

type CppVariable struct {
	Name         string
	Type         string
	Value        string
	Line         int
	IsInitialized bool
}

type CppFunction struct {
	Name       string
	ReturnType string
	Parameters []string
	Line       int
}

func AnalyzeSemantic(code string) models.SemanticResult {
	lines := strings.Split(code, "\n")
	variables := 0
	functions := 0
	var errors []string
	var declaredVars []CppVariable
	var declaredFuncs []CppFunction
	var usedVars []string

	for lineNum, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "//") || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "using namespace") {
			continue
		}
		
		// Analizar declaraciones de funciones
		if isFunctionDeclaration(line) {
			functions++
			function := parseFunctionDeclaration(line, lineNum+1)
			if function.Name != "" {
				declaredFuncs = append(declaredFuncs, function)
			}
		}
		
		// Analizar declaraciones de variables
		if isVariableDeclaration(line) {
			variableCount := countVariablesInDeclaration(line)
			variables += variableCount
			
			parsedVars := parseVariableDeclaration(line, lineNum+1)
			for _, variable := range parsedVars {
				if variable.Name != "" {
					// Verificar si ya existe la variable
					if isVariableAlreadyDeclared(variable.Name, declaredVars) {
						errors = append(errors, 
							"Línea "+strconv.Itoa(lineNum+1)+": Variable '"+variable.Name+
							"' ya fue declarada anteriormente")
					} else {
						declaredVars = append(declaredVars, variable)
						
						// Verificar compatibilidad de tipos en asignación
						if variable.IsInitialized {
							if !isTypeCompatible(variable.Type, variable.Value) {
								errors = append(errors, 
									"Línea "+strconv.Itoa(lineNum+1)+": Error de tipo - Variable '"+variable.Name+
									"' de tipo "+variable.Type+" no puede ser asignada con valor de tipo "+inferValueType(variable.Value))
							}
						}
					}
				}
			}
		}
		
		// Analizar asignaciones a variables existentes
		if isAssignment(line) && !isVariableDeclaration(line) {
			varName, value := parseAssignment(line)
			if varName != "" {
				if !isVariableAlreadyDeclared(varName, declaredVars) {
					errors = append(errors, 
						"Línea "+strconv.Itoa(lineNum+1)+": Variable '"+varName+
						"' usada pero no declarada")
				} else {
					// Verificar compatibilidad de tipos
					varType := getVariableType(varName, declaredVars)
					if !isTypeCompatible(varType, value) {
						errors = append(errors, 
							"Línea "+strconv.Itoa(lineNum+1)+": Error de tipo - No se puede asignar "+
							inferValueType(value)+" a variable de tipo "+varType)
					}
				}
			}
		}
		
		// Extraer variables usadas en la línea (excluyendo string literals)
		varsInLine := extractVariablesFromLine(line, declaredFuncs)
		usedVars = append(usedVars, varsInLine...)
	}
	
	// Verificar variables usadas pero no declaradas
	uniqueUsedVars := removeDuplicates(usedVars)
	for _, usedVar := range uniqueUsedVars {
		if !isVariableAlreadyDeclared(usedVar, declaredVars) && 
		   !isCppBuiltinOrKeyword(usedVar) && 
		   !isFunctionName(usedVar, declaredFuncs) {
			errors = append(errors, "Variable '"+usedVar+"' usada pero no declarada")
		}
	}

	// Verificar variables declaradas pero no usadas (excluyendo main que es especial)
	for _, declaredVar := range declaredVars {
		if !contains(usedVars, declaredVar.Name) && declaredVar.Name != "main" {
			errors = append(errors, "Variable '"+declaredVar.Name+"' declarada pero no usada")
		}
	}

	return models.SemanticResult{
		Variables: variables,
		Functions: functions,
		Errors:    errors,
	}
}

func countVariablesInDeclaration(line string) int {
	// Contar cuántas variables se declaran en una línea
	// Ejemplo: int a, b, c; = 3 variables
	if !isVariableDeclaration(line) {
		return 0
	}
	
	// Remover el tipo y quedarnos con las variables
	re := regexp.MustCompile(`^\s*(int|float|double|char|bool|string)\s+(.+);?\s*$`)
	matches := re.FindStringSubmatch(line)
	
	if len(matches) >= 3 {
		varPart := matches[2]
		varPart = strings.TrimSuffix(varPart, ";")
		
		// Dividir por comas para contar variables
		vars := strings.Split(varPart, ",")
		return len(vars)
	}
	
	return 1
}

func isFunctionDeclaration(line string) bool {
	// Buscar patrones de declaración de función
	patterns := []string{
		`(int|void|float|double|char|bool|string)\s+\w+\s*\([^)]*\)\s*\{?`,
	}
	
	for _, pattern := range patterns {
		if matched, _ := regexp.MatchString(pattern, line); matched {
			return true
		}
	}
	return false
}

func parseFunctionDeclaration(line string, lineNum int) CppFunction {
	// Extraer información de la función
	re := regexp.MustCompile(`(int|void|float|double|char|bool|string)\s+(\w+)\s*\(([^)]*)\)`)
	matches := re.FindStringSubmatch(line)
	
	if len(matches) >= 3 {
		returnType := strings.TrimSpace(matches[1])
		name := strings.TrimSpace(matches[2])
		params := strings.TrimSpace(matches[3])
		
		var parameters []string
		if params != "" {
			parameters = strings.Split(params, ",")
			for i, param := range parameters {
				parameters[i] = strings.TrimSpace(param)
			}
		}
		
		return CppFunction{
			Name:       name,
			ReturnType: returnType,
			Parameters: parameters,
			Line:       lineNum,
		}
	}
	
	return CppFunction{}
}

func parseVariableDeclaration(line string, lineNum int) []CppVariable {
	var variables []CppVariable
	
	// Patrón para declaraciones: tipo var1 = val1, var2 = val2;
	re := regexp.MustCompile(`^\s*(int|float|double|char|bool|string)\s+(.+);?\s*$`)
	matches := re.FindStringSubmatch(line)
	
	if len(matches) >= 3 {
		varType := strings.TrimSpace(matches[1])
		varsPart := strings.TrimSpace(matches[2])
		varsPart = strings.TrimSuffix(varsPart, ";")
		
		// Dividir por comas para manejar múltiples variables
		varDecls := strings.Split(varsPart, ",")
		
		for _, varDecl := range varDecls {
			varDecl = strings.TrimSpace(varDecl)
			
			if strings.Contains(varDecl, "=") {
				// Variable con inicialización
				parts := strings.SplitN(varDecl, "=", 2)
				name := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])
				
				variables = append(variables, CppVariable{
					Name:          name,
					Type:          varType,
					Value:         value,
					Line:          lineNum,
					IsInitialized: true,
				})
			} else {
				// Variable sin inicialización
				name := strings.TrimSpace(varDecl)
				variables = append(variables, CppVariable{
					Name:          name,
					Type:          varType,
					Value:         "",
					Line:          lineNum,
					IsInitialized: false,
				})
			}
		}
	}
	
	return variables
}

func isAssignment(line string) bool {
	// Buscar patrones de asignación (excluyendo declaraciones)
	return strings.Contains(line, "=") && 
		   !strings.Contains(line, "==") && 
		   !strings.Contains(line, "!=") &&
		   !strings.Contains(line, "<=") &&
		   !strings.Contains(line, ">=") &&
		   !isVariableDeclaration(line)
}

func parseAssignment(line string) (string, string) {
	if !strings.Contains(line, "=") {
		return "", ""
	}
	
	parts := strings.SplitN(line, "=", 2)
	if len(parts) >= 2 {
		varName := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		value = strings.TrimSuffix(value, ";")
		
		// Extraer solo el nombre de la variable (sin tipo)
		varParts := strings.Fields(varName)
		if len(varParts) > 0 {
			varName = varParts[len(varParts)-1]
		}
		
		return varName, value
	}
	
	return "", ""
}

func isTypeCompatible(varType, value string) bool {
	// Si es un string literal, no validar como variable
	if isStringLiteral(value) {
		return varType == "string"
	}
	
	valueType := inferValueType(value)
	
	switch varType {
	case "int":
		return valueType == "int"
	case "float", "double":
		return valueType == "float" || valueType == "int"
	case "char":
		return valueType == "char"
	case "bool":
		return valueType == "bool"
	case "string":
		return valueType == "string"
	}
	
	return false
}

func isStringLiteral(value string) bool {
	value = strings.TrimSpace(value)
	return (strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\""))
}

func inferValueType(value string) string {
	value = strings.TrimSpace(value)
	
	// String literal
	if isStringLiteral(value) {
		return "string"
	}
	
	// Character literal
	if (strings.HasPrefix(value, "'") && strings.HasSuffix(value, "'")) {
		return "char"
	}
	
	// Boolean
	if value == "true" || value == "false" {
		return "bool"
	}
	
	// Integer
	if matched, _ := regexp.MatchString(`^\d+$`, value); matched {
		return "int"
	}
	
	// Float
	if matched, _ := regexp.MatchString(`^\d+\.\d+$`, value); matched {
		return "float"
	}
	
	return "identifier"
}

func isVariableAlreadyDeclared(varName string, declaredVars []CppVariable) bool {
	for _, v := range declaredVars {
		if v.Name == varName {
			return true
		}
	}
	return false
}

func getVariableType(varName string, declaredVars []CppVariable) string {
	for _, v := range declaredVars {
		if v.Name == varName {
			return v.Type
		}
	}
	return ""
}

func isFunctionName(name string, functions []CppFunction) bool {
	for _, f := range functions {
		if f.Name == name {
			return true
		}
	}
	return false
}

func extractVariablesFromLine(line string, declaredFuncs []CppFunction) []string {
	var variables []string
	
	// Remover string literals para no analizarlos
	cleanLine := removeStringLiteralsForSemantic(line)
	
	// Extraer identificadores (posibles variables)
	re := regexp.MustCompile(`\b[a-zA-Z_][a-zA-Z0-9_]*\b`)
	matches := re.FindAllString(cleanLine, -1)
	
	for _, match := range matches {
		if !isCppBuiltinOrKeyword(match) && 
		   !isFunctionName(match, declaredFuncs) && 
		   match != "main" { // main es función especial
			variables = append(variables, match)
		}
	}
	
	return variables
}

func removeStringLiteralsForSemantic(line string) string {
	// Remover string literals para análisis semántico
	stringRegex := regexp.MustCompile(`"[^"]*"`)
	return stringRegex.ReplaceAllString(line, `""`)
}

func removeDuplicates(slice []string) []string {
	keys := make(map[string]bool)
	var result []string
	
	for _, item := range slice {
		if !keys[item] {
			keys[item] = true
			result = append(result, item)
		}
	}
	
	return result
}

func isCppBuiltinOrKeyword(word string) bool {
	keywords := []string{
		"int", "float", "double", "char", "bool", "void", "string",
		"if", "else", "while", "for", "do", "switch", "case", "default",
		"break", "continue", "return", "goto", "sizeof", "typedef",
		"struct", "union", "enum", "class", "public", "private", "protected",
		"virtual", "static", "const", "volatile", "extern", "register",
		"auto", "signed", "unsigned", "long", "short", "inline",
		"template", "typename", "namespace", "using", "new", "delete",
		"this", "try", "catch", "throw", "true", "false", "nullptr",
		"main", "std", "cout", "cin", "endl", "include", "define",
	}
	
	for _, keyword := range keywords {
		if word == keyword {
			return true
		}
	}
	return false
}

// Función auxiliar reutilizada
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}