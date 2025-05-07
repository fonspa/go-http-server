package main

import (
	"encoding/json"
	"net/http"
	"slices"
	"strings"
)

const (
	chirpMaxLen = 140
)

type chirp struct {
	Body string `json:"body"`
}

type validResponse struct {
	CleanedBody string `json:"cleaned_body"`
}

type errorResponse struct {
	Error string `json:"error"`
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

func respondWithError(w http.ResponseWriter, code int, msg string) {
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
	cleanedBody := removeProfaneWords(chirp.Body)
	respondWithJSON(w, http.StatusOK, validResponse{CleanedBody: cleanedBody})
}

func removeProfaneWords(msg string) string {
	profaneWords := []string{"kerfuffle", "sharbert", "fornax"}
	ret := []string{}
	fields := strings.Fields(msg)
	for _, w := range fields {
		if slices.Contains(profaneWords, strings.ToLower(w)) {
			ret = append(ret, "****")
		} else {
			ret = append(ret, w)
		}
	}
	return strings.Join(ret, " ")
}
