package main

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
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

func postChirpsHandler(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}

	params := parameters{}
	params, err := decodeJsonBody(r.Body, params)
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

	db := r.Context().Value(contextKeyDB).(*DB)
	chirp, err := db.CreateChirp(cleanedBody)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create chirp")
		return
	}

	respondWithJson(w, http.StatusCreated, chirp)
}

func getChirpsHandler(w http.ResponseWriter, r *http.Request) {
	db := r.Context().Value(contextKeyDB).(*DB)
	chirps, err := db.GetChirps()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't get chirps")
		return
	}
	respondWithJson(w, http.StatusOK, chirps)
}

func getSingleChirpHandler(w http.ResponseWriter, r *http.Request) {
	db := r.Context().Value(contextKeyDB).(*DB)
	chirpID := chi.URLParam(r, "chirpID")
	id, err := strconv.Atoi(chirpID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid chirp id")
		return
	}
	chirps, err := db.GetChirp(id)
	if err != nil {
		if errors.Is(err, &ChirpNotFound{}) {
			respondWithError(w, http.StatusNotFound, "Couldn't find chirp")
			return
		} else {
			respondWithError(w, http.StatusInternalServerError, "Couldn't get chirp")
			return
		}
	}
	respondWithJson(w, http.StatusOK, chirps)
}

func createUserHandler(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	params := parameters{}
	params, err := decodeJsonBody(r.Body, params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters")
		return
	}

	password, err := bcrypt.GenerateFromPassword([]byte(params.Password), bcrypt.DefaultCost)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't hash password")
		return
	}

	db := r.Context().Value(contextKeyDB).(*DB)
	user, err := db.CreateUser(params.Email, string(password))
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create chirp")
		return
	}

	respondWithJson(w, http.StatusCreated, user)
}

func (cfg *apiConfig) login(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email            string `json:"email"`
		Password         string `json:"password"`
		ExpiresInSeconds *int   `json:"expires_in_seconds"`
	}

	params := parameters{}
	params, err := decodeJsonBody(r.Body, params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters")
		return
	}

	db := r.Context().Value(contextKeyDB).(*DB)
	user, err := db.FindUserByEmail(params.Email)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "User does not exist")
		return

	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(params.Password))
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Password is not correct")
		return
	}

	expiresAdd := 24 * 3600
	if params.ExpiresInSeconds == nil || *params.ExpiresInSeconds > 24*3600 {
		expiresAdd = 24 * 3600
	} else {
		expiresAdd = *params.ExpiresInSeconds
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:    "chirpy",
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(expiresAdd))),
		Subject:   strconv.Itoa(user.ID),
	})

	signedToken, err := token.SignedString(cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't sign JWT Token")
		return
	}

	respondWithJson(w, http.StatusOK, struct {
		ID    int    `json:"id"`
		Email string `json:"email"`
		Token string `json:"token"`
	}{
		ID:    user.ID,
		Email: user.Email,
		Token: signedToken,
	})

}
