package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"github.com/joho/godotenv"
	"github.com/landanqrew/go-serve-intro/internal/api"
	_ "github.com/lib/pq"
)



func middlewareLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
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
	err := godotenv.Load("./secrets/.env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	dbURL := os.Getenv("DB_URL")
	platform := os.Getenv("PLATFORM")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	mux := &http.ServeMux{}
	cfg := api.GetAPIConfig(db)
	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	fmt.Printf("Starting server on port %s\n", server.Addr)

	mux.Handle("/app/", http.StripPrefix("/app/", cfg.MiddlewareMetricsInc(http.FileServer(http.Dir(".")))))
	mux.HandleFunc("GET /api/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("OK"))
	})
	mux.HandleFunc("POST /api/validate_chirp", func(w http.ResponseWriter, r *http.Request) {
		cfg.ValidateChirpRequest(w, r)
	})
	mux.HandleFunc("GET /api/chirps", func(w http.ResponseWriter, r *http.Request) {
		cfg.HandleGetAllChirps(w, r)
	})
	mux.HandleFunc("GET /api/chirps/{id}", func(w http.ResponseWriter, r *http.Request) {
		cfg.HandleGetChirpByID(w, r)
	})
	mux.HandleFunc("POST /api/chirps", func(w http.ResponseWriter, r *http.Request) {
		cfg.HandleCreateChirp(w, r)
	})
	mux.HandleFunc("PUT /api/chirps", func(w http.ResponseWriter, r *http.Request) {
		cfg.HandleUpdateChirp(w, r)
	})
	mux.HandleFunc("DELETE /api/chirps/{chirp_id}", func(w http.ResponseWriter, r *http.Request) {
		cfg.HandleDeleteChirp(w, r)
	})
	mux.HandleFunc("POST /api/users", func(w http.ResponseWriter, r *http.Request) {
		cfg.HandleCreateUser(w, r)
	})
	mux.HandleFunc("PUT /api/users", func(w http.ResponseWriter, r *http.Request) {
		cfg.HandleUpdateUser(w, r)
	})
	mux.HandleFunc("POST /api/login", func(w http.ResponseWriter, r *http.Request) {
		cfg.HandleAuthenticateUser(w, r)
	})
	mux.HandleFunc("POST /api/refresh", func(w http.ResponseWriter, r *http.Request) {
		cfg.HandleTokenRefresh(w, r)
	})
	mux.HandleFunc("POST /api/revoke", func(w http.ResponseWriter, r *http.Request) {
		cfg.HandleTokenRevoke(w, r)
	})
	mux.HandleFunc("GET /admin/metrics", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(generateAdminMetricsHTML(cfg.GetFileserverHits())))
	})
	mux.HandleFunc("POST /admin/reset", func(w http.ResponseWriter, r *http.Request) {
		if platform != "dev" {
			w.WriteHeader(http.StatusForbidden)
			w.Header().Set("Content-Type", "text/plain")
			w.Write([]byte("Forbidden"))
			return
		}
		cfg.ResetFileserverHits()
		cfg.DeleteAllUsers(context.Background())
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("OK"))
	})
	mux.HandleFunc("POST /api/polka/webhooks", func(w http.ResponseWriter, r *http.Request) {
		cfg.HandleUpdateUserSetChirpyRed(w, r)
	})

	err = server.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}


