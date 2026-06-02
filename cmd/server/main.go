package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	fmt.Println("Starting CraveCompass API...")
	
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy"}`))
	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}