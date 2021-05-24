package path

import (
	"errors"
	"os"
)

var (
	ErrInvalidPath = errors.New("provided path invalid")
)

func IsFileExists(path string) bool {
	pathInfo, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	if mode := pathInfo.Mode(); !mode.IsRegular() {
		return false
	}

	return true
}
