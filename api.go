package main

import (
	"internal/auth"
	"internal/database"
	"net/http"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
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

	password, err := auth.HashPassword(params.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't hash password")
		return
	}

	db := r.Context().Value(contextKeyDB).(*database.DB)
	user, err := db.CreateUser(params.Email, password)
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
	type response struct {
		// database.User
		ID    int    `json:"id"`
		Email string `json:"email"`
		Token string `json:"token"`
	}

	params := parameters{}
	params, err := decodeJsonBody(r.Body, params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters")
		return
	}

	db := r.Context().Value(contextKeyDB).(*database.DB)
	user, err := db.GetUserByEmail(params.Email)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "User does not exist")
		return
	}

	err = auth.CheckPasswordHash(params.Password, user.Password)
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

	respondWithJson(w, http.StatusOK, response{
		ID:    user.ID,
		Email: user.Email,
		Token: signedToken,
	})

}
