// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build (linux && amd64 && test_docker) || (linux && arm64 && test_docker)
// +build linux,amd64,test_docker linux,arm64,test_docker

package utils

import (
	"debug/elf"
	"encoding/json"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var (
	musl125 = `musl libc (x86_64)
Version 1.2.5
Dynamic Program Loader
Usage: /lib/ld-musl-x86_64.so.1 [options] [--] pathname`

	glibc235 = `ldd (Ubuntu GLIBC 2.35-0ubuntu3.8) 2.35
Copyright (C) 2022 Free Software Foundation, Inc.
This is free software; see the source for copying conditions.  There is NO
warranty; not even for MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.
Written by Roland McGrath and Ulrich Drepper.`

	glic212 = `ldd (GNU libc) 2.12
Copyright (C) 2010 Free Software Foundation, Inc.
This is free software; see the source for copying conditions.  There is NO
warranty; not even for MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.
Written by Roland McGrath and Ulrich Drepper.`

	glibc236 = `ldd (Debian GLIBC 2.36-9+deb12u8) 2.36
Copyright (C) 2022 Free Software Foundation, Inc.
This is free software; see the source for copying conditions.  There is NO
warranty; not even for MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.
Written by Roland McGrath and Ulrich Drepper.`

	daemonJSON = `{
    "default-runtime": "runc",
    "runtimes": {
        "gvisor": {
            "runtimeType": "io.containerd.runsc.v1",
            "options": {
                "TypeUrl": "io.containerd.runsc.v1.options",
                "ConfigPath": "/etc/containerd/runsc.toml"
            }
        },
        "dkrunc": {
            "path": "/path/to/dkrunc"
        }
    }
}`
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
			case 1:
				err := setDockerRunc(c.path, runcPath)
				if c.failed {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}
				val, err := loadDockerDaemonConfig(c.path)
				assert.NoError(t, err)
				assert.Equal(t, c.newCfg, val)
			case 2:
				err := unsetDockerRunc(c.path)
				assert.NoError(t, err)
				val, err := loadDockerDaemonConfig(c.path)
				assert.NoError(t, err)
				assert.Equal(t, c.newCfg, val)
			}
		})
	}
}

func TestUtils(t *testing.T) {
	v1, v2, ok := libcInfo(glibc235)
	assert.Equal(t, v1, glibc)
	assert.Equal(t, v2, "2.35")
	assert.True(t, ok)

	v1, v2, ok = libcInfo(glic212)
	assert.Equal(t, v1, glibc)
	assert.Equal(t, v2, "2.12")
	assert.True(t, ok)

	v1, v2, ok = libcInfo(glibc236)
	assert.Equal(t, v1, glibc)
	assert.Equal(t, v2, "2.36")
	assert.True(t, ok)

	v1, v2, ok = libcInfo(musl125)
	assert.Equal(t, v1, muslc)
	assert.Equal(t, v2, "1.2.5")
	assert.True(t, ok)
}

func TestLddInfo(t *testing.T) {
	syms := []elf.Symbol{
		{Library: "libc.so.6", Version: "GLIBC_2.35"},
		{Library: "ld-linux-x86-64.so.2", Version: "GLIBC_2.35"},
	}
	v, err := requiredGLIBCVersion(syms)
	assert.NoError(t, err)
	assert.Equal(t, Version{2, 35, 0}, v)
}
