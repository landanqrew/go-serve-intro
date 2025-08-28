package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync/atomic"

	"github.com/joho/godotenv"
	"github.com/landanqrew/go-serve-intro/internal/database"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	dbQueries *database.Queries
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		fmt.Printf("new hit count %d\n", cfg.fileserverHits.Load())
		next.ServeHTTP(w, r)
	})
}

func middlewareLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) getFileserverHits() int32 {
	return cfg.fileserverHits.Load()
}

func (cfg *apiConfig) resetFileserverHits() {
	cfg.fileserverHits = atomic.Int32{}
	fmt.Printf("fileserverHits reset. current val %d\n", cfg.getFileserverHits())
}

func generateAdminMetricsHTML(hits int32) string {
	return fmt.Sprintf(`<html>
  <body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited %d times!</p>
  </body>
</html>`, hits)
}

func main() {
	err := godotenv.Load("./secrets/db_config.env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	mux := &http.ServeMux{}
	cfg := apiConfig{
		fileserverHits: atomic.Int32{},
		dbQueries: database.New(db),
	}
	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}
	

	fmt.Printf("Starting server on port %s\n", server.Addr)

	
	mux.Handle("/app/", http.StripPrefix("/app/", cfg.middlewareMetricsInc(http.FileServer(http.Dir(".")))))
	mux.HandleFunc("GET /api/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("OK"))
	})
	mux.HandleFunc("POST /api/validate_chirp", func(w http.ResponseWriter, r *http.Request) {
		validateChirp(w, r)
	})
	mux.HandleFunc("GET /admin/metrics",  func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(generateAdminMetricsHTML(cfg.getFileserverHits())))
	})
	mux.HandleFunc("POST /admin/reset",  func(w http.ResponseWriter, r *http.Request) {
		cfg.resetFileserverHits()
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("OK"))
	})


	err = server.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}

func validateChirp(w http.ResponseWriter, r *http.Request) {
	type Chirp struct {
		Body string `json:"body"`
	}

	type ChirpError struct {
		Error string `json:"error"`
	}
	type ValidChirpResponse struct {
		CleanedBody string `json:"cleaned_body"`
	}

	// validate content type
	postBody := &Chirp{}
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

	replaceMap := map[string]bool{
		"kerfuffle": true,
		"sharbert": true,
		"fornax": true,
	}
	cleanedBody := postBody.Body
	splitStrings := strings.Split(cleanedBody, " ")
	for _, word := range splitStrings {
		for key, _ := range replaceMap {
			if strings.Contains(strings.ToLower(word), key) {
				cleanedBody = strings.ReplaceAll(cleanedBody, word, "****")
			}
		}
	}

	// return valid chirp response
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	jsonResponse, _ := json.Marshal(ValidChirpResponse{CleanedBody: cleanedBody})
	w.Write(jsonResponse)
}