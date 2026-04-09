// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package openfile

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDidRotateTruncated(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "app.log")
	require.NoError(t, os.WriteFile(path, []byte("hello"), 0o640))

	f, err := os.Open(path)
	require.NoError(t, err)
	defer f.Close() //nolint:errcheck

	rotated, err := DidRotate(f, 10)
	require.NoError(t, err)
	require.True(t, rotated)
}

func TestDidRotateRecreated(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "app.log")
	require.NoError(t, os.WriteFile(path, []byte("line-1"), 0o640))

	f, err := os.Open(path)
	require.NoError(t, err)
	defer f.Close() //nolint:errcheck

	oldPath := filepath.Join(dir, "app.log.1")
	require.NoError(t, os.Rename(path, oldPath))
	require.NoError(t, os.WriteFile(path, []byte("line-2"), 0o640))

	rotated, err := DidRotate(f, 0)
	require.NoError(t, err)
	require.True(t, rotated)
}

func TestFileExistsStatErrorNoPanic(t *testing.T) {
	longName := strings.Repeat("a", 8192)
	path := filepath.Join(t.TempDir(), fmt.Sprintf("%s.log", longName))

	require.False(t, FileExists(path))
}
