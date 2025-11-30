package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/landanqrew/go-serve-intro/internal/database"
)

type APIConfig struct {
	fileserverHits atomic.Int32
	dbQueries      *database.Queries
	tokenSecret    string
}

func deriveResponseJson[T any](w http.ResponseWriter, r *http.Request) (T, error) {
	type jsonReadError struct {
		Error string `json:"error"`
	}
	var t T

	if r.Header.Get("Content-Type") != "application/json" {
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("ContentType", "application/json")
		readError := jsonReadError{Error: fmt.Sprintf("content type (%s) is not 'application/json'", r.Header.Get("Content-Type"))}
		jsonResponse, err := json.Marshal(readError)
		if err != nil {
			errorBytes := []byte(fmt.Sprintf(`{"error":"%v"}`, readError.Error))
			w.Write(errorBytes)
			return t, err
		}
		w.Write(jsonResponse)
		return t, fmt.Errorf("content type (%s) is not 'application/json'", r.Header.Get("Content-Type"))
	}

	err := json.NewDecoder(r.Body).Decode(&t)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("ContentType", "application/json")
		readError := jsonReadError{Error: fmt.Sprintf("error decoding json: %v", err)}
		jsonResponse, err := json.Marshal(readError)
		if err != nil {
			errorBytes := []byte(fmt.Sprintf(`{"error":"%v"}`, readError.Error))
			w.Write(errorBytes)
			return t, err
		}
		w.Write(jsonResponse)
		return t, err
	}
	return t, nil
}



func GetAPIConfig(db *sql.DB) *APIConfig {
	return &APIConfig{
		fileserverHits: atomic.Int32{},
		dbQueries:      database.New(db),
		tokenSecret:    os.Getenv("TOKEN_SECRET"),
	}
}