// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package io

import (
	"time"

	"github.com/GuanceCloud/cliutils/metrics"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

var (
	inputsFeedVec,
	flushVec,
	inputsFilteredPtsVec *prometheus.CounterVec

	feedCost,
	inputsFeedPtsVec,
	inputsCollectLatencyVec *prometheus.SummaryVec

	queuePtsVec,
	flushWorkersVec,
	inputsLastFeedVec,
	ioChanCap,
	ioChanLen *prometheus.GaugeVec
)

func InputsFeedVec() *prometheus.CounterVec {
	return inputsFeedVec
}

func InputsFeedPtsVec() *prometheus.SummaryVec {
	return inputsFeedPtsVec
}

func InputsLastFeedVec() *prometheus.GaugeVec {
	return inputsLastFeedVec
}

func InputsCollectLatencyVec() *prometheus.SummaryVec {
	return inputsCollectLatencyVec
}

func metricsSetup() {
	feedCost = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace: "datakit",
			Subsystem: "io",
			Name:      "feed_cost_seconds",
			Help:      "IO feed waiting(on block mode) seconds",

			Objectives: map[float64]float64{
				0.5:  0.05,
				0.9:  0.01,
				0.99: 0.001,
			},
		},
		[]string{
			"category",
			"from",
		},
	)

	inputsFeedPtsVec = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace: "datakit",
			Subsystem: "io",
			Name:      "feed_point",
			Help:      "Input feed point",

			Objectives: map[float64]float64{
				0.5:  0.05,
				0.9:  0.01,
				0.99: 0.001,
			},
		},
		[]string{
			"name",
			"category",
		},
	)

	flushWorkersVec = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "datakit",
			Subsystem: "io",
			Name:      "flush_workers",
			Help:      "IO flush workers",
		},
		[]string{"category"},
	)

	flushVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "datakit",
			Subsystem: "io",
			Name:      "flush_total",
			Help:      "IO flush total",
		},
		[]string{"category"},
	)

	queuePtsVec = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "datakit",
			Subsystem: "io",
			Name:      "queue_points",
			Help:      "IO module queued(cached) points",
		},
		[]string{
			"category",
		})

	inputsFilteredPtsVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "datakit",
			Subsystem: "io",
			Name:      "input_filter_point_total",
			Help:      "Input filtered point total",
		},
		[]string{
			"name",
			"category",
		},
	)

	inputsFeedVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "datakit",
			Subsystem: "io",
			Name:      "feed_total",
			Help:      "Input feed total",
		},
		[]string{
			"name",
			"category",
		},
	)

	inputsLastFeedVec = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "datakit",
			Subsystem: "io",
			Name:      "last_feed_timestamp_seconds",
			Help:      "Input last feed time(according to DataKit local time)",
		},
		[]string{
			"name",
			"category",
		},
	)

	inputsCollectLatencyVec = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace: "datakit",
			Subsystem: "input",
			Name:      "collect_latency_seconds",
			Help:      "Input collect latency",

			Objectives: map[float64]float64{
				0.5:  0.05,
				0.9:  0.01,
				0.99: 0.001,
			},
		},
		[]string{
			"name",
			"category",
		},
	)

	ioChanLen = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "datakit",
			Subsystem: "io",
			Name:      "chan_usage",
			Help:      "IO channel usage(length of the channel)",
		},
		[]string{
			"category",
		},
	)

	ioChanCap = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "datakit",
			Subsystem: "io",
			Name:      "chan_capacity",
			Help:      "IO channel capacity",
		},
		[]string{
			"category",
		},
	)

	// add more...
}

//nolint:gochecknoinits
func init() {
	metricsSetup()
	metrics.MustRegister(Metrics()...)
}

func Metrics() []prometheus.Collector {
	return []prometheus.Collector{
		inputsFeedVec,
		inputsFeedPtsVec,
		inputsFilteredPtsVec,
		inputsLastFeedVec,
		inputsCollectLatencyVec,
		queuePtsVec,
		ioChanLen,
		ioChanCap,
		flushVec,
		flushWorkersVec,
		feedCost,
	}
}

func MetricsReset() {
	inputsFeedVec.Reset()
	inputsFeedPtsVec.Reset()
	inputsFilteredPtsVec.Reset()

	inputsCollectLatencyVec.Reset()

	queuePtsVec.Reset()
	inputsLastFeedVec.Reset()
	ioChanCap.Reset()
	ioChanLen.Reset()
	flushVec.Reset()
	flushWorkersVec.Reset()
}

// A CollectorStatus used to describe a input's status.
type CollectorStatus struct {
	Name        string `json:"name"`
	Count       int64  `json:"count"`
	Version     string `json:"version,omitempty"`
	LastTime    int64  `json:"last_time,omitempty"`
	LastErr     string `json:"last_err,omitempty"`
	LastErrTime int64  `json:"last_err_time,omitempty"`
}

// FeedMetrics extract all inputs feed metrics from mfs.
func FeedMetrics(mfs []*dto.MetricFamily, ignoreErrBefore time.Duration) (res []*CollectorStatus) {
	if len(mfs) == 0 {
		return
	}

	get := func(name string) *CollectorStatus {
		for _, x := range res {
			if name == x.Name {
				return x
			}
		}
		return nil
	}

	// we first get the input list.
	for _, mf := range mfs {
		switch mf.GetName() {
		case "datakit_inputs_instance": // get collect count(feed count)
			for _, m := range mf.Metric {
				lps := m.GetLabel() // must with these labels: category/name
				if len(lps) == 1 {
					inputName := lps[0].GetValue()
					cs := get(inputName)
					if cs == nil {
						cs = &CollectorStatus{
							Name:  inputName,
							Count: int64(m.GetCounter().GetValue()),
						}
						res = append(res, cs)
					}
				}
			}
		default: // pass
		}
	}

	// "datakit_inputs_instance" would not exist when datakit is not running(unit tests),
	// so we should use "datakit_io_feed_total" here.
	if len(res) == 0 {
		for _, mf := range mfs {
			switch mf.GetName() {
			case "datakit_io_feed_total": // get collect count(feed count)
				for _, m := range mf.Metric {
					lps := m.GetLabel() // must with these labels: category/name
					if len(lps) == 2 {
						inputName := lps[0].GetValue()
						cs := get(inputName)
						if cs == nil {
							cs = &CollectorStatus{
								Name:  inputName,
								Count: int64(m.GetCounter().GetValue()),
							}
							res = append(res, cs)
						}
					}
				}
			default: // pass
			}
		}
	}

	// then append error info(if any) into each inputs.
	for _, mf := range mfs {
		switch mf.GetName() {
		case "datakit_last_err": // get last collect error if any
			for _, m := range mf.Metric {
				lps := m.GetLabel()
				if len(lps) == 3 { // must with these labels: category/error/source
					cs := get(lps[2].GetValue()) // label `source' is the input-name
					if cs != nil {
						cs.LastErrTime = int64(m.GetGauge().GetValue())

						if time.Since(time.Unix(cs.LastErrTime, 0)) > ignoreErrBefore {
							cs.LastErrTime = 0 // ignore errors 30s(default) ago
						} else {
							cs.LastErr = lps[1].GetValue()
						}
					}
				}
			}

		case "datakit_io_last_feed": // get last collect time
			for _, m := range mf.Metric {
				lps := m.GetLabel() // must with these labels: category/name
				if len(lps) == 2 {
					cs := get(lps[1].GetValue())
					if cs != nil {
						cs.LastTime = int64(m.GetGauge().GetValue())
					}
				}
			}

		default:
			// pass
		}
	}

	return res
}
