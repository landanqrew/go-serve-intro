package api

import (
	"fmt"
	"net/http"
	"sync/atomic"
)

func (cfg *APIConfig) GetFileserverHits() int32 {
	return cfg.fileserverHits.Load()
}

func (cfg *APIConfig) ResetFileserverHits() {
	cfg.fileserverHits = atomic.Int32{}
	fmt.Printf("fileserverHits reset. current val %d\n", cfg.GetFileserverHits())
}

func (cfg *APIConfig) MiddlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		fmt.Printf("new hit count %d\n", cfg.fileserverHits.Load())
		next.ServeHTTP(w, r)
	})
}