// Package stats is used to record metrics
package stats

import (
	"github.com/prometheus/client_golang/prometheus"
)

var registry = prometheus.NewRegistry()

func GetRegistry() *prometheus.Registry {
	return registry
}

func MustRegister(c prometheus.Collector) {
	registry.MustRegister(c)
}
