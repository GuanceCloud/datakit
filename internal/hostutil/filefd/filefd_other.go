//go:build !linux
// +build !linux

package filefd

func collect() (map[string]int64, error) {
	info := make(map[string]int64)
	info["allocated"] = -1
	info["maximum"] = -1
	return info, nil
}
