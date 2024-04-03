//go:build !linux
// +build !linux

// Package dumpstd dump stderr to file
package dumpstd

func DumpStderr2File(dir string) error {
	return nil
}
