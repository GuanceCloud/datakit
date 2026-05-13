// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package downloader

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func makeTarGz(t *testing.T, name string, mode int64, content []byte) *bytes.Buffer {
	t.Helper()

	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gw)

	require.NoError(t, tw.WriteHeader(&tar.Header{
		Name: name,
		Mode: mode,
		Size: int64(len(content)),
	}))
	_, err := tw.Write(content)
	require.NoError(t, err)
	require.NoError(t, tw.Close())
	require.NoError(t, gw.Close())

	return &buf
}

func TestExtractReplacesRegularFile(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "apm_launcher.so")
	require.NoError(t, os.WriteFile(target, []byte("old launcher"), 0o755))

	buf := makeTarGz(t, "apm_launcher.so", 0o755, []byte("new launcher"))

	require.NoError(t, Extract(buf, dir))

	data, err := os.ReadFile(target)
	require.NoError(t, err)
	assert.Equal(t, []byte("new launcher"), data)

	info, err := os.Stat(target)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o755), info.Mode().Perm())
}

type errReader struct{}

func (errReader) Read(_ []byte) (int, error) {
	return 0, errors.New("copy failed")
}

func TestExtractRegularFileKeepsExistingFileOnCopyError(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("windows keeps the legacy remove-and-write behavior")
	}

	dir := t.TempDir()
	target := filepath.Join(dir, "apm_launcher.so")
	require.NoError(t, os.WriteFile(target, []byte("old launcher"), 0o755))

	require.Error(t, extractRegularFile(errReader{}, target, 0o755))

	data, err := os.ReadFile(target)
	require.NoError(t, err)
	assert.Equal(t, []byte("old launcher"), data)

	matches, err := filepath.Glob(filepath.Join(dir, ".dl.tmp.*"))
	require.NoError(t, err)
	assert.Empty(t, matches)
}
