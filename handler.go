package main

import (
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

func (cfg *apiConfig) resetHandler(w http.ResponseWriter, request *http.Request) {
	cfg.fileserverHits.Store(0)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	msg := fmt.Sprintf("Reset hits: %v", cfg.fileserverHits.Load())
	w.Write([]byte(msg))
}

func NoCacheHandler(w http.ResponseWriter, request *http.Request) {
	w.Header().Set("Cache-Control", "no-cache")
}
