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

func TestBashUserCacheIsBounded(t *testing.T) {
	tracer := NewBashTracer()

	for i := 0; i < bashUserCacheLimit+16; i++ {
		tracer.cacheUser(uint32(i), "user")
	}

	tracer.userMu.RLock()
	defer tracer.userMu.RUnlock()
	if got := len(tracer.userCache); got > bashUserCacheLimit {
		t.Fatalf("user cache entries = %d, want <= %d", got, bashUserCacheLimit)
	}
	if _, ok := tracer.userCache[uint32(bashUserCacheLimit+15)]; !ok {
		t.Fatal("expected latest uid to stay cached after cap enforcement")
	}
}
