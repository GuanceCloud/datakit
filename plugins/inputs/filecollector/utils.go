package filecollector

import (

	// nolint:gosec
	"crypto/md5"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
)

func updateFileInfo(path string) {
	mtx.Lock()
	defer mtx.Unlock()
	fileNames, err := filepath.Glob(filepath.Join(path, "*"))
	if err != nil {
		return
	}
	for _, name := range fileNames {
		fileInfoMap[name] = ""
	}
}

func addFileInfo(name, md5str string) {
	mtx.Lock()
	defer mtx.Unlock()
	fileInfoMap[name] = md5str
}

func getFileMd5(filename string) (string, error) {
	f, err := os.Open(filepath.Clean(filename))
	if err != nil {
		return "", err
	}

	defer f.Close() //nolint:errcheck,gosec

	fileMd5 := md5.New() //nolint:gosec
	if _, err := io.Copy(fileMd5, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(fileMd5.Sum(nil)), nil
}
