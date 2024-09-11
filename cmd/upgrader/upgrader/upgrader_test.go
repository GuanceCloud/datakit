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

		router.GET("/datakit/:file", func(c *gin.Context) {
			switch c.Request.URL.Path {
			case "/datakit/install.sh", "/datakit/install.ps1":
				c.Data(200, "", []byte("echo abc"))

			case "/datakit/version": // online version info
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

		specifiedVersion := "4.5.6"

		router.GET("/datakit/:file", func(c *gin.Context) {
			switch c.Request.URL.Path {
			case fmt.Sprintf("/datakit/install-%s.sh", specifiedVersion),
				fmt.Sprintf("/datakit/install-%s.ps1", specifiedVersion):
				c.Data(200, "", []byte("echo abc"))

			case "/datakit/version": // online version info
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

		upToDateVersion := "1.2.3"
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

		router.GET("/datakit/:file", func(c *gin.Context) {
			switch c.Request.URL.Path {
			case "/datakit/version": // online version info
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

		upToDateVersion := "1.2.3"
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

		router.GET("/datakit/:file", func(c *gin.Context) {
			switch c.Request.URL.Path {
			case "/datakit/version": // online version info
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

		upToDateVersion := "1.2.3"
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
		ossRouter.GET("/datakit/:file", func(c *gin.Context) {
			switch c.Request.URL.Path {
			case "/datakit/install.sh", "/datakit/install.ps1":
				c.Data(200, "", []byte("echo abc"))
			case "/datakit/version": // online version info
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
