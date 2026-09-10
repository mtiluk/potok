package blobstore

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func newTestStore(t *testing.T) *Store {
	t.Helper()
	return New(t.TempDir())
}

type erroringReader struct {
	data []byte
	err  error
}

func (r *erroringReader) Read(p []byte) (int, error) {
	if len(r.data) > 0 {
		n := copy(p, r.data)
		r.data = r.data[n:]
		return n, nil
	}
	return 0, r.err
}

func TestPutAndGet(t *testing.T) {
	s := newTestStore(t)
	content := []byte("hello, potok")

	if err := s.Put("vault-a", "blob-1", bytes.NewReader(content)); err != nil {
		t.Fatalf("Put() error: %v", err)
	}

	rc, err := s.Get("vault-a", "blob-1")
	if err != nil {
		t.Fatalf("Get() error: %v", err)
	}
	defer rc.Close()

	got, err := io.ReadAll(rc)
	if err != nil {
		t.Fatalf("ReadAll() error: %v", err)
	}
	if !bytes.Equal(got, content) {
		t.Errorf("Get() content = %q, want %q", got, content)
	}
}

func TestGetNotFound(t *testing.T) {
	s := newTestStore(t)

	_, err := s.Get("vault-a", "does-not-exist")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("Get() error = %v, want ErrNotFound", err)
	}
}

func TestPutVaultIsolation(t *testing.T) {
	s := newTestStore(t)

	if err := s.Put("vault-a", "blob-1", bytes.NewReader([]byte("from a"))); err != nil {
		t.Fatalf("Put(vault-a) error: %v", err)
	}
	if err := s.Put("vault-b", "blob-1", bytes.NewReader([]byte("from b"))); err != nil {
		t.Fatalf("Put(vault-b) error: %v", err)
	}

	rc, err := s.Get("vault-a", "blob-1")
	if err != nil {
		t.Fatalf("Get(vault-a) error: %v", err)
	}
	gotA, _ := io.ReadAll(rc)
	rc.Close()

	rc, err = s.Get("vault-b", "blob-1")
	if err != nil {
		t.Fatalf("Get(vault-b) error: %v", err)
	}
	gotB, _ := io.ReadAll(rc)
	rc.Close()

	if string(gotA) != "from a" {
		t.Errorf("vault-a blob-1 = %q, want %q", gotA, "from a")
	}
	if string(gotB) != "from b" {
		t.Errorf("vault-b blob-1 = %q, want %q", gotB, "from b")
	}
}

func TestPutOverwriteSameID(t *testing.T) {
	s := newTestStore(t)
	content := []byte("same content")

	if err := s.Put("vault-a", "blob-1", bytes.NewReader(content)); err != nil {
		t.Fatalf("first Put() error: %v", err)
	}
	if err := s.Put("vault-a", "blob-1", bytes.NewReader(content)); err != nil {
		t.Fatalf("second Put() error: %v", err)
	}

	rc, err := s.Get("vault-a", "blob-1")
	if err != nil {
		t.Fatalf("Get() error: %v", err)
	}
	defer rc.Close()

	got, _ := io.ReadAll(rc)
	if !bytes.Equal(got, content) {
		t.Errorf("Get() content = %q, want %q", got, content)
	}
}

func TestPutAtomicOnReadFailure(t *testing.T) {
	dir := t.TempDir()
	s := New(dir)

	readErr := errors.New("boom: connection reset")
	r := &erroringReader{data: []byte("partial"), err: readErr}

	err := s.Put("vault-a", "blob-1", r)
	if err == nil {
		t.Fatal("Put() error = nil, want non-nil")
	}
	if !errors.Is(err, readErr) {
		t.Errorf("Put() error = %v, want it to wrap %v", err, readErr)
	}

	if _, err := s.Get("vault-a", "blob-1"); !errors.Is(err, ErrNotFound) {
		t.Errorf("Get() after failed Put() error = %v, want ErrNotFound", err)
	}

	blobsDir := filepath.Join(dir, "vaults", "vault-a", "blobs")
	entries, err := os.ReadDir(blobsDir)
	if err != nil && !os.IsNotExist(err) {
		t.Fatalf("ReadDir(%s) error: %v", blobsDir, err)
	}
	if len(entries) != 0 {
		t.Errorf("blobs dir has %d leftover entries after failed Put(), want 0: %v", len(entries), entries)
	}
}

func TestPutWritesUnderVaultAndBlobID(t *testing.T) {
	dir := t.TempDir()
	s := New(dir)

	if err := s.Put("vault-a", "blob-1", bytes.NewReader([]byte("x"))); err != nil {
		t.Fatalf("Put() error: %v", err)
	}

	want := filepath.Join(dir, "vaults", "vault-a", "blobs", "blob-1")
	if _, err := os.Stat(want); err != nil {
		t.Errorf("expected blob at %s, stat error: %v", want, err)
	}
}
