package filepathtoolkit

import (
	"errors"
	"os"
)

func FileExists(path string) bool {
	_, err := os.Stat(path)
	if err != nil && errors.Is(err, os.ErrNotExist) {
		return false
	}
	return true
}
