package models

type AnalysisResult struct {
	LexicalAnalysis  LexicalResult  `json:"lexical_analysis"`
	SyntaxAnalysis   SyntaxResult   `json:"syntax_analysis"`
	SemanticAnalysis SemanticResult `json:"semantic_analysis"`
}

type LexicalResult struct {
	Summary map[string]int `json:"summary"`
	Total   int            `json:"total"`
}

type SyntaxResult struct {
	IsValid bool     `json:"is_valid"`
	Errors  []string `json:"errors"`
}

type SemanticResult struct {
	Variables int      `json:"variables_count"`
	Functions int      `json:"functions_count"`
	Errors    []string `json:"errors"`
}