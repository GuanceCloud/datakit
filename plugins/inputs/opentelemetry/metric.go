// Package opentelemetry is metric

package opentelemetry

import (
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	metricpb "go.opentelemetry.io/proto/otlp/metrics/v1"
)

type point struct {
	typeName  string
	startTime uint64
	unitTime  uint64
	val       interface{}
}

func getData(metric *metricpb.Metric) []*point {
	ps := make([]*point, 0)
	switch t := metric.GetData().(type) {
	case *metricpb.Metric_IntGauge: // 弃用
	case *metricpb.Metric_Gauge:
		for _, p := range t.Gauge.DataPoints {
			point := &point{}
			if double, ok := p.Value.(*metricpb.NumberDataPoint_AsDouble); ok {
				point.val = double.AsDouble
				point.typeName = "double"
			} else if i, ok := p.Value.(*metricpb.NumberDataPoint_AsInt); ok {
				point.val = i.AsInt
				point.typeName = "int"
			}
			point.startTime = p.StartTimeUnixNano
			point.unitTime = p.TimeUnixNano
			ps = append(ps, point)
		}
	case *metricpb.Metric_IntSum: // 弃用
	case *metricpb.Metric_Sum:
		for _, p := range t.Sum.DataPoints {
			point := &point{}
			if double, ok := p.Value.(*metricpb.NumberDataPoint_AsDouble); ok {
				point.val = double.AsDouble
				point.typeName = "double"
			} else if i, ok := p.Value.(*metricpb.NumberDataPoint_AsInt); ok {
				point.val = i.AsInt
				point.typeName = "int"
			}
			//	t.Sum.AggregationTemporality
			point.startTime = p.StartTimeUnixNano
			point.unitTime = p.TimeUnixNano
			ps = append(ps, point)
		}
	case *metricpb.Metric_IntHistogram: // 弃用
	case *metricpb.Metric_Histogram:
		for _, p := range t.Histogram.DataPoints {
			point := &point{}
			point.val = p.Sum
			point.typeName = "histogram"
			point.startTime = p.StartTimeUnixNano
			point.unitTime = p.TimeUnixNano
			ps = append(ps, point)
		}
	case *metricpb.Metric_ExponentialHistogram:
		for _, p := range t.ExponentialHistogram.DataPoints {
			point := &point{
				typeName:  "ExponentialHistogram",
				startTime: p.StartTimeUnixNano,
				unitTime:  p.TimeUnixNano,
				val:       p.Sum,
			}
			ps = append(ps, point)
		}
	case *metricpb.Metric_Summary:
		for _, p := range t.Summary.DataPoints {
			point := &point{
				typeName:  "summary",
				startTime: p.StartTimeUnixNano,
				unitTime:  p.TimeUnixNano,
				val:       p.Sum,
			}
			ps = append(ps, point)
		}
	default:
	}
	return ps
}

type otelResourceMetric struct {
	name        string
	attributes  map[string]string
	description string
	dataType    string // int bool int64 float
	startTime   uint64
	unitTime    uint64
	data        interface{}
	// Exemplar 可获取 spanid 等
}

type DKMetric struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
	ts     time.Time
}

func (m *DKMetric) LineProto() (*io.Point, error) {
	return io.MakePoint(m.name, m.tags, m.fields, m.ts)
}

//nolint:lll
func (m *DKMetric) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   inputName,
		Type:   "metric",
		Desc:   "opentelemetry 指标",
		Fields: map[string]interface{}{},
		Tags:   map[string]interface{}{},
	}
}
