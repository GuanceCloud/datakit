// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package flameshot

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseProcCgroupV2Path(t *testing.T) {
	content := "0::/kubepods.slice/pod123/container456\n"
	path, err := parseProcCgroupV2Path(content)
	assert.NoError(t, err)
	assert.Equal(t, "/kubepods.slice/pod123/container456", path)
}

func TestParseProcCgroupV1Path(t *testing.T) {
	content := "5:memory:/kubepods/besteffort/pod123/container456\n"
	path, err := parseProcCgroupV1Path(content)
	assert.NoError(t, err)
	assert.Equal(t, "/kubepods/besteffort/pod123/container456", path)
}

func TestResolveCgroupV2Dir(t *testing.T) {
	procRoot := t.TempDir()
	cgroupRoot := t.TempDir()
	pidDir := filepath.Join(procRoot, "1234")
	assert.NoError(t, os.MkdirAll(pidDir, 0o755))
	assert.NoError(t, os.WriteFile(filepath.Join(pidDir, "cgroup"), []byte("0::/kubepods.slice/pod123/container456\n"), 0o644))
	assert.NoError(t, os.MkdirAll(filepath.Join(cgroupRoot, "kubepods.slice/pod123/container456"), 0o755))
	assert.NoError(t, os.WriteFile(filepath.Join(cgroupRoot, "kubepods.slice/pod123/container456", "memory.current"), []byte("1\n"), 0o644))
	assert.NoError(t, os.WriteFile(filepath.Join(cgroupRoot, "kubepods.slice/pod123/container456", "memory.max"), []byte("2\n"), 0o644))
	assert.NoError(t, os.WriteFile(filepath.Join(cgroupRoot, "kubepods.slice/pod123/container456", "memory.events"), []byte("oom_kill 0\n"), 0o644))

	dir, err := resolveCgroupV2Dir(procRoot, cgroupRoot, 1234)
	assert.NoError(t, err)
	assert.Equal(t, filepath.Join(cgroupRoot, "kubepods.slice/pod123/container456"), dir)
}

func TestResolveCgroupV1Dir(t *testing.T) {
	procRoot := t.TempDir()
	cgroupRoot := t.TempDir()
	pidDir := filepath.Join(procRoot, "1234")
	assert.NoError(t, os.MkdirAll(pidDir, 0o755))
	assert.NoError(t, os.WriteFile(filepath.Join(pidDir, "cgroup"), []byte("5:memory:/kubepods/besteffort/pod123/container456\n"), 0o644))
	assert.NoError(t, os.MkdirAll(filepath.Join(cgroupRoot, "memory/kubepods/besteffort/pod123/container456"), 0o755))
	assert.NoError(t, os.WriteFile(filepath.Join(cgroupRoot, "memory/kubepods/besteffort/pod123/container456", "memory.usage_in_bytes"), []byte("1\n"), 0o644))
	assert.NoError(t, os.WriteFile(filepath.Join(cgroupRoot, "memory/kubepods/besteffort/pod123/container456", "memory.limit_in_bytes"), []byte("2\n"), 0o644))
	assert.NoError(t, os.WriteFile(filepath.Join(cgroupRoot, "memory/kubepods/besteffort/pod123/container456", "memory.oom_control"), []byte("oom_kill 0\n"), 0o644))

	dir, err := resolveCgroupV1Dir(procRoot, cgroupRoot, 1234)
	assert.NoError(t, err)
	assert.Equal(t, filepath.Join(cgroupRoot, "memory/kubepods/besteffort/pod123/container456"), dir)
}

func TestResolveCgroupV2DirFallbackToProcessRoot(t *testing.T) {
	procRoot := t.TempDir()
	cgroupRoot := t.TempDir()
	pidDir := filepath.Join(procRoot, "1234")
	processCgroupRoot := filepath.Join(pidDir, "root", strings.TrimPrefix(cgroupRoot, "/"))
	assert.NoError(t, os.MkdirAll(pidDir, 0o755))
	assert.NoError(t, os.WriteFile(filepath.Join(pidDir, "cgroup"), []byte("0::/../container-id\n"), 0o644))
	assert.NoError(t, os.MkdirAll(processCgroupRoot, 0o755))
	assert.NoError(t, os.WriteFile(filepath.Join(processCgroupRoot, "memory.current"), []byte("1\n"), 0o644))
	assert.NoError(t, os.WriteFile(filepath.Join(processCgroupRoot, "memory.max"), []byte("2\n"), 0o644))
	assert.NoError(t, os.WriteFile(filepath.Join(processCgroupRoot, "memory.events"), []byte("oom_kill 0\n"), 0o644))

	dir, err := resolveCgroupV2Dir(procRoot, cgroupRoot, 1234)
	assert.NoError(t, err)
	assert.Equal(t, processCgroupRoot, dir)
}

func TestReadCgroupV2MemoryStats(t *testing.T) {
	dir := t.TempDir()
	assert.NoError(t, os.WriteFile(filepath.Join(dir, "memory.current"), []byte("1048576\n"), 0o644))
	assert.NoError(t, os.WriteFile(filepath.Join(dir, "memory.max"), []byte("2097152\n"), 0o644))
	assert.NoError(t, os.WriteFile(filepath.Join(dir, "memory.events"), []byte("low 0\nhigh 1\nmax 2\noom 3\noom_kill 4\n"), 0o644))

	stats, err := readCgroupV2MemoryStats(dir)
	assert.NoError(t, err)
	assert.Equal(t, uint64(1048576), stats.Current)
	assert.Equal(t, uint64(2097152), stats.Max)
	assert.Equal(t, uint64(4), stats.OOMKill)
}

func TestReadCgroupV1MemoryStats(t *testing.T) {
	dir := t.TempDir()
	assert.NoError(t, os.WriteFile(filepath.Join(dir, "memory.usage_in_bytes"), []byte("1048576\n"), 0o644))
	assert.NoError(t, os.WriteFile(filepath.Join(dir, "memory.limit_in_bytes"), []byte("2097152\n"), 0o644))
	assert.NoError(t, os.WriteFile(filepath.Join(dir, "memory.oom_control"), []byte("oom_kill 7\nunder_oom 0\n"), 0o644))

	stats, err := readCgroupV1MemoryStats(dir)
	assert.NoError(t, err)
	assert.Equal(t, uint64(1048576), stats.Current)
	assert.Equal(t, uint64(2097152), stats.Max)
	assert.Equal(t, uint64(7), stats.OOMKill)
}

func TestResolveCgroupWatcherTargetPreferV2(t *testing.T) {
	procRoot := t.TempDir()
	cgroupRoot := t.TempDir()
	pidDir := filepath.Join(procRoot, "1234")
	assert.NoError(t, os.MkdirAll(pidDir, 0o755))
	assert.NoError(t, os.WriteFile(filepath.Join(pidDir, "cgroup"), []byte("0::/kubepods.slice/pod123/container456\n5:memory:/kubepods/besteffort/pod123/container456\n"), 0o644))
	assert.NoError(t, os.MkdirAll(filepath.Join(cgroupRoot, "kubepods.slice/pod123/container456"), 0o755))
	assert.NoError(t, os.WriteFile(filepath.Join(cgroupRoot, "kubepods.slice/pod123/container456", "memory.current"), []byte("1\n"), 0o644))
	assert.NoError(t, os.WriteFile(filepath.Join(cgroupRoot, "kubepods.slice/pod123/container456", "memory.max"), []byte("2\n"), 0o644))
	assert.NoError(t, os.WriteFile(filepath.Join(cgroupRoot, "kubepods.slice/pod123/container456", "memory.events"), []byte("oom_kill 0\n"), 0o644))

	dir, version, err := resolveCgroupWatcherTarget(procRoot, cgroupRoot, 1234)
	assert.NoError(t, err)
	assert.Equal(t, cgroupVersionV2, version)
	assert.Equal(t, filepath.Join(cgroupRoot, "kubepods.slice/pod123/container456"), dir)
}
