package config

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync/atomic"

	"github.com/OferRavid/chirpy/internal/database"
)

type ApiConfig struct {
	FileserverHits atomic.Int32
	DbQueries      *database.Queries
}

func (cfg *ApiConfig) MiddlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.FileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (cfg *ApiConfig) MetricsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write(fmt.Appendf(
		[]byte{},
		`
<html>
  <body>
    <h1>Welcome, Chirpy Admin</h1>
	<p>Chirpy has been visited %d times!</p>
  </body>
</html>`,
		cfg.FileserverHits.Load(),
	))
}

func (cfg *ApiConfig) ResetHandler(w http.ResponseWriter, r *http.Request) {
	cfg.FileserverHits.Store(0)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hits reset to 0"))
}

func ValidateChirpHandler(w http.ResponseWriter, r *http.Request) {
	type request struct {
		Body string `json:"body"`
	}

	decoder := json.NewDecoder(r.Body)
	req := request{}
	err := decoder.Decode(&req)
	if err != nil {
		log.Printf("Error decoding request's body: %s", err)
		w.WriteHeader(500)
		return
	}

	type response struct {
		Cleaned_body string `json:"cleaned_body"`
		Error        string `json:"error"`
	}

	resp := response{}
	statusCode := 0
	if len(req.Body) > 140 {
		resp.Error = "Chirp is too long"
		statusCode = 400
	} else {
		resp.Cleaned_body = cleaner(req.Body)
		statusCode = 200
	}

	dat, err := json.Marshal(resp)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	w.Write(dat)
}

func cleaner(textToClean string) string {
	const censoredWord = "****"
	words := strings.Split(textToClean, " ")
	var censored []string
	for _, word := range words {
		cleanWord := word
		for _, badWord := range []string{"kerfuffle", "sharbert", "fornax"} {
			if strings.ToLower(word) == badWord {
				cleanWord = censoredWord
				break
			}
		}
		censored = append(censored, cleanWord)
	}
	return strings.Join(censored, " ")
}
