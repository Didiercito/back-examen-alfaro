package services

import (
	"regexp"
	"strings"
	"github.com/didiercito/api-go-examen2/models"
)

func AnalyzeLexical(code string) models.LexicalResult {
	summary := map[string]int{
		"PR": 0, "ID": 0, "Numeros": 0, "Simbolos": 0, "Error": 0,
	}

	keywords := []string{"def", "if", "else", "print", "True", "False"}
	
	// Tokenizar básico
	tokens := strings.Fields(strings.ReplaceAll(code, "\n", " "))
	
	for _, token := range tokens {
		token = strings.TrimSpace(token)
		if token == "" {
			continue
		}
		
		// Separar símbolos del token
		cleanToken := regexp.MustCompile(`[^a-zA-Z0-9_]`).ReplaceAllString(token, "")
		symbols := regexp.MustCompile(`[^a-zA-Z0-9_\s]`).FindAllString(token, -1)
		
		// Contar símbolos
		summary["Simbolos"] += len(symbols)
		
		if cleanToken != "" {
			if contains(keywords, cleanToken) {
				summary["PR"]++
			} else if matched, _ := regexp.MatchString(`^\d+$`, cleanToken); matched {
				summary["Numeros"]++
			} else if matched, _ := regexp.MatchString(`^[a-zA-Z_][a-zA-Z0-9_]*$`, cleanToken); matched {
				summary["ID"]++
			} else {
				summary["Error"]++
			}
		}
	}

	total := summary["PR"] + summary["ID"] + summary["Numeros"] + summary["Simbolos"] + summary["Error"]
	
	return models.LexicalResult{Summary: summary, Total: total}
}

// Función auxiliar
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}