package main

import (
	"errors"
	"fmt"
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

func (cfg *apiConfig) resetHandler(w http.ResponseWriter, r *http.Request) {
	if cfg.platform != "dev" {
		respondWithError(w, http.StatusForbidden, "Not allowed", errors.New("action is forbidden outside the development platform"))
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
