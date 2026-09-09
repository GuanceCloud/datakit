// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package plval

import (
	"testing"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/offload"
)

func newTestOffloadWorker(t *testing.T) *offload.OffloadWorker {
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

func TestOffloadLeasePinsExactGenerationUntilSendCompletes(t *testing.T) {
	replaceOffloadWorker(nil)
	t.Cleanup(func() { replaceOffloadWorker(nil) })
	old := newTestOffloadWorker(t)
	replaceOffloadWorker(old)
	lease, ok := AcquireOffload()
	if !ok || lease.Worker() != old {
		t.Fatal("did not acquire the published offload generation")
	}

	replaceOffloadWorker(nil)
	if err := old.Send(point.Logging, nil); err != nil {
		t.Fatalf("leased offload generation stopped early: %v", err)
	}
	lease.Release()

	deadline := time.Now().Add(time.Second)
	for old.Send(point.Logging, nil) == nil {
		if time.Now().After(deadline) {
			t.Fatal("retired offload generation did not stop after lease release")
		}
		time.Sleep(time.Millisecond)
	}
}

func TestPreparedPlValServicesCommitClearsAndStopsOldOffload(t *testing.T) {
	replaceOffloadWorker(nil)
	t.Cleanup(func() { replaceOffloadWorker(nil) })
	old := newTestOffloadWorker(t)
	replaceOffloadWorker(old)

	published, replaced := (&PreparedPlValServices{replaceOffload: true}).Commit()
	if !replaced || published != nil {
		t.Fatalf("unexpected commit result worker=%p replaced=%v", published, replaced)
	}
	if _, ok := GetOffload(); ok {
		t.Fatal("disabled candidate retained the old offload worker")
	}
	if err := old.Send(point.Logging, nil); err == nil {
		t.Fatal("retired offload worker still accepted data")
	}
}

func TestPreparedPlValServicesInvalidOffloadPreservesCurrent(t *testing.T) {
	replaceOffloadWorker(nil)
	t.Cleanup(func() { replaceOffloadWorker(nil) })
	old := newTestOffloadWorker(t)
	replaceOffloadWorker(old)

	published, replaced := (&PreparedPlValServices{replaceOffload: false}).Commit()
	if replaced || published != nil {
		t.Fatalf("unexpected commit result worker=%p replaced=%v", published, replaced)
	}
	current, ok := GetOffload()
	if !ok || current != old {
		t.Fatal("invalid replacement changed the current offload worker")
	}
}
