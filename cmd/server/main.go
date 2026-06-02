package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/CraveCompass/api/internal/application"
	"github.com/CraveCompass/api/internal/infrastructure/database"
	"github.com/CraveCompass/api/internal/infrastructure/memory"
	apiHTTP "github.com/CraveCompass/api/internal/interfaces/http"
	"github.com/CraveCompass/api/internal/interfaces/ws"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL environment variable is not set")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("Starting CraveCompass API on port %s...\n", port)

	dbPool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}
	defer dbPool.Close()

	fmt.Println("Connected to PostgreSQL/PostGIS successfully.")

	restaurantRepo := database.NewPostgresRestaurantRepo(dbPool)
	sessionRepo := memory.NewInMemorySessionRepo()
	wsHub := ws.NewHub()

	createSessionUC := application.NewCreateSessionUseCase(restaurantRepo, sessionRepo)
	submitVoteUC := application.NewSubmitVoteUseCase(sessionRepo)

	sessionHandler := apiHTTP.NewSessionHandler(createSessionUC)
	wsHandler := ws.NewWSHandler(wsHub, submitVoteUC)

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy"}`))
	})
	http.HandleFunc("/sessions", sessionHandler.HandleCreateRoom)
	http.HandleFunc("/ws", wsHandler.HandleConnection)

	log.Fatal(http.ListenAndServe(":"+port, nil))
}
