package main

import (
	"log"
	"net/http"
	"github.com/didiercito/api-go-examen2/handlers"
	"github.com/gorilla/mux"
)

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/analyze", handlers.AnalyzeCode).Methods("POST", "OPTIONS")
	
	log.Println("🚀 Servidor iniciado en puerto 8080")
	log.Println("📡 Endpoint: POST /analyze")
	log.Fatal(http.ListenAndServe(":8080", r))
}