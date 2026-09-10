package http

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"net/http"

	"github.com/michaeltukdev/Potok/internal/server/blobstore"
	"github.com/michaeltukdev/Potok/internal/server/store"
)

func (h *Handler) PutBlob(w http.ResponseWriter, r *http.Request) {
	user, ok := UserFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	vaultName := r.PathValue("name")
	blobID := r.PathValue("id")

	if vaultName == "" || blobID == "" {
		http.Error(w, "vault name and blob id are required", http.StatusBadRequest)
		return
	}

	vault, err := h.store.VaultByName(r.Context(), user.ID, vaultName)
	if err != nil {
		if err == store.ErrVaultNotFound {
			http.Error(w, "vault not found", http.StatusNotFound)
			return
		}

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	sum := sha256.Sum256(body)
	if hex.EncodeToString(sum[:]) != blobID {
		http.Error(w, "blob id does not match content hash", http.StatusBadRequest)
		return
	}

	exists, err := h.store.HasBlob(r.Context(), vault.ID, blobID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if exists {
		http.Error(w, "Blob already exists", http.StatusOK)
		return
	}

	if err := h.blobStore.Put(vault.ID, blobID, bytes.NewReader(body)); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := h.store.PutBlob(r.Context(), vault.ID, blobID, int64(len(body))); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (h *Handler) GetBlob(w http.ResponseWriter, r *http.Request) {
	user, ok := UserFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	vaultName := r.PathValue("name")
	blobID := r.PathValue("id")

	if vaultName == "" || blobID == "" {
		http.Error(w, "vault name and blob id are required", http.StatusBadRequest)
		return
	}

	vault, err := h.store.VaultByName(r.Context(), user.ID, vaultName)
	if err != nil {
		if err == store.ErrVaultNotFound {
			http.Error(w, "vault not found", http.StatusNotFound)
			return
		}

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	blob, err := h.blobStore.Get(vault.ID, blobID)
	if err != nil {
		if errors.Is(err, blobstore.ErrNotFound) {
			http.Error(w, "blob not found", http.StatusNotFound)
			return
		}

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer blob.Close()

	w.Header().Set("Content-Type", "application/octet-stream")
	if _, err := io.Copy(w, blob); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
