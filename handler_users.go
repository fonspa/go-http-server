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

const (
	accessTokenDuration  = 1 * time.Hour
	refreshTokenDuration = 60 * 24 * time.Hour
)

type userPayload struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// NOTE: We voluntarily omit the hashed user's password in there...
type userResponse struct {
	ID           uuid.UUID `json:"id"`
	CreateAt     time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Email        string    `json:"email"`
	Password     string    `json:"-"`
	Token        string    `json:"token"`
	RefreshToken string    `json:"refresh_token"`
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
	userToken, err := auth.MakeJWT(user.ID, cfg.jwtSecret, accessTokenDuration)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "unable to create JWT for user")
		return
	}
	refreshToken, _ := auth.MakeRefreshToken()
	_, err = cfg.db.CreateRefreshToken(r.Context(), database.CreateRefreshTokenParams{
		Token:     refreshToken,
		UserID:    user.ID,
		ExpiresAt: time.Now().UTC().Add(refreshTokenDuration),
	})
	if err != nil {
		log.Printf("unable to create refresh token db record: %v", err)
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondWithJSON(w, http.StatusOK, userResponse{
		ID:           user.ID,
		CreateAt:     user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
		Email:        user.Email,
		Token:        userToken,
		RefreshToken: refreshToken,
	})
}

func (cfg *apiConfig) handlerRefreshToken(w http.ResponseWriter, r *http.Request) {
	type response struct {
		Token string `json:"token"`
	}
	// Verify that we have a refresh token in the header
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		log.Printf("no bearer token in header: %v", err)
		respondWithError(w, http.StatusBadRequest, "invalid bearer token")
		return
	}
	// Lookup refresh token in DB
	dbRefreshToken, err := cfg.db.GetUserFromRefreshToken(r.Context(), token)
	if err != nil {
		log.Printf("couldn't get refresh token record from db: %v", err)
		respondWithError(w, http.StatusUnauthorized, "unknown refresh token")
		return
	}
	// Make sure it's not expired
	if time.Now().UTC().After(dbRefreshToken.ExpiresAt) {
		log.Printf("user '%s' refresh token expired!", dbRefreshToken.UserID.String())
		respondWithError(w, http.StatusUnauthorized, "expired refresh token")
		return
	}
	// Check if the refresh token is revoked
	if dbRefreshToken.RevokedAt.Valid {
		if time.Now().UTC().After(dbRefreshToken.RevokedAt.Time) {
			log.Println("user refresh token revoked")
			respondWithError(w, http.StatusUnauthorized, "refresh token revoked")
			return
		}
	}
	// Make a new JWT for that user
	accessToken, err := auth.MakeJWT(dbRefreshToken.UserID, cfg.jwtSecret, accessTokenDuration)
	if err != nil {
		log.Printf("unable to create JWT: %v", err)
		respondWithError(w, http.StatusInternalServerError, "unable to create JWT")
		return
	}
	respondWithJSON(w, http.StatusOK, response{
		Token: accessToken,
	})
}

func (cfg *apiConfig) handlerRevokeRefreshToken(w http.ResponseWriter, r *http.Request) {
	// Get refresh token from header
	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid bearer token")
	}
	// Revoke the refresh Token
	if err = cfg.db.RevokeRefreshToken(r.Context(), refreshToken); err != nil {
		log.Printf("Unable to revoke refresh token: %v", err)
		respondWithError(w, http.StatusInternalServerError, "unable to revoke token")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
