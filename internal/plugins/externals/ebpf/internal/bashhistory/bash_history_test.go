//go:build linux
// +build linux

package bashhistory

import (
	"testing"
	"time"
)

func TestReadlineCallBackDropsShortRecord(t *testing.T) {
	tracer := NewBashTracer()

	done := make(chan struct{})
	go func() {
		tracer.readlineCallBack(0, nil, nil, nil)
		tracer.readlineCallBack(0, make([]byte, bashEventSize-1), nil, nil)
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("readlineCallBack blocked on short record")
	}
}
