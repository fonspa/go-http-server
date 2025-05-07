package main

import (
	"database/sql"
	"fmt"
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
	fmt.Println("db URL: ", dbURL)
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
	}

	mux := http.NewServeMux()
	mux.Handle("/app/", http.StripPrefix("/app", apiCfg.middlewareMetricsInc(http.FileServer(http.Dir(rootPath)))))
	mux.HandleFunc("GET /api/healthz", handlerHealthz)
	mux.HandleFunc("GET /admin/metrics", apiCfg.handlerDisplayMetrics)
	mux.HandleFunc("POST /admin/reset", apiCfg.handlerResetMetrics)
	mux.HandleFunc("POST /api/validate_chirp", handlerValidateChirp)
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}
	// main func blocks until the server is shut down
	log.Fatal(srv.ListenAndServe())
}
