// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package offload

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/GuanceCloud/cliutils/point"
)

type countingReceiver struct{ calls atomic.Int32 }

func (receiver *countingReceiver) Send(uint64, point.Category, []*point.Point) error {
	receiver.calls.Add(1)
	return nil
}

func TestOffloadStopIsIdempotentAndUnblocksSend(t *testing.T) {
	worker := &OffloadWorker{
		ch:       newDataChan(),
		stopChan: make(chan struct{}),
		sender:   &countingReceiver{},
	}
	worker.Stop()
	worker.Stop()

	done := make(chan error, 1)
	go func() { done <- worker.Send(point.Logging, []*point.Point{nil}) }()
	select {
	case err := <-done:
		if err == nil {
			t.Fatal("send on stopped worker succeeded")
		}
	case <-time.After(time.Second):
		t.Fatal("send on stopped worker blocked")
	}
}

func TestOffloadCustomerStopsAndFlushesOnce(t *testing.T) {
	receiver := &countingReceiver{}
	worker := &OffloadWorker{
		ch:       newDataChan(),
		stopChan: make(chan struct{}),
		sender:   receiver,
	}
	done := make(chan error, 1)
	go func() { done <- worker.Customer(context.Background(), point.Logging) }()
	worker.Stop()
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(time.Second):
		t.Fatal("customer did not stop")
	}
	if got := receiver.calls.Load(); got != 1 {
		t.Fatalf("flush calls = %d, want 1", got)
	}
}
