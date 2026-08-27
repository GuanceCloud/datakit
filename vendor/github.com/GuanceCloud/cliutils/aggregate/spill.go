// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package aggregate

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"
)

const (
	spillFilePrefix = "spill_"
	spillDirMarker  = ".cliutils-payload-spiller"
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
// Owned spill files are cleared on open because their lifetime never exceeds
// the process lifetime; unrelated directory entries are preserved.
type FilePayloadSpiller struct {
	dir string
	seq atomic.Int64
}

// NewFilePayloadSpiller opens a process-local payload spill directory. It only
// removes files created by a previous spiller instance after verifying the
// directory ownership marker; unrelated files and directories are preserved.
func NewFilePayloadSpiller(dir string) (*FilePayloadSpiller, error) {
	if strings.TrimSpace(dir) == "" {
		return nil, errors.New("spill dir is empty")
	}
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return nil, fmt.Errorf("resolve spill dir: %w", err)
	}
	if absDir == filepath.Clean(string(filepath.Separator)) {
		return nil, errors.New("spill dir must not be the filesystem root")
	}
	if err := os.MkdirAll(absDir, 0o700); err != nil {
		return nil, fmt.Errorf("create spill dir: %w", err)
	}
	resolvedDir, err := filepath.EvalSymlinks(absDir)
	if err != nil {
		return nil, fmt.Errorf("resolve spill dir symlinks: %w", err)
	}
	if resolvedDir == filepath.Clean(string(filepath.Separator)) {
		return nil, errors.New("spill dir must not resolve to the filesystem root")
	}

	markerPath := filepath.Join(resolvedDir, spillDirMarker)
	markerInfo, markerErr := os.Lstat(markerPath)
	switch {
	case markerErr == nil && !markerInfo.Mode().IsRegular():
		return nil, errors.New("spill dir ownership marker is not a regular file")
	case markerErr != nil && !os.IsNotExist(markerErr):
		return nil, fmt.Errorf("inspect spill dir ownership marker: %w", markerErr)
	case os.IsNotExist(markerErr):
		marker, createErr := os.OpenFile(markerPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
		if createErr != nil {
			return nil, fmt.Errorf("create spill dir ownership marker: %w", createErr)
		}
		if closeErr := marker.Close(); closeErr != nil {
			return nil, fmt.Errorf("close spill dir ownership marker: %w", closeErr)
		}
		return &FilePayloadSpiller{dir: resolvedDir}, nil
	}

	entries, err := os.ReadDir(resolvedDir)
	if err != nil {
		return nil, fmt.Errorf("read spill dir: %w", err)
	}
	for _, entry := range entries {
		if entry.IsDir() || !validSpillKey(entry.Name()) {
			continue
		}
		if err := os.Remove(filepath.Join(resolvedDir, entry.Name())); err != nil && !os.IsNotExist(err) {
			return nil, fmt.Errorf("remove stale spill file %q: %w", entry.Name(), err)
		}
	}

	return &FilePayloadSpiller{dir: resolvedDir}, nil
}

// Put stores one payload and returns its opaque spill key.
func (f *FilePayloadSpiller) Put(payload []byte) (string, error) {
	if f == nil {
		return "", errors.New("file payload spiller is nil")
	}

	key := fmt.Sprintf("%s%x_%d", spillFilePrefix, time.Now().UnixNano(), f.seq.Add(1))
	path := filepath.Join(f.dir, key)
	if err := os.WriteFile(path, payload, 0o600); err != nil {
		return "", fmt.Errorf("write spill payload: %w", err)
	}

	return key, nil
}

// Get loads a payload previously stored by Put.
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

// Delete removes a payload key. Deleting a missing key is successful.
func (f *FilePayloadSpiller) Delete(key string) error {
	if f == nil || !validSpillKey(key) {
		return nil
	}

	if err := os.Remove(filepath.Join(f.dir, key)); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("delete spill payload: %w", err)
	}

	return nil
}

// Close releases spiller resources. FilePayloadSpiller currently keeps no open handles.
func (f *FilePayloadSpiller) Close() error {
	return nil
}

// validSpillKey restricts the key character set to prevent path traversal.
func validSpillKey(key string) bool {
	if !strings.HasPrefix(key, spillFilePrefix) || len(key) == len(spillFilePrefix) || len(key) > 64 {
		return false
	}

	for i := len(spillFilePrefix); i < len(key); i++ {
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
