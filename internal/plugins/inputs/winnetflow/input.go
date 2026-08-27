// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package winnetflow collects per-process L4 network flow metrics on Windows
// via Event Tracing for Windows (ETW). It mirrors the measurement schema of
// the Linux eBPF netflow collector so existing dashboards keep working.
package winnetflow

import (
	"fmt"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/GuanceCloud/cliutils"
	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/point"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/export/doc"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/metrics"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

const (
	inputName       = "winnetflow"
	metricName      = "netflow"
	netflowSource   = inputName + "/" + metricName
	httpflowSource  = inputName + "/" + httpflowMetricName
	defaultInterval = time.Minute
	minInterval     = time.Minute
	maxInterval     = 5 * time.Minute

	defaultEtwBufferSizeKB = 64
	defaultEtwMinBuffers   = 8
	defaultEtwMaxBuffers   = 256
	minEtwBufferSizeKB     = 4
	maxEtwBufferSizeKB     = 1024
	minEtwBuffers          = 2
	maxEtwBuffers          = 4096
	maxEtwSessionMemoryKB  = 256 * 1024
	collectorRestartDelay  = 5 * time.Second
	collectorMaxRetryDelay = time.Minute
)

var (
	_ inputs.InputV2   = (*Input)(nil)
	_ inputs.ReadEnv   = (*Input)(nil)
	_ inputs.GetENVDoc = (*Input)(nil)
	_ inputs.Singleton = (*Input)(nil)
	l                  = logger.DefaultSLogger(inputName)
)

const sampleConfig = `
[[inputs.winnetflow]]
  # Flow aggregation and report interval.
  interval = "60s"

  # Optional ETW session tuning. Buffer size is in KB; unsafe values are
  # clamped to documented limits and a 256 MiB per-session memory budget.
  # etw_buffer_size_kb = 64
  # etw_min_buffers    = 8
  # etw_max_buffers    = 256

  # Max flows tracked per interval; excess new flows are dropped and counted.
  # max_flows = 65536

  # Collect L7 HTTP request metrics from the HTTP.sys ETW provider (IIS,
  # HttpListener and other HTTP.sys clients) as the "httpflow" measurement.
  # enable_httpflow = true
  # Max in-flight HTTP requests tracked; excess new requests are dropped and
  # counted in the periodic summary.
  # max_http_requests = 65536
  # Max request path length collected; longer paths are truncated and flagged.
  # httpflow_path_limit = 256

  [inputs.winnetflow.tags]
  # some_tag = "some_value"
`

// flowSource is the platform-specific ETW acquisition layer.
type flowSource interface {
	start() error
	stop()
	setOnFatal(func(string))
}

type collectorFailure struct {
	collector flowSource
	http      bool
	message   string
}

type Input struct {
	Interval          string            `toml:"interval"`
	EtwBufferSizeKB   int               `toml:"etw_buffer_size_kb"`
	EtwMinBuffers     int               `toml:"etw_min_buffers"`
	EtwMaxBuffers     int               `toml:"etw_max_buffers"`
	MaxFlows          int               `toml:"max_flows"`
	EnableHTTPFlow    bool              `toml:"enable_httpflow"`
	MaxHTTPRequests   int               `toml:"max_http_requests"`
	HTTPFlowPathLimit int               `toml:"httpflow_path_limit"`
	Tags              map[string]string `toml:"tags"`

	semStop         *cliutils.Sem
	feeder          dkio.Feeder
	tagger          datakit.GlobalTagger
	agg             *flowAggregator
	collector       flowSource
	httpAgg         *httpAggregator
	httpCollector   flowSource
	includeLoopback bool          // enables local-only ETW integration tests
	testInterval    time.Duration // bypasses the production minimum in accelerated ETW tests
	testStarted     chan struct{} // signals that collector startup attempts have completed
}

func NewInput() *Input {
	return &Input{
		semStop:           cliutils.NewSem(),
		feeder:            dkio.DefaultFeeder(),
		tagger:            datakit.DefaultGlobalTagger(),
		Tags:              map[string]string{},
		EnableHTTPFlow:    true,
		MaxHTTPRequests:   defaultMaxHTTPRequests,
		HTTPFlowPathLimit: defaultHTTPPathLimit,
	}
}

func (ipt *Input) Run() {
	l = logger.SLogger(inputName)

	if runtime.GOOS != datakit.OSWindows {
		l.Errorf("%s input is only supported on Windows", inputName)
		return
	}

	interval := ipt.interval()
	if ipt.testInterval > 0 {
		interval = ipt.testInterval
	}
	processWarmer := newProcessNameWarmer()
	defer processWarmer.stop()

	mergedTags := inputs.MergeTags(ipt.tagger.HostTags(), ipt.Tags, "")
	ipt.agg = newFlowAggregator(interval, ipt.maxFlows())
	ipt.agg.includeLoopback = ipt.includeLoopback
	ipt.agg.extraTags = mergedTags
	ipt.agg.warmProcessName = processWarmer.enqueue
	if ipt.EnableHTTPFlow {
		ipt.httpAgg = newHTTPAggregator(ipt.httpMaxRequests())
		ipt.httpAgg.includeLoopback = ipt.includeLoopback
		ipt.httpAgg.extraTags = mergedTags
		ipt.httpAgg.warmProcessName = processWarmer.enqueue
	}

	cfg, adjusted := normalizeETWConfig(ipt.EtwBufferSizeKB, ipt.EtwMinBuffers, ipt.EtwMaxBuffers)
	if adjusted {
		l.Warnf("adjusted ETW config to buffer_size_kb=%d min_buffers=%d max_buffers=%d",
			cfg.bufferSizeKB, cfg.minBuffers, cfg.maxBuffers)
	}
	fatalCh := make(chan collectorFailure, 4)
	reportFatal := func(f collectorFailure) {
		l.Errorf("%s: %s", inputName, f.message)
		select {
		case fatalCh <- f:
		default:
			l.Warnf("%s collector restart already pending", inputName)
		}
	}

	var l4StartedAt, httpStartedAt time.Time
	startL4 := func() error {
		collector, err := newETWCollector(ipt.agg, cfg)
		if err != nil {
			return fmt.Errorf("create ETW collector: %w", err)
		}
		collector.setOnFatal(func(msg string) {
			reportFatal(collectorFailure{collector: collector, message: msg})
		})
		if err := collector.start(); err != nil {
			return fmt.Errorf("start ETW collector: %w", err)
		}
		ipt.collector = collector
		l4StartedAt = time.Now()
		return nil
	}
	startHTTP := func() error {
		collector, err := newHTTPCollector(ipt.httpAgg, cfg, ipt.httpPathLimit(), ipt.httpMaxRequests())
		if err != nil {
			return fmt.Errorf("create HTTP.sys ETW collector: %w", err)
		}
		collector.setOnFatal(func(msg string) {
			reportFatal(collectorFailure{collector: collector, http: true, message: msg})
		})
		if err := collector.start(); err != nil {
			return fmt.Errorf("start HTTP.sys ETW collector: %w", err)
		}
		ipt.httpCollector = collector
		httpStartedAt = time.Now()
		l.Infof("HTTP.sys ETW collector started (httpflow measurement enabled)")
		return nil
	}
	stopCollectors := func() {
		if ipt.httpCollector != nil {
			ipt.httpCollector.stop()
			ipt.httpCollector = nil
		}
		if ipt.collector != nil {
			ipt.collector.stop()
			ipt.collector = nil
		}
	}
	defer func() {
		// Stop producers first so collector.stop can drain every copied ETW
		// event, then publish the final partial interval.
		stopCollectors()
		ipt.flushPoints()
	}()

	var l4RetryC, httpRetryC <-chan time.Time
	l4RetryDelay := collectorRestartDelay
	httpRetryDelay := collectorRestartDelay
	scheduleL4Retry := func() {
		l4RetryC = time.After(l4RetryDelay)
		l4RetryDelay = nextCollectorRetryDelay(l4RetryDelay)
	}
	scheduleHTTPRetry := func() {
		httpRetryC = time.After(httpRetryDelay)
		httpRetryDelay = nextCollectorRetryDelay(httpRetryDelay)
	}

	if err := startL4(); err != nil {
		l.Errorf("failed to start ETW collector: %v", err)
		ipt.feedLastError(netflowSource, err.Error())
		scheduleL4Retry()
	} else {
		l4RetryDelay = collectorRestartDelay
	}

	if ipt.httpAgg != nil {
		if err := startHTTP(); err != nil {
			l.Warnf("%v (httpflow temporarily disabled)", err)
			ipt.feedLastError(httpflowSource, err.Error())
			scheduleHTTPRetry()
		} else {
			httpRetryDelay = collectorRestartDelay
		}
	}

	l.Infof("%s input started, interval=%s", inputName, interval)
	if ipt.testStarted != nil {
		close(ipt.testStarted)
	}

	tick := time.NewTicker(interval)
	defer tick.Stop()

	summaryEvery := int(10 * time.Minute / interval)
	if summaryEvery < 1 {
		summaryEvery = 1
	}
	ticks := 0

	restartL4 := func() {
		if ipt.collector != nil {
			ipt.collector.stop()
			ipt.collector = nil
		}
		if err := startL4(); err != nil {
			l.Errorf("restart ETW collector failed: %v", err)
			ipt.feedLastError(netflowSource, err.Error())
			scheduleL4Retry()
			return
		}
		l4RetryC = nil
		l.Infof("ETW collector restarted")
	}
	restartHTTP := func() {
		if ipt.httpCollector != nil {
			ipt.httpCollector.stop()
			ipt.httpCollector = nil
		}
		if err := startHTTP(); err != nil {
			l.Warnf("restart HTTP.sys ETW collector failed: %v", err)
			ipt.feedLastError(httpflowSource, err.Error())
			scheduleHTTPRetry()
			return
		}
		httpRetryC = nil
	}

	for {
		select {
		case <-datakit.Exit.Wait():
			l.Infof("%s input exiting", inputName)
			return
		case <-ipt.semStop.Wait():
			l.Infof("%s input stopped", inputName)
			return
		case failure := <-fatalCh:
			if failure.http {
				ipt.feedLastError(httpflowSource, failure.message)
				if failure.collector == ipt.httpCollector {
					ipt.httpCollector.stop()
					ipt.httpCollector = nil
					scheduleHTTPRetry()
				}
			} else {
				ipt.feedLastError(netflowSource, failure.message)
				if failure.collector == ipt.collector {
					ipt.collector.stop()
					ipt.collector = nil
					scheduleL4Retry()
				}
			}
		case <-l4RetryC:
			restartL4()
		case <-httpRetryC:
			restartHTTP()
		case <-tick.C:
			ipt.flushPoints()
			// A collector that survives one reporting interval is healthy enough
			// to reset its exponential restart backoff.
			now := time.Now()
			if ipt.collector != nil && collectorSurvivedInterval(l4StartedAt, now, interval) {
				l4RetryDelay = collectorRestartDelay
			}
			if ipt.httpCollector != nil && collectorSurvivedInterval(httpStartedAt, now, interval) {
				httpRetryDelay = collectorRestartDelay
			}

			ticks++
			if ticks%summaryEvery == 0 {
				ipt.logSummary()
			}
		}
	}
}

func (ipt *Input) flushPoints() {
	if ipt.agg != nil {
		pts := ipt.agg.flush()
		if len(pts) > 0 {
			ipt.feedPoints(netflowSource, pts)
		}
	}
	if ipt.httpAgg != nil {
		httpPts := ipt.httpAgg.flush()
		if len(httpPts) > 0 {
			ipt.feedPoints(httpflowSource, httpPts)
		}
	}
}

func (ipt *Input) feedPoints(source string, pts []*point.Point) {
	if err := ipt.feeder.Feed(point.Network, pts,
		dkio.WithSource(source), dkio.WithInput(inputName)); err != nil {
		l.Warnf("feed %d %s points failed: %v", len(pts), source, err)
		ipt.feedLastError(source, err.Error())
	}
}

func (ipt *Input) feedLastError(source, message string) {
	ipt.feeder.FeedLastError(message,
		metrics.WithLastErrorInput(inputName),
		metrics.WithLastErrorSource(source),
		metrics.WithLastErrorCategory(point.Network),
	)
}

func nextCollectorRetryDelay(current time.Duration) time.Duration {
	next := current * 2
	if next > collectorMaxRetryDelay {
		return collectorMaxRetryDelay
	}
	return next
}

func collectorSurvivedInterval(started, now time.Time, interval time.Duration) bool {
	return !started.IsZero() && !now.Before(started.Add(interval))
}

// logSummary reports collector self-telemetry: decoded/dropped events, parse
// errors and live ETW session counters. It is emitted roughly every 10 minutes.
func (ipt *Input) logSummary() {
	if ipt.collector != nil {
		if r, ok := ipt.collector.(interface{ stats() collectorStats }); ok {
			st := r.stats()
			if st.dropped > 0 || st.parseErrors > 0 || st.tcbEvicted > 0 ||
				st.pendingEvicted > 0 || st.flowsSkipped > 0 ||
				st.session.eventsLost > 0 || st.session.realTimeBuffersLost > 0 {
				l.Warnf("%s summary: decoded=%d dropped=%d parse_errors=%d tcb_evicted=%d pending_evicted=%d flows_skipped=%d invalid_filtered=%d loopback_filtered=%d session{events_lost=%d realtime_buffers_lost=%d buffers=%d free=%d}",
					inputName, st.decoded, st.dropped, st.parseErrors, st.tcbEvicted, st.pendingEvicted, st.flowsSkipped, st.invalidFiltered, st.loopbackFiltered,
					st.session.eventsLost, st.session.realTimeBuffersLost,
					st.session.numberOfBuffers, st.session.freeBuffers)
			} else {
				l.Infof("%s summary: decoded=%d dropped=%d parse_errors=%d tcb_evicted=%d pending_evicted=%d flows_skipped=%d invalid_filtered=%d loopback_filtered=%d session{buffers=%d free=%d}",
					inputName, st.decoded, st.dropped, st.parseErrors, st.tcbEvicted, st.pendingEvicted, st.flowsSkipped, st.invalidFiltered, st.loopbackFiltered,
					st.session.numberOfBuffers, st.session.freeBuffers)
			}
		}
	}
	if ipt.httpCollector != nil {
		if r, ok := ipt.httpCollector.(interface{ stats() httpStats }); ok {
			st := r.stats()
			format := "%s http summary: decoded=%d dropped=%d parse_errors=%d completed=%d missed_conn=%d missed_req=%d evicted_req=%d dropped_req=%d requests_skipped=%d invalid_filtered=%d loopback_filtered=%d session{events_lost=%d realtime_buffers_lost=%d buffers=%d free=%d}"
			args := []interface{}{inputName, st.decoded, st.dropped, st.parseErrors, st.completed,
				st.missedConn, st.missedReq, st.evictedReq, st.droppedReq, st.requestsSkipped, st.invalidFiltered, st.loopbackFiltered,
				st.session.eventsLost, st.session.realTimeBuffersLost,
				st.session.numberOfBuffers, st.session.freeBuffers}
			if httpStatsHasAnomaly(st) {
				l.Warnf(format, args...)
			} else {
				l.Infof(format, args...)
			}
		}
	}
}

func httpStatsHasAnomaly(st httpStats) bool {
	return st.dropped > 0 || st.parseErrors > 0 || st.missedConn > 0 ||
		st.missedReq > 0 || st.evictedReq > 0 || st.droppedReq > 0 ||
		st.requestsSkipped > 0 || st.session.eventsLost > 0 ||
		st.session.realTimeBuffersLost > 0
}

func nonZeroInt(v, def int) int {
	if v <= 0 {
		return def
	}
	return v
}

func clampInt(v, min, max int) int {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

// normalizeETWConfig keeps one ETW session inside a predictable memory budget.
// Windows may still raise MinimumBuffers based on CPU count, but explicit user
// settings cannot overflow uint32, invert min/max, or request multi-gigabyte
// sessions that would compete with the rest of Datakit.
func normalizeETWConfig(bufferSizeKB, minBuffers, maxBuffers int) (etwConfig, bool) {
	invalidInput := bufferSizeKB < 0 || minBuffers < 0 || maxBuffers < 0
	bufferSizeKB = nonZeroInt(bufferSizeKB, defaultEtwBufferSizeKB)
	minBuffers = nonZeroInt(minBuffers, defaultEtwMinBuffers)
	maxBuffers = nonZeroInt(maxBuffers, defaultEtwMaxBuffers)
	originalBufferSizeKB := bufferSizeKB
	originalMinBuffers := minBuffers
	originalMaxBuffers := maxBuffers

	bufferSizeKB = clampInt(bufferSizeKB, minEtwBufferSizeKB, maxEtwBufferSizeKB)
	minBuffers = clampInt(minBuffers, minEtwBuffers, maxEtwBuffers)
	maxBuffers = clampInt(maxBuffers, minEtwBuffers, maxEtwBuffers)

	maxByMemory := maxEtwSessionMemoryKB / bufferSizeKB
	if minBuffers > maxByMemory {
		minBuffers = maxByMemory
	}
	if maxBuffers > maxByMemory {
		maxBuffers = maxByMemory
	}
	if maxBuffers < minBuffers {
		maxBuffers = minBuffers
	}

	return etwConfig{
		bufferSizeKB: uint32(bufferSizeKB),
		minBuffers:   uint32(minBuffers),
		maxBuffers:   uint32(maxBuffers),
	}, invalidInput || bufferSizeKB != originalBufferSizeKB || minBuffers != originalMinBuffers || maxBuffers != originalMaxBuffers
}

func (ipt *Input) interval() time.Duration {
	if ipt.Interval == "" {
		return defaultInterval
	}
	d, err := time.ParseDuration(ipt.Interval)
	if err != nil || d <= 0 {
		l.Warnf("invalid interval %q, use default %s", ipt.Interval, defaultInterval)
		return defaultInterval
	}
	if d < minInterval {
		return minInterval
	}
	if d > maxInterval {
		return maxInterval
	}
	return d
}

func (ipt *Input) maxFlows() int {
	v := nonZeroInt(ipt.MaxFlows, defaultMaxFlows)
	if v < minMaxFlows {
		return minMaxFlows
	}
	if v > maxMaxFlows {
		return maxMaxFlows
	}
	return v
}

func (ipt *Input) httpMaxRequests() int {
	v := nonZeroInt(ipt.MaxHTTPRequests, defaultMaxHTTPRequests)
	if v < minMaxHTTPRequests {
		return minMaxHTTPRequests
	}
	if v > maxMaxHTTPRequests {
		return maxMaxHTTPRequests
	}
	return v
}

func (ipt *Input) httpPathLimit() int {
	v := nonZeroInt(ipt.HTTPFlowPathLimit, defaultHTTPPathLimit)
	if v < minHTTPPathLimit {
		return minHTTPPathLimit
	}
	if v > maxHTTPPathLimit {
		return maxHTTPPathLimit
	}
	return v
}

func (ipt *Input) Terminate() {
	if ipt.semStop != nil {
		ipt.semStop.Close()
	}
}

func (*Input) Catalog() string {
	return "network"
}

func (*Input) SampleConfig() string {
	return sampleConfig
}

func (*Input) AvailableArchs() []string {
	return []string{datakit.OSLabelWindows}
}

func (*Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{&netflowMeasurement{}, &httpflowMeasurement{}}
}

func (*Input) Singleton() {}

// GetENVDoc returns the supported env variables for documentation.
func (ipt *Input) GetENVDoc() []*inputs.ENVInfo {
	//nolint:lll
	infos := []*inputs.ENVInfo{
		{FieldName: "Interval"},
		{FieldName: "EtwBufferSizeKB", Type: doc.Int, Example: "64", Desc: "ETW session buffer size in KB (4-1024).", DescZh: "ETW 会话缓冲区大小（KB，范围 4-1024）。"},
		{FieldName: "EtwMinBuffers", Type: doc.Int, Example: "8", Desc: "Minimum number of ETW session buffers (2-4096).", DescZh: "ETW 会话最小缓冲区数量（范围 2-4096）。"},
		{FieldName: "EtwMaxBuffers", Type: doc.Int, Example: "256", Desc: "Maximum ETW buffers, also limited to a 256 MiB session budget.", DescZh: "ETW 会话最大缓冲区数量，同时受每会话 256 MiB 内存预算限制。"},
		{FieldName: "MaxFlows", Type: doc.Int, Example: "65536", Desc: "Maximum concurrent flows tracked per interval; excess flows are dropped and counted.", DescZh: "每个周期跟踪的最大并发流数，超出部分丢弃并计数。"},
		{FieldName: "EnableHTTPFlow", Type: doc.Boolean, Example: "true", Desc: "Enable L7 httpflow collection from the HTTP.sys ETW provider.", DescZh: "是否启用 HTTP.sys ETW Provider 的 L7 httpflow 采集。"},
		{FieldName: "MaxHTTPRequests", Type: doc.Int, Example: "65536", Desc: "Maximum concurrent in-flight HTTP requests tracked; excess requests are dropped and counted.", DescZh: "跟踪的最大并发进行中 HTTP 请求数，超出部分丢弃并计数。"},
		{FieldName: "HTTPFlowPathLimit", Type: doc.Int, Example: "256", Desc: "Maximum collected request path length; longer paths are truncated and flagged.", DescZh: "收集的最大请求路径长度，超长部分被截断并标记。"},
		{FieldName: "Tags", Type: doc.String, Example: "`'tag1=value1,tag2=value2'`"},
	}
	return doc.SetENVDoc("ENV_INPUT_WINNETFLOW_", infos)
}

// ReadEnv reads env vars, mainly for Kubernetes deployments.
func (ipt *Input) ReadEnv(envs map[string]string) {
	// ENV_INPUT_WINNETFLOW_TAGS : "a=b,c=d"
	if tagsStr, ok := envs["ENV_INPUT_WINNETFLOW_TAGS"]; ok {
		tags := config.ParseGlobalTags(tagsStr)
		for k, v := range tags {
			ipt.Tags[k] = v
		}
	}

	// ENV_INPUT_WINNETFLOW_INTERVAL : time.Duration
	if str, ok := envs["ENV_INPUT_WINNETFLOW_INTERVAL"]; ok {
		if d, err := time.ParseDuration(str); err != nil {
			l.Warnf("parse ENV_INPUT_WINNETFLOW_INTERVAL: %s, ignore", err)
		} else {
			ipt.Interval = d.String()
		}
	}

	for envName, field := range map[string]*int{
		"ENV_INPUT_WINNETFLOW_ETW_BUFFER_SIZE_KB":  &ipt.EtwBufferSizeKB,
		"ENV_INPUT_WINNETFLOW_ETW_MIN_BUFFERS":     &ipt.EtwMinBuffers,
		"ENV_INPUT_WINNETFLOW_ETW_MAX_BUFFERS":     &ipt.EtwMaxBuffers,
		"ENV_INPUT_WINNETFLOW_MAX_FLOWS":           &ipt.MaxFlows,
		"ENV_INPUT_WINNETFLOW_MAX_HTTP_REQUESTS":   &ipt.MaxHTTPRequests,
		"ENV_INPUT_WINNETFLOW_HTTPFLOW_PATH_LIMIT": &ipt.HTTPFlowPathLimit,
	} {
		if str, ok := envs[envName]; ok {
			if v, err := strconv.Atoi(str); err != nil || v <= 0 {
				l.Warnf("parse %s: %q, ignore", envName, str)
			} else {
				*field = v
			}
		}
	}

	// ENV_INPUT_WINNETFLOW_ENABLE_HTTPFLOW : bool
	if str, ok := envs["ENV_INPUT_WINNETFLOW_ENABLE_HTTPFLOW"]; ok {
		switch strings.ToLower(strings.TrimSpace(str)) {
		case "true", "1", "yes", "on":
			ipt.EnableHTTPFlow = true
		case "false", "0", "no", "off":
			ipt.EnableHTTPFlow = false
		default:
			l.Warnf("parse ENV_INPUT_WINNETFLOW_ENABLE_HTTPFLOW: %q, ignore", str)
		}
	}
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return NewInput()
	})
}
