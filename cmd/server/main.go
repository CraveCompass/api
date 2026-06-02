package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/CraveCompass/api/internal/application"
	"github.com/CraveCompass/api/internal/infrastructure/memory"
	apiHTTP "github.com/CraveCompass/api/internal/interfaces/http"
)

func main() {
	fmt.Println("Starting CraveCompass API on port 8080...")

	restaurantRepo := memory.NewInMemoryRestaurantRepo()
	sessionRepo := memory.NewInMemorySessionRepo()

	createSessionUC := application.NewCreateSessionUseCase(restaurantRepo, sessionRepo)

	sessionHandler := apiHTTP.NewSessionHandler(createSessionUC)

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy"}`))
	})
	http.HandleFunc("/sessions", sessionHandler.HandleCreateRoom)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
