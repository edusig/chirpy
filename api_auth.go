package main

import (
	"internal/auth"
	"log"
	"net/http"
)

func (cfg *apiConfig) login(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	type response struct {
		// database.User
		ID           int    `json:"id"`
		Email        string `json:"email"`
		Token        string `json:"token"`
		RefreshToken string `json:"refresh_token"`
	}

	params := parameters{}
	params, err := decodeJsonBody(r.Body, params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters")
		return
	}

	user, err := cfg.database.GetUserByEmail(params.Email)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "User does not exist")
		return
	}

	err = auth.CheckPasswordHash(params.Password, user.Password)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Password is not correct")
		return
	}

	accessToken, refreshToken, err := auth.GenerateJWTTokens(user.ID, cfg.jwtSecret)
	if err != nil {
		log.Print(err)
		respondWithError(w, http.StatusInternalServerError, "Couldn't sign JWT Token")
		return
	}

	respondWithJson(w, http.StatusOK, response{
		ID:           user.ID,
		Email:        user.Email,
		Token:        accessToken,
		RefreshToken: refreshToken,
	})

}

func (cfg *apiConfig) refreshJWTHandler(w http.ResponseWriter, r *http.Request) {
	type response struct {
		Token string `json:"token"`
	}

	authHeader := r.Header.Get("Authorization")
	token, err := auth.ValidateJWTToken(authHeader, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid JWT token")
		return
	}
	issuer, err := token.Claims.GetIssuer()
	if err != nil || issuer == "chirpy-access" {
		respondWithError(w, http.StatusInternalServerError, "Invalid JWT token type")
		return
	}

	isRevoked, err := cfg.database.GetTokenIsRevoked(token.Raw)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't check if token is revoked")
		return
	}

	if isRevoked {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	newAccessToken, err := auth.GenerateAccessTokenFromRefresh(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't generate refreshed access token")
		return
	}

	respondWithJson(w, 200, response{
		Token: newAccessToken,
	})
}

func (cfg *apiConfig) revokeJWTHandler(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	token, err := auth.ValidateJWTToken(authHeader, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid JWT token")
		return
	}
	issuer, err := token.Claims.GetIssuer()
	if err != nil || issuer == "chirpy-access" {
		respondWithError(w, http.StatusInternalServerError, "Invalid JWT token type")
		return
	}

	err = cfg.database.AddRevokedToken(token.Raw)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't revoke token")
		return
	}

	w.WriteHeader(200)
}
