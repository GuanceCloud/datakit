package exporter

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	dto "github.com/prometheus/client_model/go"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/pkg/stats"
)

const (
	summaryLogName   = "ebpf_monitor"
	summaryInputName = "ebpf-monitor"
	summaryWarnLevel = "warn"
)

type summarySnapshot struct {
	runtimeMemory     map[string]float64
	runtimeRatio      map[string]float64
	goroutines        float64
	senderQueueLength float64
	aggEntries        map[string]float64
	cacheEntries      map[string]float64
	nicGroupCount     map[string]float64
	nicGroupRouteCnt  map[string]float64
	perfLostTotal     map[string]float64
	perfReadErrTotal  map[string]float64
	tpacketStatsTotal map[string]float64
	bpfMapFillRatio   map[string]float64
	senderRequests    map[string]float64
	kernelFunctions   []KernelFunctionStatus
	kernelFnFailed    []KernelFunctionStatus
}

func StartSummaryLogger(ctx context.Context, interval time.Duration, tags map[string]string) {
	if ctx == nil {
		return
	}
	if interval <= 0 {
		interval = time.Minute
	}

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		var prev *summarySnapshot
		for {
			select {
			case <-ticker.C:
				cur, err := collectSummarySnapshot()
				if err != nil {
					log.Warnf("collect summary snapshot failed: %v", err)
					continue
				}
				if pt := buildSummaryPoint(cur, prev, tags); pt != nil {
					if err := FeedPoint(summaryInputName, point.Logging, []*point.Point{pt}); err != nil {
						log.Warnf("feed summary log failed: %v", err)
					}
				}
				prev = cur
			case <-ctx.Done():
				return
			}
		}
	}()
}

func collectSummarySnapshot() (*summarySnapshot, error) {
	mfs, err := stats.GetRegistry().Gather()
	if err != nil {
		return nil, err
	}

	s := &summarySnapshot{
		runtimeMemory:     map[string]float64{},
		runtimeRatio:      map[string]float64{},
		aggEntries:        map[string]float64{},
		cacheEntries:      map[string]float64{},
		nicGroupCount:     map[string]float64{},
		nicGroupRouteCnt:  map[string]float64{},
		perfLostTotal:     map[string]float64{},
		perfReadErrTotal:  map[string]float64{},
		tpacketStatsTotal: map[string]float64{},
		bpfMapFillRatio:   map[string]float64{},
		senderRequests:    map[string]float64{},
	}

	s.kernelFunctions = SnapshotKernelFunctionStatus()
	s.kernelFnFailed = FailedKernelFunctionStatus(s.kernelFunctions)

	for _, mf := range mfs {
		switch mf.GetName() {
		case "dkebpf_runtime_memory_bytes":
			for _, m := range mf.Metric {
				s.runtimeMemory[labelValue(m, "type")] = metricValue(mf, m)
			}
		case "dkebpf_runtime_memory_ratio":
			for _, m := range mf.Metric {
				s.runtimeRatio[labelValue(m, "type")] = metricValue(mf, m)
			}
		case "dkebpf_runtime_goroutines":
			for _, m := range mf.Metric {
				s.goroutines = metricValue(mf, m)
			}
		case "dkebpf_exporter_sender_queue_length":
			for _, m := range mf.Metric {
				s.senderQueueLength = metricValue(mf, m)
			}
		case "dkebpf_exporter_agg_entries":
			for _, m := range mf.Metric {
				s.aggEntries[labelValue(m, "component")] = metricValue(mf, m)
			}
		case "dkebpf_exporter_cache_entries":
			for _, m := range mf.Metric {
				key := labelValue(m, "component") + "/" + labelValue(m, "cache")
				s.cacheEntries[key] = metricValue(mf, m)
			}
		case "dkebpf_exporter_nic_group_count":
			for _, m := range mf.Metric {
				key := labelValue(m, "group") + "/" + labelValue(m, "virtual_nic")
				s.nicGroupCount[key] = metricValue(mf, m)
			}
		case "dkebpf_exporter_nic_group_route_count":
			for _, m := range mf.Metric {
				key := labelValue(m, "group") + "/" + labelValue(m, "virtual_nic")
				s.nicGroupRouteCnt[key] = metricValue(mf, m)
			}
		case "dkebpf_exporter_perf_lost_samples_total":
			for _, m := range mf.Metric {
				key := labelValue(m, "component") + "/" + labelValue(m, "stream")
				s.perfLostTotal[key] = metricValue(mf, m)
			}
		case "dkebpf_exporter_perf_read_errors_total":
			for _, m := range mf.Metric {
				key := labelValue(m, "component") + "/" + labelValue(m, "stream")
				s.perfReadErrTotal[key] = metricValue(mf, m)
			}
		case "dkebpf_exporter_tpacket_stats_total":
			for _, m := range mf.Metric {
				key := labelValue(m, "component") + "/" + labelValue(m, "type")
				s.tpacketStatsTotal[key] = metricValue(mf, m)
			}
		case "dkebpf_exporter_bpf_map_fill_ratio":
			for _, m := range mf.Metric {
				key := labelValue(m, "component") + "/" + labelValue(m, "map")
				s.bpfMapFillRatio[key] = metricValue(mf, m)
			}
		case "dkebpf_exporter_sender_requests_total":
			for _, m := range mf.Metric {
				s.senderRequests[labelValue(m, "result")] = metricValue(mf, m)
			}
		}
	}

	return s, nil
}

func buildSummaryPoint(cur, prev *summarySnapshot, baseTags map[string]string) *point.Point {
	if cur == nil {
		return nil
	}

	tags := cloneTags(baseTags)
	tags["source"] = summaryLogName

	perfLostDelta := diffMap(cur.perfLostTotal, valueMap(prev, func(s *summarySnapshot) map[string]float64 { return s.perfLostTotal }))
	perfErrDelta := diffMap(cur.perfReadErrTotal, valueMap(prev, func(s *summarySnapshot) map[string]float64 { return s.perfReadErrTotal }))
	tpacketDelta := diffMap(cur.tpacketStatsTotal, valueMap(prev, func(s *summarySnapshot) map[string]float64 { return s.tpacketStatsTotal }))
	senderReqDelta := diffMap(cur.senderRequests, valueMap(prev, func(s *summarySnapshot) map[string]float64 { return s.senderRequests }))

	rss := cur.runtimeMemory["rss"]
	hwm := cur.runtimeMemory["hwm"]
	goTotal := cur.runtimeMemory["go_total"]
	heapAlloc := cur.runtimeMemory["heap_alloc"]
	nonGo := cur.runtimeMemory["non_go_estimate"]
	nonGoRatio := cur.runtimeRatio["non_go_vs_rss"]
	aggTotal := sumMap(cur.aggEntries)
	cacheTotal := sumMap(cur.cacheEntries)
	perfLostTotal := sumMap(perfLostDelta)
	perfErrTotal := sumMap(perfErrDelta)
	tpacketDropTotal := sumKeySuffix(tpacketDelta, "/drops")
	tpacketFreezeTotal := sumKeySuffix(tpacketDelta, "/queue_freezes")
	senderErrTotal := sumNonOK(senderReqDelta)

	level := "info"
	var reasons []string
	switch {
	case nonGoRatio >= 0.6:
		level = summaryWarnLevel
		reasons = append(reasons, "non_go_memory_high")
	case nonGoRatio >= 0.4:
		reasons = append(reasons, "non_go_memory_elevated")
	}
	if cur.senderQueueLength > 0 {
		level = summaryWarnLevel
		reasons = append(reasons, "sender_queue_backlog")
	}
	if perfLostTotal > 0 {
		level = summaryWarnLevel
		reasons = append(reasons, "perf_lost")
	}
	if tpacketDropTotal > 0 {
		level = summaryWarnLevel
		reasons = append(reasons, "tpacket_drop")
	}
	if senderErrTotal > 0 {
		level = summaryWarnLevel
		reasons = append(reasons, "sender_error")
	}
	if len(reasons) == 0 {
		reasons = append(reasons, "healthy")
	}
	tags["level"] = level

	msg := buildSummaryMessage(cur, perfLostDelta, perfErrDelta, tpacketDelta, senderReqDelta, level, reasons)

	fields := map[string]any{
		"status":                 level,
		"message":                msg,
		"rss_bytes":              int64(rss),
		"hwm_bytes":              int64(hwm),
		"go_total_bytes":         int64(goTotal),
		"heap_alloc_bytes":       int64(heapAlloc),
		"non_go_bytes":           int64(nonGo),
		"non_go_ratio":           nonGoRatio,
		"goroutines":             int64(cur.goroutines),
		"sender_queue_length":    int64(cur.senderQueueLength),
		"agg_entries_total":      int64(aggTotal),
		"cache_entries_total":    int64(cacheTotal),
		"perf_lost_total":        int64(perfLostTotal),
		"perf_read_errors":       int64(perfErrTotal),
		"tpacket_drops":          int64(tpacketDropTotal),
		"tpacket_freezes":        int64(tpacketFreezeTotal),
		"sender_errors":          int64(senderErrTotal),
		"kernel_function_total":  int64(len(cur.kernelFunctions)),
		"kernel_function_failed": int64(len(cur.kernelFnFailed)),
	}

	kvs := point.NewTags(tags)
	kvs = append(kvs, point.NewKVs(fields)...)
	return point.NewPoint(summaryLogName, kvs, append(point.CommonLoggingOptions(), point.WithTime(time.Now()))...)
}

func buildSummaryMessage(cur *summarySnapshot, perfLost, perfErr, tpacket, senderReq map[string]float64, level string, reasons []string) string {
	payload := map[string]any{
		"schema": "ebpf_monitor/v1",
		"runtime": map[string]any{
			"rss_bytes":           int64(cur.runtimeMemory["rss"]),
			"hwm_bytes":           int64(cur.runtimeMemory["hwm"]),
			"go_total_bytes":      int64(cur.runtimeMemory["go_total"]),
			"heap_alloc_bytes":    int64(cur.runtimeMemory["heap_alloc"]),
			"non_go_bytes":        int64(cur.runtimeMemory["non_go_estimate"]),
			"non_go_ratio":        cur.runtimeRatio["non_go_vs_rss"],
			"goroutines":          int64(cur.goroutines),
			"sender_queue_length": int64(cur.senderQueueLength),
		},
		"aggregator": map[string]any{
			"total_entries": int64(sumMap(cur.aggEntries)),
			"entries":       intMap(cur.aggEntries, false),
		},
		"cache": map[string]any{
			"total_entries": int64(sumMap(cur.cacheEntries)),
			"entries":       intMap(cur.cacheEntries, false),
		},
		"nic_groups": map[string]any{
			"count":       intMap(cur.nicGroupCount, false),
			"route_count": intMap(cur.nicGroupRouteCnt, false),
		},
		"exporter": map[string]any{
			"sender_request_delta": intMap(senderReq, true),
		},
		"perf": map[string]any{
			"lost_delta":       intMap(perfLost, true),
			"read_error_delta": intMap(perfErr, true),
		},
		"tpacket": map[string]any{
			"delta": intMap(tpacket, true),
		},
		"bpf": map[string]any{
			"fill_ratio": floatMap(cur.bpfMapFillRatio, false),
		},
		"kernel_functions": map[string]any{
			"total":       len(cur.kernelFunctions),
			"failed":      len(cur.kernelFnFailed),
			"failed_list": cur.kernelFnFailed,
		},
		"status": map[string]any{
			"level":   level,
			"reasons": reasons,
		},
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Sprintf(`{"schema":"ebpf_monitor/v1","status":{"level":%q,"reasons":["marshal_error:%s"]}}`, level, err)
	}
	return string(data)
}

func cloneTags(src map[string]string) map[string]string {
	dst := map[string]string{}
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

func labelValue(m *dto.Metric, name string) string {
	for _, l := range m.GetLabel() {
		if l.GetName() == name {
			return l.GetValue()
		}
	}
	return ""
}

func metricValue(mf *dto.MetricFamily, m *dto.Metric) float64 {
	switch mf.GetType() { //nolint:exhaustive
	case dto.MetricType_GAUGE:
		return m.GetGauge().GetValue()
	case dto.MetricType_COUNTER:
		return m.GetCounter().GetValue()
	default:
		return 0
	}
}

func valueMap(s *summarySnapshot, getter func(*summarySnapshot) map[string]float64) map[string]float64 {
	if s == nil {
		return nil
	}
	return getter(s)
}

func diffMap(cur, prev map[string]float64) map[string]float64 {
	out := map[string]float64{}
	for k, v := range cur {
		if prev == nil {
			out[k] = v
			continue
		}
		if pv, ok := prev[k]; ok && v >= pv {
			out[k] = v - pv
		} else {
			out[k] = v
		}
	}
	return out
}

func sumMap(m map[string]float64) float64 {
	var total float64
	for _, v := range m {
		total += v
	}
	return total
}

func sumKeySuffix(m map[string]float64, suffix string) float64 {
	var total float64
	for k, v := range m {
		if strings.HasSuffix(k, suffix) {
			total += v
		}
	}
	return total
}

func sumNonOK(m map[string]float64) float64 {
	var total float64
	for k, v := range m {
		if k != "ok" {
			total += v
		}
	}
	return total
}

func intMap(m map[string]float64, skipZero bool) map[string]int64 {
	if len(m) == 0 {
		return map[string]int64{}
	}
	out := make(map[string]int64, len(m))
	for k, v := range m {
		if skipZero && v == 0 {
			continue
		}
		out[k] = int64(v)
	}
	return out
}

func floatMap(m map[string]float64, skipZero bool) map[string]float64 {
	if len(m) == 0 {
		return map[string]float64{}
	}
	out := make(map[string]float64, len(m))
	for k, v := range m {
		if skipZero && v == 0 {
			continue
		}
		out[k] = v
	}
	return out
}
