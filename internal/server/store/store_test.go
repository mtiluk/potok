package store

import (
	"context"
	"errors"
	"testing"
)

func newTestStore(t *testing.T) *Store {
	t.Helper()
	ctx := context.Background()

	s, err := Open(ctx, ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { s.Close() })

	if err := s.Migrate(); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return s
}

func seedUser(t *testing.T, s *Store, email string) User {
	t.Helper()
	user, err := s.CreateUser(context.Background(), email, "hunter2")
	if err != nil {
		t.Fatalf("seed CreateUser(%q): %v", email, err)
	}
	return user
}

func TestCreateVault(t *testing.T) {
	s := newTestStore(t)
	user := seedUser(t, s, "a@example.com")

	vault, err := s.CreateVault(context.Background(), user.ID, "notes")
	if err != nil {
		t.Fatalf("CreateVault() error: %v", err)
	}

	if vault.ID == "" {
		t.Error("CreateVault() returned an empty ID")
	}
	if vault.Name != "notes" {
		t.Errorf("Name = %q, want %q", vault.Name, "notes")
	}
	if vault.CreatedAt.IsZero() {
		t.Error("CreatedAt is zero")
	}
	if vault.UpdatedAt.IsZero() {
		t.Error("UpdatedAt is zero")
	}
}

func TestCreateVaultDuplicateName(t *testing.T) {
	s := newTestStore(t)
	user := seedUser(t, s, "a@example.com")

	if _, err := s.CreateVault(context.Background(), user.ID, "notes"); err != nil {
		t.Fatalf("seed CreateVault: %v", err)
	}

	if _, err := s.CreateVault(context.Background(), user.ID, "notes"); err == nil {
		t.Error("CreateVault() with a name the user already has expected an error, got nil")
	}
}

func TestCreateVaultSameNameDifferentUsers(t *testing.T) {
	s := newTestStore(t)
	userA := seedUser(t, s, "a@example.com")
	userB := seedUser(t, s, "b@example.com")

	if _, err := s.CreateVault(context.Background(), userA.ID, "notes"); err != nil {
		t.Fatalf("CreateVault(userA): %v", err)
	}
	if _, err := s.CreateVault(context.Background(), userB.ID, "notes"); err != nil {
		t.Errorf("CreateVault(userB) with the same name as userA's vault: %v", err)
	}
}

func TestVaultByName(t *testing.T) {
	s := newTestStore(t)
	user := seedUser(t, s, "a@example.com")

	created, err := s.CreateVault(context.Background(), user.ID, "notes")
	if err != nil {
		t.Fatalf("seed CreateVault: %v", err)
	}

	got, err := s.VaultByName(context.Background(), user.ID, "notes")
	if err != nil {
		t.Fatalf("VaultByName() error: %v", err)
	}
	if got.ID != created.ID {
		t.Errorf("ID = %q, want %q", got.ID, created.ID)
	}
	if got.Name != "notes" {
		t.Errorf("Name = %q, want %q", got.Name, "notes")
	}
}

func TestVaultByNameNotFound(t *testing.T) {
	s := newTestStore(t)
	user := seedUser(t, s, "a@example.com")

	_, err := s.VaultByName(context.Background(), user.ID, "does-not-exist")
	if !errors.Is(err, ErrVaultNotFound) {
		t.Errorf("VaultByName() error = %v, want ErrVaultNotFound", err)
	}
}

func TestVaultByNameScopedToUser(t *testing.T) {
	s := newTestStore(t)
	userA := seedUser(t, s, "a@example.com")
	userB := seedUser(t, s, "b@example.com")

	if _, err := s.CreateVault(context.Background(), userA.ID, "notes"); err != nil {
		t.Fatalf("seed CreateVault: %v", err)
	}

	_, err := s.VaultByName(context.Background(), userB.ID, "notes")
	if !errors.Is(err, ErrVaultNotFound) {
		t.Errorf("VaultByName() for userB = %v, want ErrVaultNotFound (userA's vault must not be visible)", err)
	}
}

func TestListVaults(t *testing.T) {
	s := newTestStore(t)
	userA := seedUser(t, s, "a@example.com")
	userB := seedUser(t, s, "b@example.com")

	for _, name := range []string{"notes", "journal"} {
		if _, err := s.CreateVault(context.Background(), userA.ID, name); err != nil {
			t.Fatalf("seed CreateVault(%q): %v", name, err)
		}
	}
	if _, err := s.CreateVault(context.Background(), userB.ID, "work"); err != nil {
		t.Fatalf("seed CreateVault: %v", err)
	}

	gotA, err := s.ListVaults(context.Background(), userA.ID)
	if err != nil {
		t.Fatalf("ListVaults(userA) error: %v", err)
	}
	if len(gotA) != 2 {
		t.Errorf("ListVaults(userA) returned %d vaults, want 2", len(gotA))
	}

	gotB, err := s.ListVaults(context.Background(), userB.ID)
	if err != nil {
		t.Fatalf("ListVaults(userB) error: %v", err)
	}
	if len(gotB) != 1 {
		t.Errorf("ListVaults(userB) returned %d vaults, want 1", len(gotB))
	}
}

func TestListVaultsEmpty(t *testing.T) {
	s := newTestStore(t)
	user := seedUser(t, s, "a@example.com")

	got, err := s.ListVaults(context.Background(), user.ID)
	if err != nil {
		t.Fatalf("ListVaults() error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("ListVaults() = %d vaults, want 0", len(got))
	}
}

func TestDeleteVault(t *testing.T) {
	s := newTestStore(t)
	user := seedUser(t, s, "a@example.com")

	if _, err := s.CreateVault(context.Background(), user.ID, "notes"); err != nil {
		t.Fatalf("seed CreateVault: %v", err)
	}

	deleted, err := s.DeleteVault(context.Background(), user.ID, "notes")
	if err != nil {
		t.Fatalf("DeleteVault() error: %v", err)
	}
	if !deleted {
		t.Errorf("DeleteVault() = false, want true")
	}

	_, err = s.VaultByName(context.Background(), user.ID, "notes")
	if !errors.Is(err, ErrVaultNotFound) {
		t.Errorf("VaultByName() after delete = %v, want ErrVaultNotFound", err)
	}
}

func TestDeleteVaultScopedToUser(t *testing.T) {
	s := newTestStore(t)
	userA := seedUser(t, s, "a@example.com")
	userB := seedUser(t, s, "b@example.com")

	if _, err := s.CreateVault(context.Background(), userA.ID, "notes"); err != nil {
		t.Fatalf("seed CreateVault: %v", err)
	}

	deleted, err := s.DeleteVault(context.Background(), userB.ID, "notes")
	if err != nil {
		t.Fatalf("DeleteVault(userB) error: %v", err)
	}
	if deleted {
		t.Errorf("DeleteVault(userB) = true, want false (vault belongs to userA)")
	}

	if _, err := s.VaultByName(context.Background(), userA.ID, "notes"); err != nil {
		t.Errorf("userA's vault should still exist after userB's delete attempt, got: %v", err)
	}
}
