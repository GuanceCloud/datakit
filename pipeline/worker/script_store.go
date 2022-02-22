package worker

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/funcs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/parser"
)

const (
	DefaultScriptNS = "default"  // 内置 pl script， 优先级最低
	GitRepoScriptNS = "git_repo" // git 管理的 pl script
	RemoteScriptNS  = "remote"   // remote pl script，优先级最高
)

var scriptCentorStore = &dotPScriptStore{
	scripts: map[string]map[string]*ScriptInfo{
		DefaultScriptNS: {},
	},
}

type dotPScriptStore struct {
	sync.RWMutex
	scripts map[string](map[string]*ScriptInfo)
}

func (store *dotPScriptStore) cleanAllScriptWithNS(ns string) {
	store.Lock()
	defer store.Unlock()
	store.scripts[ns] = make(map[string]*ScriptInfo)
}

// func queryScript will return a copy of scriptInfo, but without ng.
func (store *dotPScriptStore) queryScript(name string) (*ScriptInfo, error) {
	store.RLock()
	defer store.RUnlock()
	if store.scripts != nil {
		var vPtr *ScriptInfo
		switch {
		case len(store.scripts[RemoteScriptNS]) > 0:
			if v, ok := store.scripts[RemoteScriptNS][name]; ok {
				vPtr = v
			}
		case len(store.scripts[GitRepoScriptNS]) > 0:
			if v, ok := store.scripts[GitRepoScriptNS][name]; ok {
				vPtr = v
			}
		case len(store.scripts[DefaultScriptNS]) > 0:
			if v, ok := store.scripts[DefaultScriptNS][name]; ok {
				vPtr = v
			}
		}
		if vPtr != nil {
			return &ScriptInfo{
				ns:       vPtr.ns,
				name:     vPtr.name,
				script:   vPtr.script,
				updateTS: vPtr.updateTS,
			}, nil
		}
	}
	return nil, fmt.Errorf("not found")
}

func (store *dotPScriptStore) queryScriptAndNewNg(name string) (*ScriptInfo, error) {
	inf, err := store.queryScript(name)
	if err != nil {
		return nil, err
	}
	inf.ng, err = parser.NewEngine(inf.script, funcs.FuncsMap, funcs.FuncsCheckMap, false)
	if err != nil {
		return nil, err
	}
	return inf, nil
}

func (store *dotPScriptStore) checkAndUpdate(info *ScriptInfo) (*ScriptInfo, error) {
	s, err := store.queryScript(info.name)
	if err != nil { // not found
		return nil, err
	}
	if s.updateTS == info.updateTS && s.ns == info.ns {
		return info, nil
	} else { // not equal, update ng, script, updateTS, ns, name
		info.ns = s.ns
		info.name = s.name
		info.script = s.script
		info.updateTS = s.updateTS
		info.ng, _ = parser.NewEngine(s.script,
			funcs.FuncsMap, funcs.FuncsCheckMap, false)
		return info, nil
	}
}

func (store *dotPScriptStore) appendScript(ns string, name string, script string, cover bool) error {
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

		node, err := parser.ParsePipeline(script)
		if err != nil {
			return err
		}
		_, ok := node.(parser.Stmts)
		if !ok {
			return fmt.Errorf("invalid AST, should not been here")
		}
		store.scripts[ns][name] = &ScriptInfo{
			script:   script,
			name:     name,
			ns:       ns,
			updateTS: time.Now().UnixNano(),
		}
		return nil
	}
}

func (store *dotPScriptStore) appendScriptFromDirPath(ns string, dirPath string, cover bool) {
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
			store.appendScriptFromFilePath(ns, sPath, cover)
		}
	}
}

func (store *dotPScriptStore) appendScriptFromFilePath(ns string, fp string, cover bool) {
	fp = filepath.Clean(fp)
	if v, err := os.ReadFile(fp); err == nil {
		_, sName := filepath.Split(fp)
		if err := store.appendScript(ns, sName, string(v), cover); err != nil {
			l.Errorf("script name: %s, path: %s, err: %v", sName, fp, err)
		}
	} else {
		l.Error(err)
	}
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

// Script return pipeline script content.
func (s *ScriptInfo) Script() string {
	return s.script
}

func LoadDefaultDotPScript2Store() {
	plPath := filepath.Join(datakit.InstallDir, "pipeline")
	loadDotPScript2StoreWithNS(DefaultScriptNS, nil, plPath)
}

func ReloadAllDefaultDotPScript2Store() {
	plPath := filepath.Join(datakit.InstallDir, "pipeline")
	cleanAllScriptWithNS(DefaultScriptNS)
	loadDotPScript2StoreWithNS(DefaultScriptNS, nil, plPath)
}

func LoadGitReposDotPScript2Store(filePath []string) {
	loadDotPScript2StoreWithNS(GitRepoScriptNS, filePath, "")
}

func ReloadAllGitReposDotPScript2Store(filePath []string) {
	cleanAllScriptWithNS(GitRepoScriptNS)
	loadDotPScript2StoreWithNS(GitRepoScriptNS, filePath, "")
}

func LoadRemoteDotPScript2Store(filePath []string) {
	loadDotPScript2StoreWithNS(RemoteScriptNS, filePath, "")
}

func ReloadAllRemoteDotPScript2Store(filePath []string) {
	cleanAllScriptWithNS(RemoteScriptNS)
	loadDotPScript2StoreWithNS(RemoteScriptNS, filePath, "")
}

// func LoadAllPlScript2StoreWithNS will clean current layer data and then add new script.
func loadDotPScript2StoreWithNS(ns string, filePath []string, dirPath string) {
	for _, v := range filePath {
		scriptCentorStore.appendScriptFromFilePath(ns, v, true)
	}
	if dirPath != "" {
		scriptCentorStore.appendScriptFromDirPath(ns, dirPath, true)
	}
}

func cleanAllScriptWithNS(ns string) {
	scriptCentorStore.cleanAllScriptWithNS(ns)
}
