// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package dataway

import (
	"fmt"
	"time"
)

type httpTraceStat struct {
	reuseConn bool
	idle      bool
	idleTime  time.Duration

	dnsStart   time.Time
	dnsResolve time.Duration
	tlsHSStart time.Time
	tlsHSDone  time.Duration
	connStart  time.Time
	connDone   time.Duration
	ttfbTime   time.Duration

	cost time.Duration
}

func (ts *httpTraceStat) String() string {
	if ts == nil {
		return "-"
	}

	return fmt.Sprintf("dataway httptrace: Conn: [reuse: %v,idle: %v/%s], DNS: %s, TLS: %s, Connect: %s, TTFB: %s, cost: %s",
		ts.reuseConn, ts.idle, ts.idleTime, ts.dnsResolve, ts.tlsHSDone, ts.connDone, ts.ttfbTime, ts.cost)
}
