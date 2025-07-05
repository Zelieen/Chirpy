package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

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
		UserID: params.UserID,
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
	chirpList, err := cfg.db.GetChirpList(r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error getting Chirp list", err)
		return
	}
	chirps := []Chirp{}
	for _, c := range chirpList {
		chirps = append(chirps, Chirp(c))
	}

	respondWithJSON(w, http.StatusOK, chirps)
}
