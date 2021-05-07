package kubernetes

import (
	"io"
	"strings"

	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
)

func promTextToMetrics(data io.Reader) (map[string]interface{}, error) {
	var parser expfmt.TextParser
	metrics, err := parser.TextToMetricFamilies(data)
	if err != nil {
		return nil, err
	}

	var fields = make(map[string]interface{})

	for name, metric := range metrics {
		if !strings.HasPrefix(name, "process_") {
			continue
		}

		switch metric.GetType() {

		case dto.MetricType_GAUGE:
			for _, mt := range metric.GetMetric() {
				fields[name] = mt.GetGauge().GetValue()
			}

		case dto.MetricType_COUNTER:
			for _, mt := range metric.GetMetric() {
				fields[name] = mt.GetCounter().GetValue()
			}
		}
	}

	return fields, nil
}
