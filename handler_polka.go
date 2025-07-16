package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/zelieen/Chirpy/internal/auth"
)

func (cfg *apiConfig) polkaHandler(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Event string `json:"event"`
		Data  struct {
			UserID uuid.UUID `json:"user_id"`
		} `json:"data"`
	}

	// check API key
	key, err := auth.GetAPIKey(r.Header)
	if err != nil {
		log.Printf("API key: %s", err)
		respondWithError(w, http.StatusUnauthorized, "Error no valid API key found", err)
		return
	}
	if key != cfg.polkaKey {
		log.Printf("API key: %s", key)
		respondWithError(w, http.StatusUnauthorized, "Error wrong API key", err)
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

	// check event
	if params.Event != "user.upgraded" {
		log.Printf("Unknown event: %s", params.Event)
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// upgrade the user
	err = cfg.db.UpgradeUserToRed(r.Context(), params.Data.UserID)
	if err != nil {
		log.Printf("Error upgrading the user: %s", err)
		respondWithError(w, http.StatusNotFound, "Error: User could not be upgraded", err)
		return
	}

	respondWithJSON(w, http.StatusNoContent, "")
}
