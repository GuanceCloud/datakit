// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package httpcli

import (
	"fmt"
	"net"
	"net/http"
	"sync/atomic"
)

type ConnWatcher struct {
	New      int64
	Close    int64
	Max      int64
	Idle     int64
	Active   int64
	Hijacked int64
}

func (cw *ConnWatcher) OnStateChange(conn net.Conn, state http.ConnState) {
	switch state {
	case http.StateNew:
		atomic.AddInt64(&cw.New, 1)
		atomic.AddInt64(&cw.Max, 1)

	case http.StateHijacked:
		atomic.AddInt64(&cw.Hijacked, 1)
		atomic.AddInt64(&cw.New, -1)

	case http.StateClosed:
		atomic.AddInt64(&cw.New, -1)
		atomic.AddInt64(&cw.Close, 1)

	case http.StateIdle:
		atomic.AddInt64(&cw.Idle, 1)

	case http.StateActive:
		atomic.AddInt64(&cw.Active, 1)
	}
}

func (cw *ConnWatcher) Count() int {
	return int(atomic.LoadInt64(&cw.New))
}

func (cw *ConnWatcher) String() string {
	return fmt.Sprintf("connections: new: %d, closed: %d, Max: %d, hijacked: %d, idle: %d, active: %d",
		cw.New, cw.Close, cw.Max, cw.Hijacked, cw.Idle, cw.Active)
}
