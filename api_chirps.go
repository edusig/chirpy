package main

import (
	"errors"
	"internal/auth"
	"internal/database"
	"net/http"
	"sort"
	"strconv"

	"github.com/go-chi/chi/v5"
)

func (cfg *apiConfig) createChirpsHandler(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}

	authHeader := r.Header.Get("Authorization")
	token, err := auth.ValidateJWTToken(authHeader, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid JWT token")
		return
	}
	userId, err := auth.GetUserFromTokenClaims(token)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error while decoding JWT token")
		return
	}

	params := parameters{}
	params, err = decodeJsonBody(r.Body, params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters")
		return
	}

	const maxChirpLength = 140
	if len(params.Body) > maxChirpLength {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long")
		return
	}

	cleanedBody := cleanBody(params.Body)

	chirp, err := cfg.database.CreateChirp(cleanedBody, userId)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create chirp")
		return
	}

	respondWithJson(w, http.StatusCreated, chirp)
}

func (cfg *apiConfig) getChirpsHandler(w http.ResponseWriter, r *http.Request) {
	chirps, err := cfg.database.GetChirps()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't get chirps")
		return
	}

	responseChirps := chirps
	authorParam := r.URL.Query().Get("author_id")
	if authorParam != "" {
		authorId, err := strconv.Atoi(authorParam)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid author id")
			return
		}
		responseChirps = make([]database.Chirp, 0)
		for _, chirp := range chirps {
			if chirp.AuthorId == authorId {
				responseChirps = append(responseChirps, chirp)
			}
		}
	}

	sortDirection := "asc"
	sortParam := r.URL.Query().Get("sort")
	if sortParam != "" {
		sortDirection = sortParam
	}

	sort.Slice(responseChirps, func(i, j int) bool {
		a, b := responseChirps[i], responseChirps[j]
		if sortDirection == "asc" {
			return a.ID < b.ID
		}
		return a.ID > b.ID
	})

	respondWithJson(w, http.StatusOK, responseChirps)
}

func (cfg *apiConfig) getSingleChirpHandler(w http.ResponseWriter, r *http.Request) {
	chirpID := chi.URLParam(r, "chirpID")
	id, err := strconv.Atoi(chirpID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid chirp id")
		return
	}
	chirps, err := cfg.database.GetChirp(id)
	if err != nil {
		if errors.Is(err, database.ErrNotExist) {
			respondWithError(w, http.StatusNotFound, "Couldn't find chirp")
			return
		} else {
			respondWithError(w, http.StatusInternalServerError, "Couldn't get chirp")
			return
		}
	}
	respondWithJson(w, http.StatusOK, chirps)
}

func (cfg *apiConfig) deleteSingleChirpHandler(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	token, err := auth.ValidateJWTToken(authHeader, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid JWT token")
		return
	}
	userId, err := auth.GetUserFromTokenClaims(token)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error while decoding JWT token")
		return
	}

	chirpID := chi.URLParam(r, "chirpID")
	id, err := strconv.Atoi(chirpID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid chirp id")
		return
	}
	chirp, err := cfg.database.GetChirp(id)
	if err != nil {
		if errors.Is(err, database.ErrNotExist) {
			respondWithError(w, http.StatusNotFound, "Couldn't find chirp")
			return
		} else {
			respondWithError(w, http.StatusInternalServerError, "Couldn't get chirp")
			return
		}
	}

	if chirp.AuthorId != userId {
		respondWithError(w, http.StatusForbidden, "You are not allowed to delete chirps from other users")
		return
	}

	err = cfg.database.DeleteChirp(id)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't delete chirp")
		return
	}

	respondWithJson(w, http.StatusOK, chirp)
}
