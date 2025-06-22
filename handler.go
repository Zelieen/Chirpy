package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
)

func readyHandler(w http.ResponseWriter, request *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func (cfg *apiConfig) metricHandler(w http.ResponseWriter, request *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	content, err := os.ReadFile("metrics.html")
	if err != nil {
		w.Write([]byte("Error: found no metric html file"))
	} else {
		msg := fmt.Sprintf(string(content), cfg.fileserverHits.Load())
		w.Write([]byte(msg))
	}
}

func (cfg *apiConfig) resetHandler(w http.ResponseWriter, r *http.Request) {
	if cfg.platform != "dev" {
		respondWithError(w, http.StatusForbidden, "Not allowed", errors.New("Action is forbidden outside the development platform"))
		return
	}

	// reset
	cfg.fileserverHits.Store(0)
	err := cfg.db.DeleteAllUsers(r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error deleting all users", err)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	msg := fmt.Sprintf("Reset website hits to %v\nand deleted all users from database", cfg.fileserverHits.Load())
	w.Write([]byte(msg))
}

func getProfanityList() []string {
	return []string{
		"kerfuffle",
		"sharbert",
		"fornax",
	}
}

func validateHandler(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		// these tags indicate how the keys in the JSON should be mapped to the struct fields
		// the struct fields must be exported (start with a capital letter) if you want them parsed
		Body string `json:"body"`
	}

	type returnVals struct {
		// the key will be the name of struct field unless you give it an explicit JSON tag
		Cleaned_Body string `json:"cleaned_body"`
	}
	respBody := returnVals{}

	// Decode Request
	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		// an error will be thrown if the JSON is invalid or has the wrong types
		// any missing fields will simply have their values in the struct set to their zero value
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

	respBody.Cleaned_Body = cleaned
	respondWithJSON(w, http.StatusOK, respBody)
}

func (cfg *apiConfig) userHandler(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email string `json:"email"`
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

	// create user
	user, err := cfg.db.CreateUser(r.Context(), params.Email)
	if err != nil {
		log.Printf("Error creating user: %s", err)
		respondWithError(w, http.StatusInternalServerError, "Error while creating user", err)
		return
	}

	respondWithJSON(w, http.StatusCreated, user)
}
