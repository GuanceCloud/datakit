// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package dialtesting

import (
	"math"
	"net"
	"time"
)

const (
	MaxTimeout = 10 * time.Second
	MaxHops    = 60
	MaxRetry   = 3
)

// TracerouteOption represent traceroute option.
type TracerouteOption struct {
	Hops    int
	Retry   int
	Timeout string

	timeout time.Duration
}

// Response for sent packet, may be failed response when timeout.
type Response struct {
	From         net.IP
	ResponseTime time.Duration

	fail bool
}

// RouteItem  represent each retry response.
type RouteItem struct {
	IP           string  `json:"ip"`
	ResponseTime float64 `json:"response_time"`
}

// Route is summary for each hop.
type Route struct {
	Total   int          `json:"total"`
	Failed  int          `json:"failed"`
	Loss    float64      `json:"loss"`
	AvgCost float64      `json:"avg_cost"`
	MinCost float64      `json:"min_cost"`
	MaxCost float64      `json:"max_cost"`
	StdCost float64      `json:"std_cost"`
	Items   []*RouteItem `json:"items"`
}

// Packet represent sent packet.
type Packet struct {
	ID  int
	Dst net.IP

	startTime time.Time
}

func mean(v []float64) float64 {
	var res float64 = 0
	var n int = len(v)

	if n == 0 {
		return 0
	}

	for i := 0; i < n; i++ {
		res += v[i]
	}
	return res / float64(n)
}

func variance(v []float64) float64 {
	var res float64 = 0
	m := mean(v)
	var n int = len(v)
	for i := 0; i < n; i++ {
		res += (v[i] - m) * (v[i] - m)
	}
	if n <= 1 {
		return 0
	}
	return res / float64(n-1)
}

func std(v []float64) float64 {
	if len(v) == 0 {
		return 0
	}
	return math.Sqrt(variance(v))
}
