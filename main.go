package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type apiConfig struct {
	fileserverHits int
	databasePath   string
}

func main() {
	const filepathRoot = "."
	const port = ":8080"
	const databasePath = "./database.json"

	dbg := flag.Bool("debug", false, "Enable debug mode")
	flag.Parse()

	apiCfg := apiConfig{
		fileserverHits: 0,
		databasePath:   databasePath,
	}

	if dbg != nil && *dbg {
		log.Printf("Debug mode on. Will remove previous database file")
		cleanDatabaseJson(databasePath)
	}

	r := chi.NewRouter()

	fsHandler := apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(filepathRoot))))
	r.Handle("/app", fsHandler)
	r.Handle("/app/*", fsHandler)

	apiRouter := chi.NewRouter()
	apiRouter.Use(apiCfg.middlewareDB)
	apiRouter.Get("/healthz", healthCheck)
	apiRouter.Get("/reset", apiCfg.resetHandler)
	apiRouter.Post("/chirps", postChirpsHandler)
	apiRouter.Get("/chirps", getChirpsHandler)
	apiRouter.Get("/chirps/{chirpID}", getSingleChirpHandler)
	apiRouter.Post("/users", createUserHandler)
	apiRouter.Post("/login", loginHandler)
	r.Mount("/api", apiRouter)

	adminRouter := chi.NewRouter()
	adminRouter.Get("/metrics", apiCfg.metricsHandler)
	r.Mount("/admin", adminRouter)

	corsMux := middlewareCors(r)
	server := http.Server{
		Addr:    port,
		Handler: corsMux,
	}
	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
	log.Fatal(server.ListenAndServe())
}
