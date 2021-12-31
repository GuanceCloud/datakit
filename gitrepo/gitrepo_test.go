package gitrepo

import (
	"context"
	"testing"
	"time"

	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	tu "gitlab.jiagouyun.com/cloudcare-tools/cliutils/testutil"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
)

// go test -v -timeout 30s -run ^TestGetGitClonePathFromGitURL$ gitlab.jiagouyun.com/cloudcare-tools/datakit/gitrepo
func TestGetGitClonePathFromGitURL(t *testing.T) {
	cases := []struct {
		name          string
		gitURL        string
		expect        string
		shouldBeError bool
	}{
		{
			name:          "http_test_url",
			gitURL:        "http://username:password@github.com/path/to/repository1.git",
			expect:        "/usr/local/datakit/gitrepos/repository1",
			shouldBeError: false,
		},

		{
			name:          "git_test_url",
			gitURL:        "git@github.com:path/to/repository4.git",
			expect:        "/usr/local/datakit/gitrepos/repository4",
			shouldBeError: false,
		},

		{
			name:          "ssh_test_url",
			gitURL:        "ssh://git@github.com:9000/path/to/repository5.git",
			expect:        "/usr/local/datakit/gitrepos/repository5",
			shouldBeError: false,
		},

		{
			name:          "empty_test_url",
			gitURL:        "",
			expect:        "",
			shouldBeError: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repoName, err := getGitClonePathFromGitURL(tc.gitURL)
			if err != nil && !tc.shouldBeError {
				t.Error(err)
			}
			tu.Equals(t, tc.expect, repoName)
		})
	}
}

// go test -v -timeout 30s -run ^TestIsUserNamePasswordAuth$ gitlab.jiagouyun.com/cloudcare-tools/datakit/gitrepo
func TestIsUserNamePasswordAuth(t *testing.T) {
	cases := []struct {
		name          string
		gitURL        string
		expect        bool
		shouldBeError bool
	}{
		{
			name:          "http_test_url",
			gitURL:        "http://username:password@github.com/path/to/repository.git",
			expect:        true,
			shouldBeError: false,
		},

		{
			name:          "git_test_url",
			gitURL:        "git@github.com:path/to/repository.git",
			expect:        false,
			shouldBeError: false,
		},

		{
			name:          "ssh_test_url",
			gitURL:        "ssh://git@github.com:9000/path/to/repository.git",
			expect:        false,
			shouldBeError: false,
		},

		{
			name:          "invalid_test_url",
			gitURL:        "ok",
			expect:        false,
			shouldBeError: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			isPassword, err := isUserNamePasswordAuth(tc.gitURL)
			if err != nil && !tc.shouldBeError {
				t.Error(err)
			}
			tu.Equals(t, tc.expect, isPassword)
		})
	}
}

// go test -v -timeout 30s -run ^TestGetUserNamePasswordFromGitURL$ gitlab.jiagouyun.com/cloudcare-tools/datakit/gitrepo
func TestGetUserNamePasswordFromGitURL(t *testing.T) {
	cases := []struct {
		name          string
		gitURL        string
		expect        map[string]string
		shouldBeError bool
	}{
		{
			name:   "http_test_url",
			gitURL: "http://username:password@github.com/path/to/repository.git",
			expect: map[string]string{
				"username": "password",
			},
			shouldBeError: false,
		},

		{
			name:   "git_test_url",
			gitURL: "git@github.com:path/to/repository.git",
			expect: map[string]string{
				"": "",
			},
			shouldBeError: false,
		},

		{
			name:   "ssh_test_url",
			gitURL: "ssh://git@github.com:9000/path/to/repository.git",
			expect: map[string]string{
				"": "",
			},
			shouldBeError: false,
		},

		{
			name:   "http_test_url_empty_username",
			gitURL: "http://:password@github.com/path/to/repository.git",
			expect: map[string]string{
				"": "password",
			},
			shouldBeError: true,
		},

		{
			name:   "http_test_url_empty_all",
			gitURL: "http://:@github.com/path/to/repository.git",
			expect: map[string]string{
				"": "",
			},
			shouldBeError: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			gitUserName, gitPassword, err := getUserNamePasswordFromGitURL(tc.gitURL)
			if err != nil && !tc.shouldBeError {
				t.Error(err)
			}
			mVal := map[string]string{
				gitUserName: gitPassword,
			}
			tu.Equals(t, tc.expect, mVal)
		})
	}
}

// go test -v -timeout 30s -run ^TestReloadCore$ gitlab.jiagouyun.com/cloudcare-tools/datakit/gitrepo
func TestReloadCore(t *testing.T) {
	const successRound = 5

	cases := []struct {
		name          string
		timeout       time.Duration
		shouldBeError bool
		expect        map[string]int
	}{
		{
			name:          "pass",
			timeout:       60 * time.Second,
			shouldBeError: false,
			expect: map[string]int{
				"round": successRound + 1,
			},
		},

		{
			name:          "timeout",
			timeout:       time.Millisecond,
			shouldBeError: true,
			expect: map[string]int{
				"round": 1,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctxNew, cancel := context.WithTimeout(context.Background(), tc.timeout)
			defer func() {
				cancel()
			}()
			round, err := reloadCore(ctxNew)
			if err != nil && !tc.shouldBeError {
				t.Error(err)
			}
			mVal := map[string]int{
				"round": round,
			}
			tu.Equals(t, tc.expect, mVal)
		})
	}
}

// go test -v -timeout 30s -run ^TestGetAuthMethod$ gitlab.jiagouyun.com/cloudcare-tools/datakit/gitrepo
func TestGetAuthMethod(t *testing.T) {
	cases := []struct {
		name          string
		gitUserName   string
		gitPassword   string
		c             *config.GitRepository
		expect        transport.AuthMethod
		shouldBeError bool
	}{
		{
			name:        "auth_username_password",
			gitUserName: "user",
			gitPassword: "pass",
			c:           &config.GitRepository{},
			expect: &http.BasicAuth{
				Username: "user",
				Password: "pass",
			},
			shouldBeError: false,
		},

		{
			name:          "auth_empty_ssh_path",
			c:             &config.GitRepository{},
			expect:        nil,
			shouldBeError: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			authM, err := getAuthMethod(tc.gitUserName, tc.gitPassword, tc.c)
			if err != nil && !tc.shouldBeError {
				t.Error(err)
			}
			tu.Equals(t, tc.expect, authM)
		})
	}
}
