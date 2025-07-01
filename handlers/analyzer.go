package handlers

import (
	"encoding/json"
	"net/http"
	"github.com/didiercito/api-go-examen2/models"
	"github.com/didiercito/api-go-examen2/services"
)

func AnalyzeCode(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	var req models.CodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	result := models.AnalysisResult{
		LexicalAnalysis:  services.AnalyzeLexical(req.Code),
		SyntaxAnalysis:   services.AnalyzeSyntax(req.Code),
		SemanticAnalysis: services.AnalyzeSemantic(req.Code),
	}

	json.NewEncoder(w).Encode(result)
}