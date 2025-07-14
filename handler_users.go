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
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

func MakeUserSafe(u database.User) User {
	return User{
		ID:        u.ID,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
		Email:     u.Email,
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
	const Expiry int = 3600

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
		log.Printf("Could not create refresh token: %s", err)
		respondWithError(w, http.StatusInternalServerError, "Login failed", err)
		return
	}

	respondWithJSON(w, http.StatusOK, response{
		User:         MakeUserSafe(user),
		Token:        newToken,
		RefreshToken: refreshToken,
	})
}
