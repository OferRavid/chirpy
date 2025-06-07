package config

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/OferRavid/chirpy/internal/auth"
	"github.com/OferRavid/chirpy/internal/database"
	"github.com/google/uuid"
)

const eventString = "user.upgraded"

type parameters struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (apiCfg *ApiConfig) CreateUsersHandler(w http.ResponseWriter, r *http.Request) {
	hashedPassword, email, err := getHashedPasswordAndEmail(w, r)
	if err != nil {
		return
	}

	user, err := apiCfg.DbQueries.CreateUser(
		r.Context(),
		database.CreateUserParams{
			Email:          email,
			HashedPassword: hashedPassword,
		},
	)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create user", err)
		return
	}

	respondWithJSON(
		w,
		http.StatusCreated,
		User{
			ID:          user.ID,
			CreatedAt:   user.CreatedAt,
			UpdatedAt:   user.UpdatedAt,
			Email:       user.Email,
			IsChirpyRed: user.IsChirpyRed,
		},
	)
}

func (apiCfg *ApiConfig) UpdatePasswordOrEmailHandler(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Missing or malformed token", err)
		return
	}

	user_id, err := auth.ValidateJWT(token, apiCfg.Secret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid bearerToken for user", err)
		return
	}

	hashedPassword, email, err := getHashedPasswordAndEmail(w, r)
	if err != nil {
		return
	}

	user, err := apiCfg.DbQueries.UpdateUser(
		r.Context(),
		database.UpdateUserParams{
			Email:          email,
			HashedPassword: hashedPassword,
			ID:             user_id,
		},
	)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't update user", err)
		return
	}

	respondWithJSON(w, http.StatusOK,
		User{
			ID:          user.ID,
			CreatedAt:   user.CreatedAt,
			UpdatedAt:   user.UpdatedAt,
			Email:       user.Email,
			IsChirpyRed: user.IsChirpyRed,
		},
	)
}

func (apiCfg *ApiConfig) UpdateMembershipStatusHandler(w http.ResponseWriter, r *http.Request) {
	type requestParams struct {
		Event string `json:"event"`
		Data  struct {
			UserID uuid.UUID `json:"user_id"`
		} `json:"data"`
	}

	decoder := json.NewDecoder(r.Body)
	params := requestParams{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}

	if params.Event != eventString {
		respondWithJSON(w, http.StatusNoContent, nil)
		return
	}

	err = apiCfg.DbQueries.UpdateMembership(r.Context(), params.Data.UserID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondWithError(w, http.StatusNotFound, "Couldn't find user", err)
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Couldn't update user", err)
		return
	}

	respondWithJSON(w, http.StatusNoContent, nil)
}

func getHashedPasswordAndEmail(w http.ResponseWriter, r *http.Request) (string, string, error) {
	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return "", "", err
	}

	hashedPassword, err := auth.HashPassword(params.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create hashed password", err)
		return "", "", err
	}

	return hashedPassword, params.Email, nil
}
