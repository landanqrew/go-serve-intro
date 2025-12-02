package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/landanqrew/go-serve-intro/internal/auth"
	"github.com/landanqrew/go-serve-intro/internal/database"
)

type ChirpError struct {
	Error string `json:"error"`
}

type ValidatedChirpResponse struct {
	Body string `json:"body"`
}

type CompleteChirp struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    string    `json:"user_id"`
}

type SuccessMessage struct {
	Message string `json:"message"`
}

// sortChirpsByCreatedAt sorts chirps by CreatedAt field
// sortOrder can be "asc" or "desc", defaults to "asc" if invalid
func sortChirpsByCreatedAt(chirps []CompleteChirp, sortOrder string) {
	if sortOrder == "desc" {
		sort.Slice(chirps, func(i, j int) bool {
			return chirps[i].CreatedAt.After(chirps[j].CreatedAt)
		})
	} else {
		// default to ascending order
		sort.Slice(chirps, func(i, j int) bool {
			return chirps[i].CreatedAt.Before(chirps[j].CreatedAt)
		})
	}
}

func (cfg *APIConfig) HandleChirpRequest(w http.ResponseWriter, r *http.Request) {
	method := r.Method
	switch method {
	case "POST":
		cfg.HandleCreateChirp(w, r)
	case "PUT":
		cfg.HandleUpdateChirp(w, r)
	case "DELETE":
		cfg.HandleDeleteChirp(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Header().Set("Content-Type", "application/json")
		jsonResponse, _ := json.Marshal(ChirpError{Error: "Method not allowed"})
		w.Write(jsonResponse)
	}
}

func (cfg *APIConfig) HandleGetAllChirps(w http.ResponseWriter, r *http.Request) {
	// query params
	authorID := r.URL.Query().Get("author_id")
	sortOrder := r.URL.Query().Get("sort")
	if sortOrder != "asc" && sortOrder != "desc" {
		sortOrder = "asc"
	}

	chirps, err := cfg.dbQueries.GetAllChirps(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Header().Set("Content-Type", "application/json")
		jsonResponse, _ := json.Marshal(ChirpError{Error: fmt.Sprintf("Error getting all chirps: %v", err)})
		w.Write(jsonResponse)
		return
	}

	// return all chirps
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	responseChirps := []CompleteChirp{}
	for _, chirp := range chirps {
		if authorID != "" && (chirp.UserID != authorID || chirp.UserID == "") {
			continue
		}
		responseChirps = append(responseChirps, CompleteChirp{
			ID:        chirp.ID,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			Body:      chirp.Body,
			UserID:    chirp.UserID,
		})
	}
	// sort chirps by created_at (asc by default, desc if sort_order is desc)
	if sortOrder == "desc" {
		// fmt.Println("sorting chirps in descending order")
		// fmt.Printf("before: responseChirps: %+v\n", responseChirps)
		sort.Slice(responseChirps, func(i, j int) bool {
			return responseChirps[i].CreatedAt.After(responseChirps[j].CreatedAt)
		})
		// fmt.Printf("after: responseChirps: %+v\n", responseChirps)
	}
	// sortChirpsByCreatedAt(responseChirps, sortOrder)
	responseChirpsJSON, _ := json.Marshal(responseChirps)
	w.Write(responseChirpsJSON)

}

func (cfg *APIConfig) HandleGetChirpByID(w http.ResponseWriter, r *http.Request) {
	// return chirp by id
	path := r.URL.Path
	id := strings.Split(path, "/")[len(strings.Split(path, "/"))-1]
	// fmt.Println("id:", id)
	chirp, err := cfg.dbQueries.GetChirpByID(r.Context(), id)
	if err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusNotFound)
			w.Header().Set("Content-Type", "application/json")
			jsonResponse, _ := json.Marshal(ChirpError{Error: "Chirp not found"})
			w.Write(jsonResponse)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		w.Header().Set("Content-Type", "application/json")
		jsonResponse, _ := json.Marshal(ChirpError{Error: fmt.Sprintf("Error getting chirp by id: %v", err)})
		w.Write(jsonResponse)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	responseChirp, _ := json.Marshal(CompleteChirp{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserID:    chirp.UserID,
	})
	w.Write(responseChirp)
}

func (cfg *APIConfig) HandleCreateChirp(w http.ResponseWriter, r *http.Request) {
	type ValidChirpRequest struct {
		Body string `json:"body"`
	}
	bearerToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Header().Set("Content-Type", "application/json")
		jsonResponse, _ := json.Marshal(jwtError{Error: err.Error()})
		w.Write(jsonResponse)
		return
	}

	userID, err := auth.ValidateJWT(bearerToken, cfg.tokenSecret)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Header().Set("Content-Type", "application/json")
		jsonResponse, _ := json.Marshal(jwtError{Error: err.Error()})
		w.Write(jsonResponse)
		return
	}
	fmt.Println("userID [HandleCreateChirp]:\n", userID)
	if userID == uuid.Nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Header().Set("Content-Type", "application/json")
		jsonResponse, _ := json.Marshal(jwtError{Error: "Invalid token"})
		w.Write(jsonResponse)
		return
	}

	userIDString := userID.String()

	// validate content type
	postBody := &ValidChirpRequest{}
	if r.Header.Get("Content-Type") != "application/json" {
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Content-Type", "application/json")
		jsonResponse, _ := json.Marshal(ChirpError{Error: "Content-Type must be application/json"})
		w.Write(jsonResponse)
		return
	}

	// read request body
	bodyBytes, err := io.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Header().Set("Content-Type", "application/json")
		jsonResponse, _ := json.Marshal(ChirpError{Error: "Could not read request body"})
		w.Write(jsonResponse)
		return
	}

	// unmarshal request body
	err = json.Unmarshal(bodyBytes, postBody)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Content-Type", "application/json")
		jsonResponse, _ := json.Marshal(ChirpError{Error: "Invalid JSON"})
		w.Write(jsonResponse)
		return
	}

	// validate chirp length
	if len(postBody.Body) > 140 {
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Content-Type", "application/json")
		jsonResponse, _ := json.Marshal(ChirpError{Error: "Chirp is too long"})
		w.Write(jsonResponse)
		return
	}

	// clean chirp body
	cleanedBody := cleanChirpBody(postBody.Body)

	// check if user exists
	userExists, err := cfg.checkUserExists(userIDString)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Header().Set("Content-Type", "application/json")
		jsonResponse, _ := json.Marshal(ChirpError{Error: fmt.Sprintf("Error checking user exists: %v", err)})
		w.Write(jsonResponse)
		return
	}
	if !userExists {
		w.WriteHeader(http.StatusNotFound)
		w.Header().Set("Content-Type", "application/json")
		jsonResponse, _ := json.Marshal(ChirpError{Error: "User not found"})
		w.Write(jsonResponse)
		return
	}

	// create chirp
	chirp, err := cfg.dbQueries.CreateChirp(r.Context(), database.CreateChirpParams{
		ID:        uuid.New().String(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Body:      cleanedBody,
		UserID:    userIDString,
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Header().Set("Content-Type", "application/json")
		jsonResponse, _ := json.Marshal(ChirpError{Error: fmt.Sprintf("Error creating chirp: %v", err)})
		w.Write(jsonResponse)
		return
	}

	// return valid chirp response
	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")
	fullChirp, _ := json.Marshal(CompleteChirp{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserID:    chirp.UserID,
	})
	w.Write(fullChirp)
}

func (cfg *APIConfig) HandleUpdateChirp(w http.ResponseWriter, r *http.Request) {
	type ValidChirpRequest struct {
		ID   string `json:"id"`
		Body string `json:"body"`
	}

	// validate content type
	postBody := &ValidChirpRequest{}
	if r.Header.Get("Content-Type") != "application/json" {
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Content-Type", "application/json")
		jsonResponse, _ := json.Marshal(ChirpError{Error: "Content-Type must be application/json"})
		w.Write(jsonResponse)
		return
	}

	// read request body
	bodyBytes, err := io.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Header().Set("Content-Type", "application/json")
		jsonResponse, _ := json.Marshal(ChirpError{Error: "Could not read request body"})
		w.Write(jsonResponse)
		return
	}

	// unmarshal request body
	err = json.Unmarshal(bodyBytes, postBody)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Content-Type", "application/json")
		jsonResponse, _ := json.Marshal(ChirpError{Error: "Invalid JSON"})
		w.Write(jsonResponse)
		return
	}

	// validate chirp length
	if len(postBody.Body) > 140 {
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Content-Type", "application/json")
		jsonResponse, _ := json.Marshal(ChirpError{Error: "Chirp is too long"})
		w.Write(jsonResponse)
		return
	}

	// clean chirp body
	cleanedBody := cleanChirpBody(postBody.Body)

	// check if chirp exists
	chirp, err := cfg.dbQueries.GetChirpByID(r.Context(), postBody.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusNotFound)
			w.Header().Set("Content-Type", "application/json")
			jsonResponse, _ := json.Marshal(ChirpError{Error: "Chirp not found"})
			w.Write(jsonResponse)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		w.Header().Set("Content-Type", "application/json")
		jsonResponse, _ := json.Marshal(ChirpError{Error: fmt.Sprintf("Error checking chirp exists: %v", err)})
		w.Write(jsonResponse)
		return
	}

	// update chirp
	chirp, err = cfg.dbQueries.UpdateChirp(r.Context(), database.UpdateChirpParams{
		ID:        postBody.ID,
		Body:      cleanedBody,
		UpdatedAt: time.Now(),
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Header().Set("Content-Type", "application/json")
		jsonResponse, _ := json.Marshal(ChirpError{Error: fmt.Sprintf("Error updating chirp: %v", err)})
		w.Write(jsonResponse)
		return
	}

	// return updated chirp
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	fullChirp, _ := json.Marshal(CompleteChirp{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserID:    chirp.UserID,
	})
	w.Write(fullChirp)
}

func (cfg *APIConfig) HandleDeleteChirp(w http.ResponseWriter, r *http.Request) {
	bearerToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Header().Set("Content-Type", "application/json")
		jsonResponse, _ := json.Marshal(jwtError{Error: err.Error()})
		w.Write(jsonResponse)
		return
	}

	userID, err := auth.ValidateJWT(bearerToken, cfg.tokenSecret)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Header().Set("Content-Type", "application/json")
		jsonResponse, _ := json.Marshal(jwtError{Error: err.Error()})
		w.Write(jsonResponse)
		return
	}
	if userID == uuid.Nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Header().Set("Content-Type", "application/json")
		jsonResponse, _ := json.Marshal(jwtError{Error: "Invalid token"})
		w.Write(jsonResponse)
		return
	}
	userIDString := userID.String()

	path := r.URL.Path
	id := strings.TrimSpace(strings.Split(path, "/")[len(strings.Split(path, "/"))-1])
	// fmt.Println("id:", id)
	chirp, err := cfg.dbQueries.GetChirpByID(r.Context(), id)
	if err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusNotFound)
			w.Header().Set("Content-Type", "application/json")
			jsonResponse, _ := json.Marshal(ChirpError{Error: "Chirp not found"})
			w.Write(jsonResponse)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		w.Header().Set("Content-Type", "application/json")
		jsonResponse, _ := json.Marshal(ChirpError{Error: fmt.Sprintf("Error getting chirp by id: %v", err)})
		w.Write(jsonResponse)
		return
	}

	if chirp.UserID != userIDString {
		w.WriteHeader(http.StatusForbidden)
		w.Header().Set("Content-Type", "application/json")
		jsonResponse, _ := json.Marshal(ChirpError{Error: "You are not authorized to delete this chirp"})
		w.Write(jsonResponse)
		return
	}

	// delete chirp
	err = cfg.dbQueries.DeleteChirp(r.Context(), id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Header().Set("Content-Type", "application/json")
		jsonResponse, _ := json.Marshal(ChirpError{Error: fmt.Sprintf("Error deleting chirp: %v", err)})
		w.Write(jsonResponse)
		return
	}
	// return success message
	w.WriteHeader(http.StatusNoContent) // 204
	w.Header().Set("Content-Type", "application/json")
	jsonResponse, _ := json.Marshal(SuccessMessage{Message: "Chirp deleted successfully"})
	w.Write(jsonResponse)
}

func cleanChirpBody(body string) string {
	replaceMap := map[string]bool{
		"kerfuffle": true,
		"sharbert":  true,
		"fornax":    true,
	}
	splitStrings := strings.Split(body, " ")
	for _, word := range splitStrings {
		for key, _ := range replaceMap {
			if strings.Contains(strings.ToLower(word), key) {
				body = strings.ReplaceAll(body, word, "****")
			}
		}
	}
	return body
}

func (cfg *APIConfig) ValidateChirpRequest(w http.ResponseWriter, r *http.Request) {
	type ValidChirpRequest struct {
		Body string `json:"body"`
	}

	postBody := &ValidChirpRequest{}
	if r.Header.Get("Content-Type") != "application/json" {
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Content-Type", "application/json")
		jsonResponse, _ := json.Marshal(ChirpError{Error: "Content-Type must be application/json"})
		w.Write(jsonResponse)
		return
	}
	err := json.NewDecoder(r.Body).Decode(&postBody)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Content-Type", "application/json")
		jsonResponse, _ := json.Marshal(ChirpError{Error: "Invalid JSON"})
		w.Write(jsonResponse)
		return
	}

	// validate chirp length
	if len(postBody.Body) > 140 {
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Content-Type", "application/json")
		jsonResponse, _ := json.Marshal(ChirpError{Error: "Chirp is too long"})
		w.Write(jsonResponse)
		return
	}

	// clean chirp body
	cleanedBody := cleanChirpBody(postBody.Body)

	// return validated chirp response
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	validatedChirp, _ := json.Marshal(ValidatedChirpResponse{Body: cleanedBody})
	w.Write(validatedChirp)
}
