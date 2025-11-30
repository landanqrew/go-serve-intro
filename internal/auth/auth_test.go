package auth

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestHashPassword(t *testing.T) {
	password := "password"
	hashedPassword, err := HashPassword(password)
	if err != nil {
		t.Fatalf("Error hashing password: %v", err)
	}
	fmt.Println(hashedPassword)
}

func TestCheckPasswordHash(t *testing.T) {
	password := "password"
	hashedPassword, err := HashPassword(password)
	if err != nil {
		t.Fatalf("Error hashing password: %v", err)
	}
	fmt.Println(hashedPassword)
	same, err := CheckPasswordHash(password, hashedPassword)
	if err != nil {
		t.Fatalf("Error checking password hash: %v", err)
	}
	fmt.Println(same)
	if !same {
		t.Fatalf("Password hashes do not match")
	}
	same, err = CheckPasswordHash("wrongpassword", hashedPassword)
	if err != nil {
		t.Fatalf("Error checking password hash: %v", err)
	}
	fmt.Println(same)
	if same {
		t.Fatalf("Password hashes match")
	}
}

func TestMakeJWT(t *testing.T) {
	userID := uuid.New()
	tokenSecret := "test-secret"
	expiresIn := 1 * time.Hour
	token, err := MakeJWT(userID, tokenSecret, expiresIn)
	if err != nil {
		t.Fatalf("Error making JWT: %v", err)
	}
	fmt.Println(token)
}

func TestValidateJWT(t *testing.T) {
	userID := uuid.New()
	tokenSecret := "test-secret"
	expiresIn := 1 * time.Hour
	token, err := MakeJWT(userID, tokenSecret, expiresIn)
	if err != nil {
		t.Fatalf("Error making JWT: %v", err)
	}
	fmt.Println(token)
}

func TestValidateJWTInvalidToken(t *testing.T) {
	tokenSecret := "test-secret"
	token := "invalid-token"
	_, err := ValidateJWT(token, tokenSecret)
	if err == nil {
		t.Fatalf("Expected error validating JWT, got nil")
	}
}

func TestCreateAndValidateJWT(t *testing.T) {
	userID := uuid.New()
	fmt.Println("userID [Created (prior to JWT)]:\n", userID)
	tokenSecret := "test-secret"
	expiresIn := 1 * time.Hour
	token, err := MakeJWT(userID, tokenSecret, expiresIn)
	if err != nil {
		t.Fatalf("Error making JWT: %v", err)
	}
	fmt.Println("token [TestCreateAndValidateJWT]:\n", token)
	userID, err = ValidateJWT(token, tokenSecret)
	if err != nil {
		t.Fatalf("Error validating JWT: %v", err)
	}
	fmt.Println("userID [Validated JWT]:\n", userID)
}

func TestGetValidBearerToken(t *testing.T) {
	headers := http.Header{}
	headers.Set("Authorization", "Bearer test-token")
	token, err := GetBearerToken(headers)
	if err != nil {
		t.Fatalf("Error getting bearer token: %v", err)
	}
	fmt.Println("token [TestGetValidBearerToken]:\n", token)
	if token != "test-token" {
		t.Fatalf("Expected token to be 'test-token', got '%s'", token)
	}
}

func TestGetBearerTokenNoAuthorization(t *testing.T) {
	headers := http.Header{}
	_, err := GetBearerToken(headers)
	if err == nil {
		t.Fatalf("Expected error getting bearer token, got nil")
	}
}

func TestGetInvalidBearerToken(t *testing.T) {
	headers := http.Header{}
	headers.Set("Authorization", "Bearer")
	_, err := GetBearerToken(headers)
	if err == nil {
		t.Fatalf("Expected error getting bearer token, got nil")
	}
}

func TestMakeRefreshToken(t *testing.T) {
	token, err := MakeRefreshToken()
	if err != nil {
		t.Fatalf("Error making refresh token: %v", err)
	}
	fmt.Println("token [TestMakeRefreshToken]:\n", token)
}