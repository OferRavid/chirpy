package config

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync/atomic"
	"time"

	"github.com/OferRavid/chirpy/internal/database"
	"github.com/google/uuid"
)

type ApiConfig struct {
	FileserverHits atomic.Int32
	DbQueries      *database.Queries
	Platform       string
}

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

func (apiCfg *ApiConfig) MiddlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiCfg.FileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (apiCfg *ApiConfig) MetricsHandler(w http.ResponseWriter, r *http.Request) {
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
		apiCfg.FileserverHits.Load(),
	))
}

func (apiCfg *ApiConfig) ResetHandler(w http.ResponseWriter, r *http.Request) {
	if apiCfg.Platform != "dev" {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("Forbidden"))
		return
	}
	apiCfg.FileserverHits.Store(0)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hits reset to 0"))
	err := apiCfg.DbQueries.DeleteUsers(r.Context())
	if err != nil {
		log.Printf("failed to delete users: %s", err)
		w.WriteHeader(500)
		return
	}
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

func (apiCfg *ApiConfig) CreateUserHandler(w http.ResponseWriter, r *http.Request) {
	type request struct {
		Email string `json:"email"`
	}

	decoder := json.NewDecoder(r.Body)
	req := request{}
	err := decoder.Decode(&req)
	if err != nil {
		log.Printf("Error decoding request's body: %s", err)
		w.WriteHeader(500)
		return
	}

	user, err := apiCfg.DbQueries.CreateUser(r.Context(), req.Email)
	if err != nil {
		log.Printf("failed to create user in database: err")
		w.WriteHeader(500)
		return
	}

	resp := User{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
	}

	dat, err := json.Marshal(resp)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
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
