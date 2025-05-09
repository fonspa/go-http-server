package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/fonspa/go-http-server/internal/auth"
	"github.com/fonspa/go-http-server/internal/database"
	"github.com/google/uuid"
)

type userPayload struct {
	Email            string `json:"email"`
	Password         string `json:"password"`
	ExpiresInSeconds int    `json:"expires_in_seconds"`
}

// NOTE: We voluntarily omit the hashed user's password in there...
type userResponse struct {
	ID        uuid.UUID `json:"id"`
	CreateAt  time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
	Password  string    `json:"-"`
	Token     string    `json:"token"`
}

func (cfg *apiConfig) handlerCreateUser(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var payload userPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		log.Printf("unable to decode request: %v", err)
		respondWithError(w, http.StatusInternalServerError, "unable to decode request")
		return
	}
	hashedPasswd, err := auth.HashPassword(payload.Password)
	if err != nil {
		log.Printf("unable to hash user password: %v", err)
		respondWithError(w, http.StatusInternalServerError, "unable to hash password")
		return
	}
	user, err := cfg.db.CreateUser(r.Context(), database.CreateUserParams{
		Email:          payload.Email,
		HashedPassword: hashedPasswd,
	})
	if err != nil {
		log.Printf("db error when creating new user: %v", err)
		respondWithError(w, http.StatusInternalServerError, "unable to create new user")
		return
	}
	respondWithJSON(w, http.StatusCreated, userResponse{
		ID:        user.ID,
		CreateAt:  user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
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

func (cfg *apiConfig) handlerLogin(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var payload userPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		log.Printf("unable to decode request: %v", err)
		respondWithError(w, http.StatusInternalServerError, "unable to decode request")
		return
	}
	user, err := cfg.db.GetUserByEmail(r.Context(), payload.Email)
	if err != nil {
		log.Printf("unable to lookup user's email: %v", err)
		respondWithError(w, http.StatusUnauthorized, "invalid email")
		return
	}
	if err := auth.CheckPasswordHash(user.HashedPassword, payload.Password); err != nil {
		log.Printf("unable to validate user's password: %v", err)
		respondWithError(w, http.StatusUnauthorized, "invalid user credentials")
		return
	}
	// Create JWT
	if payload.ExpiresInSeconds == 0 {
		payload.ExpiresInSeconds = 3600
	}
	payload.ExpiresInSeconds = min(payload.ExpiresInSeconds, 3600)

	userToken, err := auth.MakeJWT(user.ID, cfg.jwtSecret, time.Duration(payload.ExpiresInSeconds)*time.Second)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "unable to create JWT for user")
		return
	}
	respondWithJSON(w, http.StatusOK, userResponse{
		ID:        user.ID,
		CreateAt:  user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
		Token:     userToken,
	})
}
