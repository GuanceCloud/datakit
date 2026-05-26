// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package aggr

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/stretchr/testify/require"

	"github.com/GuanceCloud/cliutils/aggregate"
	"github.com/GuanceCloud/cliutils/point"
)

const (
	stressTracingDataDir       = "/home/songlq/ddtrace/aggr-test/testdata/tracing"
	stressDefaultDatawayURL    = "http://127.0.0.1:9528"
	stressDefaultPprofURL      = "http://127.0.0.1:6060/debug/pprof/heap"
	stressDefaultPprofDir      = "/tmp/aggr-stress-pprof"
	stressDefaultToken         = "tkn_external_aggr"
	stressDefaultReceiverAddr  = "127.0.0.1:19528"
	stressDefaultTargetTPS     = 300
	stressDefaultDuration      = 3 * time.Minute
	stressDefaultMonitorPeriod = time.Second
	stressDefaultMinTraceSpans = 30
	stressDefaultMaxTraceSpans = 80
	stressDefaultReceiverHold  = 2 * time.Minute
	stressPostDisconnectWait   = 2 * time.Minute
)

func TestExternalTracingStress1000PerSec(t *testing.T) {
	if testing.Short() {
		t.Skip("skip stress case in short mode")
	}

	if _, err := os.Stat(stressTracingDataDir); err != nil {
		t.Skipf("stress tracing testdata not found: %s (%v)", stressTracingDataDir, err)
	}

	datawayURL := stressGetEnvOrDefault("AGGR_STRESS_DATAWAY_URL", stressDefaultDatawayURL)
	token := stressGetEnvOrDefault("AGGR_STRESS_TOKEN", stressDefaultToken)
	receiverAddr := stressGetEnvOrDefault("AGGR_STRESS_RECEIVER_ADDR", stressDefaultReceiverAddr)
	pprofURL := stressGetEnvOrDefault("AGGR_STRESS_PPROF_URL", stressDefaultPprofURL)
	pprofDir := stressGetEnvOrDefault("AGGR_STRESS_PPROF_DIR", stressDefaultPprofDir)
	targetTPS := stressGetEnvInt("AGGR_STRESS_TARGET_TPS", stressDefaultTargetTPS)
	sendWorkers := stressGetEnvInt("AGGR_STRESS_SEND_WORKERS", 8)
	duration := stressGetEnvDuration("AGGR_STRESS_DURATION", stressDefaultDuration)
	monitorPeriod := stressGetEnvDuration("AGGR_STRESS_MONITOR_PERIOD", stressDefaultMonitorPeriod)
	receiverHold := stressGetEnvDuration("AGGR_STRESS_RECEIVER_HOLD", stressDefaultReceiverHold)
	postDisconnectWait := stressGetEnvDuration("AGGR_STRESS_POST_DISCONNECT_WAIT", stressPostDisconnectWait)
	datawayPID, err := stressResolveDatawayPID(stressGetEnvInt("AGGR_STRESS_DATAWAY_PID", 0))
	if err != nil {
		t.Skipf("skip stress case: %v", err)
	}

	if !stressEndpointReachable(datawayURL) {
		t.Skipf("dataway not reachable at %s", datawayURL)
	}

	minTraceSpans := stressGetEnvInt("AGGR_STRESS_MIN_TRACE_SPANS", stressDefaultMinTraceSpans)
	maxTraceSpans := stressGetEnvInt("AGGR_STRESS_MAX_TRACE_SPANS", stressDefaultMaxTraceSpans)
	if minTraceSpans < 1 {
		minTraceSpans = stressDefaultMinTraceSpans
	}
	if maxTraceSpans < minTraceSpans {
		maxTraceSpans = minTraceSpans
	}

	receiver := newForwardReceiver(t, receiverAddr, token)
	receiver.decodeTracing = false
	defer receiver.Close()
	t.Logf("stress receiver listening at %s, ensure dataway remote_host points to this URL", receiver.URL())

	monitor, err := newStressProcMonitor(datawayPID, monitorPeriod)
	require.NoError(t, err)
	monitor.Start()

	workersTotal := sendWorkers
	if workersTotal <= 0 {
		workersTotal = 1
	}
	start := time.Now()
	totalBatches, totalPoints, totalTraces, sendElapsed := func() (int, int, int, time.Duration) {
		basePts := stressLoadLPPointsFromDir(t, stressTracingDataDir)
		require.NotEmpty(t, basePts)

		traceTemplates := stressBuildTraceTemplates(basePts)
		baseTraceCount := len(traceTemplates)
		require.Greater(t, baseTraceCount, 0, "invalid tracing testdata: no trace_id")

		workerAggregators := make([]*Aggregator, workersTotal)
		for i := 0; i < workersTotal; i++ {
			workerAggregators[i] = newStressAggregator(t, datawayURL, token)
		}
		defer stressReleaseSenderResources(workerAggregators)

		var (
			totalBatches int
			totalPoints  int
			totalTraces  int
			round        uint64
		)

		deadline := start.Add(duration)
		sec := 0
		for time.Now().Before(deadline) {
			sec++
			secStart := time.Now()
			batchesTarget := (targetTPS + baseTraceCount - 1) / baseTraceCount
			if batchesTarget <= 0 {
				batchesTarget = 1
			}
			workers := sendWorkers
			if workers > batchesTarget {
				workers = batchesTarget
			}
			if workers <= 0 {
				workers = 1
			}

			jobs := make(chan uint64, batchesTarget)
			errCh := make(chan error, 1)
			var wg sync.WaitGroup
			var sentThisSecond int64
			var pointsThisSecond int64

			for i := 0; i < workers; i++ {
				workerAg := workerAggregators[i]
				wg.Add(1)
				go func() {
					defer wg.Done()
					for jobRound := range jobs {
						pts := stressCloneTracingPoints(traceTemplates, jobRound, time.Now(), minTraceSpans, maxTraceSpans)
						res, err := workerAg.Process(point.Tracing, "opentelemetry", pts)
						if err != nil {
							select {
							case errCh <- err:
							default:
							}
							return
						}
						if res == nil {
							select {
							case errCh <- fmt.Errorf("aggr process result is nil"):
							default:
							}
							return
						}
						if !res.Consumed {
							select {
							case errCh <- fmt.Errorf("tracing points were not consumed"):
							default:
							}
							return
						}

						atomic.AddInt64(&sentThisSecond, int64(baseTraceCount))
						atomic.AddInt64(&pointsThisSecond, int64(len(pts)))
					}
				}()
			}

			for i := 0; i < batchesTarget; i++ {
				jobs <- round + uint64(i)
			}
			close(jobs)
			wg.Wait()

			select {
			case err := <-errCh:
				require.NoError(t, err)
			default:
			}

			round += uint64(batchesTarget)
			sentThisSecondInt := int(sentThisSecond)
			totalTraces += sentThisSecondInt
			totalPoints += int(pointsThisSecond)
			totalBatches += batchesTarget

			cost := time.Since(secStart)
			t.Logf("stress sec=%d target_trace=%d sent_trace=%d sent_batch=%d cost=%s",
				sec, targetTPS, sentThisSecondInt, batchesTarget, cost)

			if sleep := time.Second - cost; sleep > 0 {
				time.Sleep(sleep)
			}
		}

		// Release large send-side allocations before hold period to observe post-send memory.
		runtime.GC()
		debug.FreeOSMemory()

		return totalBatches, totalPoints, totalTraces, time.Since(start)
	}()
	if pprofPath, err := stressFetchPprof(pprofURL, pprofDir, "send_done"); err != nil {
		t.Logf("fetch send_done pprof failed: %v", err)
	} else {
		t.Logf("saved send_done pprof: %s", pprofPath)
	}

	if receiverHold > 0 {
		t.Logf("hold stress receiver for %s after sending to avoid downstream send failures", receiverHold)
		time.Sleep(receiverHold)
	}
	if pprofPath, err := stressFetchPprof(pprofURL, pprofDir, "hold_done"); err != nil {
		t.Logf("fetch hold_done pprof failed: %v", err)
	} else {
		t.Logf("saved hold_done pprof: %s", pprofPath)
	}
	t.Logf("close stress receiver at elapsed=%s to simulate remote_host disconnection", time.Since(start).Truncate(time.Second))
	receiver.Close()
	if postDisconnectWait > 0 {
		t.Logf("receiver closed, wait %s before collecting gc heap profile", postDisconnectWait)
		time.Sleep(postDisconnectWait)
	}
	if pprofPath, err := stressFetchPprof(stressHeapPprofURLWithGC(pprofURL), pprofDir, "post_disconnect_gc"); err != nil {
		t.Logf("fetch post_disconnect_gc pprof failed: %v", err)
	} else {
		t.Logf("saved post_disconnect_gc pprof: %s", pprofPath)
	}
	receiver.mu.Lock()
	receiverErrs := append([]string(nil), receiver.errs...)
	receiver.mu.Unlock()
	if len(receiverErrs) > 0 {
		t.Logf("stress receiver observed non-fatal errors: %v", receiverErrs)
	}

	elapsed := time.Since(start)
	summary, err := monitor.Stop()
	require.NoError(t, err)
	require.Greater(t, summary.SampleCount, 0)
	require.Greater(t, summary.RSSMaxBytes, uint64(0))

	avgTraceRateWall := float64(totalTraces) / elapsed.Seconds()
	avgTraceRateSendWindow := float64(totalTraces) / sendElapsed.Seconds()
	t.Logf("stress summary: send_duration=%s total_duration=%s total_trace=%d total_points=%d total_batch=%d avg_trace_rate_send_window=%.2f/s avg_trace_rate_wall=%.2f/s",
		sendElapsed, elapsed, totalTraces, totalPoints, totalBatches, avgTraceRateSendWindow, avgTraceRateWall)
	t.Logf("dataway pid=%d monitor samples=%d cpu_avg=%.2f%% cpu_max=%.2f%% rss_avg=%.2fMB rss_max=%.2fMB rss_min=%.2fMB",
		datawayPID,
		summary.SampleCount,
		summary.CPUAvg,
		summary.CPUMax,
		float64(summary.RSSAvgBytes)/1024.0/1024.0,
		float64(summary.RSSMaxBytes)/1024.0/1024.0,
		float64(summary.RSSMinBytes)/1024.0/1024.0)
	t.Logf("stress receiver summary: %s", receiver.DebugSummary())
	stressPrintObservationReport(t, summary, sendElapsed, receiverHold, elapsed, receiver)

	require.Greater(t, receiver.PointCount(datawayAPIWriteMetric), 0, "missing forwarded /v1/write/metric points")
	require.Greater(t, receiver.FieldCount("trace_total_count"), 0, "missing derived field trace_total_count")
	require.Greater(t, receiver.FieldCount("trace_kept_count"), 0, "missing derived field trace_kept_count")
	require.Greater(t, receiver.MeasurementCount("tail_sampling"), 0, "missing derived measurement tail_sampling")
}

func newStressAggregator(t *testing.T, datawayURL, token string) *Aggregator {
	t.Helper()

	ag := &Aggregator{
		Endpoints:           []string{datawayURL + "?token=" + token},
		tailSamplingConfig:  newStressTailSamplingConfig(),
		tailSamplingEnabled: true,
		MaxRawBodySize:      1024 * 1024,
		Timeout:             5 * time.Second,
		Internal:            time.Second,
	}
	require.NoError(t, ag.tailSamplingConfig.Init())
	ag.initHTTP()
	ag.sendTSConfigToDW()

	return ag
}

func newStressTailSamplingConfig() *aggregate.TailSamplingConfigs {
	return &aggregate.TailSamplingConfigs{
		Version: 1002,
		Tracing: &aggregate.TraceTailSampling{
			DataTTL:  60 * time.Second,
			GroupKey: "trace_id",
			Pipelines: []*aggregate.SamplingPipeline{
				{
					Name:      "drop-never-1",
					Type:      aggregate.PipelineTypeCondition,
					Condition: `{ resource = "/__stress_no_match__/1" }`,
					Action:    aggregate.PipelineActionDrop,
				},
				{
					Name:      "drop-never-2",
					Type:      aggregate.PipelineTypeCondition,
					Condition: `{ resource = "/__stress_no_match__/2" }`,
					Action:    aggregate.PipelineActionDrop,
				},
				{
					Name:      "drop-never-3",
					Type:      aggregate.PipelineTypeCondition,
					Condition: `{ http_method = "TRACE" }`,
					Action:    aggregate.PipelineActionDrop,
				},
				{
					Name:      "drop-never-4",
					Type:      aggregate.PipelineTypeCondition,
					Condition: `{ http_status_code = "599" }`,
					Action:    aggregate.PipelineActionDrop,
				},
				{
					Name:      "drop-never-5",
					Type:      aggregate.PipelineTypeCondition,
					Condition: `{ db_name = "__stress_no_match_db__" }`,
					Action:    aggregate.PipelineActionDrop,
				},
				{
					Name:      "drop-never-6",
					Type:      aggregate.PipelineTypeCondition,
					Condition: `{ service = "__stress_no_match_service__" }`,
					Action:    aggregate.PipelineActionDrop,
				},
				{
					Name:      "drop-never-7",
					Type:      aggregate.PipelineTypeCondition,
					Condition: `{ span_id = "__stress_no_match_span__" }`,
					Action:    aggregate.PipelineActionDrop,
				},
				{
					Name:      "drop-never-8",
					Type:      aggregate.PipelineTypeCondition,
					Condition: `{ trace_id = "__stress_no_match_trace__" }`,
					Action:    aggregate.PipelineActionDrop,
				},
				{
					Name:      "keep-all-last",
					Type:      aggregate.PipelineTypeCondition,
					Condition: `{ 1 = 1 }`,
					Action:    aggregate.PipelineActionKeep,
				},
			},
		},
	}
}

func stressGetEnvOrDefault(key, def string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return def
}

func stressGetEnvInt(key string, def int) int {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return def
	}

	n, err := strconv.Atoi(raw)
	if err != nil || n <= 0 {
		return def
	}
	return n
}

func stressGetEnvDuration(key string, def time.Duration) time.Duration {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return def
	}

	d, err := time.ParseDuration(raw)
	if err != nil || d <= 0 {
		return def
	}
	return d
}

func stressPIDExists(pid int) bool {
	if pid <= 0 {
		return false
	}

	_, err := os.Stat(fmt.Sprintf("/proc/%d", pid))
	return err == nil
}

func stressResolveDatawayPID(explicitPID int) (int, error) {
	if stressPIDExists(explicitPID) {
		return explicitPID, nil
	}

	entries, err := os.ReadDir("/proc")
	if err != nil {
		return 0, fmt.Errorf("read /proc: %w", err)
	}

	found := 0
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		pid, err := strconv.Atoi(entry.Name())
		if err != nil || pid <= 0 {
			continue
		}

		cmdline, err := os.ReadFile(filepath.Join("/proc", entry.Name(), "cmdline"))
		if err != nil || len(cmdline) == 0 {
			continue
		}

		cmd := strings.ReplaceAll(string(cmdline), "\x00", " ")
		if !strings.Contains(cmd, "dataway") {
			continue
		}

		if pid > found {
			found = pid
		}
	}

	if found <= 0 {
		return 0, fmt.Errorf("dataway pid not found from /proc and AGGR_STRESS_DATAWAY_PID is empty")
	}

	return found, nil
}

func stressEndpointReachable(rawURL string) bool {
	u, err := url.Parse(rawURL)
	if err != nil || u.Host == "" {
		return false
	}

	conn, err := net.DialTimeout("tcp", u.Host, 500*time.Millisecond)
	if err != nil {
		return false
	}
	_ = conn.Close()
	return true
}

func stressFetchPprof(pprofURL, pprofDir, stage string) (string, error) {
	if strings.TrimSpace(pprofURL) == "" {
		return "", fmt.Errorf("pprof url is empty")
	}
	if strings.TrimSpace(pprofDir) == "" {
		return "", fmt.Errorf("pprof dir is empty")
	}

	if err := os.MkdirAll(pprofDir, 0o755); err != nil {
		return "", fmt.Errorf("mkdir pprof dir: %w", err)
	}

	resp, err := http.Get(pprofURL) //nolint:gosec
	if err != nil {
		return "", fmt.Errorf("get pprof %s: %w", pprofURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("get pprof %s: status=%s", pprofURL, resp.Status)
	}

	filename := fmt.Sprintf("dataway-heap-%s-%s.pb.gz", stage, time.Now().Format("20060102-150405"))
	path := filepath.Join(pprofDir, filename)
	fp, err := os.Create(path)
	if err != nil {
		return "", fmt.Errorf("create pprof file: %w", err)
	}
	defer fp.Close()

	if _, err := io.Copy(fp, resp.Body); err != nil {
		return "", fmt.Errorf("write pprof file: %w", err)
	}

	return path, nil
}

func stressHeapPprofURLWithGC(rawURL string) string {
	u, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil || u == nil {
		return rawURL
	}

	q := u.Query()
	q.Set("gc", "1")
	u.RawQuery = q.Encode()
	return u.String()
}

func stressLoadLPPointsFromDir(t *testing.T, dir string) []*point.Point {
	t.Helper()

	entries, err := os.ReadDir(dir)
	require.NoError(t, err)

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})

	dec := point.GetDecoder(point.WithDecEncoding(point.LineProtocol))
	defer point.PutDecoder(dec)

	pts := make([]*point.Point, 0, len(entries)*32)
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".lp" {
			continue
		}

		fp := filepath.Join(dir, entry.Name())
		body, err := os.ReadFile(fp)
		require.NoError(t, err)

		filePts, err := dec.Decode(body)
		require.NoErrorf(t, err, "decode lp file failed: %s", fp)
		pts = append(pts, filePts...)
	}

	return pts
}

func stressCountDistinctTraceIDs(pts []*point.Point) int {
	uniq := map[string]struct{}{}
	for _, pt := range pts {
		if pt == nil {
			continue
		}
		traceID, ok := pt.GetS("trace_id")
		if !ok || traceID == "" {
			continue
		}
		uniq[traceID] = struct{}{}
	}
	return len(uniq)
}

type stressTraceTemplate struct {
	TraceID string
	Points  []*point.Point
}

func stressBuildTraceTemplates(base []*point.Point) []stressTraceTemplate {
	groups := map[string][]*point.Point{}
	order := make([]string, 0, 32)

	for _, pt := range base {
		if pt == nil {
			continue
		}
		traceID, ok := pt.GetS("trace_id")
		if !ok || traceID == "" {
			continue
		}
		if _, exists := groups[traceID]; !exists {
			order = append(order, traceID)
		}
		groups[traceID] = append(groups[traceID], pt)
	}

	templates := make([]stressTraceTemplate, 0, len(order))
	for _, traceID := range order {
		templates = append(templates, stressTraceTemplate{
			TraceID: traceID,
			Points:  groups[traceID],
		})
	}

	return templates
}

func stressCloneTracingPoints(templates []stressTraceTemplate, round uint64, now time.Time, minTraceSpans, maxTraceSpans int) []*point.Point {
	if len(templates) == 0 {
		return nil
	}

	totalCap := 0
	spanRange := maxTraceSpans - minTraceSpans + 1
	for idx := range templates {
		targetSpans := minTraceSpans
		if spanRange > 1 {
			targetSpans += int((round + uint64(idx)) % uint64(spanRange))
		}
		totalCap += targetSpans
	}

	pts := make([]*point.Point, 0, totalCap)
	for traceIdx, tpl := range templates {
		if len(tpl.Points) == 0 {
			continue
		}

		targetSpans := minTraceSpans
		if spanRange > 1 {
			targetSpans += int((round + uint64(traceIdx)) % uint64(spanRange))
		}

		traceID := fmt.Sprintf("%s-%016x", tpl.TraceID, round)
		prevSpanID := "0"
		for spanIdx := 0; spanIdx < targetSpans; spanIdx++ {
			src := tpl.Points[spanIdx%len(tpl.Points)]
			if src == nil {
				continue
			}

			clonedMsg := proto.Clone(src.PBPoint())
			clonedPB, ok := clonedMsg.(*point.PBPoint)
			if !ok || clonedPB == nil {
				continue
			}

			pt := point.FromPB(clonedPB)
			ts := now.Add(time.Duration(traceIdx*1000+spanIdx) * time.Microsecond)
			pt.SetTime(ts)
			pt.Set("trace_id", traceID)

			spanID := stressSyntheticSpanID(round, traceIdx, spanIdx)
			pt.Set("span_id", spanID)
			if spanIdx == 0 {
				pt.Set("parent_id", "0")
			} else {
				pt.Set("parent_id", prevSpanID)
			}
			pt.Set("start", ts.UnixNano())
			prevSpanID = spanID

			pts = append(pts, pt)
		}
	}

	return pts
}

func stressSyntheticSpanID(round uint64, traceIdx, spanIdx int) string {
	return fmt.Sprintf("%016x", uint64(traceIdx+1)<<48|uint64(spanIdx+1)<<24|(round&0xffffff))
}

func stressReleaseSenderResources(ags []*Aggregator) {
	for _, ag := range ags {
		if ag == nil {
			continue
		}
		for _, ep := range ag.eps {
			if ep == nil {
				continue
			}
			ep.CloseIdleConnections()
		}
	}
}

type stressProcSample struct {
	Timestamp  time.Time
	CPUPercent float64
	RSSBytes   uint64
}

type stressProcSummary struct {
	SampleCount int
	CPUAvg      float64
	CPUMax      float64
	RSSAvgBytes uint64
	RSSMaxBytes uint64
	RSSMinBytes uint64
	Samples     []stressProcSample
}

type stressProcMonitor struct {
	pid      int
	interval time.Duration

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	mu      sync.Mutex
	samples []stressProcSample
	err     error
}

func newStressProcMonitor(pid int, interval time.Duration) (*stressProcMonitor, error) {
	if pid <= 0 {
		return nil, fmt.Errorf("invalid pid %d", pid)
	}
	if interval <= 0 {
		interval = time.Second
	}

	_, _, _, err := stressReadProcAndSystemTicks(pid) //nolint
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())
	return &stressProcMonitor{
		pid:      pid,
		interval: interval,
		ctx:      ctx,
		cancel:   cancel,
		samples:  make([]stressProcSample, 0, 64),
	}, nil
}

func (m *stressProcMonitor) Start() {
	m.wg.Add(1)
	go m.loop()
}

func (m *stressProcMonitor) Stop() (stressProcSummary, error) {
	m.cancel()
	m.wg.Wait()

	m.mu.Lock()
	defer m.mu.Unlock()

	summary := stressProcSummary{
		SampleCount: len(m.samples),
	}
	if len(m.samples) == 0 {
		return summary, m.err
	}

	var (
		cpuSum float64
		rssSum uint64
	)
	summary.RSSMinBytes = m.samples[0].RSSBytes
	for _, s := range m.samples {
		cpuSum += s.CPUPercent
		rssSum += s.RSSBytes
		if s.CPUPercent > summary.CPUMax {
			summary.CPUMax = s.CPUPercent
		}
		if s.RSSBytes > summary.RSSMaxBytes {
			summary.RSSMaxBytes = s.RSSBytes
		}
		if s.RSSBytes < summary.RSSMinBytes {
			summary.RSSMinBytes = s.RSSBytes
		}
	}
	summary.CPUAvg = cpuSum / float64(len(m.samples))
	summary.RSSAvgBytes = rssSum / uint64(len(m.samples))
	summary.Samples = append(summary.Samples, m.samples...)

	return summary, m.err
}

func stressPrintObservationReport(t *testing.T, summary stressProcSummary, sendElapsed, receiverHold,
	elapsed time.Duration, receiver *forwardReceiver,
) {
	t.Helper()

	if len(summary.Samples) == 0 {
		t.Logf("stress observation: no monitor samples collected\n")
		return
	}

	checkpoints := []time.Duration{
		0,
		30 * time.Second,
		60 * time.Second,
		90 * time.Second,
		120 * time.Second,
		150 * time.Second,
		sendElapsed,
		sendElapsed + 30*time.Second,
		sendElapsed + 60*time.Second,
		sendElapsed + 90*time.Second,
		sendElapsed + receiverHold,
		elapsed,
	}

	seen := map[time.Duration]struct{}{}
	uniq := make([]time.Duration, 0, len(checkpoints))
	for _, cp := range checkpoints {
		if cp < 0 || cp > elapsed {
			continue
		}
		if _, ok := seen[cp]; ok {
			continue
		}
		seen[cp] = struct{}{}
		uniq = append(uniq, cp)
	}

	t.Logf("\n==== Dataway Observation Report ====\n")
	t.Logf("time\tstage\trss_mb\tcpu_pct\n")
	for _, cp := range uniq {
		sample := stressNearestSample(summary.Samples, cp)
		t.Logf("%s\t%s\t%.2f\t%.2f\n",
			cp.Truncate(time.Second),
			stressObservationStage(cp, sendElapsed, receiverHold),
			float64(sample.RSSBytes)/1024.0/1024.0,
			sample.CPUPercent)
	}
	t.Logf("summary\tsamples=%d rss_avg=%.2f rss_max=%.2f rss_min=%.2f cpu_avg=%.2f cpu_max=%.2f\n",
		summary.SampleCount,
		float64(summary.RSSAvgBytes)/1024.0/1024.0,
		float64(summary.RSSMaxBytes)/1024.0/1024.0,
		float64(summary.RSSMinBytes)/1024.0/1024.0,
		summary.CPUAvg,
		summary.CPUMax)
	t.Logf("receiver\ttracing_requests=%d tracing_points=%d metric_requests=%d metric_points=%d tail_sampling_points=%d\n",
		receiver.RequestCount(point.URLTracing),
		receiver.PointCount(point.URLTracing),
		receiver.RequestCount(datawayAPIWriteMetric),
		receiver.PointCount(datawayAPIWriteMetric),
		receiver.MeasurementCount("tail_sampling"))
	t.Logf("==== End Observation Report ====\n\n")
}

func stressNearestSample(samples []stressProcSample, offset time.Duration) stressProcSample {
	base := samples[0].Timestamp
	best := samples[0]
	bestDiff := absDuration(best.Timestamp.Sub(base) - offset)
	for _, sample := range samples[1:] {
		diff := absDuration(sample.Timestamp.Sub(base) - offset)
		if diff < bestDiff {
			best = sample
			bestDiff = diff
		}
	}
	return best
}

func stressObservationStage(offset, sendElapsed, receiverHold time.Duration) string {
	switch {
	case offset == 0:
		return "baseline"
	case offset < sendElapsed:
		return "sending"
	case offset == sendElapsed:
		return "send_done"
	case offset < sendElapsed+receiverHold:
		return "receiver_hold"
	case offset == sendElapsed+receiverHold:
		return "hold_done"
	default:
		return "post_hold"
	}
}

func absDuration(d time.Duration) time.Duration {
	if d < 0 {
		return -d
	}
	return d
}

func (m *stressProcMonitor) loop() {
	defer m.wg.Done()

	prevProc, prevSys, prevRSS, err := stressReadProcAndSystemTicks(m.pid)
	if err != nil {
		m.setErr(err)
		return
	}
	m.append(stressProcSample{
		Timestamp:  time.Now(),
		CPUPercent: 0,
		RSSBytes:   prevRSS,
	})

	ticker := time.NewTicker(m.interval)
	defer ticker.Stop()

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			curProc, curSys, curRSS, err := stressReadProcAndSystemTicks(m.pid)
			if err != nil {
				m.setErr(err)
				return
			}

			dProc := curProc - prevProc
			dSys := curSys - prevSys
			cpu := 0.0
			if dSys > 0 {
				cpu = float64(dProc) / float64(dSys) * 100.0 * float64(runtime.NumCPU())
			}

			m.append(stressProcSample{
				Timestamp:  time.Now(),
				CPUPercent: cpu,
				RSSBytes:   curRSS,
			})

			prevProc = curProc
			prevSys = curSys
		}
	}
}

func (m *stressProcMonitor) append(s stressProcSample) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.samples = append(m.samples, s)
}

func (m *stressProcMonitor) setErr(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.err == nil {
		m.err = err
	}
}

func stressReadProcAndSystemTicks(pid int) (procTicks, sysTicks, rssBytes uint64, err error) {
	procTicks, rssBytes, err = stressReadProcTicksAndRSS(pid)
	if err != nil {
		return 0, 0, 0, err
	}

	sysTicks, err = stressReadSystemTicks()
	if err != nil {
		return 0, 0, 0, err
	}

	return procTicks, sysTicks, rssBytes, nil
}

func stressReadProcTicksAndRSS(pid int) (uint64, uint64, error) {
	body, err := os.ReadFile(fmt.Sprintf("/proc/%d/stat", pid))
	if err != nil {
		return 0, 0, fmt.Errorf("read /proc/%d/stat: %w", pid, err)
	}

	line := strings.TrimSpace(string(body))
	rparen := strings.LastIndex(line, ")")
	if rparen < 0 || rparen+2 >= len(line) {
		return 0, 0, fmt.Errorf("invalid /proc/%d/stat format", pid)
	}

	fields := strings.Fields(line[rparen+2:])
	if len(fields) < 13 {
		return 0, 0, fmt.Errorf("invalid /proc/%d/stat fields len=%d", pid, len(fields))
	}

	utime, err := strconv.ParseUint(fields[11], 10, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("parse utime: %w", err)
	}

	stime, err := strconv.ParseUint(fields[12], 10, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("parse stime: %w", err)
	}

	rss, err := stressReadRSSBytes(pid)
	if err != nil {
		return 0, 0, err
	}

	return utime + stime, rss, nil
}

func stressReadRSSBytes(pid int) (uint64, error) {
	body, err := os.ReadFile(fmt.Sprintf("/proc/%d/status", pid))
	if err != nil {
		return 0, fmt.Errorf("read /proc/%d/status: %w", pid, err)
	}

	for _, line := range strings.Split(string(body), "\n") {
		if !strings.HasPrefix(line, "VmRSS:") {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) < 2 {
			break
		}

		kb, err := strconv.ParseUint(parts[1], 10, 64)
		if err != nil {
			return 0, fmt.Errorf("parse VmRSS for pid=%d: %w", pid, err)
		}
		return kb * 1024, nil
	}

	return 0, fmt.Errorf("VmRSS not found for pid=%d", pid)
}

func stressReadSystemTicks() (uint64, error) {
	body, err := os.ReadFile("/proc/stat")
	if err != nil {
		return 0, fmt.Errorf("read /proc/stat: %w", err)
	}

	firstLine := strings.SplitN(string(body), "\n", 2)[0]
	fields := strings.Fields(firstLine)
	if len(fields) < 2 || fields[0] != "cpu" {
		return 0, fmt.Errorf("invalid /proc/stat cpu line")
	}

	var total uint64
	for _, f := range fields[1:] {
		v, err := strconv.ParseUint(f, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("parse /proc/stat cpu ticks: %w", err)
		}
		total += v
	}

	return total, nil
}
