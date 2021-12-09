// Package path wrap basic path functions.
package path

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	giturls "github.com/whilp/git-urls"
)

var ErrInvalidPath = errors.New("provided path invalid")

func IsFileExists(path string) bool {
	finfo, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}

	return finfo.Mode().IsRegular()
}

// IsDir returns true if the given path is an existing directory.
func IsDir(path string) bool {
	if pathAbs, err := filepath.Abs(path); err == nil {
		if fileInfo, err := os.Stat(pathAbs); !os.IsNotExist(err) && fileInfo.IsDir() {
			return true
		}
	}
	return false
}

// PathIsPureFileName returns whether the path is only has filename, no path stuff
// test/AA => false
// AA => true.
func PathIsPureFileName(s string) bool {
	return filepath.Dir(s) == "." && filepath.Base(s) == s
}

// GetGitPureName returns conf like below
// ssh://git@gitlab.jiagouyun.com:40022/wanchuan853/conf.git
// http://gitlab.jiagouyun.com/wanchuan853/conf.git
func GetGitPureName(gitURL string) (string, error) {
	uRL, err := giturls.Parse(gitURL)
	if err != nil {
		return "", err
	}
	fileName := filepath.Base(uRL.EscapedPath())
	ext := filepath.Ext(fileName)
	return strings.TrimSuffix(fileName, ext), nil
}

func GetParentDirName(str string) string {
	dir := filepath.Dir(str)
	return filepath.Base(dir)
}

// GetPureNameFromExt returns ab.py to ab.
func GetPureNameFromExt(name string) string {
	ext := filepath.Ext(name)
	nm := filepath.Base(name)
	return strings.TrimSuffix(nm, ext)
}

// func GetExecutePath() (string, error) {
// 	ex, err := os.Executable()
// 	if err != nil {
// 		return "", err
// 	}
// 	return filepath.Dir(ex), nil
// }
