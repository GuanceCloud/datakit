package dialtesting

import (
	"math"
	"net"
	"time"
)

// max config
const MAX_TIMEOUT = 10 * time.Second
const MAX_HOPS = 60
const MAX_RETRY = 3

// traceroute option
type TracerouteOption struct {
	Hops    int
	Retry   int
	Timeout string

	timeout time.Duration
}

// response for sent packet, may be failed response when timeout
type Response struct {
	From         net.IP
	ResponseTime time.Duration

	fail bool
}

// each retry response
type RouteItem struct {
	IP           string        `json:"ip"`
	ResponseTime time.Duration `json:"response_time"`
}

// route summary for each hop
type Route struct {
	Total   int           `json:"total"`
	Failed  int           `json:"failed"`
	Loss    float64       `json:"loss"`
	AvgCost time.Duration `json:"avg_cost"`
	MinCost time.Duration `json:"min_cost"`
	MaxCost time.Duration `json:"max_cost"`
	StdCost time.Duration `json:"std_cost"`
	Items   []*RouteItem  `json:"items"`
}

// sent packet
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
	var m = mean(v)
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
