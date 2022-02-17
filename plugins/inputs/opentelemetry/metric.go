// Package opentelemetry is metric

package opentelemetry

import (
	"encoding/json"
	"time"

	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	metricpb "go.opentelemetry.io/proto/otlp/metrics/v1"
)

type date struct {
	tags      map[string]string
	typeName  string
	startTime uint64
	unitTime  uint64
	val       interface{}
}

func getData(metric *metricpb.Metric) []*date {
	ps := make([]*date, 0)
	switch t := metric.GetData().(type) {
	// case *metricpb.Metric_IntGauge: // 弃用
	case *metricpb.Metric_Gauge:
		for _, p := range t.Gauge.DataPoints {
			point := &date{}
			if double, ok := p.Value.(*metricpb.NumberDataPoint_AsDouble); ok {
				point.val = double.AsDouble
				point.typeName = "double"
			} else if i, ok := p.Value.(*metricpb.NumberDataPoint_AsInt); ok {
				point.val = i.AsInt
				point.typeName = "int"
			}
			point.tags = toDatakitTags(p.Attributes)
			point.startTime = p.StartTimeUnixNano
			point.unitTime = p.TimeUnixNano
			ps = append(ps, point)
		}
	// case *metricpb.Metric_IntSum: // 弃用
	case *metricpb.Metric_Sum:
		for _, p := range t.Sum.DataPoints {
			point := &date{}
			if double, ok := p.Value.(*metricpb.NumberDataPoint_AsDouble); ok {
				point.val = double.AsDouble
				point.typeName = "double"
			} else if i, ok := p.Value.(*metricpb.NumberDataPoint_AsInt); ok {
				point.val = i.AsInt
				point.typeName = "int"
			}
			//	t.Sum.AggregationTemporality
			point.tags = toDatakitTags(p.Attributes)
			point.startTime = p.StartTimeUnixNano
			point.unitTime = p.TimeUnixNano
			ps = append(ps, point)
		}
	// case *metricpb.Metric_IntHistogram: // 弃用
	case *metricpb.Metric_Histogram:
		for _, p := range t.Histogram.DataPoints {
			point := &date{}
			point.val = p.Sum
			point.typeName = "histogram"
			point.tags = toDatakitTags(p.Attributes)
			point.startTime = p.StartTimeUnixNano
			point.unitTime = p.TimeUnixNano

			ps = append(ps, point)
		}
	case *metricpb.Metric_ExponentialHistogram:
		for _, p := range t.ExponentialHistogram.DataPoints {
			point := &date{
				typeName:  "ExponentialHistogram",
				tags:      toDatakitTags(p.Attributes),
				startTime: p.StartTimeUnixNano,
				unitTime:  p.TimeUnixNano,
				val:       p.Sum,
			}
			ps = append(ps, point)
		}
	case *metricpb.Metric_Summary:
		for _, p := range t.Summary.DataPoints {
			point := &date{
				typeName:  "summary",
				tags:      toDatakitTags(p.Attributes),
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
	Operation   string            `json:"operation"`
	Source      string            `json:"source"`
	Attributes  map[string]string `json:"attributes"`
	Resource    string            `json:"resource"`
	Description string            `json:"description"`
	ValueType   string            `json:"value_type"` // int bool int64 float
	StartTime   uint64            `json:"start_time"`
	UnitTime    uint64            `json:"unit_time"`
	Value       interface{}       `json:"value"`
	Content     string            `json:"content"`
	// Exemplar 可获取 spanid 等
}

func toDatakitMetric(rss []*metricpb.ResourceMetrics) []*otelResourceMetric {
	orms := make([]*otelResourceMetric, 0)
	for _, resourceMetrics := range rss {
		tags := toDatakitTags(resourceMetrics.Resource.Attributes)
		LibraryMetrics := resourceMetrics.GetInstrumentationLibraryMetrics()
		for _, libraryMetric := range LibraryMetrics {
			resource := libraryMetric.InstrumentationLibrary.Name
			metrices := libraryMetric.GetMetrics()
			for _, metrice := range metrices {
				/*	l.Debugf(metrice.Name)
					bts, err := json.MarshalIndent(metrice, "\t", "    ")
					if err == nil {
						l.Info(string(bts))
					}
					l.Infof("metric string=%s", metrice.String())*/
				ps := getData(metrice)
				for _, p := range ps {
					orm := &otelResourceMetric{
						Operation:   metrice.Name,
						Source:      inputName,
						Attributes:  tags,
						Resource:    resource,
						Description: metrice.Description,
						ValueType:   p.typeName,
						StartTime:   p.startTime,
						UnitTime:    p.unitTime,
						Value:       p.val,
					}
					for k, v := range p.tags {
						orm.Attributes[k] = v
					}
					bts, err := json.Marshal(orm)
					if err != nil {
						l.Errorf("marshal err=%v", err)
					} else {
						orm.Content = string(bts)
					}
					orms = append(orms, orm)
				}
			}
		}
	}
	return orms
}

func otelMetricToDkMetric(orms []*otelResourceMetric) (res []*DKMetric) {
	for _, resourceMetric := range orms {
		tags := map[string]string{
			"operation":   resourceMetric.Operation,
			"description": resourceMetric.Description,
		}
		for k, v := range resourceMetric.Attributes {
			tags[k] = v
		}
		fields := map[string]interface{}{
			"start":      resourceMetric.StartTime,
			"duration":   resourceMetric.UnitTime - resourceMetric.StartTime,
			"message":    resourceMetric.Content,
			"resource":   resourceMetric.Resource,
			"value_type": resourceMetric.ValueType,
			"value":      resourceMetric.Value,
		}
		dm := &DKMetric{
			name:   inputName,
			tags:   tags,
			fields: fields,
			ts:     time.Unix(0, int64(resourceMetric.StartTime)),
		}
		res = append(res, dm)
	}
	return res
}

type DKMetric struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
	ts     time.Time
}

func (m *DKMetric) LineProto() (*dkio.Point, error) {
	return dkio.MakePoint(m.name, m.tags, m.fields, m.ts)
}

//nolint:lll
func (m *DKMetric) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   inputName,
		Type:   "metric",
		Desc:   "opentelemetry metric 指标",
		Fields: map[string]interface{}{},
		Tags:   map[string]interface{}{},
	}
}
