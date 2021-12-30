package worker

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/funcs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/parser"
)

const DefaultScriptNs = "default"

var scriptCentorStore = &dotPScriptStore{
	scripts: map[string]map[string]*ScriptInfo{
		DefaultScriptNs: {},
	},
}

type dotPScriptStore struct {
	sync.RWMutex
	scripts map[string](map[string]*ScriptInfo)
}

func (store *dotPScriptStore) queryScript(name string) (*ScriptInfo, error) {
	store.RLock()
	defer store.RUnlock()
	if store.scripts != nil {
		if _, ok := store.scripts[DefaultScriptNs]; ok {
			if v, ok := store.scripts[DefaultScriptNs][name]; ok {
				vCpy := &ScriptInfo{
					name:     v.name,
					script:   v.script,
					updateTS: v.updateTS,
				}
				return vCpy, nil
			}
		}
	}
	return nil, fmt.Errorf("not found")
}

func (store *dotPScriptStore) queryScriptAndNewNg(name string) (*ScriptInfo, error) {
	store.RLock()
	defer store.RUnlock()
	inf, err := store.queryScript(name)
	if err != nil {
		// inf = &ScriptInfo{
		// 	name:     name,
		// 	script:   "if true {}",
		// 	updateTS: time.Now(),
		// }

		return nil, err
	}
	inf.ng, err = parser.NewEngine(inf.script, funcs.FuncsMap, funcs.FuncsCheckMap, false)
	if err != nil {
		return nil, err
	}
	return inf, nil
}

func (store *dotPScriptStore) checkAndUpdate(info *ScriptInfo) {
	store.RLock()
	defer store.RUnlock()
	s, err := store.queryScript(info.name)
	if err == nil { // not found
		return
	}
	if s.updateTS.Equal(info.updateTS) {
		return
	} else { // not equal, update ng, script, updateTS
		info.script = s.script
		info.updateTS = s.updateTS
		info.ng, _ = parser.NewEngine(s.script,
			funcs.FuncsMap, funcs.FuncsCheckMap, false)
		return
	}
}

func (store *dotPScriptStore) appendScript(name string, script string, cover bool) error {
	store.Lock()
	defer store.Unlock()

	if _, ok := store.scripts[DefaultScriptNs]; !ok {
		store.scripts[DefaultScriptNs] = map[string]*ScriptInfo{}
	}

	v, ok := store.scripts[DefaultScriptNs][name]
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
		store.scripts[DefaultScriptNs][name] = &ScriptInfo{
			script:   script,
			name:     name,
			updateTS: time.Now(),
		}
		return nil
	}
}

func (store *dotPScriptStore) appendScriptFromDirPath(dirPath string, cover bool) {
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
			store.appendScriptFromFilePath(sPath, cover)
		}
	}
}

func (store *dotPScriptStore) appendScriptFromFilePath(fp string, cover bool) {
	fp = filepath.Clean(fp)
	if v, err := os.ReadFile(fp); err == nil {
		_, sName := filepath.Split(fp)
		if err := store.appendScript(sName, string(v), cover); err != nil {
			l.Errorf("script name: %s, path: %s, err: %v", sName, fp, err)
		}
	} else {
		l.Error(err)
	}
}

type ScriptInfo struct {
	name     string // script name
	script   string // script content
	ng       *parser.Engine
	updateTS time.Time
}

// Name return pipeline script name.
func (s *ScriptInfo) Name() string {
	return s.name
}

// Script return pipeline script content.
func (s *ScriptInfo) Script() string {
	return s.script
}

// func QueryScriptStore(name string) (string, error) {
// 	if s := scriptCentorStore.queryScript(name); s != nil {
// 		return s.script, nil
// 	} else {
// 		return "", fmt.Errorf("no such pipeline scipt: %s", name)
// 	}
// }

// (存在同名脚本 && notifyExist 为 true) || 解析失败，将返回 error.
func scriptRegister(name, script string, notifyExist bool) error {
	err := scriptCentorStore.appendScript(name, script, false)
	if err != nil && errors.Is(err, ErrScriptExists) {
		if notifyExist {
			return err
		} else {
			return nil
		}
	}
	return err
}

// ScriptRegister 注册 pipeline 脚本，若 script 与已注册的不一致则返回 error
// 脚本解析失败也将返回 error.
func ScriptRegister(name, script string) error {
	return scriptRegister(name, script, true)
}

// ScriptRegisterSkipExist 忽略已注册的 pipeline 脚本.
func ScriptRegisterSkipExist(name, script string) error {
	return scriptRegister(name, script, false)
}
