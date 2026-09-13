package commands

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/fatih/color"
	"github.com/mtiluk/potok/internal/client/config"
	"github.com/mtiluk/potok/internal/client/secrets"
	"github.com/spf13/cobra"
)

func NewRemoteDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "remote-delete <name>",
		Short: "Delete a remote vault",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			vaultName := args[0]

			cfg, err := config.Load()
			if err != nil {
				return errors.New(color.RedString("Error loading config: %v", err))
			}

			apiKey, err := secrets.Get(secrets.APIKey)
			if err != nil {
				return errors.New(color.RedString("Error getting API key: %v", err))
			}

			response, err := apiRequest(cfg.ServerURL, apiKey, http.MethodDelete, "/vaults/"+vaultName)
			if err != nil {
				return err
			}
			defer response.Body.Close()

			body, err := io.ReadAll(response.Body)
			if err != nil {
				return fmt.Errorf("failed to read response: %w", err)
			}

			if response.StatusCode == http.StatusUnauthorized {
				return errors.New(color.RedString("Invalid or expired API key: %s", response.Status))
			}

			if response.StatusCode != http.StatusOK {
				return errors.New(color.RedString("Failed to delete vault: %s", response.Status))
			}

			fmt.Println(strings.TrimSpace(string(body)))
			return nil
		},
	}
}
