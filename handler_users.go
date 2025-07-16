package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/zelieen/Chirpy/internal/auth"
	"github.com/zelieen/Chirpy/internal/database"

	"github.com/google/uuid"
)

type User struct {
	ID          uuid.UUID `json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Email       string    `json:"email"`
	IsChirpyRed bool      `json:"is_chirpy_red"`
}

func MakeUserSafe(u database.User) User {
	return User{
		ID:          u.ID,
		CreatedAt:   u.CreatedAt,
		UpdatedAt:   u.UpdatedAt,
		Email:       u.Email,
		IsChirpyRed: u.IsChirpyRed,
	}
}

func (cfg *apiConfig) userHandler(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email    string `json:"email"`
		Password string `json:password`
	}

	// Decode Request
	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		respondWithError(w, http.StatusInternalServerError, "Error decoding parameters", err)
		return
	}

	// hash the password
	hash, err := auth.HashPassword(params.Password)
	if err != nil {
		log.Printf("Error during hashing: %s", err)
		respondWithError(w, http.StatusInternalServerError, "Error handling password", err)
		return
	}

	// create user
	user, err := cfg.db.CreateUser(r.Context(), database.CreateUserParams{
		Email:          params.Email,
		HashedPassword: hash,
	})
	if err != nil {
		log.Printf("Error creating user: %s", err)
		respondWithError(w, http.StatusInternalServerError, "Error while creating user", err)
		return
	}

	respondWithJSON(w, http.StatusCreated, MakeUserSafe(user))
}

func (cfg *apiConfig) loginHandler(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	type response struct {
		User
		Token        string `json:"token"`
		RefreshToken string `json:"refresh_token"`
	}
	const Expiry int = 3600 // this is declared in refreshHandler as well

	// Decode Request
	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		respondWithError(w, http.StatusInternalServerError, "Error decoding parameters", err)
		return
	}

	// get the user by email
	user, err := cfg.db.GetUserByEMail(r.Context(), params.Email)
	if err != nil {
		log.Printf("Error fetching user: %s", err)
		respondWithError(w, http.StatusUnauthorized, "Incorrect email or password", err)
		return
	}

	// check password validity
	err = auth.CheckPasswordHash(params.Password, user.HashedPassword)
	if err != nil {
		log.Printf("Password did not match: %s", err)
		respondWithError(w, http.StatusUnauthorized, "Incorrect email or password", err)
		return
	}

	// make login token
	newToken, err := auth.MakeJWT(user.ID, cfg.secret, (time.Duration(Expiry) * time.Second))
	if err != nil {
		log.Printf("Could not create token: %s", err)
		respondWithError(w, http.StatusInternalServerError, "Login failed", err)
		return
	}

	// make refresh token
	refreshToken, err := auth.MakeRefreshToken()
	if err != nil {
		log.Printf("Could not make refresh token: %s", err)
		respondWithError(w, http.StatusInternalServerError, "Login failed", err)
		return
	}

	// store refresh token in database
	err = cfg.db.CreateRefreshToken(r.Context(), database.CreateRefreshTokenParams{
		Token:  refreshToken,
		UserID: user.ID,
	})
	if err != nil {
		log.Printf("Could not store refresh token: %s", err)
		respondWithError(w, http.StatusInternalServerError, "Login failed", err)
		return
	}

	respondWithJSON(w, http.StatusOK, response{
		User:         MakeUserSafe(user),
		Token:        newToken,
		RefreshToken: refreshToken,
	})
}

func (cfg *apiConfig) refreshHandler(w http.ResponseWriter, r *http.Request) {
	const Expiry int = 3600 // this is declared in loginHandler as well

	type response struct {
		AccessToken string `json:"token"`
	}
	// check log in status
	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		log.Printf("Token bearer: %s", err)
		respondWithError(w, http.StatusInternalServerError, "Error getting the token bearer", err)
		return
	}

	// verify refresh token from database
	fullRefToken, err := cfg.db.GetRefreshToken(r.Context(), refreshToken)
	if err != nil {
		log.Printf("Database: %s\n", err)
		respondWithError(w, http.StatusUnauthorized, "Error getting the refresh token", err)
		return
	}
	if fullRefToken.RevokedAt.Valid {
		log.Printf("No valid token")
		respondWithError(w, http.StatusUnauthorized, "Refresh token was revoked and is no longer valid", err)
		return
	}
	if fullRefToken.ExpiresAt.Compare(time.Now()) < 1 {
		log.Printf("Expires at: %s\n", fullRefToken.ExpiresAt)
		respondWithError(w, http.StatusUnauthorized, "Refresh token has expired", err)
		return
	}

	// make login token
	newToken, err := auth.MakeJWT(fullRefToken.UserID, cfg.secret, (time.Duration(Expiry) * time.Second))
	if err != nil {
		log.Printf("Could not create access token: %s", err)
		respondWithError(w, http.StatusInternalServerError, "Refresh failed", err)
		return
	}

	respondWithJSON(w, http.StatusOK, response{AccessToken: newToken})
}

func (cfg *apiConfig) revokeHandler(w http.ResponseWriter, r *http.Request) {
	// check log in status
	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		log.Printf("Token bearer: %s", err)
		respondWithError(w, http.StatusInternalServerError, "Error getting the token bearer", err)
		return
	}

	// revoke the refresh token
	err = cfg.db.RevokeRefreshToken(r.Context(), refreshToken)
	if err != nil {
		log.Printf("Token from database: %s", err)
		respondWithError(w, http.StatusInternalServerError, "Could not revoke token", err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (cfg *apiConfig) updateUserHandler(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	// check log in status
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		log.Printf("Token bearer: %s", err)
		respondWithError(w, http.StatusUnauthorized, "Error getting the token bearer", err)
		return
	}
	userID, err := auth.ValidateJWT(token, cfg.secret)
	if err != nil {
		log.Printf("Invalid token: %s", err)
		respondWithError(w, http.StatusUnauthorized, "Error validating the token", err)
		return
	}

	// Decode Request
	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err = decoder.Decode(&params)
	if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		respondWithError(w, http.StatusInternalServerError, "Error decoding parameters", err)
		return
	}

	// Hash the password
	hash, err := auth.HashPassword(params.Password)
	if err != nil {
		log.Printf("Error during hashing: %s", err)
		respondWithError(w, http.StatusInternalServerError, "Error handling password", err)
		return
	}

	// Change user in database
	updatedUser, err := cfg.db.UpdateUserCredentials(r.Context(), database.UpdateUserCredentialsParams{
		ID:             userID,
		Email:          params.Email,
		HashedPassword: hash,
	})
	if err != nil {
		log.Printf("Error during updating: %s", err)
		respondWithError(w, http.StatusInternalServerError, "Error updating the user credentials", err)
		return
	}

	respondWithJSON(w, http.StatusOK, MakeUserSafe(updatedUser))
}
