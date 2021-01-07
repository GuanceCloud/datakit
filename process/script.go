package process

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
)

const (
	scriptDir          = "script"
	nginxScriptName    = "nginx.script"
	nginxScriptContent = ``
)

var (
	scriptsMap = map[string]string {
		nginxScriptName  : nginxScriptContent,
	}
)

func GenScript() {
	dir := filepath.Join(datakit.InstallDir, scriptDir)
	if err := CreateDirIfNeed(dir); err != nil {
		l.Errorf("create script dir: %v", err)
		return
	}

	for name, content := range scriptsMap {
		sName := filepath.Join(dir, name)
		if err := CreateScriptIfNeed(sName, content); err != nil {
			l.Errorf("create script %v: %v", name, err)
		}
	}
}

func CreateScriptIfNeed(name, content string) error {
	if !isDirOrFileExit(name) {
		return ioutil.WriteFile(name, []byte(content), 0666)
	}
	return nil

}

func CreateDirIfNeed(dir string) error {
	if !isDirOrFileExit(dir) {
		return os.MkdirAll(dir, 0666)
	}
	return nil
}

func isDirOrFileExit(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return true
}