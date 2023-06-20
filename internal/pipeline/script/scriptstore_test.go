// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package script

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"github.com/stretchr/testify/assert"
)

func TestScriptLoadFunc(t *testing.T) {
	case1 := map[point.Category]map[string]string{
		point.Logging: {
			"abcd": "if true {}",
		},
		point.Metric: {
			"abc": "if true {}",
			"def": "if true {}",
		},
	}

	CleanAllScript(DefaultScriptNS)
	CleanAllScript(GitRepoScriptNS)
	CleanAllScript(RemoteScriptNS)

	LoadAllScript(DefaultScriptNS, case1, nil)
	for category, v := range case1 {
		for name := range v {
			if y, ok := QueryScript(category, name); !ok {
				t.Error(category, " ", name, y)
				if y, ok := QueryScript(category, name); !ok {
					t.Error(y)
				}
			}
		}
	}

	CleanAllScript(DefaultScriptNS)
	CleanAllScript(GitRepoScriptNS)
	CleanAllScript(RemoteScriptNS)
	for k, v := range case1 {
		LoadScript(k, DefaultScriptNS, v, nil)
	}
	for category, v := range case1 {
		for name := range v {
			if _, ok := QueryScript(category, name); !ok {
				t.Error(category, " ", name)
			}
		}
	}

	CleanAllScript(DefaultScriptNS)
	CleanAllScript(GitRepoScriptNS)
	CleanAllScript(RemoteScriptNS)
	for category, v := range case1 {
		for name := range v {
			if _, ok := QueryScript(category, name); ok {
				t.Error(category, " ", name)
			}
		}
	}

	CleanAllScript(DefaultScriptNS)
	CleanAllScript(GitRepoScriptNS)
	CleanAllScript(RemoteScriptNS)

	for k, v := range case1 {
		LoadScript(k, "DefaultScriptNS", v, nil)
		ReloadAllRemoteDotPScript2StoreFromMap(k, v)
	}
	for category, v := range case1 {
		for name := range v {
			if s, ok := QueryScript(category, name); !ok || s.NS() != RemoteScriptNS {
				t.Error(category, " ", name)
			}
		}
	}

	CleanAllScript(DefaultScriptNS)
	CleanAllScript(GitRepoScriptNS)
	CleanAllScript(RemoteScriptNS)

	for k, v := range case1 {
		LoadScript(k, "aabb", v, nil)
	}
	CleanAllScript("aabb")

	_ = os.WriteFile("/tmp/nginx-xx.p", []byte(`
	json(_, time)
	set_tag(bb, "aa0")
	default_time(time)
	`), os.FileMode(0o755))
	LoadAllScriptThrFilepath(DefaultScriptNS, map[point.Category][]string{point.RUM: {"/tmp/nginx-xx.p"}})
	_ = os.Remove("/tmp/nginx-xx.p")
	if _, ok := QueryScript(point.RUM, "nginx-xx.p"); !ok {
		t.Error(point.RUM, " ", "nginx-xx.p")
	}

	LoadAllDefaultScripts2Store()
	ReloadAllGitReposDotPScript2Store(point.Logging, nil)
	_ = os.WriteFile("/tmp/nginx-time123.p", []byte(`
		json(_, time)
		set_tag(bb, "aa0")
		default_time(time)
		`), os.FileMode(0o755))
	LoadDotPScript2Store(point.Logging, "xxxx", "", []string{"/tmp/nginx-time.p123"})
	_ = os.Remove("/tmp/nginx-time123.p")
	LoadDotPScript2Store(point.Logging, "xxx", "", nil)
}

func TestCmpCategory(t *testing.T) {
	c, dc := CategoryList()
	c1, dc1 := func() (map[point.Category]struct{}, map[point.Category]struct{}) {
		ret1 := map[point.Category]struct{}{}
		ret2 := map[point.Category]struct{}{}
		for k := range _allCategory {
			ret1[k] = struct{}{}
		}
		for k := range _allDeprecatedCategory {
			ret2[k] = struct{}{}
		}
		return ret1, ret2
	}()

	assert.Equal(t, dc, dc1)

	assert.Equal(t, c, c1)
	assert.Equal(t, c1, func() map[point.Category]struct{} {
		ret := map[point.Category]struct{}{}
		for k := range CategoryDirName() {
			ret[k] = struct{}{}
		}
		return ret
	}())
}

func BenchmarkIndexMap(b *testing.B) {
	b.Run("sync.Map", func(b *testing.B) {
		type cachemap struct {
			m sync.Map
		}

		m := cachemap{}
		m.m.Store("abc.p", &PlScript{})
		m.m.Store("def.p", &PlScript{})

		var x1, x2, x3 *PlScript
		for i := 0; i < b.N; i++ {
			if v, ok := m.m.Load("abc.p"); ok {
				x1 = v.(*PlScript)
			}
			if v, ok := m.m.Load("def.p"); ok {
				x2 = v.(*PlScript)
			}
			if v, ok := m.m.Load("ddd"); ok {
				x3 = v.(*PlScript)
			}
		}
		b.Log(x1, x2, x3, false)
	})

	b.Run("map", func(b *testing.B) {
		type cachemap struct {
			m     map[string]*PlScript
			mlock sync.RWMutex
		}

		m := cachemap{
			m: map[string]*PlScript{
				"abc.p": {},
				"def.p": {},
			},
		}

		var x1, x2, x3 *PlScript
		var ok bool
		for i := 0; i < b.N; i++ {
			m.mlock.RLock()
			x1, ok = m.m["abc.p"]
			if !ok {
				b.Log()
			}
			m.mlock.RUnlock()

			m.mlock.RLock()
			x2, ok = m.m["def.p"]
			if !ok {
				b.Log()
			}
			m.mlock.RUnlock()

			m.mlock.RLock()
			x3, ok = m.m["ddd"]
			if ok {
				b.Log()
			}
			m.mlock.RUnlock()
		}
		b.Log(x1, x2, x3, ok)
	})
}

func TestPlScriptStore(t *testing.T) {
	store := NewScriptStore(point.Logging)

	store.indexUpdate(nil)

	err := store.UpdateScriptsWithNS(DefaultScriptNS, map[string]string{"abc.p": "default_time(time) ;set_tag(a, \"1\")"}, nil)
	if err != nil {
		t.Error(err)
	}

	err = store.UpdateScriptsWithNS(DefaultScriptNS, map[string]string{"abc.p": "default_time(time)"}, nil)
	if err != nil {
		t.Error(err)
	}

	err = store.UpdateScriptsWithNS(DefaultScriptNS, map[string]string{"abc.p": "default_time(time); set_tag(a, 1)"}, nil)
	if err == nil {
		t.Error("should not be nil")
	}

	err = store.UpdateScriptsWithNS(DefaultScriptNS, map[string]string{"abc.p": "default_time(time)"}, nil)
	if err != nil {
		t.Error(err)
	}

	err = store.UpdateScriptsWithNS(GitRepoScriptNS, map[string]string{"abc.p": "default_time(time)"}, nil)
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, store.Count(), 2)

	err = store.UpdateScriptsWithNS(ConfdScriptNS, map[string]string{"abc.p": "default_time(time)"}, nil)
	if err != nil {
		t.Error(err)
	}

	err = store.UpdateScriptsWithNS(RemoteScriptNS, map[string]string{"abc.p": "default_time(time)"}, nil)
	if err != nil {
		t.Error(err)
	}

	for i, ns := range plScriptNSSearchOrder {
		store.UpdateScriptsWithNS(ns, nil, nil)
		if i < len(plScriptNSSearchOrder)-1 {
			sInfo, ok := store.IndexGet("abc.p")
			if !ok {
				t.Error(fmt.Errorf("!ok"))
				return
			}
			if sInfo.ns != plScriptNSSearchOrder[i+1] {
				t.Error(sInfo.ns, plScriptNSSearchOrder[i+1])
			}
		} else {
			_, ok := store.IndexGet("abc.p")
			if ok {
				t.Error(fmt.Errorf("shoud not be ok"))
				return
			}
		}
	}
}

func TestWhichStore(t *testing.T) {
	r := whichStore(point.Metric)
	if r == nil {
		t.Fatal("err")
	}
	if r != _metricScriptStore {
		t.Fatal("not equal")
	}

	r = whichStore(point.MetricDeprecated)
	if r == nil {
		t.Fatal("err")
	}
	if r != _metricScriptStore {
		t.Fatal("not equal")
	}

	r = whichStore(point.Network)
	if r == nil {
		t.Fatal("err")
	}
	if r != _networkScriptStore {
		t.Fatal("not equal")
	}

	r = whichStore(point.KeyEvent)
	if r == nil {
		t.Fatal("err")
	}
	if r != _keyEventScriptStore {
		t.Fatal("not equal")
	}

	r = whichStore(point.Object)
	if r == nil {
		t.Fatal("err")
	}
	if r != _objectScriptStore {
		t.Fatal("not equal")
	}

	r = whichStore(point.CustomObject)
	if r == nil {
		t.Fatal("err")
	}
	if r != _customObjectScriptStore {
		t.Fatal("not equal")
	}

	r = whichStore(point.Logging)
	if r == nil {
		t.Fatal("err")
	}
	if r != _loggingScriptStore {
		t.Fatal("not equal")
	}

	r = whichStore(point.Tracing)
	if r == nil {
		t.Fatal("err")
	}
	if r != _tracingScriptStore {
		t.Fatal("not equal")
	}

	r = whichStore(point.Profiling)
	if r == nil {
		t.Fatal("err")
	}
	if r != _profilingScriptStore {
		t.Fatal("not equal")
	}

	r = whichStore(point.RUM)
	if r == nil {
		t.Fatal("err")
	}
	if r != _rumScriptStore {
		t.Fatal("not equal")
	}

	r = whichStore(point.Security)
	if r == nil {
		t.Fatal("err")
	}
	if r != _securityScriptStore {
		t.Fatal("not equal")
	}

	r = whichStore(point.UnknownCategory)
	if r == nil {
		t.Fatal("err")
	}
	if r != _loggingScriptStore {
		t.Fatal("not equal")
	}
}

func TestPlDirStruct(t *testing.T) {
	bPath := fmt.Sprintf("/tmp/%d/pipeline/", time.Now().UnixNano())
	_ = os.MkdirAll(bPath, os.FileMode(0o755))

	expt := map[point.Category]map[string]string{}
	for category, dirName := range CategoryDirName() {
		if _, ok := expt[category]; !ok {
			expt[category] = map[string]string{}
		}
		expt[category][dirName+"-xx.p"] = filepath.Join(bPath, dirName, dirName+"-xx.p")
	}

	_ = os.WriteFile(filepath.Join(bPath, "nginx-xx.p"), []byte(`
	json(_, time)
	set_tag(bb, "aa0")
	default_time(time)
	`), os.FileMode(0o755))

	expt[point.Logging]["nginx-xx.p"] = filepath.Join(bPath, "nginx-xx.p")

	for _, dirName := range CategoryDirName() {
		_ = os.MkdirAll(filepath.Join(bPath, dirName), os.FileMode(0o755))
		_ = os.WriteFile(filepath.Join(bPath, dirName, dirName+"-xx.p"), []byte(`
		json(_, time)
		set_tag(bb, "aa0")
		default_time(time)
		`), os.FileMode(0o755))
	}
	act := SearchPlFilePathFromPlStructPath(bPath)

	assert.Equal(t, expt, act)
}
