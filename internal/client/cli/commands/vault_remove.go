package commands

import (
	"errors"
	"fmt"

	"github.com/fatih/color"
	"github.com/mtiluk/potok/internal/client/config"
	"github.com/mtiluk/potok/internal/client/secrets"
	"github.com/spf13/cobra"
)

func NewVaultRemoveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "vault-remove <name>",
		Short: "Remove a vault from local config",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			cfg, err := config.Load()
			if errors.Is(err, config.ErrNotFound) {
				return errors.New(color.RedString("Not initialised, run `potok init` first"))
			}
			if err != nil {
				return err
			}

			if !cfg.RemoveVault(name) {
				return errors.New(color.RedString("Vault %q is not registered locally", name))
			}

			if err := secrets.Delete(secrets.VaultKeyName(name)); err != nil && !errors.Is(err, secrets.ErrNotFound) {
				return fmt.Errorf("failed to delete stored passphrase for vault %q: %w", name, err)
			}

			if err := config.Save(cfg); err != nil {
				return err
			}

			fmt.Println(color.GreenString("Removed vault %q and its stored passphrase.", name))
			return nil
		},
	}
}
