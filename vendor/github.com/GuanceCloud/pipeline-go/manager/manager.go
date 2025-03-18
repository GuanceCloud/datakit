// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package manager for managing pipeline scripts
package manager

import (
	"os"
	"path/filepath"

	"github.com/GuanceCloud/cliutils/point"
	"github.com/GuanceCloud/pipeline-go/lang/platypus"
	"github.com/GuanceCloud/pipeline-go/ptinput/plmap"
)

type Manager struct {
	storeMap map[point.Category]*ScriptStore
	relation *ScriptRelation
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
		relation: NewPipelineRelation(),
	}
	for _, cat := range point.AllCategories() {
		center.storeMap[cat] = NewScriptStore(cat, cfg)
	}
	return center
}

func (m *Manager) whichStore(category point.Category) (*ScriptStore, bool) {
	if category == point.MetricDeprecated {
		category = point.Metric
	}
	if v, ok := m.storeMap[category]; ok && v != nil {
		return v, ok
	}
	return nil, false
}

func (m *Manager) UpdateDefaultScript(mp map[point.Category]string) {
	for _, cat := range point.AllCategories() {
		if store, ok := m.whichStore(cat); ok {
			if v, ok := mp[cat]; ok && v != "" {
				store.SetDefaultScript(v)
			} else {
				store.SetDefaultScript("")
			}
		}
	}
}

func (m *Manager) GetScriptRelation() *ScriptRelation {
	return m.relation
}

func (m *Manager) QueryScript(category point.Category, name string,
	DisableDefaultP ...struct{}) (*platypus.PlScript, bool) {

	if v, ok := m.whichStore(category); ok {
		if ss, ok := v.IndexGet(name); ok {
			return ss, ok
		}

		if len(DisableDefaultP) == 0 {
			return v.IndexDefault()
		}
	}
	return nil, false
}

func (m *Manager) ScriptCount(category point.Category) int {
	if v, ok := m.whichStore(category); ok {
		return v.Count()
	}
	return 0
}

func (m *Manager) LoadScriptsFromWorkspace(ns, plPath string, tags map[string]string) {
	if plPath == "" {
		return
	}

	scripts, _ := ReadWorkspaceScripts(plPath)

	m.LoadScripts(ns, scripts, tags)
}

// LoadScripts is used to load and clean the script, parameter scripts example:
// {point.Logging: {ScriptName: ScriptContent},... }.
func (m *Manager) LoadScripts(ns string, scripts map[point.Category](map[string]string),
	tags map[string]string,
) {
	for _, cat := range point.AllCategories() {
		if ss, ok := scripts[cat]; ok {
			m.LoadScriptWithCat(cat, ns, ss, tags)
		} else {
			// cleanup the store for this category
			m.LoadScriptWithCat(cat, ns, map[string]string{}, tags)
		}
	}
}

func (m *Manager) LoadScriptWithCat(category point.Category, ns string,
	scripts, tags map[string]string,
) {
	if v, ok := m.whichStore(category); ok {
		v.UpdateScriptsWithNS(ns, scripts, tags)
	}
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
		point.DialTesting:  "dialtesting",
	}
}

func SearchWorkspaceScripts(basePath string) map[point.Category](map[string]string) {
	files := map[point.Category](map[string]string){}

	var err error
	files[point.Logging], err = SearchScripts(basePath)
	if err != nil {
		log.Warn(err)
	}

	for category, dirName := range CategoryDirName() {
		s, err := SearchScripts(filepath.Join(basePath, dirName))
		if err != nil {
			log.Warn(err)
		}
		if _, ok := files[category]; !ok {
			files[category] = map[string]string{}
		}
		for k, v := range s {
			files[category][k] = v
		}
	}
	return files
}

func ReadWorkspaceScripts(basePath string) (
	map[point.Category](map[string]string), map[point.Category](map[string]string),
) {
	scriptsPath := SearchWorkspaceScripts(basePath)

	scripts := map[point.Category](map[string]string){}
	for cat, ssPath := range scriptsPath {
		if _, ok := scripts[cat]; !ok {
			scripts[cat] = map[string]string{}
		}
		for _, path := range ssPath {
			if name, script, err := ReadScript(path); err == nil {
				scripts[cat][name] = script
			} else {
				log.Error(err)
			}
		}
	}

	return scripts, scriptsPath
}

func SearchScripts(dirPath string) (map[string]string, error) {
	ret := map[string]string{}
	dirPath = filepath.Clean(dirPath)

	dirEntry, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, err
	}

	for _, v := range dirEntry {
		if v.IsDir() {
			// todo: support sub dir
			continue
		}
		if sName := v.Name(); filepath.Ext(sName) == ".p" {
			ret[sName] = filepath.Join(dirPath, sName)
		}
	}
	return ret, nil
}

func ReadScripts(dirPath string) (map[string]string, map[string]string) {
	ret := map[string]string{}

	scriptsPath, err := SearchScripts(dirPath)
	if err != nil {
		log.Warn(err)
		return nil, nil
	}

	for _, path := range scriptsPath {
		if name, script, err := ReadScript(path); err == nil {
			ret[name] = script
		} else {
			log.Error(err)
		}
	}

	return ret, scriptsPath
}

func ReadScript(fp string) (string, string, error) {
	fp = filepath.Clean(fp)
	if v, err := os.ReadFile(filepath.Clean(fp)); err == nil {
		_, sName := filepath.Split(fp)
		return sName, string(v), nil
	} else {
		return "", "", err
	}
}
