module github.com/edusig/chirpy

go 1.21.1

require (
	github.com/go-chi/chi/v5 v5.0.10
	github.com/joho/godotenv v1.5.1
	internal/auth v1.0.0
	internal/database v1.0.0
)

require (
	github.com/golang-jwt/jwt/v5 v5.0.0 // indirect
	golang.org/x/crypto v0.17.0 // indirect
)

replace internal/database => ./internal/database

replace internal/auth => ./internal/auth
