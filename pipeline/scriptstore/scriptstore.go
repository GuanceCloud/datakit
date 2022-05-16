// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package scriptstore used to store pipeline script
package scriptstore

import (
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

var (
	_metricScriptStore       = NewScriptStore(datakit.Metric)
	_networkScriptStore      = NewScriptStore(datakit.Network)
	_keyEventScriptStore     = NewScriptStore(datakit.KeyEvent)
	_objectScriptStore       = NewScriptStore(datakit.Object)
	_customObjectScriptStore = NewScriptStore(datakit.CustomObject)
	_loggingScriptStore      = NewScriptStore(datakit.Logging)
	_tracingScriptStore      = NewScriptStore(datakit.Tracing)
	_rumScriptStore          = NewScriptStore(datakit.RUM)
	_securityScriptStore     = NewScriptStore(datakit.Security)

	_heartBeatScriptStore = NewScriptStore(datakit.HeartBeat)
)

func whichStore(category string) *ScriptStore {
	switch category {
	case datakit.Metric, datakit.MetricDeprecated:
		return _metricScriptStore
	case datakit.Network:
		return _networkScriptStore
	case datakit.KeyEvent:
		return _keyEventScriptStore
	case datakit.Object:
		return _objectScriptStore
	case datakit.CustomObject:
		return _customObjectScriptStore
	case datakit.Logging:
		return _loggingScriptStore
	case datakit.Tracing:
		return _tracingScriptStore
	case datakit.RUM:
		return _rumScriptStore
	case datakit.Security:
		return _securityScriptStore
	case datakit.HeartBeat:
		return _heartBeatScriptStore
	default:
		return _loggingScriptStore
	}
}

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

func InitStore() {
	l = logger.SLogger("pipeline-scriptstore")
	LoadDefaultDotPScript2Store()
}

func NSFindPriority(ns string) int {
	switch ns {
	case DefaultScriptNS:
		return 0 // lowest priority
	case GitRepoScriptNS:
		return 1
	case RemoteScriptNS:
		return 2
	default:
		return -1
	}
}

type ScriptStore struct {
	category string
	index    sync.Map
	storage  scriptStorage
}

type scriptStorage struct {
	sync.RWMutex
	scripts map[string](map[string]*PlScript)
}

type PlScript struct {
	name   string // script name
	script string // script content

	ns       string // script 所属 namespace
	category string

	ng *parser.Engine

	updateTS int64
}

func NewScript(name, script, ns, category string) (*PlScript, error) {
	ng, err := parser.NewEngine(script, funcs.FuncsMap, funcs.FuncsCheckMap, false)
	if err != nil {
		return nil, err
	}

	return &PlScript{
		script:   script,
		name:     name,
		ns:       ns,
		category: category,
		ng:       ng,
		updateTS: time.Now().UnixNano(),
	}, nil
}

func NewScriptStore(category string) *ScriptStore {
	return &ScriptStore{
		category: category,
		storage: scriptStorage{
			scripts: map[string]map[string]*PlScript{
				RemoteScriptNS:  {},
				GitRepoScriptNS: {},
				DefaultScriptNS: {},
			},
		},
	}
}

func (script *PlScript) Engine() *parser.Engine {
	return script.ng
}

func (script *PlScript) Name() string {
	return script.name
}

func (script *PlScript) Category() string {
	return script.category
}

func (script *PlScript) NS() string {
	return script.ns
}

func (store *ScriptStore) Get(name string) (*PlScript, bool) {
	if v, ok := store.index.Load(name); ok {
		if v, ok := v.(*PlScript); ok && v != nil {
			return v, ok
		}
	}
	return nil, false
}

func (store *ScriptStore) indexUpdate(script *PlScript) {
	if script == nil {
		return
	}

	curScript, ok := store.Get(script.name)
	if !ok {
		store.index.Store(script.name, script)
		return
	}

	nsCur := NSFindPriority(curScript.ns)
	nsNew := NSFindPriority(script.ns)
	if nsNew >= nsCur {
		store.index.Store(script.name, script)
	}
}

func (store *ScriptStore) indexDeleteAndBack(name, ns string, scripts map[string](map[string]*PlScript)) {
	curScript, ok := store.Get(name)
	if !ok {
		return
	}

	nsCur := NSFindPriority(curScript.ns)
	if NSFindPriority(ns) != nsCur {
		return
	}
	if nsCur == -1 {
		store.index.Delete(name)
		return
	}

	if nsCur > len(plScriptNSSearchOrder) {
		return
	}

	for _, v := range plScriptNSSearchOrder[len(plScriptNSSearchOrder)-nsCur:] {
		if v, ok := scripts[v]; ok {
			if s, ok := v[name]; ok {
				store.index.Store(s.name, s)
				return
			}
		}
	}
	store.index.Delete(name)
}

func (store *ScriptStore) UpdateScriptsWithNS(ns string, namedScript map[string]string) error {
	store.storage.Lock()
	defer store.storage.Unlock()

	if _, ok := store.storage.scripts[ns]; !ok {
		store.storage.scripts[ns] = map[string]*PlScript{}
	}

	script := map[string]*PlScript{}

	errScript := map[string]error{}

	for name, v := range namedScript {
		if s, err := NewScript(name, v, ns, store.category); err == nil && s != nil {
			script[name] = s
		} else {
			errScript[name] = err
		}
	}

	needDelete := []string{}

	// 在 storage & index 执行删除以及更新操作
	for name, curScript := range store.storage.scripts[ns] {
		if newScript, ok := script[name]; ok {
			if newScript.script != curScript.script {
				store.storage.scripts[ns][name] = newScript
				store.indexUpdate(newScript)
			}
			continue
		}
		needDelete = append(needDelete, name)
	}
	for _, name := range needDelete {
		store.indexDeleteAndBack(name, ns, store.storage.scripts)
		delete(store.storage.scripts[ns], name)
	}

	// 执行新增操作
	for name, newScript := range script {
		if _, ok := store.storage.scripts[ns][name]; !ok {
			store.storage.scripts[ns][name] = newScript
			store.indexUpdate(newScript)
		}
	}
	return nil
}

func QueryScript(name, category string) (*PlScript, bool) {
	return whichStore(category).Get(name)
}

func ReadPlScriptFromFile(fp string) (string, string, error) {
	fp = filepath.Clean(fp)
	if v, err := os.ReadFile(fp); err == nil {
		_, sName := filepath.Split(fp)
		return sName, string(v), nil
	} else {
		return "", "", err
	}
}

func ReadPlScriptFromDir(dirPath string) map[string]string {
	ret := map[string]string{}
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
			if name, script, err := ReadPlScriptFromFile(sPath); err != nil {
				ret[name] = script
			} else {
				l.Error(err)
			}
		}
	}
	return ret
}

// LoadDotPScript2Store will diff current layer data and then add new script.
func LoadDotPScript2Store(category, ns string, dirPath string, filePath []string) {
	if len(filePath) > 0 {
		namedScript := map[string]string{}
		for _, fp := range filePath {
			if name, script, err := ReadPlScriptFromFile(fp); err != nil {
				l.Error(err)
			} else {
				namedScript[name] = script
			}
		}
		if err := whichStore(category).UpdateScriptsWithNS(ns, namedScript); err != nil {
			l.Error(err)
		}
	}

	if dirPath != "" {
		namedScript := ReadPlScriptFromDir(dirPath)
		if err := whichStore(category).UpdateScriptsWithNS(ns, namedScript); err != nil {
			l.Error(err)
		}
	}
}

func LoadDefaultDotPScript2Store() {
	plPath := filepath.Join(datakit.InstallDir, "pipeline")
	LoadDotPScript2Store(datakit.Logging, DefaultScriptNS, plPath, nil)
}

func ReloadAllGitReposDotPScript2Store(category string, filePath []string) {
	LoadDotPScript2Store(category, GitRepoScriptNS, "", filePath)
}

func ReloadAllRemoteDotPScript2Store(category string, filePath []string) {
	LoadDotPScript2Store(category, RemoteScriptNS, "", filePath)
}

func ReloadAllRemoteDotPScript2StoreFromMap(category string, m map[string]string) {
	_ = whichStore(category).UpdateScriptsWithNS(RemoteScriptNS, m)
}
