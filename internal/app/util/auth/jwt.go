package auth

import (
	"fmt"
	"time"

	"github.com/denis-oreshkevich/shortener/internal/app/util/generator"
	"github.com/denis-oreshkevich/shortener/internal/app/util/logger"
	"github.com/golang-jwt/jwt/v4"
)

// Constants for JWT generation.
const (
	SecretKey = "MegaSecretKey"
	TokenExp  = time.Hour * 5
)

// GenerateToken generates new JWT.
func GenerateToken() (string, error) {
	id := generator.UUIDString()
	logger.Log.Debug(fmt.Sprintf("creating new token for sub = %s", id))
	claims := jwt.RegisteredClaims{
		Subject:   id,
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(TokenExp)),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString([]byte(SecretKey))
	if err != nil {
		return "", fmt.Errorf("signedString. %w", err)
	}
	return tokenString, nil
}
