package auth

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestCheckPasswordHash(t *testing.T) {
	// First, we need to create some hashed passwords for testing
	password1 := "correctPassword123!"
	password2 := "anotherPassword456!"
	hash1, _ := HashPassword(password1)
	hash2, _ := HashPassword(password2)

	tests := []struct {
		name     string
		password string
		hash     string
		wantErr  bool
	}{
		{
			name:     "Correct password",
			password: password1,
			hash:     hash1,
			wantErr:  false,
		},
		{
			name:     "Incorrect password",
			password: "wrongPassword",
			hash:     hash1,
			wantErr:  true,
		},
		{
			name:     "Password doesn't match different hash",
			password: password1,
			hash:     hash2,
			wantErr:  true,
		},
		{
			name:     "Empty password",
			password: "",
			hash:     hash1,
			wantErr:  true,
		},
		{
			name:     "Invalid hash",
			password: password1,
			hash:     "invalidhash",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CheckPasswordHash(tt.password, tt.hash)
			if (err != nil) != tt.wantErr {
				t.Errorf("CheckPasswordHash() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestJWTs(t *testing.T) {
	secret1 := "testSecret"
	secret2 := "other"
	userID1 := uuid.New()
	userID2 := uuid.New()
	duration1 := time.Duration(1 * time.Hour)
	duration2 := time.Duration(1 * time.Nanosecond)
	JWT1, _ := MakeJWT(userID1, secret1, duration1)
	JWT2, _ := MakeJWT(userID2, secret2, duration2)

	time.Sleep(time.Duration(5 * time.Second)) // Because token expiration is checked with leeway of 5 seconds

	tests := []struct {
		name      string
		user      uuid.UUID
		secret    string
		JWT       string
		WantedErr bool
	}{
		{
			name:      "Correct JWT",
			user:      userID1,
			secret:    secret1,
			JWT:       JWT1,
			WantedErr: false,
		},
		{
			name:      "Wrong JWT",
			user:      userID1,
			secret:    secret1,
			JWT:       JWT2,
			WantedErr: true,
		},
		{
			name:      "Wrong secret",
			user:      userID1,
			secret:    secret2,
			JWT:       JWT1,
			WantedErr: true,
		},
		{
			name:      "Wrong user",
			user:      userID2,
			secret:    secret1,
			JWT:       JWT1,
			WantedErr: true,
		},
		{
			name:      "Expired",
			user:      userID2,
			secret:    secret2,
			JWT:       JWT2,
			WantedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			userID, err := ValidateJWT(tt.JWT, tt.secret)

			if (err != nil) != tt.WantedErr {
				if (userID != tt.user) && (userID != tt.user) != tt.WantedErr {
					t.Errorf("ValidateJWT() error matching userID: %v, wantedErr %v", (userID != tt.user), tt.WantedErr)
				} else if err != nil {
					t.Errorf("ValidateJWT() error = %v, wantedErr %v", err, tt.WantedErr)
				}
			}
		})
	}
}

func TestBearerToken(t *testing.T) {
	tests := []struct {
		name      string
		header    map[string][]string
		wantedErr bool
	}{
		{
			name: "Good Token Bearer",
			header: map[string][]string{
				"Authorization": {"Bearer authstringfortesting"},
			},
			wantedErr: false,
		},
		{
			name: "Bad Token Bearer",
			header: map[string][]string{
				"Authorization": {"authstringfortesting"},
			},
			wantedErr: true,
		},
		{
			name: "Bad header",
			header: map[string][]string{
				"Authorisation": {"Bearer authstringfortesting"},
			},
			wantedErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := GetBearerToken(tt.header)
			if (err != nil) != tt.wantedErr {
				t.Errorf("GetBearerToken() error = %v, wantErr %v", err, tt.wantedErr)
			}
		})
	}
}
