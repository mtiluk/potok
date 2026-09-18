package store

import (
	"context"
	"crypto/rand"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mtiluk/potok/internal/server/auth"
	"golang.org/x/crypto/bcrypt"
)

const kdfSaltLen = 16

var ErrUserExists = fmt.Errorf("user already exists")
var ErrUserNotFound = fmt.Errorf("user not found")
var ErrVaultNotFound = fmt.Errorf("vault not found")
var ErrVaultExists = fmt.Errorf("vault already exists")
var ErrManifestNotFound = fmt.Errorf("manifest not found")
var ErrManifestConflict = fmt.Errorf("manifest generation conflict")

type Store struct {
	db *sql.DB
}

type Vault struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	KDFSalt   []byte    `json:"kdf_salt"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"password_hash"`
	IsAdmin      bool      `json:"is_admin"`
	APIKey       string    `json:"api_key"`
	CreatedAt    time.Time `json:"created_at"`
}

type Blob struct {
	VaultID   string    `json:"vault_id"`
	ID        string    `json:"id"`
	SizeBytes int64     `json:"size_bytes"`
	CreatedAt time.Time `json:"created_at"`
}

type Manifest struct {
	VaultID    string    `json:"vault_id"`
	Generation int64     `json:"generation"`
	Ciphertext []byte    `json:"-"`
	UpdatedAt  time.Time `json:"updated_at"`
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

	if _, err := db.ExecContext(ctx, "PRAGMA foreign_keys = ON; PRAGMA journal_mode = WAL; PRAGMA busy_timeout = 5000;"); err != nil {
		db.Close()
		return nil, fmt.Errorf("store: set pragmas: %w", err)
	}

	return &Store{db: db}, nil
}

func (s *Store) Close() error { return s.db.Close() }

func (s *Store) CreateVault(ctx context.Context, userID, name string) (Vault, error) {
	salt := make([]byte, kdfSaltLen)
	if _, err := rand.Read(salt); err != nil {
		return Vault{}, fmt.Errorf("store: create vault: generate kdf salt: %w", err)
	}

	vault := Vault{
		ID:        uuid.NewString(),
		Name:      name,
		KDFSalt:   salt,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	var existing string
	err := s.db.QueryRowContext(ctx, `
		SELECT id FROM vaults WHERE user_id = $1 AND name = $2 COLLATE NOCASE`,
		userID, name,
	).Scan(&existing)
	if err == nil {
		return Vault{}, ErrVaultExists
	}
	if err != sql.ErrNoRows {
		return Vault{}, fmt.Errorf("store: create vault: %w", err)
	}

	err = s.db.QueryRowContext(ctx, `
		INSERT INTO vaults (id, user_id, name, kdf_salt, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (user_id, name) DO NOTHING
		RETURNING id`,
		vault.ID, userID, name, vault.KDFSalt, vault.CreatedAt, vault.UpdatedAt,
	).Scan(&vault.ID)

	if err != nil {
		if err == sql.ErrNoRows {
			return Vault{}, ErrVaultExists
		}

		return Vault{}, fmt.Errorf("store: create vault: %w", err)
	}
	return vault, nil
}

func (s *Store) VaultByName(ctx context.Context, userID, name string) (Vault, error) {
	var vault Vault

	err := s.db.QueryRowContext(ctx, "SELECT id, name, kdf_salt, created_at, updated_at FROM vaults WHERE user_id = $1 AND name = $2 COLLATE NOCASE", userID, name).Scan(&vault.ID, &vault.Name, &vault.KDFSalt, &vault.CreatedAt, &vault.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return Vault{}, ErrVaultNotFound
		}

		return Vault{}, fmt.Errorf("store: vault by name: %w", err)
	}

	return vault, nil
}

func (s *Store) ListVaults(ctx context.Context, userID string) ([]Vault, error) {
	rows, err := s.db.QueryContext(ctx, "SELECT id, name, kdf_salt, created_at, updated_at FROM vaults WHERE user_id = $1", userID)
	if err != nil {
		return []Vault{}, fmt.Errorf("store: list vaults: %w", err)
	}
	defer rows.Close()

	var vaults []Vault

	for rows.Next() {
		var vault Vault
		if err := rows.Scan(&vault.ID, &vault.Name, &vault.KDFSalt, &vault.CreatedAt, &vault.UpdatedAt); err != nil {
			return []Vault{}, fmt.Errorf("store: list vaults: %w", err)
		}
		vaults = append(vaults, vault)
	}

	return vaults, nil
}

func (s *Store) DeleteVault(ctx context.Context, userID, name string) (bool, error) {
	res, err := s.db.ExecContext(ctx, "DELETE FROM vaults WHERE user_id = $1 AND name = $2 COLLATE NOCASE", userID, name)
	if err != nil {
		return false, fmt.Errorf("store: delete vault: %w", err)
	}

	n, err := res.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("store: delete vault: %w", err)
	}

	return n > 0, nil
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

func (s *Store) GetBlob(ctx context.Context, userID, vaultName, blobName string) (Blob, error) {
	var blob Blob
	err := s.db.QueryRowContext(ctx, `
		SELECT vault_id, id, size_bytes, created_at
		FROM blobs
		WHERE vault_id = ? AND id = ?`,
		vaultName, blobName,
	).Scan(&blob.VaultID, &blob.ID, &blob.SizeBytes, &blob.CreatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return Blob{}, nil
		}

		return Blob{}, fmt.Errorf("store: get blob: %w", err)
	}
	return blob, nil
}

func (s *Store) HasBlob(ctx context.Context, vaultID, blobID string) (bool, error) {
	var exists bool
	err := s.db.QueryRowContext(ctx, `
		SELECT EXISTS(
			SELECT 1
			FROM blobs
			WHERE vault_id = ? AND id = ?
		)`,
		vaultID, blobID,
	).Scan(&exists)

	if err != nil {
		return false, fmt.Errorf("store: has blob: %w", err)
	}
	return exists, nil
}

func (s *Store) PutBlob(ctx context.Context, vaultID, blobID string, sizeBytes int64) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO blobs (vault_id, id, size_bytes)
		VALUES (?, ?, ?)
	`, vaultID, blobID, sizeBytes)

	if err != nil {
		return fmt.Errorf("store: put blob: %w", err)
	}
	return nil
}

func (s *Store) GetManifest(ctx context.Context, vaultID string) (Manifest, error) {
	var m Manifest
	err := s.db.QueryRowContext(ctx, `
		SELECT vault_id, generation, ciphertext, updated_at
		FROM manifests
		WHERE vault_id = ?`,
		vaultID,
	).Scan(&m.VaultID, &m.Generation, &m.Ciphertext, &m.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return Manifest{}, ErrManifestNotFound
		}

		return Manifest{}, fmt.Errorf("store: get manifest: %w", err)
	}
	return m, nil
}

func (s *Store) PutManifest(ctx context.Context, vaultID string, expectedGeneration int64, ciphertext []byte) (Manifest, error) {
	newGeneration := expectedGeneration + 1
	var m Manifest
	var err error

	if expectedGeneration == 0 {
		err = s.db.QueryRowContext(ctx, `
			INSERT INTO manifests (vault_id, generation, ciphertext)
			VALUES (?, ?, ?)
			ON CONFLICT (vault_id) DO NOTHING
			RETURNING vault_id, generation, ciphertext, updated_at`,
			vaultID, newGeneration, ciphertext,
		).Scan(&m.VaultID, &m.Generation, &m.Ciphertext, &m.UpdatedAt)
	} else {
		err = s.db.QueryRowContext(ctx, `
			UPDATE manifests
			SET generation = ?, ciphertext = ?, updated_at = CURRENT_TIMESTAMP
			WHERE vault_id = ? AND generation = ?
			RETURNING vault_id, generation, ciphertext, updated_at`,
			newGeneration, ciphertext, vaultID, expectedGeneration,
		).Scan(&m.VaultID, &m.Generation, &m.Ciphertext, &m.UpdatedAt)
	}

	if err != nil {
		if err == sql.ErrNoRows {
			return Manifest{}, ErrManifestConflict
		}

		return Manifest{}, fmt.Errorf("store: put manifest: %w", err)
	}
	return m, nil
}
