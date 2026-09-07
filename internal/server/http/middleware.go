package http

import (
	"context"
	"net/http"
	"strings"

	"github.com/michaeltukdev/Potok/internal/server/store"
)

type contextKey int

const userContextKey contextKey = iota

func WithUser(ctx context.Context, u store.User) context.Context {
	return context.WithValue(ctx, userContextKey, u)
}

func UserFromContext(ctx context.Context) (store.User, bool) {
	u, ok := ctx.Value(userContextKey).(store.User)
	return u, ok
}

func (h *Handler) APIKeyAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
		if !ok || key == "" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		user, err := h.store.UserByAPIKey(r.Context(), key)
		if err != nil {
			if err == store.ErrUserNotFound {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		next.ServeHTTP(w, r.WithContext(WithUser(r.Context(), user)))
	})
}
