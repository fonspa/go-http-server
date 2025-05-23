package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/fonspa/go-http-server/internal/auth"
	"github.com/fonspa/go-http-server/internal/database"
	"github.com/google/uuid"
)

const (
	chirpMaxLen = 140
)

type chirpPayload struct {
	Body   string `json:"body"`
	UserID string `json:"user_id"`
}

type chirpResponse struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
}

func (cfg *apiConfig) handlerCreateChirp(w http.ResponseWriter, r *http.Request) {
	var chirp chirpPayload
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&chirp); err != nil {
		log.Printf("Unable to decode request: %v", err)
		respondWithError(w, http.StatusInternalServerError, "unable to decode request")
		return
	}
	tokenString, err := auth.GetBearerToken(r.Header)
	if err != nil {
		log.Printf("unable to get user bearer token: %v", err)
		respondWithError(w, http.StatusUnauthorized, "unable to get bearer token")
		return
	}
	userID, err := auth.ValidateJWT(tokenString, cfg.jwtSecret)
	if err != nil {
		log.Printf("unable to validate user's JWT: %v", err)
		respondWithError(w, http.StatusUnauthorized, "unable to validate user's JWT")
		return
	}
	cleanedMsg, err := validateChirp(chirp.Body)
	if err != nil {
		log.Printf("chirp invalid: %v", err)
		respondWithError(w, http.StatusBadRequest, err.Error())
	}
	userChirp, err := cfg.db.CreateChirp(r.Context(), database.CreateChirpParams{
		Body:   cleanedMsg,
		UserID: userID,
	})
	if err != nil {
		log.Printf("Unable to create chirp: %v", err)
		respondWithError(w, http.StatusInternalServerError, "unable to create chirp")
	}
	respondWithJSON(w, http.StatusCreated, chirpResponse{
		ID:        userChirp.ID,
		CreatedAt: userChirp.CreatedAt,
		UpdatedAt: userChirp.UpdatedAt,
		Body:      userChirp.Body,
		UserID:    userChirp.UserID,
	})
}

func (cfg *apiConfig) handlerGetChirps(w http.ResponseWriter, r *http.Request) {
	var chirps []database.Chirp
	var err error
	author_id := r.URL.Query().Get("author_id")
	if author_id == "" {
		chirps, err = cfg.db.GetAllChirps(r.Context())
		if err != nil {
			log.Printf("unable to retrieve chirps from db: %v", err)
			respondWithError(w, http.StatusInternalServerError, "unable to retrieve chirps")
			return
		}
	} else {
		userID, err := uuid.Parse(author_id)
		if err != nil {
			log.Printf("Invalid user ID: %v", err)
			respondWithError(w, http.StatusBadRequest, "invalid user ID")
			return
		}
		// Return only the chirps for the user ID author_id
		chirps, err = cfg.db.GetChirpsByUserID(r.Context(), userID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				log.Printf("user not found: %v", err)
				respondWithError(w, http.StatusNotFound, "user not found")
				return
			}
			respondWithError(w, http.StatusInternalServerError, "unable to retrieve chirps for that user")
			return
		}
	}
	sortOrder := r.URL.Query().Get("sort")
	if sortOrder == "desc" {
		slices.SortFunc(chirps, func(a, b database.Chirp) int {
			if a.CreatedAt.Before(b.CreatedAt) {
				return 1
			} else if a.CreatedAt.After(b.CreatedAt) {
				return -1
			}
			return 0
		})
	}
	var resp []chirpResponse
	for _, c := range chirps {
		resp = append(resp, chirpResponse{
			ID:        c.ID,
			CreatedAt: c.CreatedAt,
			UpdatedAt: c.UpdatedAt,
			Body:      c.Body,
			UserID:    c.UserID,
		})
	}
	respondWithJSON(w, http.StatusOK, resp)
}

func (cfg *apiConfig) handlerGetChirp(w http.ResponseWriter, r *http.Request) {
	chirpID := r.PathValue("chirpID")
	if chirpID == "" {
		log.Printf("User given chirp ID is empty")
		respondWithError(w, http.StatusBadRequest, "chirp ID invalid")
		return
	}
	id, err := uuid.Parse(chirpID)
	if err != nil {
		log.Printf("unable to parse chirp ID '%s': %v", chirpID, err)
		respondWithError(w, http.StatusBadRequest, "chirp ID invalid")
		return
	}
	dbChirp, err := cfg.db.GetChirpByID(r.Context(), id)
	if err != nil {
		log.Printf("unable to retrieve chirp from db: %v", err)
		respondWithError(w, http.StatusNotFound, "unable to retrieve chirp")
		return
	}
	respondWithJSON(w, http.StatusOK, chirpResponse{
		ID:        dbChirp.ID,
		CreatedAt: dbChirp.CreatedAt,
		UpdatedAt: dbChirp.UpdatedAt,
		Body:      dbChirp.Body,
		UserID:    dbChirp.UserID,
	})
}

func validateChirp(msg string) (string, error) {
	if len(msg) > chirpMaxLen {
		return "", errors.New("chirp is too long, max size is 140 characters")
	}
	return removeProfaneWords(msg), nil
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

func (cfg *apiConfig) handlerDeleteChirp(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		log.Printf("unable to get bearer token: %v", err)
		respondWithError(w, http.StatusUnauthorized, "invalid bearer token")
		return
	}
	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		log.Printf("unable to validate JWT: %v", err)
		respondWithError(w, http.StatusUnauthorized, "invalid access token")
		return
	}
	chirpID := r.PathValue("chirpID")
	if chirpID == "" {
		respondWithError(w, http.StatusBadRequest, "invalid chirp ID")
		return
	}
	chirpUUID, err := uuid.Parse(chirpID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid chirp ID")
		return
	}
	dbChirp, err := cfg.db.GetChirpByID(r.Context(), chirpUUID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "chirp not found")
		return
	}
	if dbChirp.UserID != userID {
		log.Printf("mismatch between userID and chirpID on deletion request")
		respondWithError(w, http.StatusForbidden, "unauthorized request")
		return
	}
	if err = cfg.db.DeleteChirp(r.Context(), dbChirp.ID); err != nil {
		log.Printf("unable to delete chirp by id: %v", err)
		respondWithError(w, http.StatusInternalServerError, "unable to delete chirp from DB")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
