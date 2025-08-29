// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package configwatcher

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewFileWatcher(t *testing.T) {
	t.Run("testfile", func(t *testing.T) {
		tmpFile, err := os.CreateTemp("", "testfile")
		assert.NoError(t, err)
		defer os.Remove(tmpFile.Name())

		watcher, err := newFileWatcher(tmpFile.Name(), withRecursive(true), withMaxDiffSize(1024))
		assert.NoError(t, err)
		assert.NotNil(t, watcher)

		states := watcher.getCurrentStates()
		assert.Equal(t, 1, len(states))
		assert.Equal(t, tmpFile.Name(), states[tmpFile.Name()].path)
		assert.Equal(t, true, states[tmpFile.Name()].exists)
	})

	t.Run("testfile", func(t *testing.T) {
		_, err := newFileWatcher("/nonexistent/file/path")
		assert.NoError(t, err)
	})

	t.Run("testdir", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "testdir")
		assert.NoError(t, err)
		defer os.RemoveAll(tmpDir)

		watcher, err := newFileWatcher(tmpDir)
		assert.NoError(t, err)

		states := watcher.getCurrentStates()
		assert.Equal(t, 0, len(states))
	})
}

func TestRecursive(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "testdir-A")
	assert.NoError(t, err)
	defer os.Remove(tmpDir)

	tmpDir2, err := os.MkdirTemp(tmpDir, "testdir-b")
	assert.NoError(t, err)
	defer os.Remove(tmpDir2)

	watcher, err := newFileWatcher(tmpDir, withRecursive(true), withMaxDiffSize(1024))
	assert.NoError(t, err)
	assert.NotNil(t, watcher)

	states := watcher.getCurrentStates()
	assert.Equal(t, 0, len(states))

	events, err := watcher.checkChanges()
	assert.NoError(t, err)
	assert.Equal(t, 0, len(events))

	// create file
	content := "test content"
	filePath := filepath.Join(tmpDir, "newfile.txt")
	err = os.WriteFile(filePath, []byte(content), 0o644)
	assert.NoError(t, err)

	filePath2 := filepath.Join(tmpDir2, "newfile2.txt")
	err = os.WriteFile(filePath2, []byte(content), 0o644)
	assert.NoError(t, err)

	time.Sleep(10 * time.Millisecond)

	// check
	events, err = watcher.checkChanges()
	assert.NoError(t, err)
	assert.Equal(t, 2, len(events))

	assert.Equal(t, filePath, events[0].path)
	assert.Equal(t, filePath2, events[1].path)
	assert.Equal(t, created, events[0].typ)
	assert.Equal(t, created, events[1].typ)
	assert.Equal(t, true, events[0].newState.exists)
	assert.Equal(t, true, events[1].newState.exists)

	assert.NotEqual(t, "", events[0].diff)
	t.Logf("diff: %s", events[0].diff)
	assert.NotEqual(t, "", events[1].diff)
	t.Logf("diff: %s", events[1].diff)

	states = watcher.getCurrentStates()
	assert.Equal(t, 2, len(states))
	assert.Equal(t, true, states[filePath].exists)
	assert.Equal(t, true, states[filePath2].exists)

	// modify
	newContent := "modified content"
	err = os.WriteFile(filePath, []byte(newContent), 0o644)
	assert.NoError(t, err)

	time.Sleep(10 * time.Millisecond)

	events, err = watcher.checkChanges()
	assert.NoError(t, err)
	assert.Equal(t, 1, len(events))
	assert.Equal(t, filePath, events[0].path)
	assert.Equal(t, modified, events[0].typ)

	assert.NotEqual(t, "", events[0].diff)
	t.Logf("diff: %s", events[0].diff)
}

func TestCheckChanges(t *testing.T) {
	t.Run("create", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "testdir")
		assert.NoError(t, err)
		defer os.RemoveAll(tmpDir)

		filePath := filepath.Join(tmpDir, "newfile.txt")
		watcher, err := newFileWatcher(filePath)
		assert.NoError(t, err)

		states := watcher.getCurrentStates()
		assert.Equal(t, 0, len(states))

		// create file
		content := "test content"
		err = os.WriteFile(filePath, []byte(content), 0o644)
		assert.NoError(t, err)

		time.Sleep(10 * time.Millisecond)

		// check
		events, err := watcher.checkChanges()
		assert.NoError(t, err)
		assert.Equal(t, 1, len(events))
		assert.Equal(t, created, events[0].typ)
		assert.Equal(t, true, events[0].newState.exists)
		assert.Equal(t, int64(len(content)), events[0].newState.size)
	})

	t.Run("delete", func(t *testing.T) {
		tmpFile, err := os.CreateTemp("", "testfile")
		assert.NoError(t, err)
		defer os.Remove(tmpFile.Name())

		content := "test content for deletion"
		_, err = tmpFile.WriteString(content)
		assert.NoError(t, err)
		tmpFile.Close()

		watcher, err := newFileWatcher(tmpFile.Name())
		assert.NoError(t, err)

		states := watcher.getCurrentStates()
		assert.Equal(t, 1, len(states))
		assert.Equal(t, true, states[tmpFile.Name()].exists)

		// remove file
		err = os.Remove(tmpFile.Name())
		assert.NoError(t, err)

		time.Sleep(10 * time.Millisecond)

		// check
		events, err := watcher.checkChanges()
		assert.NoError(t, err)
		assert.Equal(t, 1, len(events))
		assert.Equal(t, deleted, events[0].typ)
		assert.Equal(t, false, events[0].newState.exists)
	})

	t.Run("modify", func(t *testing.T) {
		tmpFile, err := os.CreateTemp("", "testfile")
		assert.NoError(t, err)
		defer os.Remove(tmpFile.Name())

		initialContent := "initial content"
		_, err = tmpFile.WriteString(initialContent)
		assert.NoError(t, err)
		tmpFile.Close()

		watcher, err := newFileWatcher(tmpFile.Name(), withMaxDiffSize(1024))
		assert.NoError(t, err)

		// modify
		newContent := "modified content"
		err = os.WriteFile(tmpFile.Name(), []byte(newContent), 0o644)
		assert.NoError(t, err)

		time.Sleep(10 * time.Millisecond)

		// check
		events, err := watcher.checkChanges()
		assert.NoError(t, err)
		assert.Equal(t, 1, len(events))
		assert.Equal(t, modified, events[0].typ)
		assert.Equal(t, int64(len(newContent)), events[0].newState.size)

		assert.NotEqual(t, "", events[0].diff)
		t.Logf("diff: %s", events[0].diff)
	})

	t.Run("no-change", func(t *testing.T) {
		tmpFile, err := os.CreateTemp("", "testfile")
		assert.NoError(t, err)
		defer os.Remove(tmpFile.Name())

		// write
		content := "test content"
		_, err = tmpFile.WriteString(content)
		assert.NoError(t, err)
		tmpFile.Close()

		watcher, err := newFileWatcher(tmpFile.Name())
		assert.NoError(t, err)

		time.Sleep(10 * time.Millisecond)

		// check
		events, err := watcher.checkChanges()
		assert.NoError(t, err)
		assert.Equal(t, 0, len(events))
	})

	t.Run("recreate", func(t *testing.T) {
		tmpFile, err := os.CreateTemp("", "testfile")
		assert.NoError(t, err)
		defer os.Remove(tmpFile.Name())

		// write
		content := "initial content"
		_, err = tmpFile.WriteString(content)
		assert.NoError(t, err)
		tmpFile.Close()

		watcher, err := newFileWatcher(tmpFile.Name())
		assert.NoError(t, err)

		err = os.Remove(tmpFile.Name())
		assert.NoError(t, err)

		time.Sleep(10 * time.Millisecond)

		// check
		events, err := watcher.checkChanges()
		assert.NoError(t, err)
		assert.Equal(t, 1, len(events))
		assert.Equal(t, deleted, events[0].typ)

		// recreate
		newContent := "recreated content"
		err = os.WriteFile(tmpFile.Name(), []byte(newContent), 0o644)
		assert.NoError(t, err)

		time.Sleep(10 * time.Millisecond)

		// check
		events, err = watcher.checkChanges()
		assert.NoError(t, err)
		assert.Equal(t, 1, len(events))
		assert.Equal(t, created, events[0].typ)
		assert.Equal(t, int64(len(newContent)), events[0].newState.size)
	})
}

func TestReset(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "testfile")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	// write
	content := "test content"
	_, err = tmpFile.WriteString(content)
	assert.NoError(t, err)
	tmpFile.Close()

	watcher, err := newFileWatcher(tmpFile.Name())
	assert.NoError(t, err)

	initialStates := watcher.getCurrentStates()
	assert.Equal(t, 1, len(initialStates))

	newContent := "modified content"
	err = os.WriteFile(tmpFile.Name(), []byte(newContent), 0o644)
	assert.NoError(t, err)

	// reset
	watcher.Reset()

	resetStates := watcher.getCurrentStates()
	assert.Equal(t, 1, len(resetStates))

	assert.NotEqual(t, initialStates[tmpFile.Name()].size, resetStates[tmpFile.Name()].size)
	assert.Equal(t, int64(len(newContent)), resetStates[tmpFile.Name()].size)
}

func TestLargeFileHandling(t *testing.T) {
	var maxHashSize int64 = 10

	tmpFile, err := os.CreateTemp("", "largefile")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	content := "test content - 0123456789"
	_, err = tmpFile.WriteString(content)
	assert.NoError(t, err)
	tmpFile.Close()

	watcher, err := newFileWatcher(tmpFile.Name(), withMaxDiffSize(maxHashSize))
	assert.NoError(t, err)
	assert.Equal(t, int64(10), watcher.maxDiffSize)

	// check states
	states := watcher.getCurrentStates()
	assert.Equal(t, 1, len(states))
	assert.Equal(t, true, states[tmpFile.Name()].exists)
	assert.Equal(t, 0, len(states[tmpFile.Name()].content))

	// modify
	newContent := "test content - 0123456789 - 0123456789"
	err = os.WriteFile(tmpFile.Name(), []byte(newContent), 0o644)
	assert.NoError(t, err)

	time.Sleep(10 * time.Millisecond)

	// check
	events, err := watcher.checkChanges()
	assert.NoError(t, err)
	assert.Equal(t, 1, len(events))
	assert.Equal(t, modified, events[0].typ)
	assert.Equal(t, "", events[0].diff)
}

func TestEdgeCases(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "testfile")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	watcher, err := newFileWatcher(tmpFile.Name())
	assert.NoError(t, err)

	states := watcher.getCurrentStates()
	assert.Equal(t, 1, len(states))
	assert.Equal(t, int64(0), states[tmpFile.Name()].size)

	content := "some content"
	err = os.WriteFile(tmpFile.Name(), []byte(content), 0o644)
	assert.NoError(t, err)

	time.Sleep(10 * time.Millisecond)

	// check
	events, err := watcher.checkChanges()
	assert.NoError(t, err)
	assert.Equal(t, 1, len(events))
	assert.Equal(t, modified, events[0].typ)

	// clear
	err = os.WriteFile(tmpFile.Name(), []byte{}, 0o644)
	assert.NoError(t, err)

	time.Sleep(10 * time.Millisecond)

	// check
	events, err = watcher.checkChanges()
	assert.NoError(t, err)
	assert.Equal(t, 1, len(events))
	assert.Equal(t, modified, events[0].typ)
	assert.Equal(t, int64(0), events[0].newState.size)
}
