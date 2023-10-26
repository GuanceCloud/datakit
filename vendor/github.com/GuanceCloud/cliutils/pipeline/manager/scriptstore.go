// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package manager

import (
	"sync"
	"time"

	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/pipeline/ptinput/plmap"
	"github.com/GuanceCloud/cliutils/pipeline/stats"
	"github.com/GuanceCloud/cliutils/point"
)

var l = logger.DefaultSLogger("pl-script")

const (
	DefaultScriptNS = "default" // 内置 pl script， 优先级最低
	GitRepoScriptNS = "gitrepo" // git 管理的 pl script
	ConfdScriptNS   = "confd"   // confd 管理的 pl script
	RemoteScriptNS  = "remote"  // remote pl script，优先级最高
)

var plScriptNSSearchOrder = [4]string{
	RemoteScriptNS, // 优先级最高的 ns
	ConfdScriptNS,
	GitRepoScriptNS,
	DefaultScriptNS,
}

func InitLog() {
	l = logger.SLogger("pl-script")
}

func InitStore(center *Manager, installDir string) {
	stats.InitLog()
	LoadDefaultScripts2Store(center, installDir)
}

func NSFindPriority(ns string) int {
	switch ns {
	case DefaultScriptNS:
		return 0 // lowest priority
	case GitRepoScriptNS:
		return 1
	case ConfdScriptNS:
		return 2
	case RemoteScriptNS:
		return 3
	default:
		return -1
	}
}

type ScriptStore struct {
	category point.Category

	storage scriptStorage

	index     map[string]*PlScript
	indexLock sync.RWMutex

	cfg ManagerCfg
}

type scriptStorage struct {
	sync.RWMutex
	scripts map[string](map[string]*PlScript)
}

func NewScriptStore(category point.Category, cfg ManagerCfg) *ScriptStore {
	return &ScriptStore{
		category: category,
		storage: scriptStorage{
			scripts: map[string]map[string]*PlScript{
				RemoteScriptNS:  {},
				ConfdScriptNS:   {},
				GitRepoScriptNS: {},
				DefaultScriptNS: {},
			},
		},
		index: map[string]*PlScript{},
		cfg:   cfg,
	}
}

func (store *ScriptStore) IndexGet(name string) (*PlScript, bool) {
	store.indexLock.RLock()
	defer store.indexLock.RUnlock()
	if v, ok := store.index[name]; ok {
		return v, ok
	}
	return nil, false
}

func (store *ScriptStore) Count() int {
	store.storage.RLock()
	defer store.storage.RUnlock()

	return len(store.storage.scripts[RemoteScriptNS]) +
		len(store.storage.scripts[ConfdScriptNS]) +
		len(store.storage.scripts[GitRepoScriptNS]) +
		len(store.storage.scripts[DefaultScriptNS])
}

func (store *ScriptStore) GetWithNs(name, ns string) (*PlScript, bool) {
	store.storage.RLock()
	defer store.storage.RUnlock()
	if s, ok := store.storage.scripts[ns][name]; ok {
		return s, ok
	}
	return nil, false
}

func (store *ScriptStore) indexStore(script *PlScript) {
	store.indexLock.Lock()
	defer store.indexLock.Unlock()

	if store.index == nil {
		store.index = map[string]*PlScript{}
	}
	store.index[script.name] = script
}

func (store *ScriptStore) indexDelete(name string) {
	store.indexLock.Lock()
	defer store.indexLock.Unlock()

	delete(store.index, name)
}

func (store *ScriptStore) indexUpdate(script *PlScript) {
	if script == nil {
		return
	}

	curScript, ok := store.IndexGet(script.name)

	if !ok {
		store.indexStore(script)

		stats.UpdateScriptStatsMeta(script.category, script.ns, script.name, script.script, true, false, "")
		stats.WriteEvent(&stats.ChangeEvent{
			Name:     script.name,
			Category: script.category,
			NS:       script.ns,
			Script:   script.script,
			Op:       stats.EventOpIndex,
			Time:     time.Now(),
		})
		return
	}

	nsCur := NSFindPriority(curScript.ns)
	nsNew := NSFindPriority(script.ns)
	if nsNew >= nsCur {
		store.indexStore(script)
		stats.UpdateScriptStatsMeta(curScript.category, curScript.ns, curScript.name, curScript.script, false, false, "")
		stats.UpdateScriptStatsMeta(script.category, script.ns, script.name, script.script, true, false, "")
		stats.WriteEvent(&stats.ChangeEvent{
			Name:      script.name,
			Category:  script.category,
			NS:        script.ns,
			NSOld:     curScript.ns,
			Script:    script.script,
			ScriptOld: curScript.script,
			Op:        stats.EventOpIndexUpdate,
			Time:      time.Now(),
		})
	}
}

func (store *ScriptStore) indexDeleteAndBack(name, ns string, scripts4back map[string](map[string]*PlScript)) {
	curScript, ok := store.IndexGet(name)
	if !ok {
		return
	}

	nsCur := NSFindPriority(curScript.ns)
	if NSFindPriority(ns) != nsCur {
		return
	}

	if nsCur > len(plScriptNSSearchOrder) {
		return
	}

	if nsCur == -1 {
		store.indexDelete(name)

		stats.WriteEvent(&stats.ChangeEvent{
			Name:     name,
			Category: curScript.category,
			NS:       curScript.ns,
			Script:   curScript.script,
			Op:       stats.EventOpIndexDelete,
			Time:     time.Now(),
		})
		return
	}

	for _, v := range plScriptNSSearchOrder[len(plScriptNSSearchOrder)-nsCur:] {
		if v, ok := scripts4back[v]; ok {
			if s, ok := v[name]; ok {
				store.indexStore(s)
				stats.UpdateScriptStatsMeta(s.category, s.ns, s.name, s.script, true, false, "")
				stats.WriteEvent(&stats.ChangeEvent{
					Name:      name,
					Category:  s.category,
					NS:        s.ns,
					NSOld:     curScript.ns,
					Script:    s.script,
					ScriptOld: curScript.script,
					Op:        stats.EventOpIndexDeleteAndBack,
					Time:      time.Now(),
				})
				return
			}
		}
	}

	store.indexDelete(name)

	stats.WriteEvent(&stats.ChangeEvent{
		Name:     name,
		Category: curScript.category,
		NS:       curScript.ns,
		Script:   curScript.script,
		Op:       stats.EventOpIndexDelete,
		Time:     time.Now(),
	})
}

func (store *ScriptStore) UpdateScriptsWithNS(ns string, namedScript map[string]string, scriptPath map[string]string) map[string]error {
	store.storage.Lock()
	defer store.storage.Unlock()

	if _, ok := store.storage.scripts[ns]; !ok {
		store.storage.scripts[ns] = map[string]*PlScript{}
	}

	retScripts, retErr := NewScripts(namedScript, scriptPath, ns, store.category,
		plmap.NewAggBuks(store.cfg.upFn, store.cfg.gTags))

	for name, err := range retErr {
		var errStr string
		if err != nil {
			errStr = err.Error()
		}
		change := stats.ChangeEvent{
			Name:         name,
			Category:     store.category,
			NS:           ns,
			Script:       namedScript[name],
			Op:           stats.EventOpCompileError,
			Time:         time.Now(),
			CompileError: errStr,
		}
		stats.UpdateScriptStatsMeta(store.category, ns, name, namedScript[name], false, true, errStr)
		store.indexDeleteAndBack(name, ns, store.storage.scripts)

		if v, ok := store.storage.scripts[ns][name]; ok {
			if v.plBuks != nil {
				v.plBuks.StopAllBukScanner()
			}
			delete(store.storage.scripts[ns], name)
		}
		stats.WriteEvent(&change)
	}

	needDelete := map[string]string{}

	// 在 storage & index 执行删除以及更新操作
	for name, curScript := range store.storage.scripts[ns] {
		if newScript, ok := retScripts[name]; ok {
			store.storage.scripts[ns][name] = newScript
			stats.UpdateScriptStatsMeta(store.category, ns, name, newScript.script, false, false, "")
			store.indexUpdate(newScript)
		}
		needDelete[name] = curScript.script
	}
	for name, script := range needDelete {
		stats.UpdateScriptStatsMeta(store.category, ns, name, script, false, true, "")
		store.indexDeleteAndBack(name, ns, store.storage.scripts)

		if v, ok := store.storage.scripts[ns][name]; ok {
			if v.plBuks != nil {
				v.plBuks.StopAllBukScanner()
			}
			delete(store.storage.scripts[ns], name)
		}
	}

	// 执行新增操作
	for name, newScript := range retScripts {
		if _, ok := store.storage.scripts[ns][name]; !ok {
			store.storage.scripts[ns][name] = newScript
			stats.UpdateScriptStatsMeta(store.category, ns, name, newScript.script, false, false, "")
			store.indexUpdate(newScript)
		}
	}

	if len(retErr) > 0 {
		return retErr
	}
	return nil
}

func (store *ScriptStore) LoadDotPScript2Store(ns string, dirPath string, filePath []string) {
	if len(filePath) > 0 {
		namedScript := map[string]string{}
		scriptPath := map[string]string{}
		for _, fp := range filePath {
			if name, script, err := ReadPlScriptFromFile(fp); err != nil {
				l.Error(err)
			} else {
				scriptPath[name] = fp
				namedScript[name] = script
			}
		}
		if err := store.UpdateScriptsWithNS(ns, namedScript, scriptPath); err != nil {
			l.Error(err)
		}
	}

	if dirPath != "" {
		namedScript, filePath := ReadPlScriptFromDir(dirPath)
		if err := store.UpdateScriptsWithNS(ns, namedScript, filePath); err != nil {
			l.Error(err)
		}
	}
}
