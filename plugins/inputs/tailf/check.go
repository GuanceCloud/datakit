package tailf

import (
	"errors"
	"net/http"
	"os"
	"path/filepath"
	// "gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
)

var (
	errIsDir = errors.New("is dir")

	errNotOnWhiteList = errors.New("not on the white list")

	typeWhiteList = map[string]byte{
		"text/plain; charset=utf-8": 0,
		"application/octet-stream":  0,
	}
)

func filterPath(paths []string) []string {
	var passList []string
	var list = fileList(paths)

	for _, f := range list {
		if checkFile(f) {
			passList = append(passList, f)
		}
	}
	return passList
}

func fileList(paths []string) []string {
	var list []string
	for _, path := range paths {
		// if path == config.Cfg.MainCfg.Log {
		// 	continue
		// }

		filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			absPath, err := filepath.Abs(p)
			if err != nil {
				return err
			}
			list = append(list, absPath)
			return nil
		})
	}
	return list
}

func checkFile(filename string) bool {
	info, err := os.Stat(filename)
	if err != nil {
		return false
	}

	if info.IsDir() {
		return false
	}

	f, err := os.Open(filename)
	defer f.Close()
	if err != nil {
		return false
	}

	buffer := make([]byte, 512)

	_, err = f.Read(buffer)
	if err != nil {
		return false
	}

	contentType := http.DetectContentType(buffer)
	if _, ok := typeWhiteList[contentType]; !ok {
		return false
	}

	return true
}
