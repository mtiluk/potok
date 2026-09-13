package commands

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/mtiluk/potok/internal/client/config"
	"github.com/mtiluk/potok/internal/client/secrets"
	"github.com/spf13/cobra"
)

func NewVaultAddCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "vault-add <name>",
		Short: "Register a local folder as a vault",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			if err := config.ValidateVaultName(name); err != nil {
				return err
			}

			path, err := resolveVaultPath(prompt("Vault path"))
			if err != nil {
				return err
			}

			cfg, err := config.Load()
			if errors.Is(err, config.ErrNotFound) {
				return errors.New(color.RedString("Not initialised, run `potok init` first"))
			}
			if err != nil {
				return err
			}

			if _, exists := cfg.Vault(name); exists {
				return errors.New(color.RedString("Vault %q is already registered", name))
			}

			passphrase, err := promptPassphrase("Encryption passphrase")
			if err != nil {
				return err
			}

			if err := secrets.Set(secrets.VaultKeyName(name), passphrase); err != nil {
				return fmt.Errorf("failed to store passphrase for vault %q: %w", name, err)
			}

			if err := cfg.AddVault(config.Vault{Name: name, Path: path}); err != nil {
				_ = secrets.Delete(secrets.VaultKeyName(name))
				return err
			}

			if err := config.Save(cfg); err != nil {
				_ = secrets.Delete(secrets.VaultKeyName(name))
				return err
			}

			fmt.Println(color.GreenString("Registered vault %q at %s.", name, path))
			fmt.Println("This only registers the vault locally — nothing is uploaded yet.")

			return nil
		},
	}
}

func resolveVaultPath(input string) (string, error) {
	if input == "" {
		return "", errors.New(color.RedString("A vault path is required"))
	}

	expanded, err := expandHome(input)
	if err != nil {
		return "", err
	}

	abs, err := filepath.Abs(expanded)
	if err != nil {
		return "", fmt.Errorf("failed to resolve path %q: %w", input, err)
	}

	info, err := os.Stat(abs)
	if errors.Is(err, fs.ErrNotExist) {
		return "", errors.New(color.RedString("Path %q does not exist", abs))
	}
	if err != nil {
		return "", fmt.Errorf("failed to read %q: %w", abs, err)
	}
	if !info.IsDir() {
		return "", errors.New(color.RedString("Path %q is not a directory", abs))
	}
	if _, err := os.ReadDir(abs); err != nil {
		return "", errors.New(color.RedString("Path %q is not readable: %v", abs, err))
	}

	if _, err := os.Stat(filepath.Join(abs, ".obsidian")); errors.Is(err, fs.ErrNotExist) {
		fmt.Println(color.YellowString("Warning: %q doesn't look like an Obsidian vault (no .obsidian folder found).", abs))
	}

	return abs, nil
}

func expandHome(path string) (string, error) {
	if path != "~" && !strings.HasPrefix(path, "~/") {
		return path, nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to resolve home directory: %w", err)
	}
	if path == "~" {
		return home, nil
	}
	return filepath.Join(home, path[2:]), nil
}
