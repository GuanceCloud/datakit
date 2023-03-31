// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package metrics implements datakit's Prometheus metrics
package metrics

import (
	"bytes"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
)

var reg = prometheus.NewRegistry()

// MustRegister add c to global registry and panic on any error.
func MustRegister(c ...prometheus.Collector) {
	reg.MustRegister(c...)
}

// Register add c to global registry.
func Register(c prometheus.Collector) error {
	return reg.Register(c)
}

// MustAddGolangMetrics enable Golang runtime metrics.
func MustAddGolangMetrics() {
	goexporter := collectors.NewGoCollector(collectors.WithGoCollectorRuntimeMetrics(collectors.MetricsAll))
	MustRegister(goexporter)
}

// Unregister remove c from global registry.
func Unregister(c prometheus.Collector) bool {
	return reg.Unregister(c)
}

// Gather collect all metrics within global registry.
func Gather() ([]*dto.MetricFamily, error) {
	return reg.Gather()
}

// MustGather collect all metrics within global registry and panic on any error.
func MustGather() []*dto.MetricFamily {
	x, err := reg.Gather()
	if err != nil {
		panic(err.Error())
	}
	return x
}

func sameLabels(got []*dto.LabelPair, wanted ...string) bool {
	if len(got) != len(wanted) {
		return false
	}

	for i, w := range wanted {
		if got[i].GetValue() != w {
			return false
		}
	}

	return true
}

// GetMetricOnLabels search mfs with wanted labels. wanted values order must be same as label names.
func GetMetricOnLabels(mfs []*dto.MetricFamily, name string, wanted ...string) *dto.Metric {
	for _, mf := range mfs {
		if *mf.Name != name {
			continue
		}

		for _, m := range mf.Metric {
			if sameLabels(m.GetLabel(), wanted...) {
				return m
			}
		}
	}

	return nil
}

// GetMetric with specific idx.
func GetMetric(mfs []*dto.MetricFamily, name string, idx int) *dto.Metric {
	for _, mf := range mfs {
		if *mf.Name == name {
			if len(mf.Metric) < idx {
				return nil
			}
			return mf.Metric[idx]
		}
	}
	return nil
}

// MetricFamily2Text convert metrics to text format.
func MetricFamily2Text(mfs []*dto.MetricFamily) string {
	buf := bytes.NewBuffer(nil)
	for _, mf := range mfs {
		if _, err := expfmt.MetricFamilyToText(buf, mf); err != nil {
			return ""
		}
	}
	return buf.String()
}
