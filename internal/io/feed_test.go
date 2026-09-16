// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package io

import (
	"context"
	stdio "io"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	T "testing"
	"time"

	"github.com/GuanceCloud/cliutils"
	"github.com/GuanceCloud/cliutils/aggregate"
	"github.com/GuanceCloud/cliutils/point"
	"github.com/GuanceCloud/pipeline-go/constants"
	"github.com/GuanceCloud/pipeline-go/lang"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/aggr"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/metrics"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/plval"
)

type mockFeederOutputer struct {
	feeds []*feedData
}

func prometheusCounterValue(t *T.T, counter prometheus.Counter) float64 {
	t.Helper()
	metric := &dto.Metric{}
	require.NoError(t, counter.Write(metric))
	return metric.GetCounter().GetValue()
}

func resetTailSamplingFallbackMetrics(t *T.T) {
	t.Helper()
	tailSamplingFallbackPkgVec.Reset()
	tailSamplingFallbackPtsVec.Reset()
	t.Cleanup(func() {
		tailSamplingFallbackPkgVec.Reset()
		tailSamplingFallbackPtsVec.Reset()
	})
}

func TestFeedCancellationPreservesCompletedDrop(t *T.T) {
	previous, _ := plval.GetManager()
	manager := plval.NewScriptManager(nil, nil)
	plval.SetManager(manager)
	t.Cleanup(func() { plval.SetManager(previous) })
	if err := manager.LoadScriptWithCatChecked(point.Logging, constants.NSDefault,
		map[string]string{"cancel.p": "if discard { drop(); exit() }\nfor ; repeat; {}\nadd_key(after, true)"}, nil); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	fd := GetFeedData()
	defer putFeedData(fd)
	fd.cat = point.Logging
	fd.disableFilter = true
	fd.pts = []*point.Point{
		newpt("cancel", nil, map[string]any{"discard": true, "repeat": false}, time.Now()),
		newpt("cancel", nil, map[string]any{"discard": false, "repeat": true}, time.Now()),
	}
	WithPipelineContext(ctx)(fd)
	after, created, _, err := (&dkIO{}).beforeFeed(fd)
	if err != nil {
		t.Fatal(err)
	}
	if ctx.Err() != context.DeadlineExceeded {
		t.Fatal("test did not cancel pipeline execution")
	}
	if len(after) != 1 || after[0] != fd.pts[1] {
		t.Fatal("cancellation restored an already dropped point")
	}
	if after[0].Get("after") != nil || len(created) != 0 {
		t.Fatal("published results after cancellation")
	}
}

func TestFeedPoolClearsPipelineContext(t *T.T) {
	fd := GetFeedData()
	WithPipelineContext(context.Background())(fd)
	putFeedData(fd)
	next := GetFeedData()
	defer putFeedData(next)
	if next.pipelineContext != nil {
		t.Fatal("pooled feed retains request context")
	}
}

func (m *mockFeederOutputer) Write(fd *feedData) error {
	m.feeds = append(m.feeds, fd)
	return nil
}

func (*mockFeederOutputer) WriteLastError(string, ...metrics.LastErrorOption) {}

func (*mockFeederOutputer) Reader(point.Category) <-chan *feedData { return nil }

func loadPipelineTestScript(t *T.T, namespace string, category point.Category, name, script string) {
	t.Helper()
	require.NoError(t, pipeline.InitPipeline(nil, nil, nil, ""))

	manager, ok := plval.GetManager()
	require.True(t, ok)
	manager.LoadScripts(namespace,
		map[point.Category]map[string]string{
			category: {name: script},
		}, nil)
	t.Cleanup(func() {
		manager.LoadScripts(namespace, map[point.Category]map[string]string{}, nil)
	})
}

func startConfiguredTestAggregator(t *T.T, handler http.Handler) *aggr.Aggregator {
	t.Helper()

	var configReadyOnce sync.Once
	configReady := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path == datakit.TailSamplingConfig {
			configReadyOnce.Do(func() { close(configReady) })
			w.WriteHeader(http.StatusOK)
			return
		}
		if handler != nil {
			handler.ServeHTTP(w, req)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(server.Close)

	aggregator := &aggr.Aggregator{
		Endpoints:                   []string{server.URL + "?token=test-token"},
		UseLocalConfig:              true,
		LocalConfigDir:              "aggr/testdata",
		LocalMetricConfigFile:       "aggr.toml",
		LocalTailSamplingConfigFile: "tail-sampling.toml",
	}

	oldExit := datakit.Exit
	testExit := cliutils.NewSem()
	datakit.Exit = testExit
	aggrDone := make(chan struct{})
	go func() {
		aggregator.StartAggr()
		close(aggrDone)
	}()
	t.Cleanup(func() {
		testExit.Close()
		select {
		case <-aggrDone:
		case <-time.After(3 * time.Second):
			t.Error("aggregator did not stop")
		}
		datakit.Exit = oldExit
	})

	select {
	case <-configReady:
	case <-time.After(3 * time.Second):
		t.Fatal("tail-sampling config was not loaded")
	}

	return aggregator
}

func installTestIO(t *T.T, aggregator *aggr.Aggregator) *mockFeederOutputer {
	t.Helper()

	oldDefIO := defIO
	output := &mockFeederOutputer{}
	defIO = getIO()
	defIO.foDataway = output
	defIO.Aggr = aggregator
	t.Cleanup(func() { defIO = oldDefIO })

	return output
}

type plcase struct {
	name string

	pts []*point.Point

	epts []*point.Point

	option *Option
}

func newpt(name string, tags map[string]string,
	fields map[string]any, tn time.Time,
) *point.Point {
	kvs := append(point.NewTags(tags), point.NewKVs(fields)...)
	return point.NewPoint(
		name, kvs, append(point.DefaultLoggingOptions(),
			point.WithTime(tn))...,
	)
}

func TestRunpl(t *T.T) {
	pipeline.InitPipeline(nil, nil, nil, "")

	tn := time.Now()
	cases := []plcase{
		{
			name: "a_with_opt",
			pts: []*point.Point{
				newpt("a_with_opt", map[string]string{
					"tag_1": "value_1",
				}, map[string]any{
					"field_1": "value_2",
				}, tn),
			},

			epts: []*point.Point{
				newpt("a_with_opt", map[string]string{
					"tag_1": "value_1",
				}, map[string]any{
					"field_1": "value_2",
					"a":       int64(1),
				}, tn),
			},

			option: &Option{
				PlOption: &lang.LogOption{
					ScriptMap: map[string]string{
						"a_with_opt": "a.p",
					},
				},
			},
		},

		{
			name: "a_status",
			pts: []*point.Point{
				newpt("a", map[string]string{
					"tag_1": "value_1",
				}, map[string]any{
					"field_1": "value_2",
				}, tn),
			},

			epts: nil, // filtered

			option: &Option{
				PlOption: &lang.LogOption{
					ScriptMap: map[string]string{
						"a_with_opt": "a.p",
					},
					IgnoreStatus: []string{"info"},
				},
			},
		},
	}

	for _, c := range cases {
		t.Run(``, func(t *T.T) {
			m, ok := plval.GetManager()
			if !ok {
				t.Error("!ok")
			}
			m.LoadScripts(constants.NSRemote,
				map[point.Category]map[string]string{
					point.Logging: {"a.p": "add_key('a', 1)"},
				}, nil)
			if _, ok := m.QueryScript(point.Logging, "a.p"); !ok {
				t.Error("!ok")
			}

			dkio := getIO()

			fo := GetFeedData()
			fo.input = "a"
			fo.cat = point.Logging
			fo.pts = c.pts
			fo.plOption = c.option.PlOption
			epts, _, _, err := dkio.beforeFeed(fo)
			if err != nil {
				t.Error(err)
			}

			assert.True(t, len(c.epts) == len(epts))

			for i, pt := range epts {
				t.Log(pt.Pretty())
				t.Log(c.epts[i].Pretty())
				pt.Equal(c.epts[i])
			}
		})
	}
}

func TestFeedRunsPipelineBeforeLoggingTailSampling(t *T.T) {
	resetTailSamplingFallbackMetrics(t)

	loadPipelineTestScript(t, constants.NSRemote, point.Logging,
		"extract-trace-id.p", "add_key('trace_id', 'trace-from-pipeline')")

	packets := make(chan *aggregate.DataPacket, 1)
	aggregator := startConfiguredTestAggregator(t, http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path == datakit.TailSampling || req.URL.Path == datakit.TailSamplingV2 {
			body, err := stdio.ReadAll(req.Body)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			packet := &aggregate.DataPacket{}
			if err := packet.Unmarshal(body); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			packets <- packet
		}
		w.WriteHeader(http.StatusOK)
	}))
	output := installTestIO(t, aggregator)

	logPoint := point.NewPoint("test_log", point.KVs{}.
		Add("message", "request completed"), point.DefaultLoggingOptions()...)
	logPoint.SetTime(time.Now())

	err := DefaultFeeder().Feed(point.Logging, []*point.Point{logPoint},
		WithSource("logging"),
		WithPipelineOption(&lang.LogOption{
			ScriptMap: map[string]string{"test_log": "extract-trace-id.p"},
		}))
	require.NoError(t, err)

	var packet *aggregate.DataPacket
	select {
	case packet = <-packets:
	case <-time.After(3 * time.Second):
		t.Fatal("pipeline-extracted trace_id did not enter tail sampling")
	}

	assert.Equal(t, "trace_id", packet.GroupKey)
	assert.Equal(t, "trace-from-pipeline", packet.RawGroupId)

	var tailPoint *point.Point
	require.NoError(t, packet.WalkRawPBPoints(func(raw []byte) bool {
		pb := &point.PBPoint{}
		if err := pb.Unmarshal(raw); err != nil {
			return false
		}
		tailPoint = point.FromPB(pb)
		return false
	}))
	require.NotNil(t, tailPoint)
	assert.Equal(t, "trace-from-pipeline", tailPoint.Get("trace_id"))

	ordinaryPoints := 0
	for _, feed := range output.feeds {
		ordinaryPoints += len(feed.pts)
	}
	assert.Zero(t, ordinaryPoints)
	assert.Equal(t, float64(0), prometheusCounterValue(t,
		tailSamplingFallbackPkgVec.WithLabelValues("logging", point.SLogging)))
	assert.Equal(t, float64(0), prometheusCounterValue(t,
		tailSamplingFallbackPtsVec.WithLabelValues("logging", point.SLogging)))
}

func TestFeedRunsMetricAggregationAfterPipeline(t *T.T) {
	loadPipelineTestScript(t, constants.NSRemote, point.Metric, "normalize-metric.p", `
set_measurement("otel_service")
add_key("jvm.buffer.memory.used", 100)
set_tag("id", "id-1")
set_tag("service_name", "svc-a")
`)

	aggregateRequests := make(chan struct{}, 1)
	aggregator := startConfiguredTestAggregator(t, http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path == datakit.Aggregate {
			aggregateRequests <- struct{}{}
		}
		w.WriteHeader(http.StatusOK)
	}))
	installTestIO(t, aggregator)

	metricPoint := point.NewPoint("raw_metric", point.KVs{}.
		Add("raw_value", 1), point.DefaultMetricOptions()...)
	metricPoint.SetTime(time.Now())

	err := DefaultFeeder().Feed(point.Metric, []*point.Point{metricPoint},
		WithSource("metric"),
		WithPipelineOption(&lang.LogOption{
			ScriptMap: map[string]string{"raw_metric": "normalize-metric.p"},
		}))
	require.NoError(t, err)

	select {
	case <-aggregateRequests:
	case <-time.After(3 * time.Second):
		t.Fatal("pipeline-normalized metric did not enter aggregation")
	}
	assert.Equal(t, "otel_service", metricPoint.Name())
	assert.Equal(t, int64(100), metricPoint.Get("jvm.buffer.memory.used"))
}

func TestFeedWritesPipelineCreatedPointsWhenTracingIsConsumed(t *T.T) {
	loadPipelineTestScript(t, constants.NSConfd, point.Tracing, "trace-create-point.p", `
create_point("pipeline_metric", {"origin": "trace"}, {"value": 1}, 0, "M")
`)

	tailRequests := make(chan struct{}, 1)
	aggregator := startConfiguredTestAggregator(t, http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path == datakit.TailSampling || req.URL.Path == datakit.TailSamplingV2 {
			tailRequests <- struct{}{}
		}
		w.WriteHeader(http.StatusOK)
	}))
	output := installTestIO(t, aggregator)

	tracePoint := point.NewPoint("opentelemetry", point.KVs{}.
		Add("trace_id", "trace-1").
		Add("span_id", "span-1").
		Add("parent_id", "0").
		Add("service", "svc-a").
		Add("resource", "/resource").
		Add("duration", int64(1000)), point.CommonLoggingOptions()...)
	tracePoint.SetTime(time.Now())

	err := DefaultFeeder().Feed(point.Tracing, []*point.Point{tracePoint},
		WithSource("tracing"),
		WithPipelineOption(&lang.LogOption{
			ScriptMap: map[string]string{"svc-a": "trace-create-point.p"},
		}))
	require.NoError(t, err)

	select {
	case <-tailRequests:
	case <-time.After(3 * time.Second):
		t.Fatal("processed trace did not enter tail sampling")
	}

	require.Len(t, output.feeds, 1)
	require.Len(t, output.feeds[0].pts, 1)
	created := output.feeds[0].pts[0]
	assert.Equal(t, point.Metric, output.feeds[0].cat)
	assert.Equal(t, "pipeline_metric", created.Name())
	assert.Equal(t, int64(1), created.Get("value"))
	assert.Equal(t, "trace", created.GetTag("origin"))
}

func TestFeedDoesNotTailSamplePointsDroppedByPipeline(t *T.T) {
	loadPipelineTestScript(t, constants.NSRemote, point.Logging, "drop-log.p", "drop()")

	tailRequests := make(chan struct{}, 1)
	aggregator := startConfiguredTestAggregator(t, http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path == datakit.TailSampling || req.URL.Path == datakit.TailSamplingV2 {
			tailRequests <- struct{}{}
		}
		w.WriteHeader(http.StatusOK)
	}))
	output := installTestIO(t, aggregator)

	logPoint := point.NewPoint("test_log", point.KVs{}.
		Add("message", "blacklisted").
		Add("trace_id", "trace-before-pipeline"), point.DefaultLoggingOptions()...)
	logPoint.SetTime(time.Now())

	err := DefaultFeeder().Feed(point.Logging, []*point.Point{logPoint},
		WithSource("logging"),
		WithPipelineOption(&lang.LogOption{
			ScriptMap: map[string]string{"test_log": "drop-log.p"},
		}))
	require.NoError(t, err)

	select {
	case <-tailRequests:
		t.Fatal("pipeline-dropped point entered tail sampling")
	default:
	}

	ordinaryPoints := 0
	for _, feed := range output.feeds {
		ordinaryPoints += len(feed.pts)
	}
	assert.Zero(t, ordinaryPoints)
}

func TestFeedFallsBackWithProcessedPointsWhenTailSamplingFails(t *T.T) {
	resetTailSamplingFallbackMetrics(t)

	loadPipelineTestScript(t, constants.NSConfd, point.Logging, "prepare-tail-log.p", `
add_key("trace_id", "trace-from-pipeline")
create_point("pipeline_metric", {"origin": "logging"}, {"value": 1}, 0, "M")
`)

	var tailRequests int32
	aggregator := startConfiguredTestAggregator(t, http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path == datakit.TailSampling || req.URL.Path == datakit.TailSamplingV2 {
			atomic.AddInt32(&tailRequests, 1)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	output := installTestIO(t, aggregator)

	logPoint := point.NewPoint("test_log", point.KVs{}.
		Add("message", "request completed"), point.DefaultLoggingOptions()...)
	logPoint.SetTime(time.Now())

	err := DefaultFeeder().Feed(point.Logging, []*point.Point{logPoint},
		WithSource("logging"),
		WithPipelineOption(&lang.LogOption{
			ScriptMap: map[string]string{"test_log": "prepare-tail-log.p"},
		}))
	require.NoError(t, err)
	require.Greater(t, atomic.LoadInt32(&tailRequests), int32(0))

	categoryPoints := map[point.Category][]*point.Point{}
	for _, feed := range output.feeds {
		categoryPoints[feed.cat] = append(categoryPoints[feed.cat], feed.pts...)
	}

	require.Len(t, categoryPoints[point.Logging], 1)
	assert.Equal(t, "trace-from-pipeline", categoryPoints[point.Logging][0].Get("trace_id"))
	require.Len(t, categoryPoints[point.Metric], 1)
	assert.Equal(t, "pipeline_metric", categoryPoints[point.Metric][0].Name())

	assert.Equal(t, float64(1), prometheusCounterValue(t,
		tailSamplingFallbackPkgVec.WithLabelValues("logging", point.SLogging)))
	assert.Equal(t, float64(1), prometheusCounterValue(t,
		tailSamplingFallbackPtsVec.WithLabelValues("logging", point.SLogging)))
}

func Test_correctPointTime(t *T.T) {
	t.Run("basic", func(t *T.T) {
		var kvs point.KVs
		kvs = kvs.Set("f1", 1)
		now := time.Unix(0, 456)

		pts := []*point.Point{
			point.NewPoint("some", kvs, point.WithTimestamp(123)),
			point.NewPoint("some", kvs, point.WithTime(now)), // no correction
		}

		after, n := correctPointTime(pts, now, 1)
		assert.Equal(t, 1, n)
		assert.Equal(t, now.UnixNano(), after[0].Time().UnixNano())
		assert.Equal(t, int64(123), after[0].Get("__orig_time").(int64))
	})
}

func TestAddMetricInputTag(t *T.T) {
	metricPt := point.NewPoint("metric", point.NewKVs(nil).Set("value", 1), point.DefaultMetricOptions()...)
	loggingPt := point.NewPoint("logging", point.NewKVs(nil).Set("message", "test"), point.DefaultLoggingOptions()...)

	feeder := new(ioFeeder)
	feeder.addMetricInputTag(&feedData{cat: point.Metric, pts: []*point.Point{metricPt}, inputName: "snmp"})
	feeder.addMetricInputTag(&feedData{cat: point.Logging, pts: []*point.Point{loggingPt}, inputName: "snmp"})

	assert.Equal(t, "dk.snmp", metricPt.GetTag(InputSourceTagKey))
	assert.Empty(t, loggingPt.GetTag(InputSourceTagKey))

	unknownPt := point.NewPoint("metric", point.NewKVs(nil).Set("value", 1), point.DefaultMetricOptions()...)
	feeder.addMetricInputTag(&feedData{cat: point.Metric, pts: []*point.Point{unknownPt}})
	assert.Equal(t, "dk."+unknownInputName, unknownPt.GetTag(InputSourceTagKey))
}

func TestPLAggFeedAddsInputTag(t *T.T) {
	oldDefIO := defIO
	output := &mockFeederOutputer{}
	defIO = getIO()
	defIO.foDataway = output
	t.Cleanup(func() { defIO = oldDefIO })

	pt := point.NewPoint("metric", point.NewKVs(nil).Set("value", 1), point.DefaultMetricOptions()...)
	assert.NoError(t, PLAggFeed(point.Metric, "aggregation", []*point.Point{pt}))

	assert.Len(t, output.feeds, 1)
	assert.Equal(t, "dk."+pipelineInputName, output.feeds[0].pts[0].GetTag(InputSourceTagKey))
}
