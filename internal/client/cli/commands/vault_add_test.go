package commands

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExpandHome(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("os.UserHomeDir() = %v", err)
	}

	tests := []struct {
		input string
		want  string
	}{
		{"~", home},
		{"~/notes", filepath.Join(home, "notes")},
		{"/absolute/path", "/absolute/path"},
		{"relative/path", "relative/path"},
		{"~notuser", "~notuser"},
	}

	for _, tt := range tests {
		got, err := expandHome(tt.input)
		if err != nil {
			t.Errorf("expandHome(%q) error = %v", tt.input, err)
			continue
		}
		if got != tt.want {
			t.Errorf("expandHome(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestResolveVaultPathRejectsEmptyInput(t *testing.T) {
	if _, err := resolveVaultPath(""); err == nil {
		t.Fatal("resolveVaultPath(\"\") = nil, want an error")
	}
}

func TestResolveVaultPathRejectsMissingPath(t *testing.T) {
	dir := t.TempDir()
	missing := filepath.Join(dir, "does-not-exist")

	if _, err := resolveVaultPath(missing); err == nil {
		t.Fatal("resolveVaultPath(missing) = nil, want an error")
	}
}

func TestResolveVaultPathRejectsFile(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "notes.txt")
	if err := os.WriteFile(file, []byte("hi"), 0o600); err != nil {
		t.Fatalf("os.WriteFile() = %v", err)
	}

	if _, err := resolveVaultPath(file); err == nil {
		t.Fatal("resolveVaultPath(file) = nil, want an error")
	}
}

func TestResolveVaultPathAcceptsDirectoryWithoutObsidianFolder(t *testing.T) {
	dir := t.TempDir()

	got, err := resolveVaultPath(dir)
	if err != nil {
		t.Fatalf("resolveVaultPath(%q) error = %v", dir, err)
	}

	want, err := filepath.Abs(dir)
	if err != nil {
		t.Fatalf("filepath.Abs() = %v", err)
	}
	if got != want {
		t.Errorf("resolveVaultPath(%q) = %q, want %q", dir, got, want)
	}
}

func TestResolveVaultPathAcceptsObsidianVault(t *testing.T) {
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, ".obsidian"), 0o700); err != nil {
		t.Fatalf("os.Mkdir() = %v", err)
	}

	if _, err := resolveVaultPath(dir); err != nil {
		t.Fatalf("resolveVaultPath(%q) error = %v", dir, err)
	}
}
