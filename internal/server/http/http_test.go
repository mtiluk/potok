package http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mtiluk/potok/internal/server/blobstore"
	"github.com/mtiluk/potok/internal/server/store"
)

func newTestStore(t *testing.T) *store.Store {
	t.Helper()
	ctx := context.Background()

	s, err := store.Open(ctx, ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { s.Close() })

	if err := s.Migrate(); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return s
}

func newTestBlobStore(t *testing.T) blobstore.Store {
	t.Helper()
	return *blobstore.New(t.TempDir())
}

func TestHealthEndpoint(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet, "/health", nil)

	response := httptest.NewRecorder()
	handler := NewHandler(nil, newTestBlobStore(t))
	handler.Health(response, req)

	if response.Code != http.StatusOK {
		t.Errorf("expected status code %d, got %d", http.StatusOK, response.Code)
	}

	if response.Body.String() != "OK" {
		t.Errorf("expected body %s, got %s", "OK", response.Body.String())
	}
}

func TestMeEndpoint(t *testing.T) {
	store := newTestStore(t)
	handler := NewHandler(store, newTestBlobStore(t))

	user, err := store.CreateUser(context.Background(), "a@example.com", "hunter2")
	if err != nil {
		t.Fatalf("seed CreateUser: %v", err)
	}

	tests := []struct {
		name       string
		authHeader string
		wantStatus int
	}{
		{
			name:       "valid api key",
			authHeader: "Bearer " + user.APIKey,
			wantStatus: http.StatusOK,
		},
		{
			name:       "missing authorization header",
			authHeader: "",
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "malformed authorization header",
			authHeader: user.APIKey,
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "unknown api key",
			authHeader: "Bearer potok_doesnotexist",
			wantStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/me", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			response := httptest.NewRecorder()
			handler.APIKeyAuth(http.HandlerFunc(handler.Me)).ServeHTTP(response, req)

			if response.Code != tt.wantStatus {
				t.Fatalf("expected status code %d, got %d (body: %s)",
					tt.wantStatus, response.Code, response.Body.String())
			}

			if tt.wantStatus != http.StatusOK {
				return
			}

			var got struct {
				ID           string `json:"id"`
				Email        string `json:"email"`
				PasswordHash string `json:"password_hash"`
			}
			if err := json.Unmarshal(response.Body.Bytes(), &got); err != nil {
				t.Fatalf("decode response: %v (body: %s)", err, response.Body.String())
			}

			if got.Email != user.Email {
				t.Errorf("Email = %q, want %q", got.Email, user.Email)
			}
			if got.PasswordHash != "" {
				t.Errorf("response leaked password_hash: %q", got.PasswordHash)
			}
		})
	}
}

func TestCreateVaultEndpoint(t *testing.T) {
	tests := []struct {
		name       string
		seed       string
		input      string
		wantStatus int
	}{
		{
			name:       "valid",
			input:      `{"name":"notes"}`,
			wantStatus: http.StatusCreated,
		},
		{
			name:       "malformed json",
			input:      `{"name":`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "missing name",
			input:      `{}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid name",
			input:      `{"name":"my vault"}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "duplicate name",
			seed:       `{"name":"notes"}`,
			input:      `{"name":"notes"}`,
			wantStatus: http.StatusConflict,
		},
		{
			name:       "duplicate name different case",
			seed:       `{"name":"Notes"}`,
			input:      `{"name":"notes"}`,
			wantStatus: http.StatusConflict,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s := newTestStore(t)
			handler := NewHandler(s, newTestBlobStore(t))

			user, err := s.CreateUser(context.Background(), "a@example.com", "hunter2")
			if err != nil {
				t.Fatalf("seed CreateUser: %v", err)
			}

			post := func(body string) *httptest.ResponseRecorder {
				req := httptest.NewRequest(http.MethodPost, "/vaults", strings.NewReader(body))
				req.Header.Set("Authorization", "Bearer "+user.APIKey)
				rec := httptest.NewRecorder()
				handler.APIKeyAuth(http.HandlerFunc(handler.CreateVault)).ServeHTTP(rec, req)
				return rec
			}

			if test.seed != "" {
				if rec := post(test.seed); rec.Code != http.StatusCreated {
					t.Fatalf("seed request failed: %d (body: %s)", rec.Code, rec.Body.String())
				}
			}

			response := post(test.input)
			if response.Code != test.wantStatus {
				t.Errorf("expected status code %d, got %d (body: %s)",
					test.wantStatus, response.Code, response.Body.String())
			}
		})
	}

	t.Run("unauthorized without middleware", func(t *testing.T) {
		handler := NewHandler(newTestStore(t), newTestBlobStore(t))
		req := httptest.NewRequest(http.MethodPost, "/vaults", strings.NewReader(`{"name":"notes"}`))
		rec := httptest.NewRecorder()
		handler.CreateVault(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Errorf("expected status code %d, got %d", http.StatusUnauthorized, rec.Code)
		}
	})
}

func TestListVaultsEndpoint(t *testing.T) {
	s := newTestStore(t)
	handler := NewHandler(s, newTestBlobStore(t))

	userA, err := s.CreateUser(context.Background(), "a@example.com", "hunter2")
	if err != nil {
		t.Fatalf("seed CreateUser(a): %v", err)
	}
	userB, err := s.CreateUser(context.Background(), "b@example.com", "hunter2")
	if err != nil {
		t.Fatalf("seed CreateUser(b): %v", err)
	}

	for _, name := range []string{"notes", "journal"} {
		if _, err := s.CreateVault(context.Background(), userA.ID, name); err != nil {
			t.Fatalf("seed CreateVault(%q): %v", name, err)
		}
	}

	list := func(apiKey string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodGet, "/vaults", nil)
		req.Header.Set("Authorization", "Bearer "+apiKey)
		rec := httptest.NewRecorder()
		handler.APIKeyAuth(http.HandlerFunc(handler.ListVaults)).ServeHTTP(rec, req)
		return rec
	}

	t.Run("returns only the caller's vaults", func(t *testing.T) {
		rec := list(userA.APIKey)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected status code %d, got %d (body: %s)", http.StatusOK, rec.Code, rec.Body.String())
		}

		var got []store.Vault
		if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
			t.Fatalf("decode response: %v (body: %s)", err, rec.Body.String())
		}
		if len(got) != 2 {
			t.Errorf("expected 2 vaults, got %d", len(got))
		}
	})

	t.Run("empty for a user with no vaults", func(t *testing.T) {
		rec := list(userB.APIKey)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected status code %d, got %d (body: %s)", http.StatusOK, rec.Code, rec.Body.String())
		}

		var got []store.Vault
		if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
			t.Fatalf("decode response: %v (body: %s)", err, rec.Body.String())
		}
		if len(got) != 0 {
			t.Errorf("expected 0 vaults, got %d", len(got))
		}
	})

	t.Run("unauthorized without middleware", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/vaults", nil)
		rec := httptest.NewRecorder()
		handler.ListVaults(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Errorf("expected status code %d, got %d", http.StatusUnauthorized, rec.Code)
		}
	})
}

func TestVaultByNameEndpoint(t *testing.T) {
	s := newTestStore(t)
	handler := NewHandler(s, newTestBlobStore(t))

	user, err := s.CreateUser(context.Background(), "a@example.com", "hunter2")
	if err != nil {
		t.Fatalf("seed CreateUser: %v", err)
	}
	if _, err := s.CreateVault(context.Background(), user.ID, "notes"); err != nil {
		t.Fatalf("seed CreateVault: %v", err)
	}

	get := func(name string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodGet, "/vaults/"+name, nil)
		req.Header.Set("Authorization", "Bearer "+user.APIKey)
		if name != "" {
			req.SetPathValue("name", name)
		}
		rec := httptest.NewRecorder()
		handler.APIKeyAuth(http.HandlerFunc(handler.VaultByName)).ServeHTTP(rec, req)
		return rec
	}

	t.Run("found", func(t *testing.T) {
		rec := get("notes")
		if rec.Code != http.StatusOK {
			t.Fatalf("expected status code %d, got %d (body: %s)", http.StatusOK, rec.Code, rec.Body.String())
		}

		var got store.Vault
		if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
			t.Fatalf("decode response: %v (body: %s)", err, rec.Body.String())
		}
		if got.Name != "notes" {
			t.Errorf("Name = %q, want %q", got.Name, "notes")
		}
	})

	t.Run("found case insensitive", func(t *testing.T) {
		for _, name := range []string{"notes", "NOTES", "Notes"} {
			rec := get(name)
			if rec.Code != http.StatusOK {
				t.Errorf("GET %q: expected status code %d, got %d (body: %s)",
					name, http.StatusOK, rec.Code, rec.Body.String())
				continue
			}

			var got store.Vault
			if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
				t.Errorf("GET %q: decode response: %v (body: %s)", name, err, rec.Body.String())
				continue
			}
			if got.Name != "notes" {
				t.Errorf("GET %q: Name = %q, want %q (stored casing)", name, got.Name, "notes")
			}
		}
	})

	t.Run("not found", func(t *testing.T) {
		rec := get("does-not-exist")
		if rec.Code != http.StatusNotFound {
			t.Errorf("expected status code %d, got %d (body: %s)", http.StatusNotFound, rec.Code, rec.Body.String())
		}
	})

	t.Run("missing name", func(t *testing.T) {
		rec := get("")
		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected status code %d, got %d (body: %s)", http.StatusBadRequest, rec.Code, rec.Body.String())
		}
	})

	t.Run("unauthorized without middleware", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/vaults/notes", nil)
		req.SetPathValue("name", "notes")
		rec := httptest.NewRecorder()
		handler.VaultByName(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Errorf("expected status code %d, got %d", http.StatusUnauthorized, rec.Code)
		}
	})
}

func TestDeleteVaultEndpoint(t *testing.T) {
	s := newTestStore(t)
	handler := NewHandler(s, newTestBlobStore(t))

	user, err := s.CreateUser(context.Background(), "a@example.com", "hunter2")
	if err != nil {
		t.Fatalf("seed CreateUser: %v", err)
	}
	if _, err := s.CreateVault(context.Background(), user.ID, "notes"); err != nil {
		t.Fatalf("seed CreateVault: %v", err)
	}

	del := func(name string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodDelete, "/vaults/"+name, nil)
		req.Header.Set("Authorization", "Bearer "+user.APIKey)
		if name != "" {
			req.SetPathValue("name", name)
		}
		rec := httptest.NewRecorder()
		handler.APIKeyAuth(http.HandlerFunc(handler.DeleteVault)).ServeHTTP(rec, req)
		return rec
	}

	t.Run("missing name", func(t *testing.T) {
		rec := del("")
		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected status code %d, got %d (body: %s)", http.StatusBadRequest, rec.Code, rec.Body.String())
		}
	})

	t.Run("deletes an existing vault", func(t *testing.T) {
		rec := del("notes")
		if rec.Code != http.StatusOK {
			t.Fatalf("expected status code %d, got %d (body: %s)", http.StatusOK, rec.Code, rec.Body.String())
		}
		if got, want := rec.Body.String(), "Vault deleted"; got != want {
			t.Errorf("body = %q, want %q", got, want)
		}

		if _, err := s.VaultByName(context.Background(), user.ID, "notes"); !errors.Is(err, store.ErrVaultNotFound) {
			t.Errorf("VaultByName() after delete = %v, want ErrVaultNotFound", err)
		}
	})

	t.Run("deletes matching name case insensitively", func(t *testing.T) {
		if _, err := s.CreateVault(context.Background(), user.ID, "Journal"); err != nil {
			t.Fatalf("seed CreateVault: %v", err)
		}

		rec := del("JOURNAL")
		if rec.Code != http.StatusOK {
			t.Fatalf("expected status code %d, got %d (body: %s)", http.StatusOK, rec.Code, rec.Body.String())
		}
		if got, want := rec.Body.String(), "Vault deleted"; got != want {
			t.Errorf("body = %q, want %q", got, want)
		}

		if _, err := s.VaultByName(context.Background(), user.ID, "journal"); !errors.Is(err, store.ErrVaultNotFound) {
			t.Errorf("VaultByName() after case-insensitive delete = %v, want ErrVaultNotFound", err)
		}
	})

	t.Run("deleting a nonexistent vault is idempotent but reports nothing was deleted", func(t *testing.T) {
		rec := del("does-not-exist")
		if rec.Code != http.StatusOK {
			t.Errorf("expected status code %d, got %d (body: %s)", http.StatusOK, rec.Code, rec.Body.String())
		}
		if got, want := rec.Body.String(), "No vault deleted"; got != want {
			t.Errorf("body = %q, want %q", got, want)
		}
	})

	t.Run("unauthorized without middleware", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/vaults/notes", nil)
		req.SetPathValue("name", "notes")
		rec := httptest.NewRecorder()
		handler.DeleteVault(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Errorf("expected status code %d, got %d", http.StatusUnauthorized, rec.Code)
		}
	})
}

func TestRegisterEndpoint(t *testing.T) {
	tests := []struct {
		name       string
		seed       string
		input      string
		wantStatus int
	}{
		{
			name:       "valid",
			input:      `{"email":"a@example.com","password":"hunter2"}`,
			wantStatus: http.StatusCreated,
		},
		{
			name:       "malformed json",
			input:      `{"email":`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "missing email",
			input:      `{"password":"hunter2"}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "duplicate email",
			seed:       `{"email":"a@example.com","password":"hunter2"}`,
			input:      `{"email":"a@example.com","password":"different"}`,
			wantStatus: http.StatusConflict,
		},
		{
			name:       "duplicate email different case",
			seed:       `{"email":"a@example.com","password":"hunter2"}`,
			input:      `{"email":"A@Example.com","password":"hunter2"}`,
			wantStatus: http.StatusConflict,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			handler := NewHandler(newTestStore(t), newTestBlobStore(t))

			post := func(body string) *httptest.ResponseRecorder {
				req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(body))
				req.Header.Set("Content-Type", "application/json")
				rec := httptest.NewRecorder()
				handler.Register(rec, req)
				return rec
			}

			if test.seed != "" {
				if rec := post(test.seed); rec.Code != http.StatusCreated {
					t.Fatalf("seed request failed: %d (body: %s)", rec.Code, rec.Body.String())
				}
			}

			response := post(test.input)
			if response.Code != test.wantStatus {
				t.Errorf("expected status code %d, got %d (body: %s)",
					test.wantStatus, response.Code, response.Body.String())
			}
		})
	}
}
