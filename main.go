package main

import (
	"flag"
	"internal/database"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
)

type apiConfig struct {
	fileserverHits int
	database       *database.DB
	jwtSecret      string
}

func main() {
	godotenv.Load()

	const filepathRoot = "."
	const port = ":8080"
	const databasePath = "./database.json"
	jwtSecret := os.Getenv("JST_SECRET")

	db, err := database.NewDB("database.json")
	if err != nil {
		log.Fatal(err)
	}

	dbg := flag.Bool("debug", false, "Enable debug mode")
	flag.Parse()
	if dbg != nil && *dbg {
		log.Printf("Debug mode on. Will remove previous database file")
		err := db.ResetDB()
		if err != nil {
			log.Fatal(err)
		}
	}

	apiCfg := apiConfig{
		fileserverHits: 0,
		database:       db,
		jwtSecret:      jwtSecret,
	}

	r := chi.NewRouter()

	fsHandler := apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(filepathRoot))))
	r.Handle("/app", fsHandler)
	r.Handle("/app/*", fsHandler)

	apiRouter := chi.NewRouter()
	apiRouter.Get("/healthz", healthCheck)
	apiRouter.Get("/reset", apiCfg.resetHandler)
	apiRouter.Post("/chirps", apiCfg.postChirpsHandler)
	apiRouter.Get("/chirps", apiCfg.getChirpsHandler)
	apiRouter.Get("/chirps/{chirpID}", apiCfg.getSingleChirpHandler)
	apiRouter.Delete("/chirps/{chirpID}", apiCfg.deleteSingleChirpHandler)
	apiRouter.Post("/users", apiCfg.createUserHandler)
	apiRouter.Put("/users", apiCfg.updateUserHandler)
	apiRouter.Post("/login", apiCfg.login)
	apiRouter.Post("/refresh", apiCfg.refreshJWTHandler)
	apiRouter.Post("/revoke", apiCfg.revokeJWTHandler)
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
