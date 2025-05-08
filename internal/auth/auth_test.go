package auth

import "testing"

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
