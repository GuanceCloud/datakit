// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build (linux && amd64 && test_docker) || (linux && arm64 && test_docker)
// +build linux,amd64,test_docker linux,arm64,test_docker

package installer

import (
	"encoding/json"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSetUnset(t *testing.T) {
	runcPath := "/tmp/dk_ctr_test_runc_" + strconv.FormatInt(time.Now().UnixNano(), 36)
	f, err := os.OpenFile(runcPath, os.O_CREATE, 0o755)
	if err != nil {
		t.Fatal(err)
	}
	f.Close()

	const (
		Set   = 1
		Unset = 2
	)
	cases := []struct {
		path   string
		cfg    map[string]any
		newCfg map[string]any
		kind   int // 0, 1, 2
		failed bool
	}{
		{
			kind: Set,
			path: "/tmp/" + "dk_ctr_test_cfg_" + strconv.FormatInt(time.Now().UnixNano(), 36),
			cfg: map[string]any{
				"default-runtime": "runc",
				"runtimes": map[string]any{
					"gvisor": map[string]any{
						"runtimeType": "io.containerd.runsc.v1",
						"options": map[string]any{
							"TypeUrl":    "io.containerd.runsc.v1.options",
							"ConfigPath": "/etc/containerd/runsc.toml",
						},
					},
				},
			},
			newCfg: map[string]any{
				"default-runtime": RuntimeDkRunc,
				"runtimes": map[string]any{
					"gvisor": map[string]any{
						"runtimeType": "io.containerd.runsc.v1",
						"options": map[string]any{
							"TypeUrl":    "io.containerd.runsc.v1.options",
							"ConfigPath": "/etc/containerd/runsc.toml",
						},
					},
					RuntimeDkRunc: map[string]any{
						"path": runcPath,
					},
				},
			},
		},
		{
			kind: Set,
			path: "/tmp/" + "dk_ctr_test_cfg_" + strconv.FormatInt(time.Now().UnixNano(), 36),
			cfg: map[string]any{
				"runtimes": map[string]any{
					"gvisor": map[string]any{
						"runtimeType": "io.containerd.runsc.v1",
						"options": map[string]any{
							"TypeUrl":    "io.containerd.runsc.v1.options",
							"ConfigPath": "/etc/containerd/runsc.toml",
						},
					},
				},
			},
			newCfg: map[string]any{
				"default-runtime": RuntimeDkRunc,
				"runtimes": map[string]any{
					"gvisor": map[string]any{
						"runtimeType": "io.containerd.runsc.v1",
						"options": map[string]any{
							"TypeUrl":    "io.containerd.runsc.v1.options",
							"ConfigPath": "/etc/containerd/runsc.toml",
						},
					},
					RuntimeDkRunc: map[string]any{
						"path": runcPath,
					},
				},
			},
		},
		{
			kind:   Unset,
			path:   "/tmp/" + "dk_ctr_test_cfg_" + strconv.FormatInt(time.Now().UnixNano(), 36),
			failed: true,
			cfg: map[string]any{
				"default-runtime": RuntimeDkRunc,
				"runtimes": map[string]any{
					"gvisor": map[string]any{
						"runtimeType": "io.containerd.runsc.v1",
						"options": map[string]any{
							"TypeUrl":    "io.containerd.runsc.v1.options",
							"ConfigPath": "/etc/containerd/runsc.toml",
						},
					},
				},
			},
			newCfg: map[string]any{
				"runtimes": map[string]any{
					"gvisor": map[string]any{
						"runtimeType": "io.containerd.runsc.v1",
						"options": map[string]any{
							"TypeUrl":    "io.containerd.runsc.v1.options",
							"ConfigPath": "/etc/containerd/runsc.toml",
						},
					},
				},
			},
		},
		{
			kind:   Unset,
			path:   "/tmp/" + "dk_ctr_test_cfg_" + strconv.FormatInt(time.Now().UnixNano(), 36),
			failed: true,
			cfg: map[string]any{
				"default-runtime": RuntimeDkRunc,
				"runtimes": map[string]any{
					"gvisor": map[string]any{
						"runtimeType": "io.containerd.runsc.v1",
						"options": map[string]any{
							"TypeUrl":    "io.containerd.runsc.v1.options",
							"ConfigPath": "/etc/containerd/runsc.toml",
						},
					},
				},
			},
			newCfg: map[string]any{
				"runtimes": map[string]any{
					"gvisor": map[string]any{
						"runtimeType": "io.containerd.runsc.v1",
						"options": map[string]any{
							"TypeUrl":    "io.containerd.runsc.v1.options",
							"ConfigPath": "/etc/containerd/runsc.toml",
						},
					},
				},
			},
		},
	}
	for _, c := range cases {
		t.Run("", func(t *testing.T) {
			f, err := os.OpenFile(c.path, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0o755)
			if err != nil {
				t.Fatal(err)
			}
			if err := json.NewEncoder(f).Encode(c.cfg); err != nil {
				t.Fatal(err)
			}
			f.Close()
			switch c.kind {
			case Set:
				err := setDockerRunc(c.path, runcPath)
				if c.failed {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}
				val, err := loadDockerDaemonConfig(c.path)
				assert.NoError(t, err)
				assert.Equal(t, c.newCfg, val)
			case Unset:
				err := unsetDockerRunc(c.path)
				assert.NoError(t, err)
				val, err := loadDockerDaemonConfig(c.path)
				assert.NoError(t, err)
				assert.Equal(t, c.newCfg, val)
			}
		})
	}
}
