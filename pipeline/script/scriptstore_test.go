// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package script

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
)

func TestScriptLoadFunc(t *testing.T) {
	case1 := map[string]map[string]string{
		datakit.Logging: {
			"abcd": "if true {}",
		},
		datakit.Metric: {
			"abc": "if true {}",
			"def": "if true {}",
		},
	}

	CleanAllScript(DefaultScriptNS)
	CleanAllScript(GitRepoScriptNS)
	CleanAllScript(RemoteScriptNS)

	LoadAllScript(DefaultScriptNS, case1)
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
		LoadScript(k, DefaultScriptNS, v)
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
		LoadScript(k, "DefaultScriptNS", v)
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
		LoadScript(k, "aabb", v)
	}
	CleanAllScript("aabb")

	_ = os.WriteFile("/tmp/nginx-xx.p", []byte(`
	json(_, time)
	set_tag(bb, "aa0")
	default_time(time)
	`), os.FileMode(0o755))
	LoadAllScriptThrFilepath(DefaultScriptNS, map[string][]string{datakit.RUM: {"/tmp/nginx-xx.p"}})
	_ = os.Remove("/tmp/nginx-xx.p")
	if _, ok := QueryScript(datakit.RUM, "nginx-xx.p"); !ok {
		t.Error(datakit.RUM, " ", "nginx-xx.p")
	}

	LoadAllDefaultScripts2Store()
	ReloadAllGitReposDotPScript2Store(datakit.Logging, nil)
	_ = os.WriteFile("/tmp/nginx-time123.p", []byte(`
		json(_, time)
		set_tag(bb, "aa0")
		default_time(time)
		`), os.FileMode(0o755))
	LoadDotPScript2Store(datakit.Logging, "xxxx", "", []string{"/tmp/nginx-time.p123"})
	_ = os.Remove("/tmp/nginx-time123.p")
	LoadDotPScript2Store(datakit.Logging, "xxx", "", nil)
}

func TestCmpCategory(t *testing.T) {
	c, dc := datakit.CategoryList()
	c1, dc1 := func() (map[string]struct{}, map[string]struct{}) {
		ret1 := map[string]struct{}{}
		ret2 := map[string]struct{}{}
		for k := range _allCategory {
			ret1[k] = struct{}{}
		}
		for k := range _allDeprecatedCategory {
			ret2[k] = struct{}{}
		}
		return ret1, ret2
	}()

	assert.Equal(t, c, c1)
	assert.Equal(t, dc, dc1)
	assert.Equal(t, c1, func() map[string]struct{} {
		ret := map[string]struct{}{}
		for k := range datakit.CategoryDirName() {
			ret[k] = struct{}{}
		}
		return ret
	}())
}

func TestPlScriptStore(t *testing.T) {
	store := NewScriptStore(datakit.Logging)

	store.indexUpdate(nil)

	err := store.UpdateScriptsWithNS(DefaultScriptNS, map[string]string{"abc.p": "default_time(time) set_tag(a, \"1\")"})
	if err != nil {
		t.Error(err)
	}

	err = store.UpdateScriptsWithNS(DefaultScriptNS, map[string]string{"abc.p": "default_time(time)"})
	if err != nil {
		t.Error(err)
	}

	err = store.UpdateScriptsWithNS(DefaultScriptNS, map[string]string{"abc.p": "default_time(time) set_tag(a, 1)"})
	if err != nil {
		t.Error(err)
	}

	err = store.UpdateScriptsWithNS(DefaultScriptNS, map[string]string{"abc.p": "default_time(time)"})
	if err != nil {
		t.Error(err)
	}

	err = store.UpdateScriptsWithNS(GitRepoScriptNS, map[string]string{"abc.p": "default_time(time)"})
	if err != nil {
		t.Error(err)
	}

	err = store.UpdateScriptsWithNS(RemoteScriptNS, map[string]string{"abc.p": "default_time(time)"})
	if err != nil {
		t.Error(err)
	}

	for i, ns := range plScriptNSSearchOrder {
		store.UpdateScriptsWithNS(ns, nil)
		if i < len(plScriptNSSearchOrder)-1 {
			sInfo, ok := store.Get("abc.p")
			if !ok {
				t.Error(fmt.Errorf("!ok"))
				return
			}
			if sInfo.ns != plScriptNSSearchOrder[i+1] {
				t.Error(sInfo.ns, plScriptNSSearchOrder[i+1])
			}
		} else {
			_, ok := store.Get("abc.p")
			if ok {
				t.Error(fmt.Errorf("shoud not be ok"))
				return
			}
		}
	}
}

func TestWhichStore(t *testing.T) {
	r := whichStore(datakit.Metric)
	if r == nil {
		t.Fatal("err")
	}
	if r != _metricScriptStore {
		t.Fatal("not equal")
	}

	r = whichStore(datakit.MetricDeprecated)
	if r == nil {
		t.Fatal("err")
	}
	if r != _metricScriptStore {
		t.Fatal("not equal")
	}

	r = whichStore(datakit.Network)
	if r == nil {
		t.Fatal("err")
	}
	if r != _networkScriptStore {
		t.Fatal("not equal")
	}

	r = whichStore(datakit.KeyEvent)
	if r == nil {
		t.Fatal("err")
	}
	if r != _keyEventScriptStore {
		t.Fatal("not equal")
	}

	r = whichStore(datakit.Object)
	if r == nil {
		t.Fatal("err")
	}
	if r != _objectScriptStore {
		t.Fatal("not equal")
	}

	r = whichStore(datakit.CustomObject)
	if r == nil {
		t.Fatal("err")
	}
	if r != _customObjectScriptStore {
		t.Fatal("not equal")
	}

	r = whichStore(datakit.Logging)
	if r == nil {
		t.Fatal("err")
	}
	if r != _loggingScriptStore {
		t.Fatal("not equal")
	}

	r = whichStore(datakit.Tracing)
	if r == nil {
		t.Fatal("err")
	}
	if r != _tracingScriptStore {
		t.Fatal("not equal")
	}

	r = whichStore(datakit.RUM)
	if r == nil {
		t.Fatal("err")
	}
	if r != _rumScriptStore {
		t.Fatal("not equal")
	}

	r = whichStore(datakit.Security)
	if r == nil {
		t.Fatal("err")
	}
	if r != _securityScriptStore {
		t.Fatal("not equal")
	}

	r = whichStore(datakit.HeartBeat)
	if r == nil {
		t.Fatal("err")
	}

	r = whichStore("")
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

	expt := map[string]map[string]string{}
	for category, dirName := range datakit.CategoryDirName() {
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

	expt[datakit.Logging]["nginx-xx.p"] = filepath.Join(bPath, "nginx-xx.p")

	for _, dirName := range datakit.CategoryDirName() {
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
