// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build !windows
// +build !windows

package hostchange

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/changes"
)

func TestFileChecker_Init(t *testing.T) {
	// Test valid configuration
	t.Run("valid configuration", func(t *testing.T) {
		fc := &FileChecker{
			Enabled: true,
			Files:   []string{"/etc/passwd"},
		}

		input := &Input{}
		err := fc.Init(input)
		assert.NoError(t, err)
		assert.Equal(t, input, fc.input)
	})

	// Test relative file path
	t.Run("relative file path", func(t *testing.T) {
		fc := &FileChecker{
			Enabled: true,
			Files:   []string{"relative/path.txt"},
		}

		input := &Input{}
		err := fc.Init(input)
		assert.NoError(t, err)
	})

	// Test disabled checker
	t.Run("disabled checker", func(t *testing.T) {
		fc := &FileChecker{
			Enabled: false,
			Files:   []string{"relative/path.txt"}, // Invalid path should be ignored when disabled
		}

		input := &Input{}
		err := fc.Init(input)
		assert.NoError(t, err)
	})
}

func TestFileChecker_shouldIgnore(t *testing.T) {
	fc := &FileChecker{
		IgnorePaths: []string{"/etc/passwd", "/tmp/"},
	}

	assert.True(t, fc.shouldIgnore("/etc/passwd"))
	assert.True(t, fc.shouldIgnore("/tmp/file.txt"))
	assert.False(t, fc.shouldIgnore("/etc/group"))
	assert.False(t, fc.shouldIgnore("/home/user/file.txt"))
}

func TestFileChecker_Collect(t *testing.T) {
	// Initialize host manifest
	err := changes.LoadHostManifest()
	require.NoError(t, err)

	// Test disabled checker
	t.Run("disabled_checker", func(t *testing.T) {
		// Create test files
		testDir, err := os.MkdirTemp("", "test_files")
		require.NoError(t, err)
		defer os.RemoveAll(testDir)

		file1 := filepath.Join(testDir, "file1.txt")
		err = os.WriteFile(file1, []byte("initial content 1"), 0o644)
		require.NoError(t, err)

		// Create test cache
		cacheDir, err := os.MkdirTemp("", "test_cache")
		require.NoError(t, err)
		defer os.RemoveAll(cacheDir)

		input := &Input{}

		fc := &FileChecker{
			Enabled: false,
			Files:   []string{file1},
			input:   input,
		}

		changes, err := fc.Collect()
		assert.NoError(t, err)
		assert.Nil(t, changes)
	})

	// Test initial collection (all files are new, but no events should be generated)
	t.Run("initial_collection", func(t *testing.T) {
		// Create test files
		testDir, err := os.MkdirTemp("", "test_files")
		require.NoError(t, err)
		defer os.RemoveAll(testDir)

		file1 := filepath.Join(testDir, "file1.txt")
		err = os.WriteFile(file1, []byte("initial content 1"), 0o644)
		require.NoError(t, err)

		file2 := filepath.Join(testDir, "file2.txt")
		err = os.WriteFile(file2, []byte("initial content 2"), 0o644)
		require.NoError(t, err)

		// Create test cache

		input := &Input{}

		fc := &FileChecker{
			Enabled: true,
			Files:   []string{file1, file2},
			input:   input,
		}

		// First collection should NOT detect any changes - only initialize cache
		changes, err := fc.Collect()
		assert.NoError(t, err)
		assert.Len(t, changes, 0) // No events should be generated on first collection
	})

	// Test new file added after initial collection
	t.Run("new_file_after_initial_collection", func(t *testing.T) {
		// Create test files
		testDir, err := os.MkdirTemp("", "test_files")
		require.NoError(t, err)
		defer os.RemoveAll(testDir)

		file1 := filepath.Join(testDir, "file1.txt")
		err = os.WriteFile(file1, []byte("initial content 1"), 0o644)
		require.NoError(t, err)

		input := &Input{}

		fc := &FileChecker{
			Enabled: true,
			Files:   []string{file1},
			input:   input,
		}

		// First collection should NOT detect any changes
		changes, err := fc.Collect()
		assert.NoError(t, err)
		assert.Len(t, changes, 0)

		// Now add a new file and update the checker configuration
		file2 := filepath.Join(testDir, "file2.txt")
		err = os.WriteFile(file2, []byte("initial content 2"), 0o644)
		require.NoError(t, err)

		fc.Files = []string{file1, file2}

		// Second collection should detect the new file
		changes, err = fc.Collect()
		assert.NoError(t, err)
		assert.Len(t, changes, 1) // One new file event

		// Verify changes contain expected content
		assert.Contains(t, changes[0].Message, "file created")
	})

	// Test file modification detection
	t.Run("file_modification_detection", func(t *testing.T) {
		// Create test files
		testDir, err := os.MkdirTemp("", "test_files")
		require.NoError(t, err)
		defer os.RemoveAll(testDir)

		file1 := filepath.Join(testDir, "file1.txt")
		err = os.WriteFile(file1, []byte("initial content 1"), 0o644)
		require.NoError(t, err)

		file2 := filepath.Join(testDir, "file2.txt")
		err = os.WriteFile(file2, []byte("initial content 2"), 0o644)
		require.NoError(t, err)

		input := &Input{}

		fc := &FileChecker{
			Enabled: true,
			Files:   []string{file1, file2},
			input:   input,
		}

		// First collection to cache the files
		_, err = fc.Collect()
		require.NoError(t, err)

		time.Sleep(100 * time.Millisecond)
		// Modify file1
		err = os.WriteFile(file1, []byte("modified content 1"), 0o644)
		require.NoError(t, err)

		// Collect changes
		changes, err := fc.Collect()
		assert.NoError(t, err)
		assert.Len(t, changes, 1)
		assert.Contains(t, changes[0].Message, file1)
	})

	// Test new file detection
	t.Run("new_file_detection", func(t *testing.T) {
		// Create test files
		testDir, err := os.MkdirTemp("", "test_files")
		require.NoError(t, err)
		defer os.RemoveAll(testDir)

		file1 := filepath.Join(testDir, "file1.txt")
		err = os.WriteFile(file1, []byte("initial content 1"), 0o644)
		require.NoError(t, err)

		file2 := filepath.Join(testDir, "file2.txt")
		err = os.WriteFile(file2, []byte("initial content 2"), 0o644)
		require.NoError(t, err)

		file3 := filepath.Join(testDir, "file3.txt")

		input := &Input{}

		fc := &FileChecker{
			Enabled: true,
			Files:   []string{file1, file2},
			input:   input,
		}

		// First collection to cache existing files
		_, err = fc.Collect()
		require.NoError(t, err)

		// Create a new file
		err = os.WriteFile(file3, []byte("new content 3"), 0o644)
		require.NoError(t, err)

		// Update file checker with new file
		fc.Files = []string{file1, file2, file3}

		// Collect changes
		changes, err := fc.Collect()
		assert.NoError(t, err)
		assert.Len(t, changes, 1)
		assert.Contains(t, changes[0].Message, file3)
	})

	// Test file deletion detection
	t.Run("file_deletion_detection", func(t *testing.T) {
		// Create test files
		testDir, err := os.MkdirTemp("", "test_files")
		require.NoError(t, err)
		defer os.RemoveAll(testDir)

		file1 := filepath.Join(testDir, "file1.txt")
		err = os.WriteFile(file1, []byte("initial content 1"), 0o644)
		require.NoError(t, err)

		file2 := filepath.Join(testDir, "file2.txt")
		err = os.WriteFile(file2, []byte("initial content 2"), 0o644)
		require.NoError(t, err)

		input := &Input{}

		fc := &FileChecker{
			Enabled: true,
			Files:   []string{file1, file2},
			input:   input,
		}

		// First collection to cache the files
		_, err = fc.Collect()
		require.NoError(t, err)

		// Delete file2
		err = os.Remove(file2)
		require.NoError(t, err)

		// Collect changes
		changes, err := fc.Collect()
		assert.NoError(t, err)
		assert.Len(t, changes, 1)
		assert.Contains(t, changes[0].Message, file2)
	})

	// Test ignore paths functionality
	t.Run("ignore_paths", func(t *testing.T) {
		// Create test files
		testDir, err := os.MkdirTemp("", "test_files")
		require.NoError(t, err)
		defer os.RemoveAll(testDir)

		file1 := filepath.Join(testDir, "file1.txt")
		err = os.WriteFile(file1, []byte("initial content 1"), 0o644)
		require.NoError(t, err)

		// Create a file that should be ignored
		ignoreFile := filepath.Join(testDir, "ignore.txt")
		err = os.WriteFile(ignoreFile, []byte("ignore content"), 0o644)
		require.NoError(t, err)

		// Create input with ignore paths
		input := &Input{}

		fc := &FileChecker{
			Enabled:     true,
			Files:       []string{file1, ignoreFile},
			IgnorePaths: []string{ignoreFile},
			input:       input,
		}

		// First collection to cache the files
		_, err = fc.Collect()
		require.NoError(t, err)

		time.Sleep(100 * time.Millisecond)
		// Modify both files
		err = os.WriteFile(file1, []byte("modified content 1"), 0o644)
		require.NoError(t, err)
		err = os.WriteFile(ignoreFile, []byte("modified ignore content"), 0o644)
		require.NoError(t, err)

		// Collect changes - should only see file1 change
		changes, err := fc.Collect()
		assert.NoError(t, err)
		assert.Len(t, changes, 1)
		assert.Contains(t, changes[0].Message, file1)
	})
}
