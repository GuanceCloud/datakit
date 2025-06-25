// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package io

import (
	T "testing"
	"time"

	"github.com/GuanceCloud/cliutils/metrics"
	"github.com/GuanceCloud/cliutils/point"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dkmetrics "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/metrics"
)

func TestInputFeedMetrics(t *T.T) {
	t.Run("do-feed-0-pts", func(t *T.T) {
		reg := prometheus.NewRegistry()
		reg.MustRegister(Metrics()...)

		cat := point.Metric
		f := DefaultFeeder()

		assert.NoError(t, f.Feed(cat, nil, WithSource(t.Name())))

		mfs, err := reg.Gather()
		assert.NoError(t, err)

		assert.Equal(t, 0.0, metrics.GetMetricOnLabels(mfs,
			"datakit_io_feed_point_total", cat.String(), t.Name()).GetCounter().GetValue())
		assert.Equal(t, 1.0, metrics.GetMetricOnLabels(mfs,
			"datakit_io_feed_total", cat.String(), t.Name()).GetCounter().GetValue())

		t.Cleanup(func() {
			MetricsReset()
		})
	})

	t.Run("do-feed-n-pts-with-feed-option", func(t *T.T) {
		reg := prometheus.NewRegistry()
		reg.MustRegister(Metrics()...)

		cat := point.Metric

		pts := point.RandPoints(100)

		feeder := DefaultFeeder()

		assert.NoError(t, feeder.Feed(cat, pts, WithSource(t.Name()), WithCollectCost(time.Second)))

		mfs, err := reg.Gather()
		assert.NoError(t, err)
		t.Logf("\n%s", metrics.MetricFamily2Text(mfs))

		assert.Equal(t, 1.0, metrics.GetMetricOnLabels(mfs,
			"datakit_io_feed_total", cat.String(), t.Name()).GetCounter().GetValue())
		assert.Equal(t, float64(1.0), metrics.GetMetricOnLabels(mfs,
			"datakit_input_collect_latency_seconds", cat.String(), t.Name()).GetSummary().GetSampleSum())
		assert.True(t, 0 < metrics.GetMetricOnLabels(mfs,
			"datakit_io_last_feed_timestamp_seconds", cat.String(), t.Name()).GetGauge().GetValue())

		t.Cleanup(func() {
			MetricsReset()
		})
	})
}

func TestFeedMetrics(t *T.T) {
	t.Run("basic", func(t *T.T) {
		reg := prometheus.NewRegistry()
		reg.MustRegister(Metrics()...)

		dkmetrics.LastErrVec.WithLabelValues("input_ipt", "abc", "cat_abc", "err_abc").Set(float64(time.Now().Unix()))
		dkmetrics.LastErrVec.WithLabelValues("input_ipt", "def", "cat_def", "err_def").Set(float64(time.Now().Unix()))

		inputsFeedVec.WithLabelValues("abc", "cat_abc").Inc()
		inputsFeedVec.WithLabelValues("def", "cat_def").Inc()

		inputsLastFeedVec.WithLabelValues("abc", "cat_abc").Set(float64(time.Now().Unix()))
		inputsLastFeedVec.WithLabelValues("def", "cat_def").Set(float64(time.Now().Unix()))

		mfs, err := reg.Gather()
		require.NoError(t, err)

		t.Logf("metrics:\n%s", metrics.MetricFamily2Text(mfs))

		arr := FeedMetrics(mfs, time.Second)
		assert.Len(t, arr, 2)

		t.Cleanup(func() {
			MetricsReset()
		})
	})

	t.Run("exclude-expired-error", func(t *T.T) {
		dkmetrics.LastErrVec.WithLabelValues("input_ipt", "abc", "cat_abc", "err_abc").Set(float64(time.Now().Unix() - 100))

		inputsFeedVec.WithLabelValues("abc", "cat_abc").Inc()

		inputsLastFeedVec.WithLabelValues("abc", "cat_abc").Set(float64(time.Now().Unix()))

		mfs, err := metrics.Gather()
		require.NoError(t, err)

		arr := FeedMetrics(mfs, time.Second*50)
		assert.Len(t, arr, 1)

		for _, cs := range arr {
			t.Logf("%+#v", cs)
		}

		assert.Equal(t, "", arr[0].LastErr)
		assert.Equal(t, int64(0), arr[0].LastErrTime)

		t.Cleanup(func() {
			dkmetrics.LastErrVec.Reset()
			inputsFeedVec.Reset()
			inputsLastFeedVec.Reset()
		})
	})
}
