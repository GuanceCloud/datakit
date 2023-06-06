// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.
// Some code modified from project Datadog (https://www.datadoghq.com/).

package snmputil

// MetricType is the representation of an aggregator metric type.
type MetricType int

// metric type constants enumeration.
const (
	GaugeType MetricType = iota
	RateType
	CountType
	MonotonicCountType
	CounterType
	HistogramType
	HistorateType
	SetType
	DistributionType

	// NumMetricTypes is the number of metric types; must be the last item here.
	NumMetricTypes
)
