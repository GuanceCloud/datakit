// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package manager for managing pipeline scripts
package manager

import (
	"os"
	"path/filepath"

	"github.com/GuanceCloud/cliutils/pipeline/manager/relation"
	"github.com/GuanceCloud/cliutils/pipeline/ptinput/plmap"
	"github.com/GuanceCloud/cliutils/point"
)

type Manager struct {
	storeMap map[point.Category]*ScriptStore
	relation *relation.ScriptRelation
}

type ManagerCfg struct {
	gTags [][2]string
	upFn  plmap.UploadFunc
}

func NewManagerCfg(upFn plmap.UploadFunc, gTags [][2]string) ManagerCfg {
	return ManagerCfg{
		upFn:  upFn,
		gTags: gTags,
	}
}

func NewManager(cfg ManagerCfg) *Manager {
	center := &Manager{
		storeMap: map[point.Category]*ScriptStore{},
		relation: relation.NewPipelineRelation(),
	}
	for _, cat := range point.AllCategories() {
		center.storeMap[cat] = NewScriptStore(cat, cfg)
	}
	return center
}

func (c *Manager) whichStore(category point.Category) (*ScriptStore, bool) {
	if category == point.MetricDeprecated {
		category = point.Metric
	}
	if v, ok := c.storeMap[category]; ok && v != nil {
		return v, ok
	}
	return nil, false
}

func (c *Manager) GetScriptRelation() *relation.ScriptRelation {
	return c.relation
}

func (c *Manager) QueryScript(category point.Category, name string) (*PlScript, bool) {
	if v, ok := c.whichStore(category); ok {
		return v.IndexGet(name)
	}
	return nil, false
}

func (c *Manager) ScriptCount(category point.Category) int {
	if v, ok := c.whichStore(category); ok {
		return v.Count()
	}
	return 0
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
	if dirPath == "" {
		return nil, nil
	}

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
	if basePath == "" {
		return nil, nil
	}

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

func LoadDefaultScripts2Store(center *Manager, rootDir string) {
	if rootDir == "" {
		return
	}

	plPath := filepath.Join(rootDir, "pipeline")
	LoadScripts2StoreFromPlStructPath(center, DefaultScriptNS, plPath)
}

func LoadScripts2StoreFromPlStructPath(center *Manager, ns, plPath string) {
	if plPath == "" {
		return
	}

	scripts, path := ReadPlScriptFromPlStructPath(plPath)

	LoadScripts(center, ns, scripts, path)
}

// LoadScripts is used to load and clean the script, parameter scripts example: {datakit.Logging: {ScriptName: ScriptContent},... }.
func LoadScripts(center *Manager, ns string, scripts, scriptPath map[point.Category](map[string]string)) {
	allCategoryScript := FillScriptCategoryMap(scripts)
	for category, m := range allCategoryScript {
		LoadScript(center, category, ns, m, scriptPath[category])
	}
}

func LoadScript(centor *Manager, category point.Category, ns string, scripts map[string]string, path map[string]string) {
	if v, ok := centor.whichStore(category); ok {
		v.UpdateScriptsWithNS(ns, scripts, path)
	}
}

func FillScriptCategoryMap(scripts map[point.Category](map[string]string)) map[point.Category](map[string]string) {
	allCategoryScript := map[point.Category](map[string]string){}
	for _, cat := range point.AllCategories() {
		allCategoryScript[cat] = map[string]string{}
	}
	for k, v := range scripts {
		for name, script := range v {
			if v, ok := allCategoryScript[k]; ok {
				v[name] = script
			}
		}
	}
	return allCategoryScript
}
