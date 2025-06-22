package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"sync/atomic"

	"internal/database"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	dbQueries      *database.Queries
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	dbQueries := database.New(db)

	const filepathRoot = "."
	const port = "8080"
	cfg := apiConfig{
		fileserverHits: atomic.Int32{},
		dbQueries:      dbQueries,
	}

	ServeMux := http.NewServeMux()
	appHandler := http.StripPrefix("/app", http.FileServer(http.Dir(filepathRoot)))
	ServeMux.Handle("/app/", cfg.middlewareMetricsInc(appHandler))

	ServeMux.HandleFunc("GET /api/healthz", readyHandler)
	ServeMux.HandleFunc("GET /admin/metrics", cfg.metricHandler)
	ServeMux.HandleFunc("POST /admin/reset", cfg.resetHandler)
	ServeMux.HandleFunc("POST /api/validate_chirp", validateHandler)

	Server := &http.Server{
		Handler: ServeMux,
		Addr:    ":" + port,
	}

	log.Printf("Serving from %s on port: %s\n", filepathRoot, port)
	log.Fatal(Server.ListenAndServe())
}

// this is building a nameless return function to inject a nameless function that builds a handler after it increased the fileserverHits
func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}
