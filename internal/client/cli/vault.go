package cli

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/fatih/color"
	"github.com/michaeltukdev/Potok/internal/client/config"
	"github.com/michaeltukdev/Potok/internal/client/secrets"
	"github.com/spf13/cobra"
)

func runInit() *cobra.Command {
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

func runVaultAdd() *cobra.Command {
	return &cobra.Command{
		Use:   "vault-add <name>",
		Short: "Register a local folder as a vault",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// // 1. Validate the name. config.ValidateVaultName already does this.
			// if err := config.ValidateVaultName(name); err != nil {
			// 	return err
			// }

			// // 2. Resolve the path to absolute and check it is a readable directory.
			// //    Warn (don't fail) if there is no .obsidian folder.

			// // 3. Load the config. config.ErrNotFound means run `potok init` first.
			// cfg, err := config.Load()
			// if errors.Is(err, config.ErrNotFound) {
			// 	return errors.New("not initialised, run `potok init` first")
			// }
			// if err != nil {
			// 	return err
			// }

			// // 4. Reject a name that is already registered, before prompting for
			// //    anything — nobody wants to type a passphrase twice and then be told.

			// // 5. Prompt for the passphrase, twice, no echo. Reject empty; reject
			// //    mismatched.

			// // 6. Store the passphrase in the keyring under secrets.VaultKey(name).

			// // 7. Add the vault to the config and save. If this fails, delete the
			// //    keyring entry so a retry isn't blocked by an orphaned passphrase.

			return nil
		},
	}
}
