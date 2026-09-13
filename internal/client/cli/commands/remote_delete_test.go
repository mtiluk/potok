package commands

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mtiluk/potok/internal/client/config"
	"github.com/mtiluk/potok/internal/client/secrets"
	"github.com/zalando/go-keyring"
)

func setupRemoteDelete(t *testing.T, serverURL string) {
	t.Helper()
	t.Setenv("POTOK_CONFIG_DIR", t.TempDir())
	keyring.MockInit()

	if err := config.Save(&config.Config{ServerURL: serverURL}); err != nil {
		t.Fatalf("config.Save() = %v", err)
	}
	if err := secrets.Set(secrets.APIKey, "test-key"); err != nil {
		t.Fatalf("secrets.Set() = %v", err)
	}
}

func TestRemoteDeleteReportsServerResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete || r.URL.Path != "/vaults/notes" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer test-key" {
			t.Errorf("Authorization header = %q, want %q", got, "Bearer test-key")
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Vault deleted"))
	}))
	defer server.Close()

	setupRemoteDelete(t, server.URL)

	if err := executeCmd(t, NewRemoteDeleteCmd(), "notes"); err != nil {
		t.Fatalf("remote-delete notes = %v", err)
	}
}

func TestRemoteDeleteReportsNoVaultDeleted(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("No vault deleted"))
	}))
	defer server.Close()

	setupRemoteDelete(t, server.URL)

	if err := executeCmd(t, NewRemoteDeleteCmd(), "missing"); err != nil {
		t.Fatalf("remote-delete missing = %v", err)
	}
}

func TestRemoteDeleteFailsOnUnauthorized(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
	}))
	defer server.Close()

	setupRemoteDelete(t, server.URL)

	err := executeCmd(t, NewRemoteDeleteCmd(), "notes")
	if err == nil {
		t.Fatal("remote-delete notes = nil, want an error")
	}
}

func TestRemoteDeleteFailsOnServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	}))
	defer server.Close()

	setupRemoteDelete(t, server.URL)

	err := executeCmd(t, NewRemoteDeleteCmd(), "notes")
	if err == nil {
		t.Fatal("remote-delete notes = nil, want an error")
	}
}

func TestRemoteDeleteFailsWhenNotInitialised(t *testing.T) {
	t.Setenv("POTOK_CONFIG_DIR", t.TempDir())
	keyring.MockInit()

	err := executeCmd(t, NewRemoteDeleteCmd(), "notes")
	if err == nil {
		t.Fatal("remote-delete notes = nil, want an error when not initialised")
	}
}
