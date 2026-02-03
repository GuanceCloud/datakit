// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build !windows
// +build !windows

package hostchange

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	xxhash "github.com/cespare/xxhash/v2"
)

// GetHashCode calculates the xxhash hash value of the given data.
func GetHashCode(data []byte) uint64 {
	return xxhash.Sum64(data)
}

// GetFileHash calculate file xxhash hash value.
func GetFileHash(filePath string) (uint64, error) {
	file, err := os.Open(filePath) // nolint:gosec
	if err != nil {
		return 0, fmt.Errorf("failed to open file: %w", err)
	}

	defer file.Close() //nolint:errcheck,gosec

	hash := xxhash.New()
	buf := make([]byte, 1024*1024) // 1MB buffer
	for {
		n, err := file.Read(buf)
		if err != nil && err != io.EOF {
			return 0, fmt.Errorf("failed to read file: %w", err)
		}
		if n == 0 {
			break
		}
		if _, err := hash.Write(buf[:n]); err != nil {
			return 0, fmt.Errorf("failed to write to hash: %w", err)
		}
	}

	return hash.Sum64(), nil
}

// GetFileModTime returns the modification time of the specified file.
func GetFileModTime(path string) (time.Time, error) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to stat file: %w", err)
	}

	return fileInfo.ModTime(), nil
}

// MaxTime return the later of two times.
func MaxTime(a, b time.Time) time.Time {
	if a.After(b) {
		return a
	}
	return b
}

// getString safely retrieves a string value from a map, trying multiple keys if available.
func getString(m map[string]interface{}, keys ...string) string {
	for _, k := range keys {
		if v, ok := m[k].(string); ok {
			return v
		}
	}
	return ""
}

// isContainString checks if the string contains any of the provided substrings.
func isContainString(s string, substrs ...string) bool {
	for _, substr := range substrs {
		if strings.Contains(s, substr) {
			return true
		}
	}
	return false
}

// ReadFile reads the entire content of a file into a byte slice.
func ReadFile(path string) ([]byte, error) {
	return os.ReadFile(path) // nolint:gosec
}
