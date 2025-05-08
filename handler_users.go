package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
)

type createUserPayload struct {
	Email string `json:"email"`
}

type createUserResponse struct {
	ID        string `json:"id"`
	CreateAt  string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
	Email     string `json:"email"`
}

func (cfg *apiConfig) handlerCreateUser(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var payload createUserPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		log.Printf("unable to decode request: %v", err)
		respondWithError(w, http.StatusInternalServerError, "unable to decode request")
		return
	}
	user, err := cfg.db.CreateUser(r.Context(), payload.Email)
	if err != nil {
		log.Printf("db error when creating new user: %v", err)
		respondWithError(w, http.StatusInternalServerError, "unable to create new user")
		return
	}
	respondWithJSON(w, http.StatusCreated, createUserResponse{
		ID:        user.ID.String(),
		CreateAt:  user.CreatedAt.Format(time.RFC3339),
		UpdatedAt: user.UpdatedAt.Format(time.RFC3339),
		Email:     user.Email,
	})
}

func (cfg *apiConfig) handlerDeleteAllUsers(w http.ResponseWriter, r *http.Request) {
	if cfg.platform != "dev" {
		log.Print("deleting users is only allowed in dev mode!")
		w.WriteHeader(http.StatusForbidden)
		return
	}
	if err := cfg.db.DeleteUsers(r.Context()); err != nil {
		log.Printf("unable to delete all users: %v", err)
		respondWithError(w, http.StatusInternalServerError, "unable to delete users")
		return
	}
	w.WriteHeader(http.StatusOK)
}
