package auth

import (
	"errors"
	"net/http"
	"strings"
)

func GetAPIKey(headers http.Header) (string, error) {
	authHeader := headers.Get("Authorization")
	if authHeader == "" {
		return "", errors.New("authorization header not found")
	}
	apiKey, found := strings.CutPrefix(authHeader, "ApiKey ")
	if !found {
		return "", errors.New("invalid API key")
	}
	return apiKey, nil
}

func GetBearerToken(headers http.Header) (string, error) {
	authHeader := headers.Get("Authorization")
	if authHeader == "" {
		return "", errors.New("no Authorization header found")
	}
	tokenString, ok := strings.CutPrefix(authHeader, "Bearer ")
	if !ok {
		return "", errors.New("no auth token found")
	}
	return tokenString, nil
}
