package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"

	repo "github.com/adk2004/auction_service/internal/adapters/postgresql/sqlc"
	"github.com/adk2004/auction_service/internal/handlers"
	"github.com/adk2004/auction_service/internal/services"
	"github.com/adk2004/auction_service/router"
	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using environment variables")
	}

	ctx := context.Background()

	conn, err := pgx.Connect(ctx, os.Getenv("GOOSE_DBSTRING"))
	if err != nil {
		slog.Error("DB connection failed", "error", err)
		os.Exit(1)
	}
	defer conn.Close(ctx)

	repo := repo.New(conn)
	aucService, err := services.NewAuctionService(repo, conn)
	if err != nil {
		slog.Error("Failed to initialize auction service", "error", err)
		os.Exit(1)
	}
	aucHandler := handlers.NewAuctionHandler(aucService)

	r := router.SetupRouter(aucHandler)

	log.Printf("Starting server on port 5001...")
	if err := http.ListenAndServe(":5001", r); err != nil {
		panic(err)
	}
}
