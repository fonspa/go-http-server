package main

import (
	"encoding/json"
	"net/http"
)

const (
	chirpMaxLen = 140
)

type chirp struct {
	Body string `json:"body"`
}

type valid struct {
	Valid bool `json:"valid"`
}

func respondWithJSON(w http.ResponseWriter, code int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	data, err := json.Marshal(payload)
	if err != nil {
		w.WriteHeader(500)
		return
	}
	w.WriteHeader(code)
	w.Write(data)
}

type errorResponse struct {
	Error string `json:"error"`
}

func respondWithError(w http.ResponseWriter, code int, msg string) {
	w.WriteHeader(code)
	w.Header().Set("Content-Type", "application/json")
	respondWithJSON(w, code, errorResponse{Error: msg})
}

func handlerValidateChirp(w http.ResponseWriter, r *http.Request) {
	var chirp chirp
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&chirp); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}
	if len(chirp.Body) > chirpMaxLen {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long")
		return
	}
	respondWithJSON(w, http.StatusOK, valid{Valid: true})
}
