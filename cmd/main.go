package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	repo "github.com/adk2004/auction_service/internal/adapters/postgresql/sqlc"
	"github.com/adk2004/auction_service/internal/handlers"
	"github.com/adk2004/auction_service/internal/services"
	"github.com/adk2004/auction_service/router"
	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found")
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	conn, err := pgx.Connect(ctx, os.Getenv("GOOSE_DBSTRING"))
	if err != nil {
		slog.Error("DB connection failed", "error", err)
		os.Exit(1)
	}
	defer conn.Close(context.Background())

	repo := repo.New(conn)

	aucService, err := services.NewAuctionService(repo, conn)
	if err != nil {
		slog.Error("Failed to initialize auction service", "error", err)
		os.Exit(1)
	}
	defer aucService.Cancel()

	aucHandler := handlers.NewAuctionHandler(aucService)
	r := router.SetupRouter(aucHandler)

	server := &http.Server{
		Addr:    ":5001",
		Handler: r,
	}

	go func() {
		log.Println("Starting server on port 5001...")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Server error", "error", err)
			stop()
		}
	}()

	<-ctx.Done()
	log.Println("Shutting down...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		slog.Error("Graceful shutdown failed", "error", err)
	}
}
