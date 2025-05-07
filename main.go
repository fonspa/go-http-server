package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
)

const (
	rootPath = "."
	port     = "8080"
)

type apiConfig struct {
	fileserverHits atomic.Int32
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) handlerDisplayMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write(fmt.Appendf(nil,
		`
		<html>
		<body>
			<h1>Welcome, Chirpy Admin</h1>
			<p>Chirpy has been visited %d times!</p>
		</body>
		</html>
		`,
		cfg.fileserverHits.Load()))
}

func (cfg *apiConfig) handlerResetMetrics(w http.ResponseWriter, r *http.Request) {
	cfg.fileserverHits.Store(0)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hits reset to 0"))
}

type Chirp struct {
	Body string `json:"body"`
}

func main() {
	apiCfg := apiConfig{
		fileserverHits: atomic.Int32{},
	}
	mux := http.NewServeMux()
	mux.Handle("/app/", http.StripPrefix("/app", apiCfg.middlewareMetricsInc(http.FileServer(http.Dir(rootPath)))))
	mux.HandleFunc("GET /api/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(http.StatusText(http.StatusOK)))
	})
	mux.HandleFunc("GET /admin/metrics", apiCfg.handlerDisplayMetrics)
	mux.HandleFunc("POST /admin/reset", apiCfg.handlerResetMetrics)
	mux.HandleFunc("POST /api/validate_chirp", func(w http.ResponseWriter, r *http.Request) {
		var chirp Chirp
		defer r.Body.Close()
		if err := json.NewDecoder(r.Body).Decode(&chirp); err != nil {
			w.WriteHeader(500)
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{
			"error":"Something went wrong"
			}`))
			return
		}
		if len(chirp.Body) > 140 {
			w.WriteHeader(400)
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{
			"error":"Chirp is too long"
			}`))
			return
		}
		w.WriteHeader(200)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
		"valid":true
		}`))
	})
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}
	// main func blocks until the server is shut down
	log.Fatal(srv.ListenAndServe())
}
