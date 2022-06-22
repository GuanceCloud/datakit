//go:build !linux
// +build !linux

package tailer

//nolint
func getFileKey(file string) string {
	return file
}
