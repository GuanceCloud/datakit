// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package jsonfields

import (
	"github.com/GuanceCloud/cliutils/metrics"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	resultConverted   = "converted"
	resultInvalidJSON = "invalid_json"
	resultNonObject   = "non_object"
	resultNoFields    = "no_fields"
)

var conversionCounter = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: "datakit",
		Subsystem: "logtail",
		Name:      "json_as_fields_total",
		Help:      "Total number of JSON-as-fields conversion outcomes",
	},
	[]string{"result"},
)

func observe(result string) {
	conversionCounter.WithLabelValues(result).Inc()
}

//nolint:gochecknoinits
func init() {
	metrics.MustRegister(conversionCounter)
}
