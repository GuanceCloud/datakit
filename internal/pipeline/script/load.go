// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package script for managing pipeline scripts
package script

import (
	"os"
	"path/filepath"

	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
)

func QueryScript(category point.Category, name string) (*PlScript, bool) {
	return whichStore(category).IndexGet(name)
}

func ScriptCount(category point.Category) int {
	return whichStore(category).Count()
}

func ReadPlScriptFromFile(fp string) (string, string, error) {
	fp = filepath.Clean(fp)
	if v, err := os.ReadFile(filepath.Clean(fp)); err == nil {
		_, sName := filepath.Split(fp)
		return sName, string(v), nil
	} else {
		return "", "", err
	}
}

func SearchPlFilePathFormDir(dirPath string) map[string]string {
	ret := map[string]string{}
	dirPath = filepath.Clean(dirPath)
	if dirEntry, err := os.ReadDir(dirPath); err != nil {
		l.Warn(err)
	} else {
		for _, v := range dirEntry {
			if v.IsDir() {
				continue
			}
			sName := v.Name()
			if filepath.Ext(sName) != ".p" {
				continue
			}
			ret[sName] = filepath.Join(dirPath, sName)
		}
	}
	return ret
}

func ReadPlScriptFromDir(dirPath string) (map[string]string, map[string]string) {
	ret := map[string]string{}
	retPath := map[string]string{}
	dirPath = filepath.Clean(dirPath)
	if dirEntry, err := os.ReadDir(dirPath); err != nil {
		l.Warn(err)
	} else {
		for _, v := range dirEntry {
			if v.IsDir() {
				continue
			}
			sName := v.Name()
			if filepath.Ext(sName) != ".p" {
				continue
			}
			sPath := filepath.Join(dirPath, sName)
			if name, script, err := ReadPlScriptFromFile(sPath); err == nil {
				ret[name] = script
				retPath[name] = sPath
			} else {
				l.Error(err)
			}
		}
	}
	return ret, retPath
}

func CategoryDirName() map[point.Category]string {
	return map[point.Category]string{
		point.Metric:       "metric",
		point.Network:      "network",
		point.KeyEvent:     "keyevent",
		point.Object:       "object",
		point.CustomObject: "custom_object",
		point.Logging:      "logging",
		point.Tracing:      "tracing",
		point.RUM:          "rum",
		point.Security:     "security",
		point.Profiling:    "profiling",
	}
}

func SearchPlFilePathFromPlStructPath(basePath string) map[point.Category](map[string]string) {
	files := map[point.Category](map[string]string){}

	files[point.Logging] = SearchPlFilePathFormDir(basePath)

	for category, dirName := range CategoryDirName() {
		s := SearchPlFilePathFormDir(filepath.Join(basePath, dirName))
		if _, ok := files[category]; !ok {
			files[category] = map[string]string{}
		}
		for k, v := range s {
			files[category][k] = v
		}
	}
	return files
}

func ReadPlScriptFromPlStructPath(basePath string) (
	map[point.Category](map[string]string), map[point.Category](map[string]string),
) {
	scripts := map[point.Category](map[string]string){}
	scriptsPath := map[point.Category](map[string]string){}

	scripts[point.Logging], scriptsPath[point.Logging] = ReadPlScriptFromDir(basePath)

	for category, dirName := range CategoryDirName() {
		s, p := ReadPlScriptFromDir(filepath.Join(basePath, dirName))
		if _, ok := scripts[category]; !ok {
			scripts[category] = map[string]string{}
		}
		if _, ok := scriptsPath[category]; !ok {
			scriptsPath[category] = map[string]string{}
		}

		for k, v := range s {
			scripts[category][k] = v
		}
		for k, v := range p {
			scriptsPath[category][k] = v
		}
	}
	return scripts, scriptsPath
}

// LoadDotPScript2Store will diff current layer data and then update.
func LoadDotPScript2Store(category point.Category, ns string, dirPath string, filePath []string) {
	whichStore(category).LoadDotPScript2Store(ns, dirPath, filePath)
}

func LoadAllDefaultScripts2Store() {
	plPath := filepath.Join(datakit.InstallDir, "pipeline")
	LoadAllScripts2StoreFromPlStructPath(DefaultScriptNS, plPath)
}

func LoadAllScripts2StoreFromPlStructPath(ns, plPath string) {
	scripts, path := ReadPlScriptFromPlStructPath(plPath)

	LoadAllScript(ns, scripts, path)
}

func LoadScript(category point.Category, ns string, scripts map[string]string, path map[string]string) {
	_ = whichStore(category).UpdateScriptsWithNS(ns, scripts, path)
}

func FillScriptCategoryMap(scripts map[point.Category](map[string]string)) map[point.Category](map[string]string) {
	allCategoryScript := map[point.Category](map[string]string){}
	for k := range _allCategory {
		allCategoryScript[k] = map[string]string{}
	}
	for k, v := range scripts {
		for name, script := range v {
			allCategoryScript[k][name] = script
		}
	}
	return allCategoryScript
}

func FillScriptCategoryMapFp(scripts map[point.Category]([]string)) map[point.Category]([]string) {
	allCategoryScript := map[point.Category]([]string){}
	for k := range _allCategory {
		allCategoryScript[k] = []string{}
	}
	for k, v := range scripts {
		allCategoryScript[k] = append(allCategoryScript[k], v...)
	}
	return allCategoryScript
}

// LoadAllScript is used to load and clean the script, parameter scripts example: {datakit.Logging: {ScriptName: ScriptContent},... }.
func LoadAllScript(ns string, scripts, scriptPath map[point.Category](map[string]string)) {
	allCategoryScript := FillScriptCategoryMap(scripts)
	for category, m := range allCategoryScript {
		_ = whichStore(category).UpdateScriptsWithNS(ns, m, scriptPath[category])
	}
}

// LoadAllScriptThrFilepath is used to load and clean  the script, parameter scripts example: {datakit.Logging: [filepath1,..],... }.
func LoadAllScriptThrFilepath(ns string, scripts map[point.Category]([]string)) {
	allCategoryScript := FillScriptCategoryMapFp(scripts)
	for category, filePath := range allCategoryScript {
		LoadDotPScript2Store(category, GitRepoScriptNS, "", filePath)
	}
}

// CleanAllScript is used to clean up all scripts.
func CleanAllScript(ns string) {
	allCategoryScript := FillScriptCategoryMap(nil)
	for category, m := range allCategoryScript {
		_ = whichStore(category).UpdateScriptsWithNS(ns, m, nil)
	}
}

// ReloadAllGitReposDotPScript2Store Deprecated.
func ReloadAllGitReposDotPScript2Store(category point.Category, filePath []string) {
	LoadDotPScript2Store(category, GitRepoScriptNS, "", filePath)
}

// ReloadAllRemoteDotPScript2StoreFromMap Deprecated.
func ReloadAllRemoteDotPScript2StoreFromMap(category point.Category, m map[string]string) {
	_ = whichStore(category).UpdateScriptsWithNS(RemoteScriptNS, m, nil)
}
