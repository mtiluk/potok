package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"

	nethttp "net/http"

	"github.com/mtiluk/potok/internal/server/blobstore"
	"github.com/mtiluk/potok/internal/server/config"
	httpapi "github.com/mtiluk/potok/internal/server/http"
	"github.com/mtiluk/potok/internal/server/store"
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
	handler := httpapi.NewHandler(conn, *blobstore.New(cfg.DataDir))
	nethttp.HandleFunc("GET /health", handler.Health)
	nethttp.HandleFunc("POST /register", handler.Register)
	nethttp.Handle("POST /vaults", handler.APIKeyAuth(nethttp.HandlerFunc(handler.CreateVault)))
	nethttp.Handle("GET /vaults", handler.APIKeyAuth(nethttp.HandlerFunc(handler.ListVaults)))
	nethttp.Handle("GET /vaults/{name}", handler.APIKeyAuth(nethttp.HandlerFunc(handler.VaultByName)))
	nethttp.Handle("DELETE /vaults/{name}", handler.APIKeyAuth(nethttp.HandlerFunc(handler.DeleteVault)))
	nethttp.Handle("PUT /vaults/{name}/blobs/{id}", handler.APIKeyAuth(nethttp.HandlerFunc(handler.PutBlob)))
	nethttp.Handle("GET /vaults/{name}/blobs/{id}", handler.APIKeyAuth(nethttp.HandlerFunc(handler.GetBlob)))
	nethttp.Handle("GET /me", handler.APIKeyAuth(nethttp.HandlerFunc(handler.Me)))
	log.Fatal(nethttp.ListenAndServe(cfg.Addr, nil))
}
