package auth

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/alexedwards/argon2id"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func HashPassword(password string) (string, error) {
	hashedPassword, err := argon2id.CreateHash(password, argon2id.DefaultParams)
	if err != nil {
		return "", err
	}
	return hashedPassword, nil
}

func CheckPasswordHash(password, hash string) (bool, error) {
	same, err := argon2id.ComparePasswordAndHash(password, hash)
	if err != nil {
		return false, err
	}
	return same, nil
}

func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
  token := jwt.NewWithClaims(
		jwt.SigningMethodHS256, 
		jwt.RegisteredClaims{
			Issuer: "chirpy",
			IssuedAt: jwt.NewNumericDate(time.Now().UTC()),
			ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(expiresIn)),
			Subject: userID.String(),
		})
	
	signedToken, err := token.SignedString([]byte(tokenSecret))	
	if err != nil {
		return "", err
	}
	return signedToken, nil
}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	claims := jwt.RegisteredClaims{}

	parsed, err := jwt.ParseWithClaims(
		tokenString, 
		&claims,
		func(token *jwt.Token) (any, error) {
			return []byte(tokenSecret), nil
		},
	)
	if err != nil {
		return uuid.Nil, fmt.Errorf("error parsing token: %w", err)
	}
	
	parsedClaims, ok := parsed.Claims.(*jwt.RegisteredClaims)
	if !ok {
		return uuid.Nil, errors.New("error casting claims")
	}
	return uuid.Parse(parsedClaims.Subject)
}

func GetBearerToken(headers http.Header) (string, error) {
	bearerToken := headers.Get("Authorization")
	if !strings.HasPrefix(bearerToken, "Bearer") {
		return "", errors.New("invalid bearer token")
	}
	if bearerToken == "" {
		return "", errors.New("no bearer token found")
	}

	trimmedToken := strings.TrimSpace(strings.TrimPrefix(bearerToken, "Bearer"))
	if trimmedToken == "" {
		return "", errors.New("empty bearer token")
	}
	return trimmedToken, nil
}

func GetApiKey(headers http.Header) (string, error) {
	apiKey := headers.Get("Authorization")
	if !strings.HasPrefix(apiKey, "ApiKey") {
		return "", errors.New("invalid api key")
	}
	if apiKey == "" {
		return "", errors.New("no api key found")
	}

	trimmedApiKey := strings.TrimSpace(strings.TrimPrefix(apiKey, "ApiKey"))
	if trimmedApiKey == "" {
		return "", errors.New("empty api key")
	}
	return trimmedApiKey, nil
}

func MakeRefreshToken() (string, error) {
	byteSequence := make([]byte, 32)
	res, err := rand.Read(byteSequence)
	if err != nil {
		return "", err
	}
	if res != len(byteSequence) {
		return "", errors.New("failed to read random bytes")
	}
	base64Encoded := base64.URLEncoding.EncodeToString(byteSequence)
	return base64Encoded, nil
}