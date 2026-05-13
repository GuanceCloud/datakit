// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/opencontainers/runtime-spec/specs-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dkRuncUtils "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/apminject/dkrunc/utils"
	injUtils "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/apminject/utils"
)

type testInjectDirs struct {
	dirInject    string
	dirSubInject string
	dirSubLib    string
	dirSubLog    string
}

func setTestInjectDirs(t *testing.T) testInjectDirs {
	t.Helper()

	oldDirInject := dkRuncUtils.DirInject
	oldDirSubInject := dkRuncUtils.DirInjectSubInject
	oldDirSubLib := dkRuncUtils.DirInjectSubLib
	oldDirSubLog := dkRuncUtils.DirInjectSubLog

	apmDir := filepath.Join(t.TempDir(), "apm_inject")
	dirs := testInjectDirs{
		dirInject:    apmDir,
		dirSubInject: filepath.Join(apmDir, "inject"),
		dirSubLib:    filepath.Join(apmDir, "lib"),
		dirSubLog:    filepath.Join(apmDir, "log"),
	}
	dkRuncUtils.DirInject = dirs.dirInject
	dkRuncUtils.DirInjectSubInject = dirs.dirSubInject
	dkRuncUtils.DirInjectSubLib = dirs.dirSubLib
	dkRuncUtils.DirInjectSubLog = dirs.dirSubLog
	t.Cleanup(func() {
		dkRuncUtils.DirInject = oldDirInject
		dkRuncUtils.DirInjectSubInject = oldDirSubInject
		dkRuncUtils.DirInjectSubLib = oldDirSubLib
		dkRuncUtils.DirInjectSubLog = oldDirSubLog
	})

	require.NoError(t, os.MkdirAll(dirs.dirSubInject, 0o755))
	require.NoError(t, os.MkdirAll(filepath.Join(dirs.dirSubLib, "java"), 0o755))
	require.NoError(t, os.MkdirAll(dirs.dirSubLog, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dirs.dirSubLib, "java", "dd-java-agent.jar"), []byte("jar"), 0o644))
	return dirs
}

func prepareJavaSpec(t *testing.T) (*Spec, testInjectDirs) {
	t.Helper()

	root := t.TempDir()
	binDir := filepath.Join(root, "bin")
	require.NoError(t, os.MkdirAll(binDir, 0o755))

	javaPath := filepath.Join(binDir, "java")
	require.NoError(t, os.WriteFile(javaPath, []byte("#!/bin/sh\necho 'openjdk version \"17.0.12\"' >&2\n"), 0o755))

	dirs := setTestInjectDirs(t)

	spec := &Spec{
		Root: &specs.Root{Path: root},
		Process: &specs.Process{
			Args: []string{"/bin/java", "-jar", "app.jar"},
			Env:  []string{"PATH=/bin"},
		},
	}

	return spec, dirs
}

func TestTryModProcSpecAcceptsAbsoluteJavaPath(t *testing.T) {
	spec, dirs := prepareJavaSpec(t)

	got, err := tryModProcSpec(spec, spec.Root.Path)
	require.NoError(t, err)

	assert.Contains(t, got.Process.Env,
		"JAVA_TOOL_OPTIONS=-javaagent:"+filepath.Join(dirs.dirSubLib, "java", "dd-java-agent.jar"))
}

func TestTryModProcSpecDetectsLastJavaAgentArg(t *testing.T) {
	spec, _ := prepareJavaSpec(t)
	spec.Process.Args = []string{"java", "-javaagent:/opt/dd-java-agent.jar"}

	_, err := tryModProcSpec(spec, spec.Root.Path)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "already injected")
}

func TestTryInjectCurrentSpecWithoutUDSStillInjects(t *testing.T) {
	tmpDir := t.TempDir()
	bundle := filepath.Join(tmpDir, "bundle")
	require.NoError(t, os.MkdirAll(filepath.Join(bundle, "rootfs"), 0o755))

	dirs := setTestInjectDirs(t)

	spec := &Spec{
		Root: &specs.Root{Path: "rootfs"},
		Process: &specs.Process{
			Args: []string{"/bin/true"},
			Env:  []string{"PATH=/bin"},
		},
	}
	configPath := filepath.Join(bundle, "config.json")
	data, err := json.Marshal(spec)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(configPath, data, 0o644))

	oldArgs := os.Args
	os.Args = []string{"dkrunc", "create", "--bundle", bundle, "container-id"}
	t.Cleanup(func() {
		os.Args = oldArgs
	})

	eventRec := &EventRec{}
	tryInjectCurrentSpec(eventRec, &injUtils.AgentAddr{
		DkHost: injUtils.DefaultDKHost,
		DkPort: injUtils.DefaultDKPort,
	})
	require.Empty(t, eventRec.Errors)

	got, err := loadSpec(configPath)
	require.NoError(t, err)
	assert.Contains(t, got.Process.Env, "LD_PRELOAD="+filepath.Join(dirs.dirSubInject, "apm_launcher.so"))
	assert.NotContains(t, strings.Join(got.Process.Env, "\n"), injUtils.EnvDKSocketAddr+"=")
	assert.Len(t, got.Process.Env, 2)
	assert.Len(t, got.Mounts, 3)
}
