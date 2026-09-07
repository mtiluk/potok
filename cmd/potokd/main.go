package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"

	nethttp "net/http"

	"github.com/michaeltukdev/Potok/internal/server/config"
	httpapi "github.com/michaeltukdev/Potok/internal/server/http"
	"github.com/michaeltukdev/Potok/internal/server/store"
)

func main() {
	slog.Info("Starting Potok server...")
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "LoadConfig() error: %v\n", err)
		os.Exit(1)
	}
	slog.Info("Config loaded successfully", "addr", cfg.Addr, "dataDir", cfg.DataDir, "databaseURL", cfg.DatabaseURL)

	conn, err := store.Open(context.Background(), cfg.DatabaseURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Open() error: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close()
	slog.Info("Database connection established")

	slog.Info("Running database migrations")
	if err := conn.Migrate(); err != nil {
		fmt.Fprintf(os.Stderr, "Migrate() error: %v\n", err)
		os.Exit(1)
	}
	slog.Info("Database migrations completed successfully")

	slog.Info("Starting HTTP server")
	handler := httpapi.NewHandler(conn)
	nethttp.HandleFunc("GET /health", handler.Health)
	nethttp.HandleFunc("POST /register", handler.Register)
	nethttp.Handle("GET /me", handler.APIKeyAuth(nethttp.HandlerFunc(handler.Me)))
	log.Fatal(nethttp.ListenAndServe(cfg.Addr, nil))
}
