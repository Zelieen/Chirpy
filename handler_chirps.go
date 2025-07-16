package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/zelieen/Chirpy/internal/auth"
	"github.com/zelieen/Chirpy/internal/database"

	"github.com/google/uuid"
)

type Chirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
}

func getProfanityList() []string {
	return []string{
		"kerfuffle",
		"sharbert",
		"fornax",
	}
}

func (cfg *apiConfig) chirpHandler(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body   string    `json:"body"`
		UserID uuid.UUID `json:"user_id"`
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

	// check log in status
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		log.Printf("Token bearer: %s", err)
		respondWithError(w, http.StatusInternalServerError, "Error getting the token bearer", err)
		return
	}
	tokenUser, err := auth.ValidateJWT(token, cfg.secret)
	if err != nil {
		log.Printf("Invalid token: %s", err)
		respondWithError(w, http.StatusUnauthorized, "Error validating the token", err)
		return
	}

	// check Chirp length
	if len(params.Body) > 140 {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long", nil)
		return
	}

	// replace "profanity"
	words := strings.Split(params.Body, " ")
	for i, w := range words {
		for _, p := range getProfanityList() {
			if strings.ToLower(w) == p {
				words[i] = "****"
				continue
			}
		}
	}
	cleaned := strings.Join(words, " ")

	// create Chirp
	chirp, err := cfg.db.CreateChirp(r.Context(), database.CreateChirpParams{
		Body:   cleaned,
		UserID: tokenUser,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error creating Chirp", err)
		return
	}

	respondWithJSON(w, http.StatusCreated, Chirp(chirp))
}

func (cfg *apiConfig) chirpGetHandler(w http.ResponseWriter, r *http.Request) {
	chirpID, err := uuid.Parse(r.PathValue("chirpID"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Not a valid chirp id", err)
		return
	}

	// get Chirp by ID
	chirp, err := cfg.db.GetChirpByID(r.Context(), chirpID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Chirp not found", err)
		return
	}

	respondWithJSON(w, http.StatusOK, Chirp(chirp))
}

func (cfg *apiConfig) chirpListHandler(w http.ResponseWriter, r *http.Request) {
	// get Chirp list
	author := r.URL.Query().Get("author_id")
	chirpList := []database.Chirp{}
	if author == "" { // get all chirps
		completeList, err := cfg.db.GetChirpList(r.Context())
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Error getting Chirp list", err)
			return
		}
		chirpList = append(chirpList, completeList...)
	} else { // filter chirps from author
		author_id, err := uuid.Parse(author)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Error invalid user id", err)
			return
		}
		authorList, err := cfg.db.GetChirpsByAuthor(r.Context(), author_id)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Error getting Chirp list from user", err)
			return
		}
		chirpList = append(chirpList, authorList...)
	}

	// fill the response list
	chirps := []Chirp{}
	for _, c := range chirpList {
		chirps = append(chirps, Chirp(c))
	}

	respondWithJSON(w, http.StatusOK, chirps)
}

func (cfg *apiConfig) chirpDeleteHandler(w http.ResponseWriter, r *http.Request) {
	chirpID, err := uuid.Parse(r.PathValue("chirpID"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Not a valid chirp id", err)
		return
	}

	// check log in status
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		log.Printf("Token bearer: %s", err)
		respondWithError(w, http.StatusUnauthorized, "Error getting the token bearer", err)
		return
	}
	tokenUser, err := auth.ValidateJWT(token, cfg.secret)
	if err != nil {
		log.Printf("Invalid token: %s", err)
		respondWithError(w, http.StatusUnauthorized, "Error validating the token", err)
		return
	}

	// check chirp exists
	chirp, err := cfg.db.GetChirpByID(r.Context(), chirpID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Chirp not found", err)
		return
	}

	// check authorship
	if chirp.UserID != tokenUser {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	// delete Chirp
	err = cfg.db.DeleteChirpByID(r.Context(), chirpID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error deleting Chirp", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
