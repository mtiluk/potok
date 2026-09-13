package commands

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/fatih/color"
	"github.com/mtiluk/potok/internal/client/config"
	"github.com/mtiluk/potok/internal/client/secrets"
	"github.com/spf13/cobra"
)

type remoteVault struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func NewRemoteListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "remote-list",
		Short: "List all remote vaults",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {

			// 1. Config load
			cfg, err := config.Load()
			if err != nil {
				return errors.New(color.RedString("Error loading config: %v", err))
			}

			// 2. API key check
			apiKey, err := secrets.Get(secrets.APIKey)
			if err != nil {
				return errors.New(color.RedString("Error getting API key: %v", err))
			}

			var vaults []remoteVault
			response, err := apiRequestJSON(cfg.ServerURL, apiKey, http.MethodGet, "/vaults", &vaults)
			if err != nil {
				return err
			}

			if response.StatusCode == http.StatusUnauthorized {
				return errors.New(color.RedString("Invalid or expired API key: %s", response.Status))
			}

			if response.StatusCode != http.StatusOK {
				return errors.New(color.RedString("Failed to list vaults: %s", response.Status))
			}

			for _, vault := range vaults {
				fmt.Println(vault.Name)
			}

			return nil
		},
	}
}
