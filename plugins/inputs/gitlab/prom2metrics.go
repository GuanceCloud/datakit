package gitlab

import (
	"io"
	"strings"

	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
)

const prefixName = "gitlab_"

var ignoreList = map[string]interface{}{
	"http_requests_total": nil,
}

type samplePoint struct {
	tags   map[string]string
	fields map[string]interface{}
}

func promTextToMetrics(data io.Reader) ([]*samplePoint, error) {
	var parser expfmt.TextParser
	metrics, err := parser.TextToMetricFamilies(data)
	if err != nil {
		return nil, err
	}

	var pts []*samplePoint
	var prom prom

	for name, metric := range metrics {
		if _, ok := ignoreList[name]; ok {
			continue
		}

		switch metric.GetType() {
		case dto.MetricType_COUNTER:
			pts = append(pts, prom.counter(name, metric.Metric)...)

		case dto.MetricType_HISTOGRAM:
			pts = append(pts, prom.histogram(name, metric.Metric)...)

		case dto.MetricType_GAUGE:
			l.Debugf("ignore gauge")
		case dto.MetricType_SUMMARY:
			l.Debugf("ignore summary")
		case dto.MetricType_UNTYPED:
			l.Debugf("ignore untyped")
		}
	}

	return pts, nil
}

type prom struct{}

func (p *prom) counter(name string, metrics []*dto.Metric) []*samplePoint {
	var pts []*samplePoint
	for _, m := range metrics {
		if m.GetCounter() == nil {
			continue
		}
		pts = append(pts, &samplePoint{
			tags:   labelToTags(m.GetLabel()),
			fields: map[string]interface{}{strings.TrimPrefix(name, prefixName): m.GetCounter().GetValue()},
		})
	}
	return pts
}

func (p *prom) histogram(name string, metrics []*dto.Metric) []*samplePoint {
	var pts []*samplePoint
	for _, m := range metrics {
		if m.GetHistogram() == nil {
			continue
		}
		pts = append(pts, &samplePoint{
			tags: labelToTags(m.GetLabel()),
			fields: map[string]interface{}{
				strings.TrimPrefix(name, prefixName) + "_count": float64(m.GetHistogram().GetSampleCount()),
				strings.TrimPrefix(name, prefixName) + "_sum":   m.GetHistogram().GetSampleSum(),
			},
		})
	}
	return pts
}

func labelToTags(label []*dto.LabelPair) map[string]string {
	if len(label) == 0 {
		return nil
	}
	tags := make(map[string]string, len(label))
	for _, lab := range label {
		tags[lab.GetName()] = lab.GetValue()
	}
	return tags
}
