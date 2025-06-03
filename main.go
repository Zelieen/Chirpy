package main

import (
	"log"
	"net/http"
)

func main() {
	const filepathRoot = "."
	const port = "8080"

	ServeMux := http.NewServeMux()
	ServeMux.Handle("/app/", http.StripPrefix("/app", http.FileServer(http.Dir(filepathRoot))))

	ServeMux.HandleFunc("/healthz", readyhandler)

	Server := &http.Server{
		Handler: ServeMux,
		Addr:    ":" + port,
	}

	log.Printf("Serving from %s on port: %s\n", filepathRoot, port)
	log.Fatal(Server.ListenAndServe())
}
