// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build !windows
// +build !windows

package hostchange

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestParseCrontab tests the ParseCrontab function with different scenarios
func TestParseCrontab(t *testing.T) {
	t.Run("SystemCrontab", func(t *testing.T) {
		// Create a temporary system crontab file
		tempContent := `# Comment line
PATH=/usr/bin:/bin

# Regular crontab entry
0 1 * * * root /usr/bin/backup
30 2 * * 1-5 admin /usr/bin/update
`

		file, err := os.CreateTemp("", "system_crontab")
		assert.NoError(t, err)
		defer os.Remove(file.Name())

		_, err = file.WriteString(tempContent)
		assert.NoError(t, err)
		file.Close()

		// Parse the crontab file
		content, err := ReadFile(file.Name())
		assert.NoError(t, err)
		crontabs, err := ParseCrontab(file.Name(), content)
		assert.NoError(t, err)

		// Verify the parsed crontab entries
		assert.Len(t, crontabs, 2)

		// Check first entry
		assert.Equal(t, "root", crontabs[0].User)
		assert.Equal(t, "0 1 * * *", crontabs[0].Schedule)
		assert.Equal(t, "/usr/bin/backup", crontabs[0].Command)

		// Check second entry
		assert.Equal(t, "admin", crontabs[1].User)
		assert.Equal(t, "30 2 * * 1-5", crontabs[1].Schedule)
		assert.Equal(t, "/usr/bin/update", crontabs[1].Command)
	})

	t.Run("UserCrontab", func(t *testing.T) {
		// Save original path so we can restore it later
		originalSpoolPath := DirPathSpoolCrontabs
		defer func() {
			DirPathSpoolCrontabs = originalSpoolPath
		}()

		// Create a temporary directory to simulate /var/spool/cron/crontabs/
		tempDir, err := os.MkdirTemp("", "crontabs")
		assert.NoError(t, err)
		defer os.RemoveAll(tempDir)

		// Update the path variable to use our temporary directory
		DirPathSpoolCrontabs = tempDir

		// Create a user crontab file in the temporary directory
		// The filename represents the username
		tempContent := `# User crontab for testuser
PATH=/usr/bin:/bin

# Regular crontab entry
* * * * * /usr/bin/monitor
30 2 * * 1-5 /usr/bin/user_backup
`

		// Create a file named 'testuser' in our temporary directory
		userCronPath := filepath.Join(tempDir, "testuser")
		err = os.WriteFile(userCronPath, []byte(tempContent), 0o644)
		assert.NoError(t, err)

		// Parse the user crontab file
		content, err := ReadFile(userCronPath)
		assert.NoError(t, err)
		crontabs, err := ParseCrontab(userCronPath, content)
		assert.NoError(t, err)

		// Verify the parsed crontab entries
		assert.Len(t, crontabs, 2)

		// Check first entry
		assert.Equal(t, "testuser", crontabs[0].User)
		assert.Equal(t, "* * * * *", crontabs[0].Schedule)
		assert.Equal(t, "/usr/bin/monitor", crontabs[0].Command)

		// Check second entry
		assert.Equal(t, "testuser", crontabs[1].User)
		assert.Equal(t, "30 2 * * 1-5", crontabs[1].Schedule)
		assert.Equal(t, "/usr/bin/user_backup", crontabs[1].Command)
	})

	t.Run("InvalidFile", func(t *testing.T) {
		// Try to parse a non-existent file
		// This should not be called with ParseCrontab directly, but we'll test the error handling
		_, err := ReadFile("/non/existent/file")
		assert.Error(t, err)
	})
}

// TestGetCrontabKey tests generating keys for crontab entries
func TestGetCrontabKey(t *testing.T) {
	// Create two different crontab entries
	cron1 := &Crontab{
		User:     "root",
		Schedule: "0 1 * * *",
		Command:  "/usr/bin/backup",
	}

	cron2 := &Crontab{
		User:     "admin",
		Schedule: "30 2 * * 1-5",
		Command:  "/usr/bin/update",
	}

	// Create a crontab entry with FilePath
	cron3 := &Crontab{
		User:     "root",
		Schedule: "0 1 * * *",
		Command:  "/usr/bin/backup",
		FilePath: "/etc/crontab",
	}

	// Generate keys
	key1 := getCrontabKey(cron1)
	key2 := getCrontabKey(cron2)
	key3 := getCrontabKey(cron3)

	// Verify keys are unique
	assert.NotEqual(t, key1, key2)
	assert.NotEqual(t, key1, key3)
	assert.NotEqual(t, key2, key3)

	// Verify keys are generated based on the correct string representation
	// The actual key is a hash of this string, not the string itself
	stringKey1 := "0 1 * * *:root:/usr/bin/backup"
	expectedHash1 := GetHashCode([]byte(stringKey1))
	assert.Equal(t, expectedHash1, key1)

	// Verify key with FilePath
	stringKey3 := "0 1 * * *:root:/usr/bin/backup:/etc/crontab"
	expectedHash3 := GetHashCode([]byte(stringKey3))
	assert.Equal(t, expectedHash3, key3)
}

// TestGetCrontabFiles tests getting all crontab files from system paths
func TestGetCrontabFiles(t *testing.T) {
	// Save original paths so we can restore them later
	originalFilePathCrontab := FilePathCrontab
	originalDirPathCronD := DirPathCronD
	originalDirPathSpoolCrontabs := DirPathSpoolCrontabs
	defer func() {
		FilePathCrontab = originalFilePathCrontab
		DirPathCronD = originalDirPathCronD
		DirPathSpoolCrontabs = originalDirPathSpoolCrontabs
	}()

	// Create temporary directories to simulate the crontab file structure
	tempDir, err := os.MkdirTemp("", "crontab_test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create subdirectories
	cronDDir := filepath.Join(tempDir, "cron.d")
	crontabsDir := filepath.Join(tempDir, "crontabs")
	assert.NoError(t, os.Mkdir(cronDDir, 0o755))
	assert.NoError(t, os.Mkdir(crontabsDir, 0o755))

	// Create test files
	crontabFile := filepath.Join(tempDir, "crontab")
	assert.NoError(t, os.WriteFile(crontabFile, []byte("# Test crontab"), 0o644))

	// Create files in /etc/cron.d/
	for i := 0; i < 3; i++ {
		filename := fmt.Sprintf("cronjob%d", i)
		filePath := filepath.Join(cronDDir, filename)
		assert.NoError(t, os.WriteFile(filePath, []byte(fmt.Sprintf("# Test cronjob %d", i)), 0o644))
	}

	// Create files in /var/spool/cron/crontabs/
	users := []string{"root", "user1", "user2"}
	for _, user := range users {
		filePath := filepath.Join(crontabsDir, user)
		assert.NoError(t, os.WriteFile(filePath, []byte(fmt.Sprintf("# Test crontab for %s", user)), 0o644))
	}

	// Update the path variables to use our temporary directories
	FilePathCrontab = crontabFile
	DirPathCronD = cronDDir
	DirPathSpoolCrontabs = crontabsDir

	// Test getCrontabFiles function
	files, err := getCrontabFiles()
	assert.NoError(t, err)

	// Verify the expected files are in the result
	assert.Contains(t, files, crontabFile)

	// Verify all cron.d files are included
	for i := 0; i < 3; i++ {
		filename := fmt.Sprintf("cronjob%d", i)
		expectedPath := filepath.Join(cronDDir, filename)
		assert.Contains(t, files, expectedPath)
	}

	// Verify all user crontab files are included
	for _, user := range users {
		expectedPath := filepath.Join(crontabsDir, user)
		assert.Contains(t, files, expectedPath)
	}

	// Verify the total number of files
	// 1 (crontab) + 3 (cron.d) + 3 (crontabs) = 7 files
	assert.Len(t, files, 7)
}

// TestCrontabChecker tests the CrontabChecker functionality
func TestCrontabChecker(t *testing.T) {
	t.Run("Init", func(t *testing.T) {
		// Create a CrontabChecker instance
		cc := &CrontabChecker{
			Enabled: true,
		}

		// Create an Input instance
		input := &Input{}

		// Test initialization
		err := cc.Init(input)
		assert.NoError(t, err)

		// Verify initialization
		assert.Equal(t, input, cc.input)
	})

	// Note: The Init_WithCustomPath test has been removed because CrontabPath field is no longer used

	t.Run("Init_NilInput", func(t *testing.T) {
		// Create a CrontabChecker instance
		cc := &CrontabChecker{}

		// Test initialization with nil input (should not error as per current implementation)
		err := cc.Init(nil)
		assert.NoError(t, err) // Current implementation allows nil input
	})

	t.Run("Collect_Disabled", func(t *testing.T) {
		// Create a CrontabChecker instance that is disabled
		cc := &CrontabChecker{
			Enabled: false,
		}

		// Create an Input instance
		input := &Input{}

		// Initialize
		err := cc.Init(input)
		assert.NoError(t, err)

		// Test Collect
		changes, err := cc.Collect()
		assert.NoError(t, err)
		assert.Empty(t, changes) // Should return empty slice when disabled
	})

	t.Run("Collect_Enabled", func(t *testing.T) {
		// Create a CrontabChecker instance that is enabled
		cc := &CrontabChecker{
			Enabled: true,
		}

		// Create an Input instance
		input := &Input{}

		// Initialize
		err := cc.Init(input)
		assert.NoError(t, err)

		// Test Collect (this will try to read real system crontab files)
		// We're mainly testing that it doesn't panic and returns a valid response
		changes, err := cc.Collect()
		// The error might be nil or an error if the crontab files don't exist, both are acceptable
		if err != nil {
			// It's okay if we get an error when running on a non-Linux system
			assert.Contains(t, err.Error(), "failed to fetch crontab info and events")
		} else {
			// If no error, we should get a valid changes slice
			assert.NotNil(t, changes)
		}
	})
}
