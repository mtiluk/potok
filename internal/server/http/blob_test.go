package http

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mtiluk/potok/internal/server/blobstore"
	"github.com/mtiluk/potok/internal/server/store"
)

func newBlobTestHandler(t *testing.T, s *store.Store) *Handler {
	t.Helper()
	return &Handler{
		store:     s,
		blobStore: *blobstore.New(t.TempDir()),
	}
}

func hashHex(content []byte) string {
	sum := sha256.Sum256(content)
	return hex.EncodeToString(sum[:])
}

func TestPutBlobEndpoint(t *testing.T) {
	s := newTestStore(t)
	handler := newBlobTestHandler(t, s)

	user, err := s.CreateUser(context.Background(), "a@example.com", "hunter2")
	if err != nil {
		t.Fatalf("seed CreateUser: %v", err)
	}
	if _, err := s.CreateVault(context.Background(), user.ID, "notes"); err != nil {
		t.Fatalf("seed CreateVault: %v", err)
	}
	otherUser, err := s.CreateUser(context.Background(), "b@example.com", "hunter2")
	if err != nil {
		t.Fatalf("seed CreateUser(other): %v", err)
	}

	put := func(apiKey, vaultName, blobID string, body []byte) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodPut, "/vaults/"+vaultName+"/blobs/"+blobID, bytes.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+apiKey)
		if vaultName != "" {
			req.SetPathValue("name", vaultName)
		}
		if blobID != "" {
			req.SetPathValue("id", blobID)
		}
		rec := httptest.NewRecorder()
		handler.APIKeyAuth(http.HandlerFunc(handler.PutBlob)).ServeHTTP(rec, req)
		return rec
	}

	content := []byte("hello, potok")
	id := hashHex(content)

	t.Run("stores a new blob and records it", func(t *testing.T) {
		rec := put(user.APIKey, "notes", id, content)
		if rec.Code != http.StatusCreated {
			t.Fatalf("expected status code %d, got %d (body: %s)", http.StatusCreated, rec.Code, rec.Body.String())
		}

		vault, err := s.VaultByName(context.Background(), user.ID, "notes")
		if err != nil {
			t.Fatalf("VaultByName: %v", err)
		}
		has, err := s.HasBlob(context.Background(), vault.ID, id)
		if err != nil {
			t.Fatalf("HasBlob: %v", err)
		}
		if !has {
			t.Error("HasBlob() = false after a successful Put, want true")
		}

		rc, err := handler.blobStore.Get(vault.ID, id)
		if err != nil {
			t.Fatalf("blobStore.Get: %v", err)
		}
		defer rc.Close()
	})

	t.Run("short-circuits an already-stored blob", func(t *testing.T) {
		rec := put(user.APIKey, "notes", id, content)
		if rec.Code != http.StatusOK {
			t.Errorf("expected status code %d, got %d (body: %s)", http.StatusOK, rec.Code, rec.Body.String())
		}
		if got, want := rec.Body.String(), "Blob already exists\n"; got != want {
			t.Errorf("body = %q, want %q", got, want)
		}
	})

	t.Run("rejects content that doesn't match the claimed id", func(t *testing.T) {
		rec := put(user.APIKey, "notes", id, []byte("tampered content"))
		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected status code %d, got %d (body: %s)", http.StatusBadRequest, rec.Code, rec.Body.String())
		}
	})

	t.Run("missing vault name", func(t *testing.T) {
		rec := put(user.APIKey, "", id, content)
		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected status code %d, got %d (body: %s)", http.StatusBadRequest, rec.Code, rec.Body.String())
		}
	})

	t.Run("missing blob id", func(t *testing.T) {
		rec := put(user.APIKey, "notes", "", content)
		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected status code %d, got %d (body: %s)", http.StatusBadRequest, rec.Code, rec.Body.String())
		}
	})

	t.Run("vault not found", func(t *testing.T) {
		rec := put(user.APIKey, "does-not-exist", id, content)
		if rec.Code != http.StatusNotFound {
			t.Errorf("expected status code %d, got %d (body: %s)", http.StatusNotFound, rec.Code, rec.Body.String())
		}
	})

	t.Run("cannot write to another user's vault", func(t *testing.T) {
		rec := put(otherUser.APIKey, "notes", id, content)
		if rec.Code != http.StatusNotFound {
			t.Errorf("expected status code %d, got %d (body: %s)", http.StatusNotFound, rec.Code, rec.Body.String())
		}
	})

	t.Run("unauthorized without middleware", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPut, "/vaults/notes/blobs/"+id, bytes.NewReader(content))
		req.SetPathValue("name", "notes")
		req.SetPathValue("id", id)
		rec := httptest.NewRecorder()
		handler.PutBlob(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Errorf("expected status code %d, got %d", http.StatusUnauthorized, rec.Code)
		}
	})
}

func TestGetBlobEndpoint(t *testing.T) {
	s := newTestStore(t)
	handler := newBlobTestHandler(t, s)

	user, err := s.CreateUser(context.Background(), "a@example.com", "hunter2")
	if err != nil {
		t.Fatalf("seed CreateUser: %v", err)
	}
	if _, err := s.CreateVault(context.Background(), user.ID, "notes"); err != nil {
		t.Fatalf("seed CreateVault: %v", err)
	}
	otherUser, err := s.CreateUser(context.Background(), "b@example.com", "hunter2")
	if err != nil {
		t.Fatalf("seed CreateUser(other): %v", err)
	}

	content := []byte("hello again, potok")
	id := hashHex(content)

	seedReq := httptest.NewRequest(http.MethodPut, "/vaults/notes/blobs/"+id, bytes.NewReader(content))
	seedReq.Header.Set("Authorization", "Bearer "+user.APIKey)
	seedReq.SetPathValue("name", "notes")
	seedReq.SetPathValue("id", id)
	seedRec := httptest.NewRecorder()
	handler.APIKeyAuth(http.HandlerFunc(handler.PutBlob)).ServeHTTP(seedRec, seedReq)
	if seedRec.Code != http.StatusCreated {
		t.Fatalf("seed Put failed: %d (body: %s)", seedRec.Code, seedRec.Body.String())
	}

	get := func(apiKey, vaultName, blobID string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodGet, "/vaults/"+vaultName+"/blobs/"+blobID, nil)
		req.Header.Set("Authorization", "Bearer "+apiKey)
		if vaultName != "" {
			req.SetPathValue("name", vaultName)
		}
		if blobID != "" {
			req.SetPathValue("id", blobID)
		}
		rec := httptest.NewRecorder()
		handler.APIKeyAuth(http.HandlerFunc(handler.GetBlob)).ServeHTTP(rec, req)
		return rec
	}

	t.Run("returns the stored bytes", func(t *testing.T) {
		rec := get(user.APIKey, "notes", id)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected status code %d, got %d (body: %s)", http.StatusOK, rec.Code, rec.Body.String())
		}
		if !bytes.Equal(rec.Body.Bytes(), content) {
			t.Errorf("body = %q, want %q", rec.Body.Bytes(), content)
		}
	})

	t.Run("blob not found", func(t *testing.T) {
		rec := get(user.APIKey, "notes", hashHex([]byte("never uploaded")))
		if rec.Code != http.StatusNotFound {
			t.Errorf("expected status code %d, got %d (body: %s)", http.StatusNotFound, rec.Code, rec.Body.String())
		}
	})

	t.Run("vault not found", func(t *testing.T) {
		rec := get(user.APIKey, "does-not-exist", id)
		if rec.Code != http.StatusNotFound {
			t.Errorf("expected status code %d, got %d (body: %s)", http.StatusNotFound, rec.Code, rec.Body.String())
		}
	})

	t.Run("missing vault name", func(t *testing.T) {
		rec := get(user.APIKey, "", id)
		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected status code %d, got %d (body: %s)", http.StatusBadRequest, rec.Code, rec.Body.String())
		}
	})

	t.Run("missing blob id", func(t *testing.T) {
		rec := get(user.APIKey, "notes", "")
		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected status code %d, got %d (body: %s)", http.StatusBadRequest, rec.Code, rec.Body.String())
		}
	})

	t.Run("cannot read another user's vault blobs", func(t *testing.T) {
		rec := get(otherUser.APIKey, "notes", id)
		if rec.Code != http.StatusNotFound {
			t.Errorf("expected status code %d, got %d (body: %s)", http.StatusNotFound, rec.Code, rec.Body.String())
		}
	})

	t.Run("unauthorized without middleware", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/vaults/notes/blobs/"+id, nil)
		req.SetPathValue("name", "notes")
		req.SetPathValue("id", id)
		rec := httptest.NewRecorder()
		handler.GetBlob(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Errorf("expected status code %d, got %d", http.StatusUnauthorized, rec.Code)
		}
	})
}
