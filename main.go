package main

import (
	"log"
	"net/http"
	"sync/atomic"
)

const (
	rootPath = "."
	port     = "8080"
)

func main() {
	apiCfg := apiConfig{
		fileserverHits: atomic.Int32{},
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
