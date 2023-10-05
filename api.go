package main

import (
	"errors"
	"internal/database"
	"net/http"
	"strings"
)

func healthCheck(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)
	w.Write([]byte("OK"))
}

func (cfg *apiConfig) resetHandler(w http.ResponseWriter, r *http.Request) {
	cfg.fileserverHits = 0
	w.WriteHeader(200)
}

func (cfg *apiConfig) polkaWebhookHandler(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Event string `json:"event"`
		Data  struct {
			UserID int `json:"user_id"`
		} `json:"data"`
	}

	authHeader := r.Header.Get("Authorization")
	apiKey := strings.Replace(authHeader, "ApiKey ", "", 1)
	if apiKey != cfg.polkaApiKey {
		respondWithError(w, http.StatusUnauthorized, "Invalid api key")
		return
	}

	params := parameters{}
	params, err := decodeJsonBody(r.Body, params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters")
		return
	}

	if params.Event != "user.upgraded" {
		w.WriteHeader(200)
		return
	}

	err = cfg.database.UpgradeUser(params.Data.UserID)
	if err != nil {
		if errors.Is(err, database.ErrNotExist) {
			respondWithError(w, http.StatusNotFound, "User not found")
		} else {
			respondWithError(w, http.StatusInternalServerError, "Error while upgrading user")
		}
		return
	}

	w.WriteHeader(200)
}
