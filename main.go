package main

import (
	"fmt"
	"net/http"
)

type apiConfig struct {
	fileserverHits int
}

func main() {
	apiCfg := apiConfig{fileserverHits: 0}
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", healthCheck)
	mux.HandleFunc("/metrics", apiCfg.metricsHandler)
	mux.HandleFunc("/reset", apiCfg.resetHandler)
	mux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app/", http.FileServer(http.Dir(".")))))
	corsMux := middlewareCors(mux)
	server := http.Server{
		Addr:    ":8080",
		Handler: corsMux,
	}
	server.ListenAndServe()
}

func middlewareCors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func healthCheck(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)
	w.Write([]byte("OK"))
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits++
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) metricsHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(fmt.Sprintf("Hits: %v", cfg.fileserverHits)))
}

func (cfg *apiConfig) resetHandler(w http.ResponseWriter, r *http.Request) {
	cfg.fileserverHits = 0
	w.WriteHeader(200)
}
