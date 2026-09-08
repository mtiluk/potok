package commands

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/fatih/color"
	"github.com/michaeltukdev/Potok/internal/client/config"
	"github.com/michaeltukdev/Potok/internal/client/secrets"
	"github.com/spf13/cobra"
)

type Vault struct {
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

			fmt.Println(color.GreenString("✓") + " Config loaded successfully")
			fmt.Println()

			// 2. API key check
			apiKey, err := secrets.Get(secrets.APIKey)
			if err != nil {
				return errors.New(color.RedString("Error getting API key: %v", err))
			}

			response, err := apiRequest(cfg.ServerURL, apiKey, http.MethodGet, "/vaults")
			if err != nil {
				return err
			}

			defer response.Body.Close()

			if response.StatusCode == http.StatusUnauthorized {
				return errors.New(color.RedString("Invalid or expired API key: %s", response.Status))
			}

			var vaults []Vault
			if err := json.NewDecoder(response.Body).Decode(&vaults); err != nil {
				return errors.New(color.RedString("Failed to decode vaults: %v", err))
			}
			for _, vault := range vaults {
				fmt.Println(vault.Name)
			}

			return nil
		},
	}
}
