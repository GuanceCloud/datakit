// Package path wrap basic path functions.
package path

import (
	"errors"
	"os"
)

var ErrInvalidPath = errors.New("provided path invalid")

func IsFileExists(path string) bool {
	finfo, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}

	return finfo.Mode().IsRegular()
}
