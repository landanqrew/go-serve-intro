package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/landanqrew/go-serve-intro/internal/auth"
	"github.com/landanqrew/go-serve-intro/internal/database"
)

type userResponse struct {
	ID string `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email string `json:"email"`
}

func (cfg *APIConfig) HandleCreateUser(w http.ResponseWriter, r *http.Request) {
	type createUserParams struct {
		Email string `json:"email"`
		Password string `json:"password"`
	}
	type jsonReadError struct {
		Error string `json:"error"`
	}
	type databaseError struct {
		Error string `json:"error"`
	}
	type hashError struct {
		Error string `json:"error"`
	}
	params, err := deriveResponseJson[createUserParams](w, r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Content-Type", "application/json")
		jsonResponse, _ := json.Marshal(jsonReadError{Error: err.Error()})
		w.Write(jsonResponse)
		return
	}

	// hash password
	hashedPassword, err := auth.HashPassword(params.Password)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Header().Set("Content-Type", "application/json")
		jsonResponse, _ := json.Marshal(hashError{Error: err.Error()})
		w.Write(jsonResponse)
		return
	}
	fmt.Println("hashedPassword [HandleCreateUser]:", hashedPassword)

	user, err := cfg.dbQueries.CreateUser(r.Context(), database.CreateUserParams{
		ID:        uuid.New().String(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Email:     params.Email,
		HashedPassword: hashedPassword,
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Header().Set("Content-Type", "application/json")
		jsonResponse, _ := json.Marshal(databaseError{Error: err.Error()})
		w.Write(jsonResponse)
		return
	}
	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")
	userResponse := struct {
		ID        string    `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Email     string    `json:"email"`
	}{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
	}
	json.NewEncoder(w).Encode(userResponse)
}

func (cfg *APIConfig) HandleUpdateUserPassword(w http.ResponseWriter, r *http.Request) {
	type updateUserPasswordParams struct {
		ID string `json:"id"`
		Password string `json:"password"`
	}
	type jsonReadError struct {
		Error string `json:"error"`
	}
	type databaseError struct {
		Error string `json:"error"`
	}
	type hashError struct {
		Error string `json:"error"`
	}
	params, err := deriveResponseJson[updateUserPasswordParams](w, r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
	}

	// hash password
	hashedPassword, err := auth.HashPassword(params.Password)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Header().Set("Content-Type", "application/json")
		jsonResponse, _ := json.Marshal(hashError{Error: err.Error()})
		w.Write(jsonResponse)
		return
	}

	// update password
	user, err := cfg.dbQueries.UpdateUserPasswordByID(r.Context(), database.UpdateUserPasswordByIDParams{
		ID: params.ID,
		HashedPassword: hashedPassword,
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Header().Set("Content-Type", "application/json")
		jsonResponse, _ := json.Marshal(databaseError{Error: err.Error()})
		w.Write(jsonResponse)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	userResponse := struct {
		ID string `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Email string `json:"email"`
	}{
		ID: user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email: user.Email,
	}
	res, _ :=json.Marshal(userResponse)
	w.Write(res)
}

func (cfg *APIConfig) HandleAuthenticateUser(w http.ResponseWriter, r *http.Request) {
	type authenticateUserParams struct {
		Email string `json:"email"`
		Password string `json:"password"`
		ExpiresInSeconds int `json:"expires_in_seconds,omitempty"`
	}
	type jsonReadError struct {
		Error string `json:"error"`
	}
	type databaseError struct {
		Error string `json:"error"`
	}
	type unauthorizedError struct {
		Error string `json:"error"`
	}
	type hashError struct {
		Error string `json:"error"`
	}
	params, err := deriveResponseJson[authenticateUserParams](w, r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Content-Type", "application/json")
		jsonResponse, _ := json.Marshal(jsonReadError{Error: err.Error()})
		w.Write(jsonResponse)
		return
	}
	// set default expires in seconds to 3600 if not provided
	if params.ExpiresInSeconds == 0 {
		params.ExpiresInSeconds = 3600
	}
	
	users, err := cfg.dbQueries.GetUsersByEmail(r.Context(), params.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusUnauthorized)
			w.Header().Set("Content-Type", "application/json")
			jsonResponse, _ := json.Marshal(unauthorizedError{Error: "Invalid email or password"})
			w.Write(jsonResponse)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		w.Header().Set("Content-Type", "application/json")
		jsonResponse, _ := json.Marshal(databaseError{Error: err.Error()})
		w.Write(jsonResponse)
		return
	}

	for _, user := range users {
		same, err := auth.CheckPasswordHash(params.Password, user.HashedPassword)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Header().Set("Content-Type", "application/json")
			jsonResponse, _ := json.Marshal(hashError{Error: err.Error()})
			w.Write(jsonResponse)
			return
		}
		if same {
			// authorized
			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "application/json")
			userResponse := userResponse{
				ID: user.ID,
				CreatedAt: user.CreatedAt,
				UpdatedAt: user.UpdatedAt,
				Email: user.Email,
			}
			res, _ := json.Marshal(userResponse)
			w.Write(res)
			return
		}
	}
	// not authorized
	w.WriteHeader(http.StatusUnauthorized)
	w.Header().Set("Content-Type", "application/json")
	jsonResponse, _ := json.Marshal(unauthorizedError{Error: "Invalid email or password"})
	w.Write(jsonResponse)
}

func (cfg *APIConfig) checkUserExists(userID string) (bool, error) {
	_, err := cfg.dbQueries.GetUserByID(context.Background(), userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, fmt.Errorf("error checking user exists: %w", err)
	}
	return true, nil
}

func (cfg *APIConfig) DeleteAllUsers(ctx context.Context) {	
	cfg.dbQueries.DeleteAllUsers(ctx)
}