// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package script

import (
	"os"
	"path/filepath"
	"sync"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/stats"
)

var l = logger.DefaultSLogger("pl-script")

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
	_profilingScriptStore    = NewScriptStore(datakit.Profile)

	// TODO: If you add a store, please add the relevant content in the whichStore function.

	_allCategory = map[string]*ScriptStore{
		datakit.Metric:       _metricScriptStore,
		datakit.Network:      _networkScriptStore,
		datakit.KeyEvent:     _keyEventScriptStore,
		datakit.Object:       _objectScriptStore,
		datakit.CustomObject: _customObjectScriptStore,
		datakit.Logging:      _loggingScriptStore,
		datakit.Tracing:      _tracingScriptStore,
		datakit.RUM:          _rumScriptStore,
		datakit.Security:     _securityScriptStore,
		datakit.Profile:      _profilingScriptStore,
	}

	_allDeprecatedCategory = map[string]*ScriptStore{
		datakit.MetricDeprecated: _metricScriptStore,
	}
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
	case datakit.Profile:
		return _profilingScriptStore
	default:
		l.Warn("unsuppored category: %s", category)
		return _loggingScriptStore
	}
}

const (
	DefaultScriptNS = "default" // 内置 pl script， 优先级最低
	GitRepoScriptNS = "gitrepo" // git 管理的 pl script
	RemoteScriptNS  = "remote"  // remote pl script，优先级最高
)

var plScriptNSSearchOrder = [3]string{
	RemoteScriptNS, // 优先级最高的 ns
	GitRepoScriptNS,
	DefaultScriptNS,
}

func InitStore() {
	l = logger.SLogger("pl-script")
	stats.InitStats()
	LoadAllDefaultScripts2Store()
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

func (store *ScriptStore) Get(name string) (*PlScript, bool) {
	if v, ok := store.index.Load(name); ok {
		if v, ok := v.(*PlScript); ok && v != nil {
			return v, ok
		}
	}
	return nil, false
}

func (store *ScriptStore) GetWithNs(name, ns string) (*PlScript, bool) {
	store.storage.RLock()
	defer store.storage.RUnlock()
	if s, ok := store.storage.scripts[ns][name]; ok {
		return s, ok
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
		store.index.Store(script.name, script)

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
	curScript, ok := store.Get(name)
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
		store.index.Delete(name)

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
				store.index.Store(s.name, s)
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

	store.index.Delete(name)

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

	retScripts, retErr := NewScripts(namedScript, scriptPath, ns, store.category)

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
		delete(store.storage.scripts[ns], name)
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
		delete(store.storage.scripts[ns], name)
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

func QueryScript(category, name string) (*PlScript, bool) {
	return whichStore(category).Get(name)
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

func SearchPlFilePathFromPlStructPath(basePath string) map[string](map[string]string) {
	files := map[string](map[string]string){}

	files[datakit.Logging] = SearchPlFilePathFormDir(basePath)

	for category, dirName := range datakit.CategoryDirName() {
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

func ReadPlScriptFromPlStructPath(basePath string) (map[string](map[string]string), map[string](map[string]string)) {
	scripts := map[string](map[string]string){}
	scriptsPath := map[string](map[string]string){}

	scripts[datakit.Logging], scriptsPath[datakit.Logging] = ReadPlScriptFromDir(basePath)

	for category, dirName := range datakit.CategoryDirName() {
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
func LoadDotPScript2Store(category, ns string, dirPath string, filePath []string) {
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

func LoadScript(category, ns string, scripts map[string]string, path map[string]string) {
	_ = whichStore(category).UpdateScriptsWithNS(ns, scripts, path)
}

func FillScriptCategoryMap(scripts map[string](map[string]string)) map[string](map[string]string) {
	allCategoryScript := map[string](map[string]string){}
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

func FillScriptCategoryMapFp(scripts map[string]([]string)) map[string]([]string) {
	allCategoryScript := map[string]([]string){}
	for k := range _allCategory {
		allCategoryScript[k] = []string{}
	}
	for k, v := range scripts {
		allCategoryScript[k] = append(allCategoryScript[k], v...)
	}
	return allCategoryScript
}

// LoadAllScript is used to load and clean the script, parameter scripts example: {datakit.Logging: {ScriptName: ScriptContent},... }.
func LoadAllScript(ns string, scripts, scriptPath map[string](map[string]string)) {
	allCategoryScript := FillScriptCategoryMap(scripts)
	for category, m := range allCategoryScript {
		_ = whichStore(category).UpdateScriptsWithNS(ns, m, scriptPath[category])
	}
}

// LoadAllScriptThrFilepath is used to load and clean  the script, parameter scripts example: {datakit.Logging: [filepath1,..],... }.
func LoadAllScriptThrFilepath(ns string, scripts map[string]([]string)) {
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
func ReloadAllGitReposDotPScript2Store(category string, filePath []string) {
	LoadDotPScript2Store(category, GitRepoScriptNS, "", filePath)
}

// ReloadAllRemoteDotPScript2StoreFromMap Deprecated.
func ReloadAllRemoteDotPScript2StoreFromMap(category string, m map[string]string) {
	_ = whichStore(category).UpdateScriptsWithNS(RemoteScriptNS, m, nil)
}
