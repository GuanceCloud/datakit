//go:build linux
// +build linux

package sysmonitor

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"syscall"

	ddfp "github.com/DataDog/gopsutil/process/filepath"

	"github.com/GuanceCloud/cliutils/logger"
)

var log = logger.DefaultSLogger("ebpf")

func SetLogger(nl *logger.Logger) {
	log = nl
}

type fileKey struct {
	dev uint64
	ino uint64
}

func diff(old, cur map[string]struct{}) (map[string]struct{}, map[string]struct{}) {
	add := map[string]struct{}{}
	del := map[string]struct{}{}
	for k := range cur {
		if _, ok := old[k]; !ok {
			add[k] = struct{}{}
		}
	}

	for k := range old {
		if _, ok := cur[k]; !ok {
			del[k] = struct{}{}
		}
	}

	return del, add
}

func ShortID(binPath ...string) string {
	var str bytes.Buffer
	for _, s := range binPath {
		str.WriteString(s)
	}

	sha1Val := sha256.Sum256(str.Bytes())
	return strconv.FormatUint(
		binary.BigEndian.Uint64(sha1Val[:]), 36)
}

func resolveBinPath(pid int, fpath string) string {
	resolver := ddfp.NewResolver(HostProc())
	resolver = resolver.LoadPIDMounts(HostProc(strconv.Itoa(pid)))
	return resolver.Resolve(fpath)
}

// GetEnv retrieves the environment variable key. If it does not exist it returns the default.
// Copy from vendor/github.com/shirou/gopsutil/v3/internal/common/common.go:common.GetEnv.
func GetEnv(key string, dfault string, combineWith ...string) string {
	value := os.Getenv(key)
	if value == "" {
		value = dfault
	}

	switch len(combineWith) {
	case 0:
		return value
	case 1:
		return filepath.Join(value, combineWith[0])
	default:
		all := make([]string, len(combineWith)+1)
		all[0] = value
		copy(all[1:], combineWith)
		return filepath.Join(all...)
	}
}

func FileInfo(fp string) (fileKey, error) {
	if f, err := os.Stat(fp); err == nil {
		v := f.Sys()
		if sysF, ok := v.(*syscall.Stat_t); ok {
			return fileKey{
				dev: sysF.Dev,
				ino: sysF.Ino,
			}, nil
		}
	} else {
		return fileKey{}, err
	}

	return fileKey{}, fmt.Errorf("get file info failed")
}

// HostProc returns the value of the host proc path.
// Context from vendor/github.com/shirou/gopsutil/v3/internal/common/common.go:common.HostProc.
func HostProc(combineWith ...string) string {
	return GetEnv("HOST_PROC", "/proc", combineWith...)
}

func HostRoot(combineWith ...string) string {
	return GetEnv("HOST_ROOT", "/", combineWith...)
}
