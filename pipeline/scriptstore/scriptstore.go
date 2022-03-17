// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package scriptstore used to store pipeline script
package scriptstore

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/funcs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/parser"
)

var l = logger.DefaultSLogger("pipeline-scriptstore")

const (
	DefaultScriptNS = "default"  // 内置 pl script， 优先级最低
	GitRepoScriptNS = "git_repo" // git 管理的 pl script
	RemoteScriptNS  = "remote"   // remote pl script，优先级最高
)

var plScriptNSSearchOrder = [3]string{
	RemoteScriptNS, // 优先级最高的 ns
	GitRepoScriptNS,
	DefaultScriptNS,
}

var scriptCentorStore = &DotPScriptStore{
	scripts: map[string]map[string]*ScriptInfo{
		RemoteScriptNS:  {},
		GitRepoScriptNS: {},
		DefaultScriptNS: {},
	},
}

type DotPScriptStore struct {
	sync.RWMutex
	scripts map[string](map[string]*ScriptInfo)
}

func (store *DotPScriptStore) cleanAllScriptWithNS(ns string) {
	store.Lock()
	defer store.Unlock()
	store.scripts[ns] = make(map[string]*ScriptInfo)
}

// func queryScript will return a copy of scriptInfo .
func (store *DotPScriptStore) queryScript(name string, oldInfo *ScriptInfo) (*ScriptInfo, error) {
	store.RLock()
	defer store.RUnlock()

	if store.scripts == nil {
		return nil, fmt.Errorf("not found")
	}

	for _, ns := range plScriptNSSearchOrder {
		if len(store.scripts[ns]) == 0 {
			continue
		}
		if vPtr, ok := store.scripts[ns][name]; ok {
			if oldInfo != nil {
				if vPtr.name != oldInfo.name {
					return nil, fmt.Errorf("name %s != %s", vPtr.name, oldInfo.name)
				}
				if vPtr.updateTS == oldInfo.updateTS && vPtr.ns == oldInfo.ns {
					return oldInfo, nil
				}
			}
			return &ScriptInfo{
				ns:       vPtr.ns,
				ng:       vPtr.ng.Copy(),
				name:     vPtr.name,
				script:   vPtr.script,
				updateTS: vPtr.updateTS,
			}, nil
		}
	}

	return nil, fmt.Errorf("not found")
}

// QueryScript the first parameter is the script name, the second parameter can be nil.
// you can use this function to get the script information stored in the script store.
func QueryScript(name string, needUpdate *ScriptInfo) (*ScriptInfo, error) {
	return scriptCentorStore.queryScript(name, needUpdate)
}

func (store *DotPScriptStore) appendScript(ns string, name string, script string, cover bool) error {
	store.Lock()
	defer store.Unlock()

	if _, ok := store.scripts[ns]; !ok {
		store.scripts[ns] = map[string]*ScriptInfo{}
	}

	v, ok := store.scripts[ns][name]
	if ok && !cover {
		if v.Script() == script {
			return nil
		} else {
			return ErrScriptExists
		}
	} else {
		// (ok && cover) || (!ok)
		if cover && v != nil && v.Script() == script {
			return nil
		}

		ng, err := parser.NewEngine(script, funcs.FuncsMap, funcs.FuncsCheckMap, false)
		if err != nil {
			return err
		}

		store.scripts[ns][name] = &ScriptInfo{
			script:   script,
			name:     name,
			ns:       ns,
			ng:       ng,
			updateTS: time.Now().UnixNano(),
		}
		return nil
	}
}

func AppendScript(ns string, name string, script string, cover bool) error {
	return scriptCentorStore.appendScript(ns, name, script, cover)
}

func (store *DotPScriptStore) appendScriptFromDirPath(ns string, dirPath string, cover bool) {
	dirPath = filepath.Clean(dirPath)
	if dirEntry, err := os.ReadDir(dirPath); err != nil {
		l.Error(err)
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
			if err := store.appendScriptFromFilePath(ns, sPath, cover); err != nil {
				l.Error(err)
			}
		}
	}
}

func (store *DotPScriptStore) appendScriptFromFilePath(ns string, fp string, cover bool) error {
	fp = filepath.Clean(fp)
	if v, err := os.ReadFile(fp); err == nil {
		_, sName := filepath.Split(fp)
		if err := store.appendScript(ns, sName, string(v), cover); err != nil {
			return fmt.Errorf("script name: %s, path: %s, err: %w", sName, fp, err)
		}
	} else {
		return err
	}
	return nil
}

type ScriptInfo struct {
	name     string // script name
	ns       string // script 所属 namespace
	script   string // script content
	ng       *parser.Engine
	updateTS int64
}

// Name return pipeline script name.
func (s *ScriptInfo) Name() string {
	return s.name
}

func (s *ScriptInfo) NameSpace() string {
	return s.ns
}

func (s *ScriptInfo) Engine() *parser.Engine {
	return s.ng
}

// Script return pipeline script content.
func (s *ScriptInfo) Script() string {
	return s.script
}

func InitStore() {
	l = logger.SLogger("pipeline-scriptstore")
	LoadDefaultDotPScript2Store()
}

func LoadDefaultDotPScript2Store() {
	plPath := filepath.Join(datakit.InstallDir, "pipeline")
	LoadDotPScript2StoreWithNS(DefaultScriptNS, nil, plPath)
}

func ReloadAllDefaultDotPScript2Store() {
	plPath := filepath.Join(datakit.InstallDir, "pipeline")
	CleanAllScriptWithNS(DefaultScriptNS)
	LoadDotPScript2StoreWithNS(DefaultScriptNS, nil, plPath)
}

func LoadGitReposDotPScript2Store(filePath []string) {
	LoadDotPScript2StoreWithNS(GitRepoScriptNS, filePath, "")
}

func ReloadAllGitReposDotPScript2Store(filePath []string) {
	CleanAllScriptWithNS(GitRepoScriptNS)
	LoadDotPScript2StoreWithNS(GitRepoScriptNS, filePath, "")
}

func LoadRemoteDotPScript2Store(filePath []string) {
	LoadDotPScript2StoreWithNS(RemoteScriptNS, filePath, "")
}

func ReloadAllRemoteDotPScript2Store(filePath []string) {
	CleanAllScriptWithNS(RemoteScriptNS)
	LoadDotPScript2StoreWithNS(RemoteScriptNS, filePath, "")
}

// LoadDotPScript2StoreWithNS will clean current layer data and then add new script.
func LoadDotPScript2StoreWithNS(ns string, filePath []string, dirPath string) {
	for _, v := range filePath {
		if err := scriptCentorStore.appendScriptFromFilePath(ns, v, true); err != nil {
			l.Error(err)
		}
	}
	if dirPath != "" {
		scriptCentorStore.appendScriptFromDirPath(ns, dirPath, true)
	}
}

func CleanAllScriptWithNS(ns string) {
	scriptCentorStore.cleanAllScriptWithNS(ns)
}
