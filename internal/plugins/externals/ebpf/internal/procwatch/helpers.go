//go:build linux
// +build linux

package procwatch

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/GuanceCloud/cliutils/logger"
	"github.com/dgraph-io/ristretto"
)

var log = logger.DefaultSLogger("ebpf")

func SetLogger(nl *logger.Logger) {
	log = nl
}

func deltaSet(oldSet, newSet map[string]struct{}) (removed map[string]struct{}, added map[string]struct{}) {
	added = make(map[string]struct{})
	removed = make(map[string]struct{})

	for item := range newSet {
		if _, ok := oldSet[item]; !ok {
			added[item] = struct{}{}
		}
	}

	for item := range oldSet {
		if _, ok := newSet[item]; !ok {
			removed[item] = struct{}{}
		}
	}

	return removed, added
}

func shortID(parts ...string) string {
	var buf bytes.Buffer
	for _, part := range parts {
		buf.WriteString(part)
	}

	sum := sha256.Sum256(buf.Bytes())
	return strconv.FormatUint(binary.BigEndian.Uint64(sum[:]), 36)
}

func resolveHostBinaryPath(pid int, procPath string) string {
	if pid <= 0 || procPath == "" {
		return ""
	}
	resolved := sharedResolverCache.Resolve(HostProc(strconv.Itoa(pid)), procPath)
	if candidate := normalizeProcfsBinaryPath(resolved); candidate != "" && isRegularFile(candidate) {
		return candidate
	}
	if candidate := resolveViaProcRootPortal(HostProc(strconv.Itoa(pid), "root"), procPath); candidate != "" {
		return candidate
	}
	if rootPath, err := readProcessRootPath(pid); err == nil && rootPath != "" {
		candidate := normalizeProcfsBinaryPath(filepath.Join(rootPath, procPath))
		if candidate != "" && isRegularFile(candidate) {
			return candidate
		}
	}
	return sharedResolverCache.Resolve(HostProc(strconv.Itoa(pid)), procPath)
}

func resolveHostBinaryPathFromProcFD(dirfd int, pid int, procPath string) string {
	if pid <= 0 || procPath == "" {
		return ""
	}
	resolved := sharedResolverCache.Resolve(HostProc(strconv.Itoa(pid)), procPath)
	if candidate := normalizeProcfsBinaryPath(resolved); candidate != "" && isRegularFile(candidate) {
		return candidate
	}
	if dirfd >= 0 {
		procRootPortal := filepath.Join("/proc/self/fd", strconv.Itoa(dirfd), "root")
		if candidate := resolveViaProcRootPortal(procRootPortal, procPath); candidate != "" {
			return candidate
		}
		if rootPath, err := readProcessRootPathFromProcFD(dirfd, pid); err == nil && rootPath != "" {
			candidate := normalizeProcfsBinaryPath(filepath.Join(rootPath, procPath))
			if candidate != "" && isRegularFile(candidate) {
				return candidate
			}
		}
	}
	return resolveHostBinaryPath(pid, procPath)
}

func resolveViaProcRootPortal(procRoot string, procPath string) string {
	if procRoot == "" || procPath == "" {
		return ""
	}
	trimmedPath := strings.TrimPrefix(procPath, string(filepath.Separator))
	candidate := normalizeProcfsBinaryPath(filepath.Join(procRoot, trimmedPath))
	if candidate != "" && isRegularFile(candidate) {
		if resolved, err := filepath.EvalSymlinks(candidate); err == nil {
			if resolved = normalizeProcfsBinaryPath(resolved); resolved != "" {
				return resolved
			}
		}
		return candidate
	}
	return ""
}

func envOrDefault(key string, fallback string, join ...string) string {
	value := os.Getenv(key)
	if value == "" {
		value = fallback
	}

	if len(join) == 0 {
		return value
	}

	parts := make([]string, 0, len(join)+1)
	parts = append(parts, value)
	parts = append(parts, join...)
	return filepath.Join(parts...)
}

func HostProc(parts ...string) string {
	return envOrDefault("HOST_PROC", "/proc", parts...)
}

func HostRoot(parts ...string) string {
	return envOrDefault("HOST_ROOT", "/", parts...)
}

func newCache(maxCost int64, counters int64) *ristretto.Cache {
	cache, err := ristretto.NewCache(&ristretto.Config{
		MaxCost:     maxCost,
		NumCounters: counters,
		BufferItems: 64,
		Metrics:     true,
	})
	if err != nil {
		panic(fmt.Errorf("create ristretto cache: %w", err))
	}

	return cache
}
