// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package plval

import (
	"testing"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"github.com/GuanceCloud/pipeline-go/constants"
)

func TestSetManagerRetiresStateAfterLastLease(t *testing.T) {
	original, _ := GetManager()
	old := NewScriptManager(nil, nil)
	old.LoadScripts(constants.NSRemote, map[point.Category]map[string]string{
		point.Logging: {"old.p": `add_key(version, 1)`},
	}, nil)
	if old.ScriptCount(point.Logging) != 1 {
		t.Fatal("old manager fixture did not load")
	}
	SetManager(old)
	t.Cleanup(func() { SetManager(original) })

	lease, ok := old.Acquire()
	if !ok {
		t.Fatal("could not lease old manager")
	}
	SetManager(NewScriptManager(nil, nil))
	// Retirement must not invalidate state held by the in-flight batch.
	if lease.Manager().ScriptCount(point.Logging) != 1 {
		t.Fatal("old manager state changed before its lease drained")
	}
	lease.Release()

	deadline := time.Now().Add(time.Second)
	for old.ScriptCount(point.Logging) != 0 {
		if time.Now().After(deadline) {
			t.Fatal("old manager state was not cleaned after its lease drained")
		}
		time.Sleep(time.Millisecond)
	}
}
