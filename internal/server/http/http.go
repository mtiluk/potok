package http

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/michaeltukdev/Potok/internal/server/store"
)

type Store interface {
	CreateVault(ctx context.Context, name string, wrappedKey []byte) (store.Vault, error)
	VaultByName(ctx context.Context, name string) (store.Vault, error)
	ListVaults(ctx context.Context) ([]store.Vault, error)
	DeleteVault(ctx context.Context, name string) error
	CreateUser(ctx context.Context, email, password string) (store.User, error)
	UserByAPIKey(ctx context.Context, apiKey string) (store.User, error)
}

type Handler struct {
	store Store
}

func NewHandler(s Store) *Handler {
	return &Handler{store: s}
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("OK"))
}

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	user, ok := UserFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(struct {
		ID        string    `json:"id"`
		Email     string    `json:"email"`
		IsAdmin   bool      `json:"is_admin"`
		CreatedAt time.Time `json:"created_at"`
	}{
		ID:        user.ID,
		Email:     user.Email,
		IsAdmin:   user.IsAdmin,
		CreatedAt: user.CreatedAt,
	})
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if body.Email == "" || body.Password == "" {
		http.Error(w, "email and password are required", http.StatusBadRequest)
		return
	}

	user, err := h.store.CreateUser(r.Context(), body.Email, body.Password)
	if err != nil {
		if err == store.ErrUserExists {
			http.Error(w, "user already exists", http.StatusConflict)
			return
		}

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write(fmt.Appendf(nil, "User created: %v", user))
}
