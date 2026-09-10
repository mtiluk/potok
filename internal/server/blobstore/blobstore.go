package blobstore

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
)

var ErrNotFound = errors.New("not found")

type Store struct {
	dataDir string
}

func New(dataDir string) *Store {
	return &Store{dataDir: dataDir}
}

func validID(s string) bool {
	if s == "" || s == "." || s == ".." {
		return false
	}

	if strings.ContainsAny(s, `/\`) || strings.ContainsRune(s, 0) {
		return false
	}
	return true
}

func (s *Store) Put(vaultID, blobID string, r io.Reader) error {
	if !validID(vaultID) || !validID(blobID) {
		return errors.New("invalid ID")
	}

	dir := filepath.Join(s.dataDir, "vaults", vaultID, "blobs")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	tmp, err := os.CreateTemp(dir, blobID+".tmp-*")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()

	if _, err := io.Copy(tmp, r); err != nil {
		tmp.Close()
		os.Remove(tmpPath)
		return err
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpPath)
		return err
	}

	path := filepath.Join(dir, blobID)
	if err := os.Rename(tmpPath, path); err != nil {
		os.Remove(tmpPath)
		return err
	}

	return nil
}

func (s *Store) Get(vaultID, blobID string) (io.ReadCloser, error) {
	if !validID(vaultID) || !validID(blobID) {
		return nil, errors.New("invalid ID")
	}

	path := filepath.Join(s.dataDir, "vaults", vaultID, "blobs", blobID)
	blob, err := os.Open(path)
	if err != nil {
		return nil, ErrNotFound
	}

	return blob, nil
}
