// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package winnetflow

import (
	"net/url"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
	"unicode/utf8"

	"github.com/GuanceCloud/cliutils/point"
)

const (
	httpflowMetricName = "httpflow"

	// Keep the default generous, but cap pathological configurations because
	// request keys can include multi-KiB paths and are duplicated during flush.
	defaultMaxHTTPRequests = 65536
	minMaxHTTPRequests     = 1024
	maxMaxHTTPRequests     = 131072

	defaultHTTPPathLimit = 256
	minHTTPPathLimit     = 16
	maxHTTPPathLimit     = 4096
)

// httpEvent is a completed HTTP transaction produced by the ETW layer. For
// HTTP.sys the collector is always the server side: src is the local (server)
// endpoint and dst is the remote (client) endpoint, direction is "incoming".
type httpEvent struct {
	ts           time.Time // completion time
	start        time.Time // request received time
	family       string
	pid          uint32
	srcIP        string
	srcPort      uint32
	dstIP        string
	dstPort      uint32
	method       string
	path         string
	status       uint16
	bytesWritten uint64
	truncated    bool
}

type httpFlowKey struct {
	family  string
	pid     uint32
	srcIP   string
	srcPort uint32
	dstIP   string
	dstPort uint32
	method  string
	path    string
	status  int
}

type httpFlowStats struct {
	count        int
	latencySumNS int64 // sum of request durations in nanoseconds
	bytesRead    uint64
	bytesWritten uint64
	truncated    bool
}

// httpAggregator groups completed HTTP transactions per interval and flushes
// them as httpflow points with the same schema as the Linux eBPF httpflow
// collector (count, average latency, path/method/status grouping).
type httpAggregator struct {
	mu               sync.Mutex
	data             map[httpFlowKey]*httpFlowStats
	maxRequests      int
	requestsSkipped  uint64
	invalidFiltered  atomic.Uint64
	loopbackFiltered atomic.Uint64
	includeLoopback  bool // test-only escape hatch for local ETW smoke tests
	extraTags        map[string]string
	warmProcessName  func(uint32)
	warmPIDs         map[uint32]struct{}
}

func newHTTPAggregator(maxRequests int) *httpAggregator {
	return &httpAggregator{
		data:        make(map[httpFlowKey]*httpFlowStats),
		maxRequests: maxRequests,
		warmPIDs:    make(map[uint32]struct{}),
	}
}

func (a *httpAggregator) onEvent(ev *httpEvent) {
	if ev == nil {
		return
	}
	switch filterEndpoints(ev.srcIP, ev.srcPort, ev.dstIP, ev.dstPort, a.includeLoopback) {
	case endpointAccepted:
	case endpointInvalid:
		a.invalidFiltered.Add(1)
		return
	case endpointLoopback:
		a.loopbackFiltered.Add(1)
		return
	}

	srcPort, dstPort := normalizeClientPort(directionIncoming, ev.srcPort, ev.dstPort)

	key := httpFlowKey{
		family:  ev.family,
		pid:     ev.pid,
		srcIP:   ev.srcIP,
		srcPort: srcPort,
		dstIP:   ev.dstIP,
		dstPort: dstPort,
		method:  ev.method,
		path:    ev.path,
		status:  int(ev.status),
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	st := a.data[key]
	if st == nil {
		if a.maxRequests > 0 && len(a.data) >= a.maxRequests {
			a.requestsSkipped++
			return
		}
		st = &httpFlowStats{}
		a.data[key] = st
		a.warmPIDLocked(ev.pid)
	}

	st.count++
	if latency := ev.ts.Sub(ev.start); latency > 0 {
		st.latencySumNS += latency.Nanoseconds()
	}
	st.bytesWritten += ev.bytesWritten
	st.truncated = st.truncated || ev.truncated
}

func (a *httpAggregator) warmPIDLocked(pid uint32) {
	if a.warmProcessName == nil || pid == 0 {
		return
	}
	if _, warmed := a.warmPIDs[pid]; warmed {
		return
	}
	a.warmPIDs[pid] = struct{}{}
	a.warmProcessName(pid)
}

// flush snapshots and resets the aggregator, producing httpflow points.
func (a *httpAggregator) flush() []*point.Point {
	a.mu.Lock()
	if len(a.data) == 0 {
		a.mu.Unlock()
		return nil
	}

	ts := time.Now()
	data := a.data
	a.data = make(map[httpFlowKey]*httpFlowStats, boundedMapHint(len(a.data), defaultMaxHTTPRequests))
	a.warmPIDs = make(map[uint32]struct{}, processNameCacheHint(len(a.warmPIDs)))
	a.mu.Unlock()

	pts := make([]*point.Point, 0, len(data))
	processNames := make(map[uint32]string, processNameCacheHint(len(data)))
	for key, st := range data {
		fields := map[string]any{
			"method":        key.method,
			"http_version":  "",
			"path":          key.path,
			"status_code":   int64(key.status),
			"latency":       avgLatencyNS(st.latencySumNS, st.count),
			"bytes_read":    int64(st.bytesRead),
			"bytes_written": int64(st.bytesWritten),
			"truncated":     st.truncated,
			"count":         int64(st.count),
		}

		tags := map[string]string{
			"family":       key.family,
			"direction":    directionIncoming,
			"transport":    "tcp",
			"pid":          strconv.FormatUint(uint64(key.pid), 10),
			"src_ip":       key.srcIP,
			"dst_ip":       key.dstIP,
			"src_port":     flowPortTag(key.srcPort),
			"dst_port":     flowPortTag(key.dstPort),
			"process_name": flushProcessName(processNames, key.pid),
			"src_ip_type":  endpointIPType(key.srcIP),
			"dst_ip_type":  endpointIPType(key.dstIP),
			"dst_nat_ip":   "N/A",
			"dst_nat_port": "N/A",
		}
		addClientServerInfo(tags, fields)

		// httpflow carries string fields (method/path), so use the logging
		// option set like the Linux eBPF httpflow exporter does.
		opts := point.CommonLoggingOptions()
		opts = append(opts, point.WithTime(ts))
		if len(a.extraTags) > 0 {
			opts = append(opts, point.WithExtraTags(a.extraTags))
		}
		pts = append(pts, point.NewPoint(httpflowMetricName,
			append(point.NewTags(tags), point.NewKVs(fields)...), opts...))
	}

	return pts
}

func avgLatencyNS(sum int64, count int) int64 {
	if count == 0 {
		return 0
	}
	return sum / int64(count)
}

func (a *httpAggregator) requestsSkippedCount() uint64 {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.requestsSkipped
}

func (a *httpAggregator) filterCounts() (invalid, loopback uint64) {
	return a.invalidFiltered.Load(), a.loopbackFiltered.Load()
}

// httpRequestPath extracts only the URL path reported by HTTP.sys. Query
// strings are intentionally excluded because they commonly contain secrets
// and unbounded values that create excessive aggregation cardinality.
func httpRequestPath(uri string, limit int) (string, bool) {
	u, err := url.Parse(uri)
	if err != nil {
		u = nil
	}
	path := "/"
	if u != nil {
		if u.Path != "" {
			path = u.Path
		}
	}
	if limit > 0 && len(path) > limit {
		end := limit
		for end > 0 && !utf8.ValidString(path[:end]) {
			end--
		}
		return path[:end], true
	}
	return path, false
}

// httpVerbName maps the HTTP_VERB enumeration (http.h) to the method string.
// Verbs outside the enumeration (PATCH, custom) are recovered from the ASCII
// Verb field of events 4/8.
var httpVerbNames = map[uint32]string{
	3:  "OPTIONS",
	4:  "GET",
	5:  "HEAD",
	6:  "POST",
	7:  "PUT",
	8:  "DELETE",
	9:  "TRACE",
	10: "CONNECT",
	11: "TRACK",
	12: "MOVE",
	13: "COPY",
	14: "PROPFIND",
	15: "PROPPATCH",
	16: "MKCOL",
	17: "LOCK",
	18: "UNLOCK",
	19: "SEARCH",
	20: "QUERY",
}

func httpVerbName(v uint32) string {
	if s, ok := httpVerbNames[v]; ok {
		return s
	}
	return ""
}
