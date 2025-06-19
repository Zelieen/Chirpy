package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
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

func (cfg *apiConfig) resetHandler(w http.ResponseWriter, request *http.Request) {
	cfg.fileserverHits.Store(0)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	msg := fmt.Sprintf("Reset hits: %v", cfg.fileserverHits.Load())
	w.Write([]byte(msg))
}

func validateHandler(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		// these tags indicate how the keys in the JSON should be mapped to the struct fields
		// the struct fields must be exported (start with a capital letter) if you want them parsed
		Body string `json:"body"`
	}

	type returnVals struct {
		// the key will be the name of struct field unless you give it an explicit JSON tag
		Error string `json:"error"`
		Valid bool   `json:"valid"`
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
		w.WriteHeader(500)
		respBody.Error = "Error decoding request"
		respBody.Valid = false
		// Encode Response
		dat, err := json.Marshal(respBody)
		if err != nil {
			log.Printf("Error marshalling JSON: %s", err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(dat)
		return
	}

	// check Chirp length
	if len(params.Body) > 140 {
		respBody.Error = "Chirp is too long"
		respBody.Valid = false
		w.WriteHeader(400)
	} else {
		respBody.Valid = true
		w.WriteHeader(200)
	}

	// Encode Response
	dat, err := json.Marshal(respBody)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(dat)
}
