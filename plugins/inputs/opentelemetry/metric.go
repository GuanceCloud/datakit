// Package opentelemetry is metric

package opentelemetry

import (
	"encoding/json"
	"strconv"
	"time"

	commonpb "go.opentelemetry.io/proto/otlp/common/v1"

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
	Operation   string            `json:"operation"`   // metric.name
	Source      string            `json:"source"`      // inputName ： opentelemetry
	Attributes  map[string]string `json:"attributes"`  // tags
	Resource    string            `json:"resource"`    // global.Meter name
	Description string            `json:"description"` // metric.Description
	StartTime   uint64            `json:"start_time"`  // start time
	UnitTime    uint64            `json:"unit_time"`   // end time

	ValueType string      `json:"value_type"` // double | int | histogram | ExponentialHistogram | summary
	Value     interface{} `json:"value"`      // 5种类型 对应的值：int | float

	Content string `json:"content"` //

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

// toDatakitTags : make attributes to tags
func toDatakitTags(attr []*commonpb.KeyValue) map[string]string {
	m := make(map[string]string, len(attr))
	for _, kv := range attr {
		key := replace(kv.Key) // 统一将`.`换成 `_`
		switch t := kv.GetValue().Value.(type) {
		case *commonpb.AnyValue_StringValue:
			m[key] = kv.GetValue().GetStringValue()
		case *commonpb.AnyValue_BoolValue:
			m[key] = strconv.FormatBool(t.BoolValue)
		case *commonpb.AnyValue_IntValue:
			m[key] = strconv.FormatInt(t.IntValue, 10)
		case *commonpb.AnyValue_DoubleValue:
			m[key] = strconv.FormatFloat(t.DoubleValue, 'f', 2, 64)
		case *commonpb.AnyValue_ArrayValue:
			m[key] = t.ArrayValue.String()
		case *commonpb.AnyValue_KvlistValue:
			tags := toDatakitTags(t.KvlistValue.Values)
			for s, s2 := range tags {
				m[s] = s2
			}
		case *commonpb.AnyValue_BytesValue:
			m[key] = string(t.BytesValue)
		default:
			m[key] = kv.Value.GetStringValue()
		}
	}
	return m
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
