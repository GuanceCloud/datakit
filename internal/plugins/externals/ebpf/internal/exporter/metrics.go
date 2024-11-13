package exporter

import (
	"github.com/prometheus/client_golang/prometheus"
)

var ePtsVec = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: "dkebpf",
		Subsystem: "exporter",
		Name:      "points_total",
		Help:      "The number of data points processed by the exporter",
	},
	[]string{"name", "category"},
)
