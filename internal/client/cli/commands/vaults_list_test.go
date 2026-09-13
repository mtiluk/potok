package commands

import (
	"bytes"
	"io"
	"os"
	"testing"
	"time"

	"github.com/mtiluk/potok/internal/client/config"
)

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	orig := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe() = %v", err)
	}
	os.Stdout = w
	defer func() { os.Stdout = orig }()

	fn()

	if err := w.Close(); err != nil {
		t.Fatalf("close pipe: %v", err)
	}
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatalf("read pipe: %v", err)
	}
	return buf.String()
}

func TestVaultsListFailsWhenNotInitialised(t *testing.T) {
	t.Setenv("POTOK_CONFIG_DIR", t.TempDir())

	err := executeCmd(t, NewVaultsListCmd())
	if err == nil {
		t.Fatal("vaults-list = nil, want an error when not initialised")
	}
}

func TestVaultsListReportsNoneRegistered(t *testing.T) {
	t.Setenv("POTOK_CONFIG_DIR", t.TempDir())
	if err := config.Save(&config.Config{ServerURL: "https://potok.example.com"}); err != nil {
		t.Fatalf("config.Save() = %v", err)
	}

	out := captureStdout(t, func() {
		if err := executeCmd(t, NewVaultsListCmd()); err != nil {
			t.Fatalf("vaults-list = %v", err)
		}
	})
	if !bytes.Contains([]byte(out), []byte("No vaults registered")) {
		t.Errorf("output = %q, want it to mention no vaults are registered", out)
	}
}

func TestVaultsListShowsRegisteredVaults(t *testing.T) {
	t.Setenv("POTOK_CONFIG_DIR", t.TempDir())

	synced := time.Date(2026, 1, 2, 15, 4, 5, 0, time.UTC)
	cfg := &config.Config{
		ServerURL: "https://potok.example.com",
		Vaults: []config.Vault{
			{Name: "notes", Path: "/home/user/notes"},
			{Name: "journal", Path: "/home/user/journal", LastSyncedAt: &synced},
		},
	}
	if err := config.Save(cfg); err != nil {
		t.Fatalf("config.Save() = %v", err)
	}

	out := captureStdout(t, func() {
		if err := executeCmd(t, NewVaultsListCmd()); err != nil {
			t.Fatalf("vaults-list = %v", err)
		}
	})

	for _, want := range []string{"notes", "/home/user/notes", "never", "journal", "/home/user/journal"} {
		if !bytes.Contains([]byte(out), []byte(want)) {
			t.Errorf("output = %q, want it to contain %q", out, want)
		}
	}
}

func TestLastSyncedString(t *testing.T) {
	if got := lastSyncedString(nil); got != "never" {
		t.Errorf("lastSyncedString(nil) = %q, want %q", got, "never")
	}

	ts := time.Date(2026, 1, 2, 15, 4, 5, 0, time.UTC)
	got := lastSyncedString(&ts)
	if got == "never" || got == "" {
		t.Errorf("lastSyncedString(%v) = %q, want a formatted timestamp", ts, got)
	}
}
