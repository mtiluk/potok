package config

import (
	"strings"
	"testing"
)

func setAllEnv(t *testing.T) {
	t.Helper()
	t.Setenv("ADDR", "8080")
	t.Setenv("DATA_DIR", "/home/")
	t.Setenv("DATABASE_URL", "file:potok.db")
}

func TestLoadConfigMissingRequiredEnv(t *testing.T) {
	requiredVars := []string{"ADDR", "DATA_DIR", "DATABASE_URL"}
	for _, v := range requiredVars {
		t.Setenv(v, "")
	}

	for _, missing := range requiredVars {
		t.Run("missing_"+missing, func(t *testing.T) {
			setAllEnv(t)
			t.Setenv(missing, "")

			_, err := LoadConfig()
			if err == nil {
				t.Fatalf("LoadConfig() expected error when %s is missing, got nil", missing)
			}
			if !strings.Contains(err.Error(), missing) {
				t.Errorf("error %q should mention the missing variable %s", err, missing)
			}
		})
	}
}

func TestLoadConfig(t *testing.T) {
	setAllEnv(t)

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() unexpected error: %v", err)
	}

	if cfg.Addr != "8080" {
		t.Errorf("Addr = %q, want %q", cfg.Addr, "8080")
	}

	if cfg.DataDir != "/home/" {
		t.Errorf("DataDir = %q, want %q", cfg.DataDir, "/home/")
	}

	if cfg.DatabaseURL != "file:potok.db" {
		t.Errorf("DatabaseURL = %q, want %q", cfg.DatabaseURL, "file:potok.db")
	}
}
