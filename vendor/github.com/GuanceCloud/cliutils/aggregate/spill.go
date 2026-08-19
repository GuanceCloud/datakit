package aggregate

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync/atomic"
	"time"
)

// PayloadSpiller stores oversized payloads on disk.
// It backs the time-wheel tiered cache: large packets (a few packets holding
// most of the memory footprint) spill to disk while the time wheel keeps metadata only.
type PayloadSpiller interface {
	// Put writes a payload and returns a key usable with Get/Delete later.
	Put(payload []byte) (key string, err error)
	// Get reads back a payload by key.
	Get(key string) ([]byte, error)
	// Delete removes the data for a key (idempotent).
	Delete(key string) error
	// Close closes the storage.
	Close() error
}

// FilePayloadSpiller is a directory-plus-file implementation of PayloadSpiller.
// The directory is cleared on open to remove spill data left by a previous run
// (spill lifetime never exceeds the process lifetime).
type FilePayloadSpiller struct {
	dir string
	seq atomic.Int64
}

func NewFilePayloadSpiller(dir string) (*FilePayloadSpiller, error) {
	if err := os.RemoveAll(dir); err != nil {
		return nil, fmt.Errorf("clear spill dir: %w", err)
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, fmt.Errorf("create spill dir: %w", err)
	}

	return &FilePayloadSpiller{dir: dir}, nil
}

func (f *FilePayloadSpiller) Put(payload []byte) (string, error) {
	if f == nil {
		return "", errors.New("file payload spiller is nil")
	}

	key := fmt.Sprintf("%x_%d", time.Now().UnixNano(), f.seq.Add(1))
	path := filepath.Join(f.dir, key)
	if err := os.WriteFile(path, payload, 0o600); err != nil {
		return "", fmt.Errorf("write spill payload: %w", err)
	}

	return key, nil
}

func (f *FilePayloadSpiller) Get(key string) ([]byte, error) {
	if f == nil {
		return nil, errors.New("file payload spiller is nil")
	}
	if !validSpillKey(key) {
		return nil, fmt.Errorf("invalid spill key %q", key)
	}

	payload, err := os.ReadFile(filepath.Join(f.dir, key))
	if err != nil {
		return nil, fmt.Errorf("read spill payload: %w", err)
	}

	return payload, nil
}

func (f *FilePayloadSpiller) Delete(key string) error {
	if f == nil || !validSpillKey(key) {
		return nil
	}

	if err := os.Remove(filepath.Join(f.dir, key)); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("delete spill payload: %w", err)
	}

	return nil
}

func (f *FilePayloadSpiller) Close() error {
	return nil
}

// validSpillKey restricts the key character set to prevent path traversal.
func validSpillKey(key string) bool {
	if key == "" || len(key) > 64 {
		return false
	}

	for i := 0; i < len(key); i++ {
		c := key[i]
		switch {
		case c >= '0' && c <= '9':
		case c >= 'a' && c <= 'f':
		case c == '_':
		default:
			return false
		}
	}

	return true
}
