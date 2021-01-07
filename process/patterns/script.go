package patterns

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
)

const (
	PatternDir = "pattern"
)

func MkPatternsFile() {
	dir := filepath.Join(datakit.InstallDir, PatternDir)
	if err := CreateDirIfNotExist(dir); err != nil {
		return
	}

	for name, contents := range GlobalPatterns {
		content := ""
		fName := filepath.Join(dir, name)
		for _, rule := range contents {
			content += strings.Join(rule, " ")
			content += "\n"
		}
		CreatePatternFile(fName, content)
	}
}

func CreatePatternFile(name, content string) error {
	ioutil.WriteFile(name, []byte(content), 0666)
	return nil
}

func CreateDirIfNotExist(dir string) error {
	_, err := os.Stat(dir)
	if err == nil {
		return nil
	}
	if os.IsNotExist(err) {
		return os.MkdirAll(dir, 0666)
	}
	return nil
}