package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/zelieen/Chirpy/internal/database"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	db             *database.Queries
	platform       string
	secret         string
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal("DB_URL must be set")
	}
	platform := os.Getenv("PLATFORM")
	if platform == "" {
		log.Fatal("PLATFORM must be set")
	}
	secret := os.Getenv("SECRET")
	if secret == "" {
		log.Fatal("SECRET must be set")
	}
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Error opening the database: %s", err)
	}
	dbQueries := database.New(db)

	const filepathRoot = "."
	const port = "8080"
	cfg := apiConfig{
		fileserverHits: atomic.Int32{},
		db:             dbQueries,
		platform:       platform,
		secret:         secret,
	}

	ServeMux := http.NewServeMux()
	appHandler := http.StripPrefix("/app", http.FileServer(http.Dir(filepathRoot)))
	ServeMux.Handle("/app/", cfg.middlewareMetricsInc(appHandler))

	ServeMux.HandleFunc("GET /api/healthz", readyHandler)
	ServeMux.HandleFunc("GET /admin/metrics", cfg.metricHandler)
	ServeMux.HandleFunc("POST /admin/reset", cfg.resetHandler)
	ServeMux.HandleFunc("POST /api/chirps", cfg.chirpHandler)
	ServeMux.HandleFunc("GET /api/chirps", cfg.chirpListHandler)
	ServeMux.HandleFunc("GET /api/chirps/{chirpID}", cfg.chirpGetHandler)
	ServeMux.HandleFunc("DELETE /api/chirps/{chirpID}", cfg.chirpDeleteHandler)
	ServeMux.HandleFunc("POST /api/users", cfg.userHandler)
	ServeMux.HandleFunc("PUT /api/users", cfg.updateUserHandler)
	ServeMux.HandleFunc("POST /api/login", cfg.loginHandler)
	ServeMux.HandleFunc("POST /api/refresh", cfg.refreshHandler)
	ServeMux.HandleFunc("POST /api/revoke", cfg.revokeHandler)

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
