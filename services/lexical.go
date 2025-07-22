package services

import (
	"regexp"
	"strings"
	"github.com/didiercito/api-go-examen2/models"
)

func AnalyzeLexical(code string) models.LexicalResult {
	summary := map[string]int{
		"PR": 0,      // Palabras reservadas
		"ID": 0,      // Identificadores
		"Numeros": 0, // Números
		"Simbolos": 0,// Símbolos y operadores
		"Error": 0,   // Errores léxicos
	}

	// Palabras reservadas de C++
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

	// Remover comentarios primero
	cleanCode := removeComments(code)
	
	// Contar string literals primero y removerlos del análisis posterior
	stringLiterals := countAndRemoveStringLiterals(cleanCode)
	summary["Simbolos"] += stringLiterals.count
	codeWithoutStrings := stringLiterals.cleanCode

	// Expresiones regulares para diferentes tipos de tokens
	regexes := map[string]*regexp.Regexp{
		"numbers":     regexp.MustCompile(`\b\d+(\.\d+)?\b`),
		"identifiers": regexp.MustCompile(`\b[a-zA-Z_][a-zA-Z0-9_]*\b`),
		"symbols":     regexp.MustCompile(`[\+\-\*/=<>!&|%^~(){}\[\];,.:?]|<<|>>|<=|>=|==|!=|&&|\|\||\+\+|--|\+=|-=|\*=|/=|%=`),
		"chars":       regexp.MustCompile(`'[^']*'`),
	}

	// Contar números
	numbers := regexes["numbers"].FindAllString(codeWithoutStrings, -1)
	summary["Numeros"] = len(numbers)

	// Contar símbolos (excluyendo los ya contados en strings)
	symbols := regexes["symbols"].FindAllString(codeWithoutStrings, -1)
	summary["Simbolos"] += len(symbols)

	// Contar caracteres literales
	char_literals := regexes["chars"].FindAllString(codeWithoutStrings, -1)
	summary["Simbolos"] += len(char_literals)

	// Identificar palabras reservadas e identificadores
	identifiers := regexes["identifiers"].FindAllString(codeWithoutStrings, -1)
	
	// Crear sets para evitar duplicados
	foundKeywords := make(map[string]bool)
	foundIdentifiers := make(map[string]bool)
	
	for _, token := range identifiers {
		if containsString(keywords, strings.ToLower(token)) || containsString(keywords, token) {
			if !foundKeywords[token] {
				foundKeywords[token] = true
				summary["PR"]++
			}
		} else if isValidIdentifier(token) {
			if !foundIdentifiers[token] {
				foundIdentifiers[token] = true
				summary["ID"]++
			}
		} else {
			summary["Error"]++
		}
	}

	// Verificar errores léxicos básicos
	errorPatterns := []*regexp.Regexp{
		regexp.MustCompile(`\d+[a-zA-Z]+`), // Números seguidos de letras
		regexp.MustCompile(`"[^"]*$`),      // Strings sin cerrar
		regexp.MustCompile(`'[^']*$`),      // Chars sin cerrar
	}

	for _, pattern := range errorPatterns {
		errors := pattern.FindAllString(code, -1)
		summary["Error"] += len(errors)
	}

	total := summary["PR"] + summary["ID"] + summary["Numeros"] + summary["Simbolos"] + summary["Error"]
	
	return models.LexicalResult{Summary: summary, Total: total}
}

type StringLiteralResult struct {
	count     int
	cleanCode string
}

func countAndRemoveStringLiterals(code string) StringLiteralResult {
	// Contar y remover string literals
	stringRegex := regexp.MustCompile(`"[^"]*"`)
	matches := stringRegex.FindAllString(code, -1)
	cleanCode := stringRegex.ReplaceAllString(code, `""`)
	
	return StringLiteralResult{
		count:     len(matches),
		cleanCode: cleanCode,
	}
}

func removeComments(code string) string {
	// Remover comentarios de línea //
	lineCommentRegex := regexp.MustCompile(`//.*`)
	code = lineCommentRegex.ReplaceAllString(code, "")
	
	// Remover comentarios de bloque /* */
	blockCommentRegex := regexp.MustCompile(`/\*[\s\S]*?\*/`)
	code = blockCommentRegex.ReplaceAllString(code, "")
	
	return code
}

func isValidIdentifier(token string) bool {
	// Un identificador válido en C++ debe empezar con letra o _
	// y contener solo letras, números y _
	matched, _ := regexp.MatchString(`^[a-zA-Z_][a-zA-Z0-9_]*$`, token)
	return matched
}

// Función auxiliar para verificar si un item está en un slice
func containsString(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}