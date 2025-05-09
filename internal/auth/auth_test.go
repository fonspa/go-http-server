package auth

import (
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestCheckPasswordHash(t *testing.T) {
	pass1 := "123456"
	pass2 := "totoplop"
	hash1, _ := HashPassword(pass1)
	hash2, _ := HashPassword(pass2)

	cases := []struct {
		name     string
		password string
		hash     string
		wantErr  bool
	}{
		{
			name:     "correct pass",
			password: pass1,
			hash:     hash1,
			wantErr:  false,
		},
		{
			name:     "incorrect pass",
			password: "totoplop",
			hash:     hash1,
			wantErr:  true,
		},
		{
			name:     "incorrect hash",
			password: pass1,
			hash:     hash2,
			wantErr:  true,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := CheckPasswordHash(c.hash, c.password)
			if (err != nil) != c.wantErr {
				t.Fatalf("want err: %v, got %v", c.wantErr, err)
			}
		})
	}
}

func TestJWT(t *testing.T) {
	userID := uuid.New()
	cases := []struct {
		name         string
		userID       uuid.UUID
		createSecret string
		parseSecret  string
		expiresIn    time.Duration
		wantID       uuid.UUID
		wantErr      bool
	}{
		{
			name:         "correct infos",
			userID:       userID,
			createSecret: "mytokensecret",
			parseSecret:  "mytokensecret",
			expiresIn:    1 * time.Hour,
			wantID:       userID,
			wantErr:      false,
		},
		{
			name:         "incorrect secret",
			userID:       userID,
			createSecret: "mytokensecret",
			parseSecret:  "toto",
			expiresIn:    1 * time.Hour,
			wantID:       uuid.Nil,
			wantErr:      true,
		},
		{
			name:         "expired token",
			userID:       userID,
			createSecret: "mytokensecret",
			parseSecret:  "mytokensecret",
			expiresIn:    1 * time.Millisecond,
			wantID:       uuid.Nil,
			wantErr:      true,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			token, err := MakeJWT(c.userID, c.createSecret, c.expiresIn)
			if err != nil {
				t.Fatalf("Error creating token: %v", err)
			}
			gotID, err := ValidateJWT(token, c.parseSecret)
			if (err != nil) != c.wantErr {
				t.Errorf("want err %v, got %v", c.wantErr, err)
			}
			if gotID != c.wantID {
				t.Errorf("want ID %v, got %v", c.wantID, gotID)
			}
		})
	}
}

func TestGetBearerToken(t *testing.T) {
	cases := []struct {
		name      string
		header    http.Header
		wantToken string
		wantErr   bool
	}{
		{
			name:      "valid header",
			header:    http.Header{"Authorization": []string{"Bearer abc"}},
			wantToken: "abc",
			wantErr:   false,
		},
		{
			name:      "no Authorization header",
			header:    http.Header{},
			wantToken: "",
			wantErr:   true,
		},
		{
			name:      "No Bearer prefix",
			header:    http.Header{"Authorization": []string{"abc"}},
			wantToken: "",
			wantErr:   true,
		},
		{
			name:      "Invalid Bearer prefix",
			header:    http.Header{"Authorization": []string{"InvalidBearer abc"}},
			wantToken: "",
			wantErr:   true,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			gotToken, err := GetBearerToken(c.header)
			if (err != nil) != c.wantErr {
				t.Errorf("want err: %v, got err: %v", c.wantErr, err)
			}
			if gotToken != c.wantToken {
				t.Errorf("want token: %s, got token: %s", c.wantToken, gotToken)
			}
		})
	}
}
