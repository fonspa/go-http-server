package main

import (
	"database/sql"
	"encoding/json"
	"errors"
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
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Email        string    `json:"email"`
	Password     string    `json:"-"`
	Token        string    `json:"token"`
	RefreshToken string    `json:"refresh_token"`
	IsChirpyRed  bool      `json:"is_chirpy_red"`
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
		ID:          user.ID,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
		Email:       user.Email,
		IsChirpyRed: user.IsChirpyRed.Bool,
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
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
		Email:        user.Email,
		Token:        userToken,
		RefreshToken: refreshToken,
		IsChirpyRed:  user.IsChirpyRed.Bool,
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

func (cfg *apiConfig) handlerUpdateUser(w http.ResponseWriter, r *http.Request) {
	// Get access token
	accessToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "invalid bearer token")
		return
	}
	userID, err := auth.ValidateJWT(accessToken, cfg.jwtSecret)
	if err != nil {
		log.Printf("unable to validate access JWT: %v", err)
		respondWithError(w, http.StatusUnauthorized, "invalid access token")
		return
	}
	defer r.Body.Close()
	var payload userPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		log.Printf("unable to decode request: %v", err)
		respondWithError(w, http.StatusInternalServerError, "unable to decode request")
		return
	}
	hashedPwd, err := auth.HashPassword(payload.Password)
	if err != nil {
		log.Printf("unable to hash password '%s': %v", payload.Password, err)
		respondWithError(w, http.StatusInternalServerError, "unable to hash password")
		return
	}
	usr, err := cfg.db.UpdateUserCredentials(r.Context(), database.UpdateUserCredentialsParams{
		Email:          payload.Email,
		HashedPassword: hashedPwd,
		ID:             userID,
	})
	if err != nil {
		log.Printf("unable to update user's credentials: %v", err)
		respondWithError(w, http.StatusInternalServerError, "unable to update user's credentials")
		return
	}
	respondWithJSON(w, http.StatusOK, userResponse{
		ID:          usr.ID,
		CreatedAt:   usr.CreatedAt,
		UpdatedAt:   usr.UpdatedAt,
		Email:       usr.Email,
		IsChirpyRed: usr.IsChirpyRed.Bool,
	})
}

func (cfg *apiConfig) handlerPolkaWebhook(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Event string `json:"event"`
		Data  struct {
			UserID string `json:"user_id"`
		} `json:"data"`
	}
	var params parameters
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		log.Printf("Unable to decode Polka webhook request: %v", err)
		respondWithError(w, http.StatusInternalServerError, "unable to parse request")
		return
	}
	if params.Event != "user.upgraded" {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	userID, err := uuid.Parse(params.Data.UserID)
	if err != nil {
		log.Printf("Unable to parse user ID: %v", err)
		respondWithError(w, http.StatusInternalServerError, "unable to parse user ID")
		return
	}
	_, err = cfg.db.UpgradeUserToRed(r.Context(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Printf("could not find user with this ID: %v", err)
			respondWithError(w, http.StatusNotFound, "user not found")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "could not upgrade user")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
