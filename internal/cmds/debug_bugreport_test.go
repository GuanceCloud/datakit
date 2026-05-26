// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package cmds

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
)

func TestBugreport_collectGitRepos(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "test-gitrepos")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a test Git repo config file
	testConfContent := `token=tkn_123456
password="secret"`
	testConfPath := filepath.Join(tmpDir, "test.conf")
	if err = os.WriteFile(testConfPath, []byte(testConfContent), 0o644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	// Create a non-conf file to ensure it's skipped
	testNonConfPath := filepath.Join(tmpDir, "test.txt")
	if err = os.WriteFile(testNonConfPath, []byte("test"), 0o644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Save original GitReposDir and restore it later
	originalGitReposDir := datakit.GitReposDir
	datakit.GitReposDir = tmpDir
	defer func() {
		datakit.GitReposDir = originalGitReposDir
	}()

	// Create datakitInfo instance
	info := &datakitInfo{}
	if err = info.init(); err != nil {
		t.Fatalf("Failed to init datakitInfo: %v", err)
	}
	defer info.clean()

	// Test collectGitRepos
	if err = info.collectGitRepos(); err != nil {
		t.Errorf("collectGitRepos failed: %v", err)
	}

	// Verify the file was copied and sanitized
	gitreposDir := filepath.Join(info.tmpDir, "gitrepos")
	copiedFile := filepath.Join(gitreposDir, "test.conf.copy")

	if _, err = os.Stat(copiedFile); os.IsNotExist(err) {
		t.Errorf("Expected copied file %s to exist", copiedFile)
	}

	// Read the copied file and verify sensitive information is sanitized
	content, err := os.ReadFile(copiedFile)
	if err != nil {
		t.Fatalf("Failed to read copied file: %v", err)
	}

	contentStr := string(content)
	if strings.Contains(contentStr, "tkn_123456") {
		t.Errorf("Token should be sanitized, but found: %s", contentStr)
	}

	if strings.Contains(contentStr, "secret") {
		t.Errorf("Password should be sanitized, but found: %s", contentStr)
	}

	if !strings.Contains(contentStr, "******") {
		t.Errorf("Expected sanitized content to contain ******, but got: %s", contentStr)
	}
}

func TestBugreport_collectLog_externals(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "test-externals")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create externals directory
	externalsDir := filepath.Join(tmpDir, "externals")
	if err = os.Mkdir(externalsDir, 0o755); err != nil {
		t.Fatalf("Failed to create externals dir: %v", err)
	}

	// Create test log files
	testLogContent := `Test log content`
	testLogPath := filepath.Join(externalsDir, "test.log")
	if err = os.WriteFile(testLogPath, []byte(testLogContent), 0o644); err != nil {
		t.Fatalf("Failed to write test log: %v", err)
	}

	// Create a non-log file to ensure it's skipped
	testNonLogPath := filepath.Join(externalsDir, "test.txt")
	if err = os.WriteFile(testNonLogPath, []byte("test"), 0o644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Save original InstallDir and restore it later
	originalInstallDir := datakit.InstallDir
	datakit.InstallDir = tmpDir
	defer func() {
		datakit.InstallDir = originalInstallDir
	}()

	// Create datakitInfo instance
	info := &datakitInfo{}
	if err = info.init(); err != nil {
		t.Fatalf("Failed to init datakitInfo: %v", err)
	}
	defer info.clean()

	// Test collectLog with externals
	if err = info.collectLog(); err != nil {
		t.Errorf("collectLog failed: %v", err)
	}

	// Verify the log file was copied
	externalsLogDir := filepath.Join(info.tmpDir, "log", "externals")
	copiedFile := filepath.Join(externalsLogDir, "test.log")

	if _, err := os.Stat(copiedFile); os.IsNotExist(err) {
		t.Errorf("Expected copied log file %s to exist", copiedFile)
	}

	// Read the copied file and verify content
	content, err := os.ReadFile(copiedFile)
	if err != nil {
		t.Fatalf("Failed to read copied log file: %v", err)
	}

	if string(content) != testLogContent {
		t.Errorf("Expected log content to be %q, but got %q", testLogContent, string(content))
	}
}

func TestBugreport_escapeString(t *testing.T) {
	info := &datakitInfo{}

	tests := []struct {
		name     string
		input    string
		kinds    []string
		expected string
	}{
		{
			name:     "dataway token",
			input:    "token=tkn_123456",
			kinds:    []string{"dataway"},
			expected: "token=******",
		},
		{
			name:     "password",
			input:    "password=\"secret\"",
			kinds:    []string{"password"},
			expected: "password = \"******\"",
		},
		{
			name:     "uri",
			input:    "https://user:pass@example.com",
			kinds:    []string{"uri"},
			expected: "https://user:******@example.com",
		},
		{
			name:     "multiple kinds",
			input:    "token=tkn_123456 password=\"secret\" https://user:pass@example.com",
			kinds:    []string{"dataway", "password", "uri"},
			expected: "token=****** password = \"******\" https://user:******@example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := info.escapeString(tt.input, tt.kinds)
			if result != tt.expected {
				t.Errorf("escapeString(%q, %v) = %q, want %q", tt.input, tt.kinds, result, tt.expected)
			}
		})
	}
}

func TestBugreport_containString(t *testing.T) {
	info := &datakitInfo{}

	tests := []struct {
		name     string
		input    string
		substrs  []string
		expected bool
	}{
		{
			name:     "contains substring",
			input:    "password123",
			substrs:  []string{"password"},
			expected: true,
		},
		{
			name:     "case insensitive",
			input:    "PASSWORD123",
			substrs:  []string{"password"},
			expected: true,
		},
		{
			name:     "does not contain",
			input:    "user123",
			substrs:  []string{"password"},
			expected: false,
		},
		{
			name:     "multiple substrs",
			input:    "token123",
			substrs:  []string{"password", "token"},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := info.containString(tt.input, tt.substrs)
			if result != tt.expected {
				t.Errorf("containString(%q, %v) = %v, want %v", tt.input, tt.substrs, result, tt.expected)
			}
		})
	}
}

func TestUploadBugReportViaDataway(t *testing.T) {
	originalOSS := *flagDebugBugreportOSS
	originalDataway := *flagDebugBugreportDataway
	defer func() {
		*flagDebugBugreportOSS = originalOSS
		*flagDebugBugreportDataway = originalDataway
	}()
	t.Run("use dataway when oss flag is empty", func(t *testing.T) {
		*flagDebugBugreportOSS = ""
		if !shouldUploadBugReportViaDataway() {
			t.Fatalf("expected shouldUploadBugReportViaDataway to return true")
		}
	})

	t.Run("do not use dataway when oss flag is set", func(t *testing.T) {
		*flagDebugBugreportOSS = "host:bucket:ak:sk"
		if shouldUploadBugReportViaDataway() {
			t.Fatalf("expected shouldUploadBugReportViaDataway to return false")
		}
	})

	t.Run("use custom dataway URLs", func(t *testing.T) {
		*flagDebugBugreportDataway = " http://dataway-1 , http://dataway-2 "

		urls := bugReportDatawayURLs()
		if len(urls) != 2 {
			t.Fatalf("expected 2 custom dataway URLs, got %d", len(urls))
		}
		if urls[0] != "http://dataway-1?token=bugreport-default" || urls[1] != "http://dataway-2?token=bugreport-default" {
			t.Fatalf("unexpected custom dataway URLs: %#v", urls)
		}

		*flagDebugBugreportDataway = ""
	})

	t.Run("keep existing custom dataway token", func(t *testing.T) {
		*flagDebugBugreportDataway = "http://dataway-1?token=tkn_custom"

		urls := bugReportDatawayURLs()
		if len(urls) != 1 {
			t.Fatalf("expected 1 custom dataway URL, got %d", len(urls))
		}
		if urls[0] != "http://dataway-1?token=tkn_custom" {
			t.Fatalf("unexpected custom dataway URL: %s", urls[0])
		}

		*flagDebugBugreportDataway = ""
	})

	const zipContent = "fake-zip-content"

	tmpFileName := ""
	if tmpFile, err := os.CreateTemp(t.TempDir(), "bugreport-*.zip"); err != nil {
		t.Fatalf("create temp zip file: %v", err)
	} else {
		tmpFileName = tmpFile.Name()
		defer func() {
			if err := os.RemoveAll(tmpFileName); err != nil {
				t.Errorf("remove temp zip file failed: %v", err)
			}
		}()
		if _, err := tmpFile.WriteString(zipContent); err != nil {
			t.Fatalf("write temp zip file: %v", err)
		}
		if err := tmpFile.Close(); err != nil {
			t.Fatalf("close temp zip file: %v", err)
		}
	}

	originalCfg := config.Cfg
	config.Cfg = config.DefaultConfig()
	defer func() {
		config.Cfg = originalCfg
	}()

	t.Run("dataway ok", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != datakit.BugReportUpload {
				t.Fatalf("unexpected upload path: %s", r.URL.Path)
			}

			if r.Header.Get("Host-Name") == "" {
				t.Fatalf("expected Host-Name header to be set")
			}

			if body, err := io.ReadAll(r.Body); err != nil {
				t.Fatalf("read request body: %v", err)
			} else {
				defer r.Body.Close()
				if string(body) != zipContent {
					t.Fatalf("unexpected request body: %q", string(body))
				}
			}

			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"content":"https://example.com/bugreport.zip"}`))
		}))
		defer ts.Close()
		config.Cfg.Dataway.URLs = []string{ts.URL + "?token=tkn_11111111111111111111"}
		if err := config.Cfg.SetHostname(); err != nil {
			t.Fatalf("set hostname: %v", err)
		}

		if _, err := uploadBugReportViaDataway(tmpFileName); err != nil {
			t.Fatalf("uploadBugReportViaDataway failed: %v", err)
		}
	})

	t.Run("custom dataway ok without config dataway", func(t *testing.T) {
		originalDatawayCfg := config.Cfg.Dataway
		defer func() {
			config.Cfg.Dataway = originalDatawayCfg
		}()

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != datakit.BugReportUpload {
				t.Fatalf("unexpected upload path: %s", r.URL.Path)
			}
			if got := r.URL.Query().Get("token"); got != defaultBugReportDatawayToken {
				t.Fatalf("unexpected token: %s", got)
			}

			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"content":"2026-04-22/host-a/dkbr_xxx.zip"}`))
		}))
		defer ts.Close()

		config.Cfg.Dataway = nil
		*flagDebugBugreportDataway = ts.URL

		if _, err := uploadBugReportViaDataway(tmpFileName); err != nil {
			t.Fatalf("uploadBugReportViaDataway failed: %v", err)
		}

		*flagDebugBugreportDataway = ""
	})

	t.Run("use config dataway without token", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if got := r.URL.Query().Get("token"); got != defaultBugReportDatawayToken {
				t.Fatalf("unexpected token: %s", got)
			}

			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"content":"2026-04-22/host-a/dkbr_xxx.zip"}`))
		}))
		defer ts.Close()

		config.Cfg.Dataway.URLs = []string{ts.URL}
		if err := config.Cfg.SetHostname(); err != nil {
			t.Fatalf("set hostname: %v", err)
		}

		if _, err := uploadBugReportViaDataway(tmpFileName); err != nil {
			t.Fatalf("uploadBugReportViaDataway failed: %v", err)
		}
	})

	t.Run("dataway failed", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("upload failed on server"))
		}))
		defer ts.Close()

		config.Cfg.Dataway.URLs = []string{ts.URL + "?token=tkn_11111111111111111111"}
		if err := config.Cfg.SetHostname(); err != nil {
			t.Fatalf("set hostname: %v", err)
		}

		if _, err := uploadBugReportViaDataway(tmpFileName); err == nil {
			t.Fatalf("expected uploadBugReportViaDataway to fail")
		} else if !strings.Contains(err.Error(), "all-retry-failed") {
			t.Fatalf("expected wrapped retry error, got: %v", err)
		}
	})

	t.Run("dataway not support bugreport", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte("<html>DataWay Endpoint</html>"))
		}))
		defer ts.Close()

		config.Cfg.Dataway.URLs = []string{ts.URL + "?token=tkn_11111111111111111111"}
		if err := config.Cfg.SetHostname(); err != nil {
			t.Fatalf("set hostname: %v", err)
		}

		if _, err := uploadBugReportViaDataway(tmpFileName); err == nil {
			t.Fatalf("expected uploadBugReportViaDataway to fail")
		} else {
			if !strings.Contains(err.Error(), "dataway does not support uploading bugreport") {
				t.Fatalf("expected unsupported bugreport error, got: %v", err)
			}
			if strings.Contains(err.Error(), "DataWay Endpoint") {
				t.Fatalf("expected error not to contain response body, got: %v", err)
			}
		}
	})
}
