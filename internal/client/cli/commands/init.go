package commands

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/fatih/color"
	"github.com/mtiluk/potok/internal/client/config"
	"github.com/mtiluk/potok/internal/client/secrets"
	"github.com/spf13/cobra"
)

func NewInitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Initialize the config",
		RunE: func(cmd *cobra.Command, args []string) error {
			color.New(color.FgCyan, color.Bold).Println("Welcome to Potok!")
			fmt.Println("Let's get you set up - you'll need your server URL and an API key.")
			fmt.Println()

			serverURL := prompt("Server URL")
			apiKey := prompt("API key")
			fmt.Println()

			if serverURL == "" || apiKey == "" {
				return errors.New(color.RedString("Both a server URL and an API key are required"))
			}

			fmt.Println("Validating API key and server URL...")
			fmt.Println()

			response, err := apiRequest(serverURL, apiKey, http.MethodGet, "/me")
			if err != nil {
				return err
			}
			defer response.Body.Close()

			if response.StatusCode == http.StatusUnauthorized {
				return errors.New(color.RedString("Unauthorized: Invalid API Key"))
			}

			if response.StatusCode != http.StatusOK {
				return fmt.Errorf("unexpected status code: %d", response.StatusCode)
			}

			cfg, err := config.Load()
			if err != nil && !errors.Is(err, config.ErrNotFound) {
				return err
			}
			if cfg == nil {
				cfg = &config.Config{}
			}
			cfg.ServerURL = serverURL

			if err := secrets.Set(secrets.APIKey, apiKey); err != nil {
				return fmt.Errorf("failed to store API key: %w", err)
			}

			if err := config.Save(cfg); err != nil {
				_ = secrets.Delete(secrets.APIKey)
				return err
			}

			fmt.Println(color.GreenString("Success! Your API key and server URL have been validated."))

			return nil
		},
	}
}
