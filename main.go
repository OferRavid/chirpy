package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/OferRavid/chirpy/internal/config"
	"github.com/OferRavid/chirpy/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	const filepathRoot = "."
	const port = "8080"

	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal("DB_URL must be set")
	}
	platform := os.Getenv("PLATFORM")
	if platform == "" {
		log.Fatal("PLATFORM must be set")
	}
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		log.Fatal("JWT_SECRET environment variable is not set")
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("failed to open db: %s\n", err)
	}
	dbQueries := database.New(db)

	apiCfg := &config.ApiConfig{
		FileserverHits: atomic.Int32{},
		DbQueries:      dbQueries,
		Platform:       platform,
		Secret:         secret,
	}

	mux := http.NewServeMux()
	fsHandler := http.StripPrefix("/app", apiCfg.MiddlewareMetricsInc(http.FileServer(http.Dir(filepathRoot))))
	mux.Handle("/app/", fsHandler)

	mux.HandleFunc("GET /api/healthz", config.StatusHandler)
	mux.HandleFunc("GET /admin/metrics", apiCfg.MetricsHandler)
	mux.HandleFunc("GET /api/chirps", apiCfg.RetrieveChirpsHandler)
	mux.HandleFunc("GET /api/chirps/{chirpID}", apiCfg.GetChirpsHandler)

	mux.HandleFunc("POST /admin/reset", apiCfg.ResetHandler)
	mux.HandleFunc("POST /api/users", apiCfg.CreateUsersHandler)
	mux.HandleFunc("POST /api/login", apiCfg.LoginHandler)
	mux.HandleFunc("POST /api/chirps", apiCfg.CreateChirpsHandler)
	mux.HandleFunc("POST /api/refresh", apiCfg.RefreshTokenHandler)
	mux.HandleFunc("POST /api/revoke", apiCfg.RevokeTokenHandler)

	mux.HandleFunc("PUT /api/users", apiCfg.UpdatePasswordOrEmailHandler)

	mux.HandleFunc("DELETE /api/chirps/{chirpID}", apiCfg.DeleteChirpsHandler)

	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
	log.Fatal(server.ListenAndServe())
}
