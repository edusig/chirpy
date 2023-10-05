package auth

import (
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// HashPassword -
func HashPassword(password string) (string, error) {
	dat, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(dat), nil
}

// CheckPasswordHash -
func CheckPasswordHash(password, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

type CustomClaims struct {
	jwt.RegisteredClaims
}

func generateJWTToken(userID int, issuer, jwtSecret string, expiration time.Duration) (string, error) {
	now := time.Now()
	expiresDate := now.Add(expiration)

	claims := jwt.RegisteredClaims{
		Issuer:    issuer,
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(expiresDate),
		Subject:   strconv.Itoa(userID),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &claims)
	return token.SignedString([]byte(jwtSecret))
}

func GenerateJWTTokens(userID int, jwtSecret string) (string, string, error) {
	accessToken, err := generateJWTToken(userID, "chirpy-access", jwtSecret, time.Hour)
	if err != nil {
		return "", "", err
	}
	refreshToken, err := generateJWTToken(userID, "chirpy-refresh", jwtSecret, time.Hour*24*60)
	if err != nil {
		return "", "", err
	}
	return accessToken, refreshToken, nil
}

func ValidateJWTToken(authHeader, jwtSecret string) (*jwt.Token, error) {
	tokenString := strings.Replace(authHeader, "Bearer ", "", 1)
	claims := jwt.RegisteredClaims{}
	token, err := jwt.ParseWithClaims(tokenString, &claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(jwtSecret), nil
	})
	if err != nil {
		return token, err
	}
	return token, nil
}

func GetUserFromTokenClaims(token *jwt.Token) (int, error) {
	id, err := token.Claims.GetSubject()
	if err != nil {
		return 0, err
	}
	userId, err := strconv.Atoi(id)
	if err != nil {
		return 0, err
	}
	return userId, nil
}

func GenerateAccessTokenFromRefresh(refreshToken *jwt.Token, jwtSecret string) (string, error) {
	userID, err := GetUserFromTokenClaims(refreshToken)
	if err != nil {
		return "", err
	}

	token, err := generateJWTToken(userID, "chirpy-access", jwtSecret, time.Hour)
	if err != nil {
		return "", err
	}
	return token, nil
}
