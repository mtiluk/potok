package commands

import (
	"errors"
	"strings"
	"testing"

	"github.com/michaeltukdev/Potok/internal/client/config"
	"github.com/michaeltukdev/Potok/internal/client/secrets"
	"github.com/spf13/cobra"
	"github.com/zalando/go-keyring"
)

func setupVaultRemove(t *testing.T) {
	t.Helper()
	t.Setenv("POTOK_CONFIG_DIR", t.TempDir())
	keyring.MockInit()
}

func saveConfigWithVault(t *testing.T) *config.Config {
	t.Helper()
	cfg := &config.Config{
		ServerURL: "https://potok.example.com",
		Vaults:    []config.Vault{{Name: "notes", Path: t.TempDir()}},
	}
	if err := config.Save(cfg); err != nil {
		t.Fatalf("config.Save() = %v", err)
	}
	return cfg
}

func executeCmd(t *testing.T, cmd *cobra.Command, args ...string) error {
	t.Helper()
	cmd.SetArgs(args)
	return cmd.Execute()
}

func TestVaultRemoveRemovesRegisteredVault(t *testing.T) {
	setupVaultRemove(t)
	saveConfigWithVault(t)

	if err := executeCmd(t, NewVaultRemoveCmd(), "notes"); err != nil {
		t.Fatalf("vault-remove notes = %v", err)
	}

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("config.Load() = %v", err)
	}
	if len(cfg.Vaults) != 0 {
		t.Errorf("Vaults = %+v, want none after removal", cfg.Vaults)
	}
}

func TestVaultRemoveDeletesStoredPassphrase(t *testing.T) {
	setupVaultRemove(t)
	saveConfigWithVault(t)

	if err := secrets.Set(secrets.VaultKeyName("notes"), "supersecret"); err != nil {
		t.Fatalf("secrets.Set() = %v", err)
	}

	if err := executeCmd(t, NewVaultRemoveCmd(), "notes"); err != nil {
		t.Fatalf("vault-remove notes = %v", err)
	}

	if _, err := secrets.Get(secrets.VaultKeyName("notes")); !errors.Is(err, secrets.ErrNotFound) {
		t.Errorf("secrets.Get() after remove = %v, want ErrNotFound", err)
	}
}

func TestVaultRemoveFailsForUnknownVault(t *testing.T) {
	setupVaultRemove(t)
	saveConfigWithVault(t)

	err := executeCmd(t, NewVaultRemoveCmd(), "missing")
	if err == nil {
		t.Fatal("vault-remove missing = nil, want an error")
	}
	if !strings.Contains(err.Error(), `"missing"`) {
		t.Errorf("error = %q, want it to mention the vault name", err)
	}

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("config.Load() = %v", err)
	}
	if len(cfg.Vaults) != 1 {
		t.Errorf("Vaults = %+v, want the original vault left untouched", cfg.Vaults)
	}
}

func TestVaultRemoveFailsWhenNotInitialised(t *testing.T) {
	setupVaultRemove(t)

	err := executeCmd(t, NewVaultRemoveCmd(), "notes")
	if err == nil {
		t.Fatal("vault-remove = nil, want an error when not initialised")
	}
	if err.Error() != "not initialised, run `potok init` first" {
		t.Errorf("error = %q, want the not-initialised message", err)
	}
}
