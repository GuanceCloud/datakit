// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package manager

import (
	"path/filepath"
	"sync"
	"time"

	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/point"
	"github.com/GuanceCloud/pipeline-go/ptinput/plmap"
	"github.com/GuanceCloud/pipeline-go/stats"
)

var log = logger.DefaultSLogger("pl-script")

const (
	NSDefault = "default" // 内置 pl script， 优先级最低
	NSGitRepo = "gitrepo" // git 管理的 pl script
	NSConfd   = "confd"   // confd 管理的 pl script
	NSRemote  = "remote"  // remote pl script，优先级最高
)

var nsSearchOrder = [4]string{
	NSRemote, // 优先级最高的 ns
	NSConfd,
	NSGitRepo,
	NSDefault,
}

func InitLog() {
	log = logger.SLogger("pl-script")
}

func InitStore(manager *Manager, installDir string, tags map[string]string) {
	stats.InitLog()

	plPath := filepath.Join(installDir, "pipeline")
	manager.LoadScriptsFromWorkspace(NSDefault, plPath, tags)
}

func NSFindPriority(ns string) int {
	switch ns {
	case NSDefault:
		return 0 // lowest priority
	case NSGitRepo:
		return 1
	case NSConfd:
		return 2
	case NSRemote:
		return 3
	default:
		return -1
	}
}

type ScriptStore struct {
	category point.Category

	storage scriptStorage

	defultScript string

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
				NSRemote:  {},
				NSConfd:   {},
				NSGitRepo: {},
				NSDefault: {},
			},
		},
		index: map[string]*PlScript{},
		cfg:   cfg,
	}
}

func (store *ScriptStore) SetDefaultScript(name string) {
	store.indexLock.Lock()
	defer store.indexLock.Unlock()
	store.defultScript = name
}

func (store *ScriptStore) GetDefaultScript() string {
	store.indexLock.RLock()
	defer store.indexLock.RUnlock()
	return store.defultScript
}

func (store *ScriptStore) IndexGet(name string) (*PlScript, bool) {
	store.indexLock.RLock()
	defer store.indexLock.RUnlock()
	if v, ok := store.index[name]; ok {
		return v, ok
	}
	return nil, false
}

func (store *ScriptStore) IndexDefault() (*PlScript, bool) {
	store.indexLock.RLock()
	defer store.indexLock.RUnlock()

	if v, ok := store.index[store.defultScript]; ok {
		return v, ok
	}
	return nil, false
}

func (store *ScriptStore) Count() int {
	store.storage.RLock()
	defer store.storage.RUnlock()

	return len(store.storage.scripts[NSRemote]) +
		len(store.storage.scripts[NSConfd]) +
		len(store.storage.scripts[NSGitRepo]) +
		len(store.storage.scripts[NSDefault])
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

		stats.WriteUpdateTime(script.tags)
		stats.WriteEvent(&stats.ChangeEvent{
			Name:     script.name,
			Category: script.category,
			NS:       script.ns,
			Script:   script.script,
			Op:       stats.EventOpIndex,
			Time:     time.Now(),
		}, script.tags)
		return
	}

	nsCur := NSFindPriority(curScript.ns)
	nsNew := NSFindPriority(script.ns)
	if nsNew >= nsCur {
		store.indexStore(script)
		stats.WriteUpdateTime(curScript.tags)
		stats.WriteUpdateTime(script.tags)
		stats.WriteEvent(&stats.ChangeEvent{
			Name:      script.name,
			Category:  script.category,
			NS:        script.ns,
			NSOld:     curScript.ns,
			Script:    script.script,
			ScriptOld: curScript.script,
			Op:        stats.EventOpIndexUpdate,
			Time:      time.Now(),
		}, script.tags)
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

	if nsCur > len(nsSearchOrder) {
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
		}, curScript.tags)
		return
	}

	for _, v := range nsSearchOrder[len(nsSearchOrder)-nsCur:] {
		if v, ok := scripts4back[v]; ok {
			if s, ok := v[name]; ok {
				store.indexStore(s)
				stats.WriteUpdateTime(s.tags)
				stats.WriteEvent(&stats.ChangeEvent{
					Name:      name,
					Category:  s.category,
					NS:        s.ns,
					NSOld:     curScript.ns,
					Script:    s.script,
					ScriptOld: curScript.script,
					Op:        stats.EventOpIndexDeleteAndBack,
					Time:      time.Now(),
				}, s.tags)
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
	}, curScript.tags)
}

func (store *ScriptStore) UpdateScriptsWithNS(ns string,
	namedScript, scriptTags map[string]string,
) map[string]error {
	store.storage.Lock()
	defer store.storage.Unlock()

	if _, ok := store.storage.scripts[ns]; !ok {
		store.storage.scripts[ns] = map[string]*PlScript{}
	}

	aggBuk := plmap.NewAggBuks(store.cfg.upFn, store.cfg.gTags)
	retScripts, retErr := NewScripts(namedScript, scriptTags, ns, store.category,
		aggBuk)

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

		sTags := map[string]string{
			"name":     name,
			"ns":       ns,
			"lang":     "platypus",
			"category": store.category.String(),
		}

		for k, v := range scriptTags {
			if _, ok := sTags[k]; !ok {
				sTags[k] = v
			}
		}

		stats.WriteUpdateTime(sTags)
		store.indexDeleteAndBack(name, ns, store.storage.scripts)

		if v, ok := store.storage.scripts[ns][name]; ok {
			if v.plBuks != nil {
				v.plBuks.StopAllBukScanner()
			}
			if v.cache != nil {
				v.cache.Stop()
			}
			if v.ptWindow != nil {
				v.ptWindow.Deprecated()
			}
			delete(store.storage.scripts[ns], name)
		}
		stats.WriteEvent(&change, sTags)
	}

	// 如果上一次的集合中的脚本不存在于当前结果，则删除
	for name, curScript := range store.storage.scripts[ns] {
		if newScript, ok := retScripts[name]; ok { // 有更新
			store.storage.scripts[ns][name] = newScript
			stats.WriteUpdateTime(newScript.tags)
			store.indexUpdate(newScript)
		} else { // 删除
			stats.WriteUpdateTime(curScript.tags)
			store.indexDeleteAndBack(name, ns, store.storage.scripts)
			delete(store.storage.scripts[ns], name)
		}

		// 清理之前一个脚本的资源
		if curScript.plBuks != nil {
			curScript.plBuks.StopAllBukScanner()
		}
		if curScript.cache != nil {
			curScript.cache.Stop()
		}
		if curScript.ptWindow != nil {
			curScript.ptWindow.Deprecated()
		}
	}

	// 执行新增操作
	for name, newScript := range retScripts {
		if _, ok := store.storage.scripts[ns][name]; !ok {
			store.storage.scripts[ns][name] = newScript
			stats.WriteUpdateTime(newScript.tags)
			store.indexUpdate(newScript)
		}
	}

	if len(retErr) > 0 {
		return retErr
	}
	return nil
}
