package file_collector

import (
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

func addFileInfo(name, MD5 string) {
	mtx.Lock()
	defer mtx.Unlock()
	fileInfoMap[name] = MD5
}

func getFileMd5(filename string) (string, error) {
	f, err := os.Open(filename)
	defer f.Close()
	if err != nil {
		return "", err
	}
	fileMd5 := md5.New()
	if _, err := io.Copy(fileMd5, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(fileMd5.Sum(nil)), nil
}
