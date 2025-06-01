package config

import (
	"log"
	"net/http"
)

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
