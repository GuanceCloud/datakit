// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package pipeline

import (
	"context"
	"testing"

	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/offload"
	plval "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/plval"
)

func newSnapshotOffloader(t *testing.T) *offload.OffloadWorker {
	t.Helper()
	worker, err := offload.NewOffloader(&offload.OffloadConfig{
		Receiver:  offload.PlOffloadRcv,
		Addresses: []string{"http://127.0.0.1:1"},
	})
	if err != nil {
		t.Fatal(err)
	}
	return worker
}

func TestRunPlWithoutScriptsReleasesCapturedOffload(t *testing.T) {
	originalManager, _ := plval.GetManager()
	worker := newSnapshotOffloader(t)
	jitInitMu.Lock()
	plval.SetManager(plval.NewScriptManager(nil, nil))
	plval.SetOffload(worker)
	jitInitMu.Unlock()
	t.Cleanup(func() {
		jitInitMu.Lock()
		plval.SetManager(originalManager)
		plval.SetOffload(nil)
		jitInitMu.Unlock()
	})

	result, err := RunPl(point.Logging, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	result.Release()
	plval.SetOffload(nil)
	if err := worker.Send(point.Logging, nil); err == nil {
		t.Fatal("empty-script return retained the retired offload generation")
	}
}

func TestPipelineLeaseCapturesManagerAndOffloadFromOnePublication(t *testing.T) {
	originalManager, _ := plval.GetManager()
	oldManager := plval.NewScriptManager(nil, nil)
	newManager := plval.NewScriptManager(nil, nil)
	oldExpected, _ := oldManager.Acquire()
	oldRaw := oldExpected.Manager()
	oldExpected.Release()
	newExpected, _ := newManager.Acquire()
	newRaw := newExpected.Manager()
	newExpected.Release()
	oldOffload := newSnapshotOffloader(t)
	newOffload := newSnapshotOffloader(t)

	jitInitMu.Lock()
	plval.SetManager(oldManager)
	plval.SetOffload(oldOffload)
	jitInitMu.Unlock()
	t.Cleanup(func() {
		jitInitMu.Lock()
		plval.SetManager(originalManager)
		plval.SetOffload(nil)
		jitInitMu.Unlock()
	})

	manager, runner, offloadLease, _, _, err := acquirePipelineLeasesContext(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer manager.Release()
	if runner != nil {
		defer runner.release()
	}
	if manager.Manager() != oldRaw || offloadLease == nil || offloadLease.Worker() != oldOffload {
		t.Fatal("pipeline acquisition did not capture the old manager/offload publication")
	}

	jitInitMu.Lock()
	plval.SetManager(newManager)
	plval.SetOffload(newOffload)
	jitInitMu.Unlock()
	if manager.Manager() != oldRaw || offloadLease.Worker() != oldOffload {
		t.Fatal("published replacement changed an in-flight manager/offload snapshot")
	}
	offloadLease.Release()
	offloadLease = nil

	manager2, runner2, offloadLease2, _, _, err := acquirePipelineLeasesContext(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer manager2.Release()
	if runner2 != nil {
		defer runner2.release()
	}
	if offloadLease2 != nil {
		defer offloadLease2.Release()
	}
	if manager2.Manager() != newRaw || offloadLease2 == nil || offloadLease2.Worker() != newOffload {
		t.Fatal("new acquisition did not capture the replacement manager/offload publication")
	}
}
