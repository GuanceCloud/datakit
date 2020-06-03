//+build windows

package hostobject

import (
	"runtime"

	"golang.org/x/sys/windows/registry"
)

func getOSInfo() *osInfo {
	oi := &osInfo{
		OSType: runtime.GOOS,
		Arch:   runtime.GOARCH,
	}

	key, err := registry.OpenKey(registry.LOCAL_MACHINE, "SOFTWARE\\Microsoft\\Windows NT\\CurrentVersion", registry.QUERY_VALUE|registry.ENUMERATE_SUB_KEYS)
	if err != nil {
		return oi
	}
	defer key.Close()

	s, _, _ := key.GetStringValue("ProductName")
	oi.Release = s

	return oi
}
