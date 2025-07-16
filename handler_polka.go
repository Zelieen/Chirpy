package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/google/uuid"
)

func (cfg *apiConfig) polkaHandler(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Event string `json:"event"`
		Data  struct {
			UserID uuid.UUID `json:"user_id"`
		} `json:"data"`
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
