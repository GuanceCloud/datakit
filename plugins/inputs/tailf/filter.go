package tailf

import (
	"net/http"
	"os"
	"path/filepath"
)

var typeWhiteList = map[string]interface{}{
	"text/plain; charset=utf-8": nil,
	// "application/octet-stream":  nil,
}

// getFileList abs filename in the paths
func getFileList(paths []string) []string {
	var list []string

	for _, path := range paths {
		_ = filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
			// ignore errors
			if err != nil {
				return nil
			}
			absPath, err := filepath.Abs(p)
			if err != nil {
				return nil
			}
			list = append(list, absPath)
			return nil
		})
	}

	return list
}

func isNotDirectory(fn string) bool {
	info, err := os.Stat(fn)
	if err != nil {
		return false
	}

	if info.IsDir() {
		return false
	}
	return true
}

func getFileContentType(fn string) (string, error) {
	f, err := os.Open(fn)
	if err != nil {
		return "", err
	}
	defer f.Close()

	buffer := make([]byte, 25)

	_, err = f.Read(buffer)
	if err != nil {
		return "", err
	}

	return http.DetectContentType(buffer), nil
}
