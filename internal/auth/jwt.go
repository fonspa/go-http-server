package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"log"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const (
	tokenIssuer = "chirpy"
)

func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
	now := time.Now().UTC()
	claims := jwt.RegisteredClaims{
		Issuer:    tokenIssuer,
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(expiresIn)),
		Subject:   userID.String(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(tokenSecret))
	if err != nil {
		log.Printf("unable to sign JWT token: %v", err)
		return "", err
	}
	return signedToken, nil
}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (any, error) {
		return []byte(tokenSecret), nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Name}))
	if err != nil {
		log.Printf("unable to parse the JWT token string: %v", err)
		return uuid.Nil, err
	}
	issuer, err := token.Claims.GetIssuer()
	if err != nil {
		log.Printf("unable to parse issuer from JWT: %v", err)
		return uuid.Nil, err
	}
	if issuer != tokenIssuer {
		return uuid.Nil, errors.New("invalid JWT issuer")
	}
	subject, err := token.Claims.GetSubject()
	if err != nil {
		log.Printf("unable to parse claims from JWT: %v", err)
		return uuid.Nil, err
	}
	id, err := uuid.Parse(subject)
	if err != nil {
		log.Printf("unable to parse uuid from JWT subject: %v", err)
		return uuid.Nil, err
	}
	return id, nil
}

func MakeRefreshToken() (string, error) {
	b := make([]byte, 32)
	rand.Read(b)
	encodedStr := hex.EncodeToString(b)
	return encodedStr, nil
}
