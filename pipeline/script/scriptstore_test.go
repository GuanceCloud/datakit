// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package script

import (
	"fmt"
	"os"
	"testing"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
)

func TestCall(t *testing.T) {
	LoadDefaultDotPScript2Store()
	ReloadAllGitReposDotPScript2Store(datakit.Logging, nil)
	ReloadAllRemoteDotPScript2Store(datakit.Logging, nil)
	_ = os.WriteFile("/tmp/nginx-time123.p", []byte(`
	json(_, time)
	set_tag(bb, "aa0")
	default_time(time)
	`), os.FileMode(0o755))
	LoadDotPScript2Store(datakit.Logging, "xxxx", "", []string{"/tmp/nginx-time.p123"})
	_ = os.Remove("/tmp/nginx-time123.p")
	LoadDotPScript2Store(datakit.Logging, "xxx", "", nil)
}

func TestPlScriptStore(t *testing.T) {
	store := NewScriptStore(datakit.Logging)

	err := store.UpdateScriptsWithNS(DefaultScriptNS, map[string]string{"abc.p": "default_time(time)"})
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
	if r != _heartBeatScriptStore {
		t.Fatal("not equal")
	}

	r = whichStore("")
	if r == nil {
		t.Fatal("err")
	}
	if r != _loggingScriptStore {
		t.Fatal("not equal")
	}
}
