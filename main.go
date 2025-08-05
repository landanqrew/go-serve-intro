package main

import (
	"log"
	"net/http"
)

func main() {
	mux := &http.ServeMux{}
	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	
	/*mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	})*/
	mux.Handle("/", http.FileServer(http.Dir(".")))

	err := server.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}