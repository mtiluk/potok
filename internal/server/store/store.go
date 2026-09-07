package store

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/michaeltukdev/Potok/internal/server/auth"
	"golang.org/x/crypto/bcrypt"
)

var ErrUserExists = fmt.Errorf("user already exists")
var ErrUserNotFound = fmt.Errorf("user not found")

type Store struct {
	db *sql.DB
}

type Vault struct {
	ID         int64     `json:"id"`
	Name       string    `json:"name"`
	WrappedKey []byte    `json:"-"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"password_hash"`
	IsAdmin      bool      `json:"is_admin"`
	APIKey       string    `json:"api_key"`
	CreatedAt    time.Time `json:"created_at"`
}

func Open(ctx context.Context, dsn string) (*Store, error) {
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("store: open: %w", err)
	}

	db.SetMaxOpenConns(1)

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("store: ping: %w", err)
	}

	for _, pragma := range []string{
		"PRAGMA foreign_keys = ON",
		"PRAGMA journal_mode = WAL",
		"PRAGMA busy_timeout = 5000",
	} {
		if _, err := db.ExecContext(ctx, pragma); err != nil {
			db.Close()
			return nil, fmt.Errorf("store: %s: %w", pragma, err)
		}
	}

	return &Store{db: db}, nil
}

// Extra
func (s *Store) Close() error { return s.db.Close() }

func (s *Store) CreateVault(ctx context.Context, name string, wrappedKey []byte) (Vault, error) {
	return Vault{}, nil
}

func (s *Store) VaultByName(ctx context.Context, name string) (Vault, error) {
	return Vault{}, nil
}

func (s *Store) ListVaults(ctx context.Context) ([]Vault, error) {
	return []Vault{}, nil
}

func (s *Store) DeleteVault(ctx context.Context, name string) error {
	return nil
}

func (s *Store) CreateUser(ctx context.Context, email, password string) (User, error) {
	email = strings.ToLower(email)

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return User{}, fmt.Errorf("store: create user: %w", err)
	}

	key, err := auth.GenerateAPIKey()
	if err != nil {
		return User{}, fmt.Errorf("store: create user: %w", err)
	}

	user := User{
		ID:           uuid.NewString(),
		Email:        email,
		PasswordHash: string(hash),
		APIKey:       key,
	}

	err = s.db.QueryRowContext(ctx, `
		INSERT INTO users (id, email, password_hash, api_key)
		VALUES (?, ?, ?, ?)
		ON CONFLICT (email) DO NOTHING
		RETURNING id, email, is_admin, created_at`,
		user.ID, user.Email, user.PasswordHash, key,
	).Scan(&user.ID, &user.Email, &user.IsAdmin, &user.CreatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return User{}, ErrUserExists
		}

		return User{}, fmt.Errorf("store: create user: %w", err)
	}
	return user, nil
}

func (s *Store) UserByAPIKey(ctx context.Context, apiKey string) (User, error) {
	if apiKey == "" {
		return User{}, ErrUserNotFound
	}

	var user User
	err := s.db.QueryRowContext(ctx, `
		SELECT id, email, password_hash, is_admin, api_key, created_at
		FROM users
		WHERE api_key = ?`,
		apiKey,
	).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.IsAdmin, &user.APIKey, &user.CreatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return User{}, ErrUserNotFound
		}

		return User{}, fmt.Errorf("store: user by api key: %w", err)
	}
	return user, nil
}
