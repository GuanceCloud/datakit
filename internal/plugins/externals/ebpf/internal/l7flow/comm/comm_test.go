//go:build linux
// +build linux

// Package comm stores connection information
package comm

import (
	"bytes"
	"encoding/json"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDump(t *testing.T) {
	d := NetwrkData{
		Conn: ConnectionInfo{
			Saddr: [4]uint32{1},
		},
		Payload: []byte("assss\n"),
	}
	d.Payload = append(d.Payload, 11)

	s, _ := json.Marshal(d)
	t.Log(string(s))
	d2 := NetwrkData{}
	json.Unmarshal(s, &d2)
	assert.Equal(t, d, d2)
	assert.Equal(t, true, bytes.Equal(d.Payload, d2.Payload))
}

func TestNetwrkDataStringShortPayloadDoesNotPanic(t *testing.T) {
	d := NetwrkData{
		Payload: []byte("12345678901"),
	}

	assert.NotPanics(t, func() {
		_ = d.String()
	})
}

func TestThreadTraceInitializesRandInnerID(t *testing.T) {
	randInnerIDMu.Lock()
	previousRand := randInnerID
	randInnerID = nil
	randInnerIDMu.Unlock()
	t.Cleanup(func() {
		randInnerIDMu.Lock()
		randInnerID = previousRand
		randInnerIDMu.Unlock()
	})

	var trace ThreadTrace
	id := trace.Insert(DIn, 1, [2]int32{2, 0}, 100)
	if id == 0 {
		t.Fatal("expected non-zero inner id")
	}
	if got := trace.GetInnerID(1, [2]int32{2, 0}, 100); got != id {
		t.Fatalf("inner id = %d, want %d", got, id)
	}
}

func TestNextInnerIDConcurrent(t *testing.T) {
	var wg sync.WaitGroup
	for i := 0; i < 32; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = nextInnerID()
		}()
	}
	wg.Wait()
}
