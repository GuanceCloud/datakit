// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package metrics

import (
	"sync/atomic"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	lastErrorVec = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "datakit",
			Name:      "lasterr",

			Help: "Datakit internal errors(with error occurred unix timestamp)",
		},
		[]string{
			"source",  // where the error comes from
			"message", // detailed error message
		},
	)

	lastErrorCountVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "datakit",
			Name:      "lasterr_total",

			Help: "Datakit internal errors(with error occurred unix timestamp)",
		},
		[]string{
			"source", // where the error comes from
		},
	)

	lastErrorCount    int64
	MaxLastErrorCount = int64(128)
)

// nolint: gochecknoinits
func init() {
	MustRegister(lastErrorVec, lastErrorCountVec)
}

// ResetLastErrors cleanup all metrics within last errors.
func ResetLastErrors() {
	lastErrorVec.Reset()
}

// AddLastErr add error msg to datakit's metric server.
func AddLastErr(source, msg string) {
	lastErrorVec.WithLabelValues(source, msg).Set(float64(time.Now().Unix()))
	lastErrorCountVec.WithLabelValues(source).Inc()

	if atomic.AddInt64(&lastErrorCount, int64(1))%MaxLastErrorCount == 0 { // clean
		ResetLastErrors()
	}
}
