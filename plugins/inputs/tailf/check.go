package tailf

import (
	"net/http"
	"os"
	"path/filepath"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
)

var typeWhiteList = map[string]byte{
	"text/plain; charset=utf-8": 0,
	// "application/octet-stream":  0,
}

func filterPath(paths []string) (list []string) {
	var fileList = getFileList(paths)

	// if errror not nil, absLog == ""
	absLog, _ := filepath.Abs(config.Cfg.MainCfg.Log)

	for _, f := range fileList {
		if f == absLog {
			continue
		}
		if whiteFile(f) {
			list = append(list, f)
		}
	}

	return
}

// getFileList Traverse all files in the paths
func getFileList(paths []string) (list []string) {

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

	return
}

func whiteFile(fn string) bool {
	if !isNotDirectory(fn) {
		return false
	}

	contentType, err := getFileContentType(fn)
	if err != nil {
		return false
	}

	if _, ok := typeWhiteList[contentType]; !ok {
		return false
	}

	return true
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

	buffer := make([]byte, 512)

	_, err = f.Read(buffer)
	if err != nil {
		return "", err
	}

	return http.DetectContentType(buffer), nil
}
