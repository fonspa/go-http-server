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
	if dbURL == "" {
		log.Fatalf("you must provide a DB_URL")
	}
	platform := os.Getenv("PLATFORM")
	if platform == "" {
		log.Fatalf("you must provide a PLATFORM")
	}
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatalf("you must provide a JWT_SECRET")
	}

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
		jwtSecret:      jwtSecret,
	}

	mux := http.NewServeMux()
	// FileServer
	mux.Handle("/app/", http.StripPrefix("/app", apiCfg.middlewareMetricsInc(http.FileServer(http.Dir(rootPath)))))
	// API GET
	mux.HandleFunc("GET /api/healthz", handlerHealthz)
	mux.HandleFunc("GET /api/chirps", apiCfg.handlerGetChirps)
	mux.HandleFunc("GET /api/chirps/{chirpID}", apiCfg.handlerGetChirp)
	// API POST
	mux.HandleFunc("POST /api/users", apiCfg.handlerCreateUser)
	mux.HandleFunc("POST /api/chirps", apiCfg.handlerCreateChirp)
	mux.HandleFunc("POST /api/login", apiCfg.handlerLogin)
	mux.HandleFunc("POST /api/refresh", apiCfg.handlerRefreshToken)
	mux.HandleFunc("POST /api/revoke", apiCfg.handlerRevokeRefreshToken)
	// API PUT
	mux.HandleFunc("PUT /api/users", apiCfg.handlerUpdateUser)
	// API DELETE
	mux.HandleFunc("DELETE /api/chirps/{chirpID}", apiCfg.handlerDeleteChirp)
	// ADMIN GET
	mux.HandleFunc("GET /admin/metrics", apiCfg.handlerDisplayMetrics)
	// ADMIN POST
	mux.HandleFunc("POST /admin/reset", apiCfg.handlerDeleteAllUsers)
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}
	// main func blocks until the server is shut down
	log.Fatal(srv.ListenAndServe())
}
