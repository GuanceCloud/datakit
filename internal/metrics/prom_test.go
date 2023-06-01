// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package metrics

import (
	"fmt"
	"strings"
	T "testing"

	"github.com/prometheus/client_golang/prometheus"
	p8s "github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
)

func BenchmarkP8s(b *T.B) {
	n := 100

	b.Run("n-histogram-vec", func(b *T.B) {
		var (
			arr []*p8s.HistogramVec
			reg = prometheus.NewRegistry()
		)

		for i := 0; i < n; i++ {
			c := p8s.NewHistogramVec(
				p8s.HistogramOpts{
					Namespace: "ns",
					Subsystem: "BenchmarkP8s",
					Name:      fmt.Sprintf("%d", i),
					Help:      "nothing",
					Buckets: []float64{
						float64(10),
						float64(100),
						float64(1000),
						float64(5000),
						float64(30000),
					},
				},
				[]string{
					"label_1",
					"label_2",
					"label_3",
				},
			)

			assert.NotNil(b, c)

			arr = append(arr, c)
		}

		for i := 0; i < n; i++ {
			reg.MustRegister(arr[i])
		}

		for i := 0; i < n; i++ {
			arr[i].WithLabelValues("v1", "v2", "v3").Observe(float64(i) * 1.0)
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			reg.Gather()
		}
	})

	b.Run("n-summary-vec-with-quantile", func(b *T.B) {
		var (
			arr []*p8s.SummaryVec
			reg = prometheus.NewRegistry()
		)

		for i := 0; i < n; i++ {
			c := p8s.NewSummaryVec(
				p8s.SummaryOpts{
					Namespace: "ns",
					Subsystem: "BenchmarkP8s",
					Name:      fmt.Sprintf("%d", i),
					Help:      "nothing",
					Objectives: map[float64]float64{
						0.5:  0.05,
						0.75: 0.0075,
						0.95: 0.005,
					},
				},
				[]string{
					"label_1",
					"label_2",
					"label_3",
				},
			)

			assert.NotNil(b, c)

			arr = append(arr, c)
		}

		for i := 0; i < n; i++ {
			reg.MustRegister(arr[i])
		}

		for i := 0; i < n; i++ {
			arr[i].WithLabelValues("v1", "v2", "v3").Observe(float64(i) * 1.0)
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			reg.Gather()
		}
	})

	b.Run("n-summary-vec", func(b *T.B) {
		var (
			arr []*p8s.SummaryVec
			reg = prometheus.NewRegistry()
		)

		for i := 0; i < n; i++ {
			c := p8s.NewSummaryVec(
				p8s.SummaryOpts{
					Namespace: "ns",
					Subsystem: "BenchmarkP8s",
					Name:      fmt.Sprintf("%d", i),
					Help:      "nothing",
				},
				[]string{
					"label_1",
					"label_2",
					"label_3",
				},
			)

			assert.NotNil(b, c)

			arr = append(arr, c)
		}

		for i := 0; i < n; i++ {
			reg.MustRegister(arr[i])
		}

		for i := 0; i < n; i++ {
			arr[i].WithLabelValues("v1", "v2", "v3").Observe(float64(i) * 1.0)
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			reg.Gather()
		}
	})

	b.Run("n-counter-vec", func(b *T.B) {
		var (
			arr []*p8s.CounterVec
			reg = prometheus.NewRegistry()
		)

		for i := 0; i < n; i++ {
			c := p8s.NewCounterVec(
				p8s.CounterOpts{
					Namespace: "ns",
					Subsystem: "BenchmarkP8s",
					Name:      fmt.Sprintf("%d", i),
					Help:      "nothing",
				},
				[]string{
					"label_1",
					"label_2",
					"label_3",
				},
			)

			assert.NotNil(b, c)

			arr = append(arr, c)
		}

		for i := 0; i < n; i++ {
			reg.MustRegister(arr[i])
		}

		for i := 0; i < n; i++ {
			arr[i].WithLabelValues("v1", "v2", "v3").Inc()
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			reg.Gather()
		}
	})

	b.Run("n-gauge-vec", func(b *T.B) {
		var (
			arr []*p8s.GaugeVec
			reg = prometheus.NewRegistry()
		)

		for i := 0; i < n; i++ {
			c := p8s.NewGaugeVec(
				p8s.GaugeOpts{
					Namespace: "ns",
					Subsystem: "BenchmarkP8s",
					Name:      fmt.Sprintf("%d", i),
					Help:      "nothing",
				},
				[]string{
					"label_1",
					"label_2",
					"label_3",
				},
			)

			assert.NotNil(b, c)

			arr = append(arr, c)
		}

		for i := 0; i < n; i++ {
			reg.MustRegister(arr[i])
		}

		for i := 0; i < n; i++ {
			arr[i].WithLabelValues("v1", "v2", "v3").Inc()
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			reg.Gather()
		}
	})

	b.Run("n-gauge-vec-with-long-label-vaule", func(b *T.B) {
		var (
			arr []*p8s.GaugeVec
			reg = prometheus.NewRegistry()
		)

		for i := 0; i < n; i++ {
			c := p8s.NewGaugeVec(
				p8s.GaugeOpts{
					Namespace: "ns",
					Subsystem: "BenchmarkP8s",
					Name:      fmt.Sprintf("%d", i),
					Help:      "nothing",
				},
				[]string{
					"label_1",
					"label_2",
					"label_3",
				},
			)

			assert.NotNil(b, c)

			arr = append(arr, c)
		}

		for i := 0; i < n; i++ {
			reg.MustRegister(arr[i])
		}

		labelValues := []string{
			strings.Repeat("1", 100),
			strings.Repeat("2", 100),
			strings.Repeat("3", 100),
		}

		for i := 0; i < n; i++ {
			arr[i].WithLabelValues(
				labelValues[0],
				labelValues[1],
				labelValues[2],
			).Inc()
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			reg.Gather()
		}
	})

	b.Run("n-gauge", func(b *T.B) {
		var (
			arr []p8s.Gauge
			reg = prometheus.NewRegistry()
		)

		for i := 0; i < n; i++ {
			c := p8s.NewGauge(
				p8s.GaugeOpts{
					Namespace: "ns",
					Subsystem: "BenchmarkP8s",
					Name:      fmt.Sprintf("%d", i),
					Help:      "nothing",
				},
			)

			assert.NotNil(b, c)

			arr = append(arr, c)
		}

		for i := 0; i < n; i++ {
			reg.MustRegister(arr[i])
		}

		for i := 0; i < n; i++ {
			arr[i].Inc()
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			reg.Gather()
		}
	})
}
