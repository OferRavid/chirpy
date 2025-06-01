package config

import (
	"net/http"
	"time"

	"github.com/OferRavid/chirpy/internal/auth"
)

func (apiCfg *ApiConfig) RevokeTokenHandler(w http.ResponseWriter, r *http.Request) {
	refresh_token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Missing bearer token in headers", err)
		return
	}

	refreshToken, err := apiCfg.DbQueries.GetRefreshTokenByToken(r.Context(), refresh_token)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Refresh token doesn't exist", err)
		return
	}
	if time.Now().After(refreshToken.ExpiresAt) || refreshToken.RevokedAt.Valid {
		respondWithError(w, http.StatusUnauthorized, "Refresh token already expired", err)
		return
	}

	err = apiCfg.DbQueries.UpdateRefreshToken(r.Context(), refreshToken.Token)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to update record", err)
	}

	respondWithJSON(w, http.StatusNoContent, nil)
}
