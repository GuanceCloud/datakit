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
	"github.com/GuanceCloud/pipeline-go/constants"
	"github.com/GuanceCloud/pipeline-go/lang"
	"github.com/GuanceCloud/pipeline-go/lang/platypus"
	"github.com/GuanceCloud/pipeline-go/stats"
)

var log = logger.DefaultSLogger("pl-script")

var nsSearchOrder = [4]string{
	constants.NSRemote, // 优先级最高的 ns
	constants.NSConfd,
	constants.NSGitRepo,
	constants.NSDefault,
}

func InitLog() {
	log = logger.SLogger("pl-script")
}

func InitStore(manager *Manager, installDir string, tags map[string]string) {
	stats.InitLog()

	plPath := filepath.Join(installDir, "pipeline")
	manager.LoadScriptsFromWorkspace(constants.NSDefault, plPath, tags)
}

func NSFindPriority(ns string) int {
	switch ns {
	case constants.NSDefault:
		return 0 // lowest priority
	case constants.NSGitRepo:
		return 1
	case constants.NSConfd:
		return 2
	case constants.NSRemote:
		return 3
	default:
		return -1
	}
}

type ScriptStore struct {
	category point.Category

	storage scriptStorage

	defultScript string

	index     map[string]*platypus.PlScript
	indexLock sync.RWMutex

	cfg ManagerCfg
}

type scriptStorage struct {
	sync.RWMutex
	scripts map[string](map[string]*platypus.PlScript)
}

func NewScriptStore(category point.Category, cfg ManagerCfg) *ScriptStore {
	return &ScriptStore{
		category: category,
		storage: scriptStorage{
			scripts: map[string]map[string]*platypus.PlScript{
				constants.NSRemote:  {},
				constants.NSConfd:   {},
				constants.NSGitRepo: {},
				constants.NSDefault: {},
			},
		},
		index: map[string]*platypus.PlScript{},
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

func (store *ScriptStore) IndexGet(name string) (*platypus.PlScript, bool) {
	store.indexLock.RLock()
	defer store.indexLock.RUnlock()
	if v, ok := store.index[name]; ok {
		return v, ok
	}
	return nil, false
}

func (store *ScriptStore) IndexDefault() (*platypus.PlScript, bool) {
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

	return len(store.storage.scripts[constants.NSRemote]) +
		len(store.storage.scripts[constants.NSConfd]) +
		len(store.storage.scripts[constants.NSGitRepo]) +
		len(store.storage.scripts[constants.NSDefault])
}

func (store *ScriptStore) GetWithNs(name, ns string) (*platypus.PlScript, bool) {
	store.storage.RLock()
	defer store.storage.RUnlock()
	if s, ok := store.storage.scripts[ns][name]; ok {
		return s, ok
	}
	return nil, false
}

func (store *ScriptStore) indexStore(script *platypus.PlScript) {
	store.indexLock.Lock()
	defer store.indexLock.Unlock()

	if store.index == nil {
		store.index = map[string]*platypus.PlScript{}
	}
	store.index[script.Name()] = script
}

func (store *ScriptStore) indexDelete(name string) {
	store.indexLock.Lock()
	defer store.indexLock.Unlock()

	delete(store.index, name)
}

func (store *ScriptStore) indexUpdate(script *platypus.PlScript) {
	if script == nil {
		return
	}

	curScript, ok := store.IndexGet(script.Name())

	if !ok {
		store.indexStore(script)

		stats.WriteUpdateTime(script.Meta())
		stats.WriteEvent(&stats.ChangeEvent{
			Name:     script.Name(),
			Category: script.Category(),
			NS:       script.NS(),
			Script:   script.Content(),
			Op:       stats.EventOpIndex,
			Time:     time.Now(),
		}, script.Meta())
		return
	}

	nsCur := NSFindPriority(curScript.NS())
	nsNew := NSFindPriority(script.NS())
	if nsNew >= nsCur {
		store.indexStore(script)
		stats.WriteUpdateTime(curScript.Meta())
		stats.WriteUpdateTime(script.Meta())
		stats.WriteEvent(&stats.ChangeEvent{
			Name:      script.Name(),
			Category:  script.Category(),
			NS:        script.NS(),
			NSOld:     curScript.NS(),
			Script:    script.Content(),
			ScriptOld: curScript.Content(),
			Op:        stats.EventOpIndexUpdate,
			Time:      time.Now(),
		}, script.Meta())
	}
}

func (store *ScriptStore) indexDeleteAndBack(name, ns string,
	scripts4back map[string](map[string]*platypus.PlScript)) {
	curScript, ok := store.IndexGet(name)
	if !ok {
		return
	}

	nsCur := NSFindPriority(curScript.NS())
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
			Category: curScript.Category(),
			NS:       curScript.NS(),
			Script:   curScript.Content(),
			Op:       stats.EventOpIndexDelete,
			Time:     time.Now(),
		}, curScript.Meta())
		return
	}

	for _, v := range nsSearchOrder[len(nsSearchOrder)-nsCur:] {
		if v, ok := scripts4back[v]; ok {
			if s, ok := v[name]; ok {
				store.indexStore(s)
				stats.WriteUpdateTime(s.Meta())
				stats.WriteEvent(&stats.ChangeEvent{
					Name:      name,
					Category:  s.Category(),
					NS:        s.NS(),
					NSOld:     curScript.NS(),
					Script:    s.Content(),
					ScriptOld: curScript.Content(),
					Op:        stats.EventOpIndexDeleteAndBack,
					Time:      time.Now(),
				}, s.Meta())
				return
			}
		}
	}

	store.indexDelete(name)

	stats.WriteEvent(&stats.ChangeEvent{
		Name:     name,
		Category: curScript.Category(),
		NS:       curScript.NS(),
		Script:   curScript.Content(),
		Op:       stats.EventOpIndexDelete,
		Time:     time.Now(),
	}, curScript.Meta())
}

func (store *ScriptStore) UpdateScriptsWithNS(ns string,
	namedScript, scriptTags map[string]string,
) map[string]error {
	store.storage.Lock()
	defer store.storage.Unlock()

	if _, ok := store.storage.scripts[ns]; !ok {
		store.storage.scripts[ns] = map[string]*platypus.PlScript{}
	}

	retScripts, retErr := platypus.NewScripts(
		namedScript,

		lang.WithMeta(scriptTags),
		lang.WithNS(ns),
		lang.WithCat(store.category),

		lang.WithAggBkt(store.cfg.upFn, store.cfg.gTags),
		lang.WithCache(),
		lang.WithPtWindow(),
	)

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
			v.Cleanup()

			delete(store.storage.scripts[ns], name)
		}
		stats.WriteEvent(&change, sTags)
	}

	// 如果上一次的集合中的脚本不存在于当前结果，则删除
	for name, sc := range store.storage.scripts[ns] {
		if newScript, ok := retScripts[name]; ok { // 有更新
			store.storage.scripts[ns][name] = newScript
			stats.WriteUpdateTime(newScript.Meta())
			store.indexUpdate(newScript)
		} else { // 删除
			stats.WriteUpdateTime(sc.Meta())
			store.indexDeleteAndBack(name, ns, store.storage.scripts)
			delete(store.storage.scripts[ns], name)
		}

		// 清理之前一个脚本的资源
		sc.Cleanup()
	}

	// 执行新增操作
	for name, newScript := range retScripts {
		if _, ok := store.storage.scripts[ns][name]; !ok {
			store.storage.scripts[ns][name] = newScript
			stats.WriteUpdateTime(newScript.Meta())
			store.indexUpdate(newScript)
		}
	}

	if len(retErr) > 0 {
		return retErr
	}
	return nil
}
