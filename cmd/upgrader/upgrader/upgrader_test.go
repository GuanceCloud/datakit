// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package upgrader

import (
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	T "testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/atomic"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/httpapi"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/version"
)

func TestUpgrader(t *T.T) {
	t.Run("upgrade-to-latest-version", func(t *T.T) {
		router := gin.New()
		router.GET("/v1/ping", func(c *gin.Context) {
			pi := pingInfo{
				Content: httpapi.Ping{
					Version: "1.2.3",
					Commit:  "fake-commit-id",
				},
			}

			j, err := json.Marshal(pi)
			assert.NoError(t, err)

			c.Data(200, "application/json", j)
		})

		router.GET("/datakit-v2/:file", func(c *gin.Context) {
			switch c.Request.URL.Path {
			case "/datakit-v2/install.sh", "/datakit-v2/install.ps1":
				c.Data(200, "", []byte("echo abc"))

			case "/datakit-v2/version": // online version info
				vi := version.VerInfo{
					VersionString: "1024.1024.1024",
					Commit:        "fake-online-version-commit",
				}
				j, err := json.Marshal(vi)
				assert.NoError(t, err)

				c.Data(200, "application/json", j)

			default:
				assert.Truef(t, false, "should not been here, get path: %s", c.Request.URL.Path)
			}
		})

		fakeServer := httptest.NewServer(router)
		defer fakeServer.Close()
		time.Sleep(time.Second)

		u := upgraderImpl{
			upgradeStatus: atomic.NewInt32(0),
			c: &MainConfig{
				DatakitAPIListen: fakeServer.Listener.Addr().String(),
				InstallerBaseURL: fakeServer.URL,
			},
		}

		assert.NoError(t, u.upgrade())
		assert.Equal(t, int32(0), u.upgradeStatus.Load())
	})

	t.Run("upgrade-to-specified-version", func(t *T.T) {
		router := gin.New()
		router.GET("/v1/ping", func(c *gin.Context) {
			pi := pingInfo{
				Content: httpapi.Ping{
					Version: "1.2.3",
					Commit:  "fake-commit-id",
				},
			}

			j, err := json.Marshal(pi)
			assert.NoError(t, err)

			c.Data(200, "application/json", j)
		})

		specifiedVersion := "1.5.6"

		router.GET("/datakit-v2/:file", func(c *gin.Context) {
			switch c.Request.URL.Path {
			case fmt.Sprintf("/datakit-v2/install-%s.sh", specifiedVersion),
				fmt.Sprintf("/datakit-v2/install-%s.ps1", specifiedVersion):
				c.Data(200, "", []byte("echo abc"))

			case "/datakit-v2/version": // online version info
				vi := version.VerInfo{
					VersionString: "1024.1024.1024",
					Commit:        "fake-online-version-commit",
				}
				j, err := json.Marshal(vi)
				assert.NoError(t, err)

				c.Data(200, "application/json", j)

			default:
				assert.Truef(t, false, "should not been here, get path: %s", c.Request.URL.Path)
			}
		})

		fakeServer := httptest.NewServer(router)
		defer fakeServer.Close()
		time.Sleep(time.Second)

		u := upgraderImpl{
			upgradeStatus: atomic.NewInt32(0),
			c: &MainConfig{
				DatakitAPIListen: fakeServer.Listener.Addr().String(),
				InstallerBaseURL: fakeServer.URL,
			},
		}

		assert.NoError(t, u.upgrade(withVersion(specifiedVersion)))
		assert.Equal(t, int32(0), u.upgradeStatus.Load())
	})

	t.Run("up-to-date", func(t *T.T) {
		router := gin.New()

		upToDateVersion := "2.1.0"
		router.GET("/v1/ping", func(c *gin.Context) {
			pi := pingInfo{
				Content: httpapi.Ping{
					Version: upToDateVersion,
					Commit:  "fake-commit-id",
				},
			}

			j, err := json.Marshal(pi)
			assert.NoError(t, err)

			c.Data(200, "application/json", j)
		})

		router.GET("/datakit-v2/:file", func(c *gin.Context) {
			switch c.Request.URL.Path {
			case "/datakit-v2/version": // online version info
				vi := version.VerInfo{
					VersionString: upToDateVersion,
					Commit:        "fake-commit-id",
				}
				j, err := json.Marshal(vi)
				assert.NoError(t, err)

				c.Data(200, "application/json", j)

			default:
				assert.Truef(t, false, "should not been here, get path: %s", c.Request.URL.Path)
			}
		})

		fakeServer := httptest.NewServer(router)
		defer fakeServer.Close()
		time.Sleep(time.Second)

		u := upgraderImpl{
			upgradeStatus: atomic.NewInt32(0),
			c: &MainConfig{
				DatakitAPIListen: fakeServer.Listener.Addr().String(),
				InstallerBaseURL: fakeServer.URL,
			},
		}

		err := u.upgrade()
		assert.Error(t, err)
		assert.ErrorIs(t, err, httpapi.ErrDKVersionUptoDate)
		assert.Equal(t, int32(0), u.upgradeStatus.Load())

		t.Logf("expected error: %s", err)
	})

	t.Run("still-upgrading", func(t *T.T) {
		u := upgraderImpl{
			upgradeStatus: atomic.NewInt32(1), // to to upgrading
		}

		err := u.upgrade()
		assert.Error(t, err)
		assert.ErrorIs(t, err, httpapi.ErrIsUpgrading)

		t.Logf("expected error: %s", err)
	})

	t.Run("with-dk-api-https", func(t *T.T) {
		router := gin.New()

		upToDateVersion := "2.1.0"
		router.GET("/v1/ping", func(c *gin.Context) {
			pi := pingInfo{
				Content: httpapi.Ping{
					Version: upToDateVersion,
					Commit:  "fake-commit-id",
				},
			}

			j, err := json.Marshal(pi)
			assert.NoError(t, err)

			c.Data(200, "application/json", j)
		})

		router.GET("/datakit-v2/:file", func(c *gin.Context) {
			switch c.Request.URL.Path {
			case "/datakit-v2/version": // online version info
				vi := version.VerInfo{
					VersionString: upToDateVersion,
					Commit:        "fake-commit-id",
				}
				j, err := json.Marshal(vi)
				assert.NoError(t, err)

				c.Data(200, "application/json", j)

			default:
				assert.Truef(t, false, "should not been here, get path: %s", c.Request.URL.Path)
			}
		})

		fakeServer := httptest.NewTLSServer(router)
		defer fakeServer.Close()
		time.Sleep(time.Second)

		u := upgraderImpl{
			upgradeStatus: atomic.NewInt32(0),
			c: &MainConfig{
				DatakitAPIListen: fakeServer.Listener.Addr().String(),
				DatakitAPIHTTPS:  true,
				InstallerBaseURL: fakeServer.URL,
			},
		}

		err := u.upgrade()
		assert.Error(t, err)
		assert.ErrorIs(t, err, httpapi.ErrDKVersionUptoDate)
		assert.Equal(t, int32(0), u.upgradeStatus.Load())

		t.Logf("expected error: %s", err)
	})

	t.Run("with-http-proxy", func(t *T.T) {
		// start datakit server
		dkRouter := gin.New()

		upToDateVersion := "2.1.0"
		dkRouter.GET("/v1/ping", func(c *gin.Context) {
			pi := pingInfo{
				Content: httpapi.Ping{
					Version: "0.1.2", // old version
					Commit:  "fake-commit-id",
				},
			}

			j, err := json.Marshal(pi)
			assert.NoError(t, err)

			c.Data(200, "application/json", j)
		})

		dkServer := httptest.NewServer(dkRouter)
		defer dkServer.Close()
		time.Sleep(time.Second)

		// start oss server
		ossRouter := gin.New()
		ossRouter.GET("/datakit-v2/:file", func(c *gin.Context) {
			switch c.Request.URL.Path {
			case "/datakit-v2/install.sh", "/datakit-v2/install.ps1":
				c.Data(200, "", []byte("echo abc"))
			case "/datakit-v2/version": // online version info
				vi := version.VerInfo{
					VersionString: upToDateVersion,
					Commit:        "fake-commit-id",
				}
				j, err := json.Marshal(vi)
				assert.NoError(t, err)

				c.Data(200, "application/json", j)

			default:
				assert.Truef(t, false, "should not been here, get path: %s", c.Request.URL.Path)
			}
		})

		ossServer := httptest.NewServer(ossRouter)
		defer ossServer.Close()

		t.Logf("start oss server at %s", ossServer.URL)

		// start proxy server
		ossURL, err := url.Parse(ossServer.URL)
		assert.NoError(t, err)
		proxyHandler := httputil.NewSingleHostReverseProxy(ossURL)
		proxy := httptest.NewServer(proxyHandler)
		defer proxy.Close()

		time.Sleep(time.Second) // wait 3 server ok

		u := upgraderImpl{
			upgradeStatus: atomic.NewInt32(0),
			c: &MainConfig{
				DatakitAPIListen: dkServer.Listener.Addr().String(),
				InstallerBaseURL: ossServer.URL,
				Proxy:            proxy.URL,
			},
		}

		assert.NoError(t, u.upgrade())
		assert.Equal(t, int32(0), u.upgradeStatus.Load())
	})

	t.Run("with-https-proxy", func(t *T.T) {
		t.Skip("TODO")
	})
}

func TestDoUpgradeRunsScriptInWritableInstallDir(t *T.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell script cwd assertion is for non-Windows upgrader")
	}

	installDir := t.TempDir()
	scriptFile := filepath.Join(installDir, "install.sh")
	script := `#!/bin/sh
set -eu
script_dir="$(cd "$(dirname "$0")" && pwd -P)"
pwd_dir="$(pwd -P)"
tmp_dir="$(cd "${TMPDIR:-}" && pwd -P)"
if [ "$pwd_dir" != "$script_dir" ]; then
	echo "unexpected pwd: $pwd_dir, want: $script_dir" >&2
	exit 23
fi
if [ "$tmp_dir" != "$script_dir" ]; then
	echo "unexpected TMPDIR: $tmp_dir, want: $script_dir" >&2
	exit 24
fi
`
	assert.NoError(t, os.WriteFile(scriptFile, []byte(script), 0o755))

	u := upgraderImpl{
		upgradeStatus: atomic.NewInt32(0),
		c: &MainConfig{
			InstallDir: installDir,
		},
	}

	assert.NoError(t, u.doUpgrade(scriptFile, ""))
}

func TestFormatUpgradeEnvsForLogRedactsSensitiveValues(t *T.T) {
	envs := []string{
		"PATH=/usr/bin:/bin",
		"DK_UPGRADE=1",
		"GUANCE_API_KEY_DF=secret-api-key",
		"OSS_SECRET_KEY=secret-oss-key",
		"HTTPS_PROXY=http://user:pass@127.0.0.1:7890",
		"PASSWORD=secret-password",
		"MALFORMED_ENV",
	}

	logText := formatUpgradeEnvsForLog(envs)

	assert.Contains(t, logText, "PATH=/usr/bin:/bin")
	assert.Contains(t, logText, "DK_UPGRADE=1")
	assert.Contains(t, logText, "GUANCE_API_KEY_DF=******")
	assert.Contains(t, logText, "OSS_SECRET_KEY=******")
	assert.Contains(t, logText, "HTTPS_PROXY=******")
	assert.Contains(t, logText, "PASSWORD=******")
	assert.Contains(t, logText, "MALFORMED_ENV")
	assert.NotContains(t, logText, "secret-api-key")
	assert.NotContains(t, logText, "secret-oss-key")
	assert.NotContains(t, logText, "user:pass")
	assert.NotContains(t, logText, "secret-password")
}
