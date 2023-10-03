package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
)

type apiConfig struct {
	fileserverHits int
	databasePath   string
	jwtSecret      string
}

func main() {
	godotenv.Load()

	const filepathRoot = "."
	const port = ":8080"
	const databasePath = "./database.json"
	jwtSecret := os.Getenv("JST_SECRET")

	dbg := flag.Bool("debug", false, "Enable debug mode")
	flag.Parse()

	apiCfg := apiConfig{
		fileserverHits: 0,
		databasePath:   databasePath,
		jwtSecret:      jwtSecret,
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
	apiRouter.Post("/login", apiCfg.login)
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
