// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package stats

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var defaultLabelNames = []string{"category", "name", "namespace"}

type RecMetric struct {
	plPtsVec,
	plDropVec,
	plErrPtsVec *prometheus.CounterVec
	plCostVec   *prometheus.SummaryVec
	plUpdateVec *prometheus.GaugeVec
	labeslNames []string
}

func (rec *RecMetric) WriteMetric(tags map[string]string, pt, ptDrop, ptError float64, cost time.Duration) {
	lbVals := selectLabels(tags, rec.labeslNames)

	if pt > 0 {
		rec.plPtsVec.WithLabelValues(lbVals...).Add(pt)
	}

	if ptDrop > 0 {
		rec.plDropVec.WithLabelValues(lbVals...).Add(ptDrop)
	}

	if ptError > 0 {
		rec.plErrPtsVec.WithLabelValues(lbVals...).Add(ptError)
	}

	if cost > 0 {
		rec.plCostVec.WithLabelValues(lbVals...).Observe(float64(cost) / float64(time.Second))
	}
}

func (rec *RecMetric) WriteUpdateTime(tags map[string]string) {
	rec.plUpdateVec.WithLabelValues(selectLabels(
		tags, rec.labeslNames)...).Set(float64(time.Now().Unix()))
}

func (rec *RecMetric) Metrics() []prometheus.Collector {
	return []prometheus.Collector{
		rec.plPtsVec,
		rec.plDropVec,
		rec.plErrPtsVec,
		rec.plCostVec,
		rec.plUpdateVec,
	}
}

func newRecMetric(namespace, subsystem string, labelNames []string) *RecMetric {
	if len(labelNames) == 0 {
		labelNames = defaultLabelNames
	}

	plPtsVec := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "point_total",
			Help:      "Pipeline processed total points",
		}, labelNames,
	)

	plDropVec := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "drop_point_total",
			Help:      "Pipeline total dropped points",
		}, labelNames,
	)

	plErrPtsVec := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "error_point_total",
			Help:      "Pipeline processed total error points",
		}, labelNames,
	)

	plCostVec := prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "cost_seconds",
			Help:      "Pipeline total running time",
		}, labelNames,
	)

	plUpdateVec := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "last_update_timestamp_seconds",
			Help:      "Pipeline last update time",
		}, labelNames,
	)

	return &RecMetric{
		plPtsVec:    plPtsVec,
		plDropVec:   plDropVec,
		plErrPtsVec: plErrPtsVec,
		plCostVec:   plCostVec,
		plUpdateVec: plUpdateVec,
		labeslNames: labelNames,
	}
}

func selectLabels(tags map[string]string, lb []string) []string {
	if len(lb) == 0 {
		lb = defaultLabelNames
	}

	lbVals := make([]string, len(lb))

	for i := range lb {
		if val, ok := tags[lb[i]]; ok {
			lbVals[i] = val
		}
	}

	return lbVals
}
