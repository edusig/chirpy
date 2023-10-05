package main

import (
	"internal/auth"
	"internal/database"
	"log"
	"net/http"
)

func (cfg *apiConfig) createUserHandler(w http.ResponseWriter, r *http.Request) {
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

	user, err := cfg.database.CreateUser(params.Email, password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create user")
		return
	}

	respondWithJson(w, http.StatusCreated, database.User{
		ID:    user.ID,
		Email: user.Email,
	})
}

func (cfg *apiConfig) updateUserHandler(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
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
	issuer, err := token.Claims.GetIssuer()
	log.Printf("JWT ISSUER %v", issuer)
	if err != nil || issuer == "chirpy-refresh" {
		respondWithError(w, http.StatusUnauthorized, "Invalid JWT token issuer")
		return
	}

	params := parameters{}
	params, err = decodeJsonBody(r.Body, params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters")
		return
	}

	hashedPassword, err := auth.HashPassword(params.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't hash password")
		return
	}

	user, err := cfg.database.UpdateUser(userId, params.Email, hashedPassword)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't update user")
		return
	}
	respondWithJson(w, 200, database.User{
		ID:    user.ID,
		Email: user.Email,
	})
}
