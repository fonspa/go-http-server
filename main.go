package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/fonspa/go-http-server/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

const (
	rootPath = "."
	port     = "8080"
)

func main() {
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	platform := os.Getenv("PLATFORM")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("unable to open the database: %v", err)
	}
	if err := db.Ping(); err != nil {
		log.Fatalf("unable to contact the DB: %v", err)
	}
	dbQueries := database.New(db)

	apiCfg := apiConfig{
		fileserverHits: atomic.Int32{},
		db:             dbQueries,
		platform:       platform,
	}

	mux := http.NewServeMux()
	// FileServer
	mux.Handle("/app/", http.StripPrefix("/app", apiCfg.middlewareMetricsInc(http.FileServer(http.Dir(rootPath)))))
	// API
	mux.HandleFunc("GET /api/healthz", handlerHealthz)
	mux.HandleFunc("POST /api/users", apiCfg.handlerCreateUser)
	mux.HandleFunc("POST /api/chirps", apiCfg.handlerCreateChirp)
	mux.HandleFunc("GET /api/chirps", apiCfg.handlerGetChirps)
	// ADMIN
	mux.HandleFunc("GET /admin/metrics", apiCfg.handlerDisplayMetrics)
	mux.HandleFunc("POST /admin/reset", apiCfg.handlerDeleteAllUsers)
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}
	// main func blocks until the server is shut down
	log.Fatal(srv.ListenAndServe())
}
