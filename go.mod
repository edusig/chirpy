module github.com/edusig/chirpy

go 1.21.1

require (
	github.com/golang-jwt/jwt/v5 v5.0.0
	github.com/joho/godotenv v1.5.1
	golang.org/x/crypto v0.13.0
	github.com/go-chi/chi/v5 v5.0.10
	internal/database v1.0.0
	internal/auth v1.0.0
)

replace internal/database => ./internal/database

replace internal/auth => ./internal/auth
