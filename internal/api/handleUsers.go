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
	Token string `json:"token,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
	IsChirpyRed bool `json:"is_chirpy_red"`
}

type updateUserEmailAndPasswordParams struct {
	UserID string `json:"user_id"`
	Email string `json:"email"`
	Password string `json:"password"`
}

type updateUserEmailParams struct {
	UserID string `json:"user_id"`
	Email string `json:"email"`
}

type updateUserPasswordParams struct {
	UserID string `json:"user_id"`
	Password string `json:"password"`
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
type jwtError struct {
	Error string `json:"error"`
}
type notFoundError struct {
	Error string `json:"error"`
}
type refreshTokenError struct {
	Error string `json:"error"`
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
	// fmt.Println("hashedPassword [HandleCreateUser]:", hashedPassword)

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
	userResponse := userResponse{
		ID: user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email: user.Email,
		IsChirpyRed: user.IsChirpyRed,
	}
	json.NewEncoder(w).Encode(userResponse)
}

func (cfg *APIConfig) HandleUpdateUserPassword(w http.ResponseWriter, r *http.Request, params updateUserPasswordParams) {

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
		ID: params.UserID,
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

func (cfg *APIConfig) HandleUpdateUserEmail(w http.ResponseWriter, r *http.Request, params updateUserEmailParams) {

	// update email
	user, err := cfg.dbQueries.UpdateUserEmailByID(r.Context(), database.UpdateUserEmailByIDParams{
		ID: params.UserID,
		Email: params.Email,
		UpdatedAt: time.Now(),
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


func (cfg *APIConfig) HandleUpdateUserEmailAndPassword(w http.ResponseWriter, r *http.Request, params updateUserEmailAndPasswordParams) {

	// hash password
	hashedPassword, err := auth.HashPassword(params.Password)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Header().Set("Content-Type", "application/json")
		jsonResponse, _ := json.Marshal(hashError{Error: err.Error()})
		w.Write(jsonResponse)
		return
	}

	// update email and password
	user, err := cfg.dbQueries.UpdateUser(r.Context(), database.UpdateUserParams{
		ID: params.UserID,
		Email: params.Email,
		HashedPassword: hashedPassword,
		UpdatedAt: time.Now(),
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

func (cfg *APIConfig) HandleUpdateUser(w http.ResponseWriter, r *http.Request) {
	type updateUserParams struct {
		Email string `json:"email,omitempty"`
		Password string `json:"password,omitempty"`
	}
	params, err := deriveResponseJson[updateUserParams](w, r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Content-Type", "application/json")
		jsonResponse, _ := json.Marshal(jsonReadError{Error: err.Error()})
		w.Write(jsonResponse)
		return
	}
	fmt.Println("params [HandleUpdateUser]:", params)

	if params.Password == "" && params.Email == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Content-Type", "application/json")
		jsonResponse, _ := json.Marshal(jsonReadError{Error: "Email or password are required"})
		w.Write(jsonResponse)
		return
	}

	// get bearer token
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Header().Set("Content-Type", "application/json")
		jsonResponse, _ := json.Marshal(unauthorizedError{Error: "Invalid Authorization Token"})
		w.Write(jsonResponse)
		return
	}
	fmt.Println("token [HandleUpdateUser]:", token)

	// get userID
	userID, err := auth.ValidateJWT(token, cfg.tokenSecret)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Header().Set("Content-Type", "application/json")
		jsonResponse, _ := json.Marshal(unauthorizedError{Error: "Invalid Authorization Token"})
		w.Write(jsonResponse)
		return
	}
	fmt.Println("userID [HandleUpdateUser]:", userID)
	// get all user data
	_, err = cfg.dbQueries.GetUserByID(r.Context(), userID.String())
	if err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusNotFound)
			w.Header().Set("Content-Type", "application/json")
			jsonResponse, _ := json.Marshal(notFoundError{Error: "User not found"})
			w.Write(jsonResponse)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		w.Header().Set("Content-Type", "application/json")
		jsonResponse, _ := json.Marshal(databaseError{Error: err.Error()})
		w.Write(jsonResponse)
		return
	}

	if params.Password != "" && params.Email != "" {
		// update email and password
		fmt.Println("update email and password [HandleUpdateUser]")
		cfg.HandleUpdateUserEmailAndPassword(w, r, updateUserEmailAndPasswordParams{
			UserID: userID.String(),
			Email: params.Email,
			Password: params.Password,
		})
		return
	}
	if params.Password != "" {
		// update password
		cfg.HandleUpdateUserPassword(w, r, updateUserPasswordParams{
			UserID: userID.String(),
			Password: params.Password,
		})
		return
	}
	if params.Email != "" {
		// update email
		cfg.HandleUpdateUserEmail(w, r, updateUserEmailParams{
			UserID: userID.String(),
			Email: params.Email,
		})
		return
	}
}

func (cfg *APIConfig) HandleAuthenticateUser(w http.ResponseWriter, r *http.Request) {
	type authenticateUserParams struct {
		Email string `json:"email"`
		Password string `json:"password"`
		ExpiresInSeconds int `json:"expires_in_seconds,omitempty"`
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
			// create JWT
			token, err := auth.MakeJWT(uuid.MustParse(user.ID), cfg.tokenSecret, time.Duration(params.ExpiresInSeconds)*time.Second)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Header().Set("Content-Type", "application/json")
				jsonResponse, _ := json.Marshal(jwtError{Error: err.Error()})
				w.Write(jsonResponse)
				return
			}

			// create refresh token
			refreshToken, err := auth.MakeRefreshToken()
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Header().Set("Content-Type", "application/json")
				jsonResponse, _ := json.Marshal(refreshTokenError{Error: err.Error()})
				w.Write(jsonResponse)
				return
			}

			// create refresh token record
			_, err = cfg.dbQueries.CreateRefreshToken(r.Context(), database.CreateRefreshTokenParams{
				Token: refreshToken,
				CreatedAt: time.Now().UTC(),
				UpdatedAt: time.Now().UTC(),
				UserID: user.ID,
				ExpiresAt: time.Now().UTC().Add(time.Duration(params.ExpiresInSeconds)*time.Second),
			})
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Header().Set("Content-Type", "application/json")
				jsonResponse, _ := json.Marshal(databaseError{Error: err.Error()})
				w.Write(jsonResponse)
				return
			}

			// write response
			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "application/json")
			userResponse := userResponse{
				ID: user.ID,
				CreatedAt: user.CreatedAt,
				UpdatedAt: user.UpdatedAt,
				Email: user.Email,
				Token: token,
				RefreshToken: refreshToken,
				IsChirpyRed: user.IsChirpyRed,
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

func (cfg *APIConfig) HandleTokenRefresh(w http.ResponseWriter, r *http.Request) {
	type tokenRefreshResponse struct {
		Token string `json:"token"`
		RefreshToken string `json:"refresh_token"`
	}

	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Header().Set("Content-Type", "application/json")
		jsonResponse, _ := json.Marshal(unauthorizedError{Error: "Invalid Authorization Token"})
		w.Write(jsonResponse)
		return
	}

	refreshTokenRecord, err := cfg.dbQueries.GetRefreshTokenByToken(r.Context(), refreshToken)
	if err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusNotFound)
			w.Header().Set("Content-Type", "application/json")
			jsonResponse, _ := json.Marshal(notFoundError{Error: "Refresh token not found"})
			w.Write(jsonResponse)
			return
		}
		w.WriteHeader(http.StatusUnauthorized)
		w.Header().Set("Content-Type", "application/json")
		jsonResponse, _ := json.Marshal(unauthorizedError{Error: "Invalid Refresh Token"})
		w.Write(jsonResponse)
		return
	}

	if refreshTokenRecord.ExpiresAt.Before(time.Now().UTC()) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Header().Set("Content-Type", "application/json")
		jsonResponse, _ := json.Marshal(unauthorizedError{Error: "Refresh token expired"})
		w.Write(jsonResponse)
		return
	}

	if refreshTokenRecord.RevokedAt.Valid {
		w.WriteHeader(http.StatusUnauthorized)
		w.Header().Set("Content-Type", "application/json")
		jsonResponse, _ := json.Marshal(unauthorizedError{Error: "Refresh token revoked"})
		w.Write(jsonResponse)
		return
	}

	// get duration between creation and expiry
	expiresInDuration := refreshTokenRecord.ExpiresAt.Sub(refreshTokenRecord.CreatedAt)

	// create JWT
	token, err := auth.MakeJWT(uuid.MustParse(refreshTokenRecord.UserID), cfg.tokenSecret, expiresInDuration)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Header().Set("Content-Type", "application/json")
		jsonResponse, _ := json.Marshal(jwtError{Error: err.Error()})
		w.Write(jsonResponse)
		return
	}

	/*
	// revoke refresh token
	_, err = cfg.dbQueries.RevokeRefreshToken(r.Context(), database.RevokeRefreshTokenParams{
		Token: refreshToken,
		RevokedAt: sql.NullTime{Time: time.Now().UTC(), Valid: true},
		UpdatedAt: time.Now().UTC(),
	})
	if err != nil {
		fmt.Println("error revoking refresh token [HandleTokenRefresh]:", err)
	}
	*/

	// create new refresh token
	newRefreshToken, err := auth.MakeRefreshToken()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Header().Set("Content-Type", "application/json")
		jsonResponse, _ := json.Marshal(refreshTokenError{Error: err.Error()})
		w.Write(jsonResponse)
		return
	}

	// create new refresh token record
	_, err = cfg.dbQueries.CreateRefreshToken(r.Context(), database.CreateRefreshTokenParams{
		Token: newRefreshToken,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		UserID: refreshTokenRecord.UserID,
		ExpiresAt: time.Now().UTC().Add(expiresInDuration),
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Header().Set("Content-Type", "application/json")
		jsonResponse, _ := json.Marshal(databaseError{Error: err.Error()})
		w.Write(jsonResponse)
		return
	}

	// write response
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tokenRefreshResponse{
		Token: token,
		RefreshToken: newRefreshToken,
	})
}

func (cfg *APIConfig) HandleTokenRevoke(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Header().Set("Content-Type", "application/json")
		jsonResponse, _ := json.Marshal(unauthorizedError{Error: "Invalid Authorization Token"})
		w.Write(jsonResponse)
		return
	}

	refreshTokenRecord, err := cfg.dbQueries.GetRefreshTokenByToken(r.Context(), refreshToken)
	if err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusNotFound)
			w.Header().Set("Content-Type", "application/json")
			jsonResponse, _ := json.Marshal(notFoundError{Error: "Refresh token not found"})
			w.Write(jsonResponse)
			return
		}
		w.WriteHeader(http.StatusUnauthorized)
		w.Header().Set("Content-Type", "application/json")
		jsonResponse, _ := json.Marshal(unauthorizedError{Error: "Invalid Refresh Token"})
		w.Write(jsonResponse)
		return
	}

	if refreshTokenRecord.ExpiresAt.Before(time.Now().UTC()) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Header().Set("Content-Type", "application/json")
		jsonResponse, _ := json.Marshal(unauthorizedError{Error: "Refresh token expired"})
		w.Write(jsonResponse)
		return
	}

	if refreshTokenRecord.RevokedAt.Valid {
		w.WriteHeader(http.StatusUnauthorized)
		w.Header().Set("Content-Type", "application/json")
		jsonResponse, _ := json.Marshal(unauthorizedError{Error: "Refresh token revoked"})
		w.Write(jsonResponse)
		return
	}

	// revoke refresh token
	_, err = cfg.dbQueries.RevokeRefreshToken(r.Context(), database.RevokeRefreshTokenParams{
		Token: refreshToken,
		RevokedAt: sql.NullTime{Time: time.Now().UTC(), Valid: true},
		UpdatedAt: time.Now().UTC(),
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Header().Set("Content-Type", "application/json")
		jsonResponse, _ := json.Marshal(databaseError{Error: err.Error()})
		w.Write(jsonResponse)
		return
	}

	// write response
	w.WriteHeader(http.StatusNoContent)
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte("")) // empty response body
}

func (cfg *APIConfig) HandleUpdateUserSetChirpyRed(w http.ResponseWriter, r *http.Request) {

	// get api key
	apiKey, err := auth.GetApiKey(r.Header)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Header().Set("Content-Type", "application/json")
		jsonResponse, _ := json.Marshal(unauthorizedError{Error: "Invalid API Key"})
		w.Write(jsonResponse)
		return
	}
	if apiKey != cfg.polkaKey {
		w.WriteHeader(http.StatusUnauthorized)
		w.Header().Set("Content-Type", "application/json")
		jsonResponse, _ := json.Marshal(unauthorizedError{Error: "Invalid API Key"})
		w.Write(jsonResponse)
		return
	}

	// decode body
	type updateUserSetChirpyRedParams struct {
		Event string `json:"event"`
		Data struct {
			UserID string `json:"user_id,omitempty"`
		} `json:"data"`
	}

	params, err := deriveResponseJson[updateUserSetChirpyRedParams](w, r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Content-Type", "application/json")
		jsonResponse, _ := json.Marshal(jsonReadError{Error: err.Error()})
		w.Write(jsonResponse)
		return
	}

	if params.Event != "user.upgraded" {
		w.WriteHeader(http.StatusNoContent)
		w.Header().Set("Content-Type", "application/json")
		jsonResponse, _ := json.Marshal(SuccessMessage{Message: "Event not supported"})
		w.Write(jsonResponse)
		return
	}

	if params.Data.UserID == "" {	
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Content-Type", "application/json")
		jsonResponse, _ := json.Marshal(jsonReadError{Error: "User ID is required"})
		w.Write(jsonResponse)
		return
	}

	_, err = cfg.dbQueries.GetUserByID(r.Context(), params.Data.UserID)
	if err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusNotFound)
			w.Header().Set("Content-Type", "application/json")
			jsonResponse, _ := json.Marshal(notFoundError{Error: "User not found"})
			w.Write(jsonResponse)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		w.Header().Set("Content-Type", "application/json")
		jsonResponse, _ := json.Marshal(databaseError{Error: err.Error()})
		w.Write(jsonResponse)
		return
	}

	_, err = cfg.dbQueries.UpdateUserSetChirpyRed(r.Context(), database.UpdateUserSetChirpyRedParams{
		ID: params.Data.UserID,
		UpdatedAt: time.Now(),
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Header().Set("Content-Type", "application/json")
		jsonResponse, _ := json.Marshal(databaseError{Error: err.Error()})
		w.Write(jsonResponse)
		return
	}

	w.WriteHeader(http.StatusNoContent)
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte("")) // empty response body
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