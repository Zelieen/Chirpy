package main

import (
	"log"
	"net/http"
	"sync/atomic"
)

type apiConfig struct {
	fileserverHits atomic.Int32
}

func main() {
	const filepathRoot = "."
	const port = "8080"

	ServeMux := http.NewServeMux()
	appHandler := http.StripPrefix("/app", http.FileServer(http.Dir(filepathRoot)))
	ServeMux.Handle("/app/", appHandler)

	ServeMux.HandleFunc("/healthz", readyhandler)

	Server := &http.Server{
		Handler: ServeMux,
		Addr:    ":" + port,
	}

	log.Printf("Serving from %s on port: %s\n", filepathRoot, port)
	log.Fatal(Server.ListenAndServe())
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	cfg.fileserverHits.Add(1)
}
