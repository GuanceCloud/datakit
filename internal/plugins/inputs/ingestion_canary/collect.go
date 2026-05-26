// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package ingestioncanary

import (
	"context"
	"fmt"
	"time"

	"github.com/GuanceCloud/cliutils/point"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/goroutine"
	ingestioncanary "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/ingestion_canary"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/compact"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/ntp"
)

const resultMeasurementName = "ingestion_canary_result"

func (ipt *Input) Collect(ptTS int64) error {
	ipt.round++

	g := goroutine.NewGroup(goroutine.Option{Name: inputName})
	for _, fn := range ipt.feedFuncs {
		fn := fn
		g.Go(func(ctx context.Context) error {
			fn()
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		l.Errorf("feed operations failed: %s", err)
		return err
	}

	// Report all result points in batch with unified timestamp
	if len(ipt.resultPoints) > 0 && ipt.ResultWorkspace != "" {
		now := ntp.Now()
		for _, pt := range ipt.resultPoints {
			pt.SetTime(now)
		}
		if err := ipt.reportToWorkspace(ipt.resultPoints...); err != nil {
			l.Errorf("report to workspace failed: %s", err)
		}
		ipt.resultPoints = ipt.resultPoints[:0] // reset result points
	}

	return nil
}

// feedMetric generates and feeds metric data.
func (ipt *Input) feedMetric() {
	ts := time.Now().Truncate(time.Millisecond)

	metricPts := ipt.canary.Metric(ts, ipt.round, ipt.Tags)
	err := ipt.feeder.Feed(point.Metric, []*point.Point{metricPts},
		dkio.WithElection(ipt.Election),
		dkio.WithCollectCost(time.Since(ts)),
		dkio.WithSource(inputName),
		dkio.DisableGlobalTags(true))
	if err != nil {
		l.Errorf("feed metric failed: %s", err)
		return
	}

	ipt.queryAndCollectResult(ingestioncanary.MetricCategory, ts, "")
}

// feedLogging generates and feeds logging data.
func (ipt *Input) feedLogging() {
	ts := time.Now().Truncate(time.Millisecond)

	storageIndex := "default"
	if ipt.Logging != nil && ipt.Logging.StorageIndex != "" {
		storageIndex = ipt.Logging.StorageIndex
	}

	loggingPts := ipt.canary.Logging(ts, ipt.round, ipt.Tags)
	err := ipt.feeder.Feed(point.Logging, []*point.Point{loggingPts},
		dkio.WithElection(ipt.Election),
		dkio.WithCollectCost(time.Since(ts)),
		dkio.WithStorageIndex(storageIndex),
		dkio.WithSource(inputName),
		dkio.DisableGlobalTags(true))
	if err != nil {
		l.Errorf("feed logging failed: %s, index: %s", err, storageIndex)
		return
	}

	ipt.queryAndCollectResult(ingestioncanary.LoggingCategory, ts, storageIndex)
}

// feedTracing generates and feeds tracing data.
func (ipt *Input) feedTracing() {
	ts := time.Now().Truncate(time.Millisecond)

	tracingPts := ipt.canary.Tracing(ts, ipt.round, ipt.Tags)
	err := ipt.feeder.Feed(point.Tracing, []*point.Point{tracingPts},
		dkio.WithElection(ipt.Election),
		dkio.WithCollectCost(time.Since(ts)),
		dkio.WithSource(inputName),
		dkio.DisableGlobalTags(true))
	if err != nil {
		l.Errorf("feed tracing failed: %s", err)
		return
	}

	ipt.queryAndCollectResult(ingestioncanary.TracingCategory, ts, "")
}

// queryAndCollectResult queries data and collects result point (for Logging, requires storage_index).
func (ipt *Input) queryAndCollectResult(category ingestioncanary.Category, feedTime time.Time, storageIndex string) {
	status := ""
	var ingestionLatency time.Duration

	timeoutCtx, cancel := context.WithTimeout(context.Background(), ipt.QueryTimeout.Duration)
	defer cancel()

	defer func() {
		if status != "" {
			ingestionLatency = time.Since(feedTime)
			pt := ipt.buildResultPoint(category, status, storageIndex, ingestionLatency.Milliseconds())
			if pt != nil {
				ipt.mu.Lock()
				ipt.resultPoints = append(ipt.resultPoints, pt)
				ipt.mu.Unlock()
			}
		}
	}()

	errorCount := 0
	for {
		found, err := ipt.canary.CheckLast(category, &ingestioncanary.Feed{
			TimeMs: feedTime.UnixMilli(),
			Round:  ipt.round,
		}, storageIndex)
		if err != nil {
			errorCount++
			l.Errorf("query failed (retry %d/%d): %v", errorCount, ipt.ErrorRetries, err)
			if errorCount >= ipt.ErrorRetries {
				status = "error"
				return
			}
		}

		if found {
			status = "ok"
			break
		}

		select {
		case <-datakit.Exit.Wait():
			return
		case <-ipt.semStop.Wait():
			return
		case <-timeoutCtx.Done():
			status = "timeout"
			return
		case <-time.After(ipt.PollInterval.Duration):
		}
	}
}

// buildResultPoint builds a result point for reporting.
func (ipt *Input) buildResultPoint(category ingestioncanary.Category, status string, storageIndex string, latencyMs int64) *point.Point {
	storageTag := ""
	if storageIndex != "" {
		storageTag = fmt.Sprintf(",storage_index=%s", storageIndex)
	}
	l.Infof("ingestion_canary_result: category=%s, status=%s%s, latency_ms=%dms", category, status, storageTag, latencyMs)

	if ipt.ResultWorkspace == "" {
		return nil
	}

	opts := append(point.DefaultMetricOptions(), point.WithTime(time.Now()))
	kvs := point.NewKVs(map[string]interface{}{
		"latency_ms": latencyMs,
	})

	kvs = kvs.AddTag("category", string(category))
	kvs = kvs.AddTag("status", status)
	for k, v := range ipt.Tags {
		kvs = kvs.AddTag(k, v)
	}
	if storageIndex != "" {
		kvs = kvs.AddTag("storage_index", storageIndex)
	}

	return point.NewPoint(resultMeasurementName, kvs, opts...)
}

// reportToWorkspace reports results to specified workspace using DataWay writer.
func (ipt *Input) reportToWorkspace(pts ...*point.Point) error {
	if ipt.resultDataway == nil {
		return fmt.Errorf("result dataway not initialized")
	}

	return ipt.resultDataway.Write(
		compact.WithPoints(pts),
		compact.WithDynamicURL(ipt.writeURL),
		compact.WithCategory(point.DynamicDWCategory),
		compact.WithNoWAL(true),
		compact.WithGzipDuringBuildBody(true),
		compact.WithHTTPHeader("X-Sub-Category", inputName),
	)
}
