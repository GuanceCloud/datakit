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

func initPatternsFile() error {
	dir := filepath.Join(datakit.InstallDir, PatternDir)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return err
	}

	for name, contents := range GlobalPatterns {
		content := ""
		fName := filepath.Join(dir, name)
		for _, rule := range contents {
			content += strings.Join(rule, " ")
			content += "\n"
		}

		if err := ioutil.WriteFile(fName, []byte(content)); err != nil {
			return err
		}
	}
}
