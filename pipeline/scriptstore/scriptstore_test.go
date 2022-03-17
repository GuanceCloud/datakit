// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package scriptstore

import (
	"os"
	"testing"
)

func TestCall(t *testing.T) {
	LoadDefaultDotPScript2Store()
	ReloadAllDefaultDotPScript2Store()
	LoadGitReposDotPScript2Store(nil)
	ReloadAllGitReposDotPScript2Store(nil)
	LoadRemoteDotPScript2Store(nil)
	ReloadAllRemoteDotPScript2Store(nil)
	_ = os.WriteFile("/tmp/nginx-time123.p", []byte(`
	json(_, time)
	set_tag(bb, "aa0")
	default_time(time)
	`), os.FileMode(0o755))
	LoadDotPScript2StoreWithNS("xxxx", []string{"/tmp/nginx-time.p123"}, "")
	_ = os.Remove("/tmp/nginx-time123.p")
	LoadDotPScript2StoreWithNS("xxx", nil, "")
}

func TestPlScriptStore(t *testing.T) {
	store := &DotPScriptStore{
		scripts: map[string]map[string]*ScriptInfo{},
	}

	err := store.appendScript(DefaultScriptNS, "abc.p", "default_time(time)", true)
	if err != nil {
		t.Error(err)
	}

	err = store.appendScript(GitRepoScriptNS, "abc.p", "default_time(time)", true)
	if err != nil {
		t.Error(err)
	}

	err = store.appendScript(RemoteScriptNS, "abc.p", "default_time(time)", true)
	if err != nil {
		t.Error(err)
	}

	err = store.appendScript(RemoteScriptNS, "abc.p", "default(time)", true)
	if err == nil {
		t.Error("err")
	}

	for _, ns := range plScriptNSSearchOrder {
		sInfo, err := store.queryScript("abc.p", nil)
		if err != nil {
			t.Error(err)
			return
		}
		if sInfo.ns != ns {
			t.Error(sInfo.ns, ns)
		}
		store.cleanAllScriptWithNS(ns)
	}
}
