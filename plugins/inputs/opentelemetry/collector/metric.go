// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package collector

import (
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	metricpb "go.opentelemetry.io/proto/otlp/metrics/v1"
)

type date struct {
	tags      *dkTags
	typeName  string
	startTime uint64
	unitTime  uint64
	val       interface{}
}

func (s *SpansStorage) getData(metric *metricpb.Metric) []*date {
	ps := make([]*date, 0)
	switch t := metric.GetData().(type) {
	// case *metricpb.Metric_IntGauge: // 弃用
	case *metricpb.Metric_Gauge:
		for _, p := range t.Gauge.DataPoints {
			point := &date{tags: newEmptyTags(s.RegexpString, s.GlobalTags)}
			if double, ok := p.Value.(*metricpb.NumberDataPoint_AsDouble); ok {
				point.val = double.AsDouble
				point.typeName = "double"
			} else if i, ok := p.Value.(*metricpb.NumberDataPoint_AsInt); ok {
				point.val = i.AsInt
				point.typeName = "int"
			}
			point.tags.setAttributesToTags(p.Attributes)
			point.startTime = p.StartTimeUnixNano
			point.unitTime = p.TimeUnixNano
			ps = append(ps, point)
		}
	// case *metricpb.Metric_IntSum: // 弃用
	case *metricpb.Metric_Sum:
		for _, p := range t.Sum.DataPoints {
			point := &date{tags: newEmptyTags(s.RegexpString, s.GlobalTags)}
			if double, ok := p.Value.(*metricpb.NumberDataPoint_AsDouble); ok {
				point.val = double.AsDouble
				point.typeName = "double"
			} else if i, ok := p.Value.(*metricpb.NumberDataPoint_AsInt); ok {
				point.val = i.AsInt
				point.typeName = "int"
			}
			//	t.Sum.AggregationTemporality
			point.tags.setAttributesToTags(p.Attributes)
			point.startTime = p.StartTimeUnixNano
			point.unitTime = p.TimeUnixNano
			ps = append(ps, point)
		}
	// case *metricpb.Metric_IntHistogram: // 弃用
	case *metricpb.Metric_Histogram:
		for _, p := range t.Histogram.DataPoints {
			point := &date{tags: newEmptyTags(s.RegexpString, s.GlobalTags)}
			point.val = p.Sum
			point.typeName = "histogram"
			point.tags.setAttributesToTags(p.Attributes)
			point.startTime = p.StartTimeUnixNano
			point.unitTime = p.TimeUnixNano

			ps = append(ps, point)
		}
	case *metricpb.Metric_ExponentialHistogram:
		for _, p := range t.ExponentialHistogram.DataPoints {
			point := &date{
				typeName:  "ExponentialHistogram",
				tags:      newEmptyTags(s.RegexpString, s.GlobalTags).setAttributesToTags(p.Attributes),
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
				tags:      newEmptyTags(s.RegexpString, s.GlobalTags).setAttributesToTags(p.Attributes),
				startTime: p.StartTimeUnixNano,
				unitTime:  p.TimeUnixNano,
				val:       p.Sum,
			}
			ps = append(ps, point)
		}
	default:
		l.Warnf("unknown metric.Data type or is deprecated Data type")
	}
	for _, p := range ps {
		// 统一处理 tag 问题
		p.tags.checkAllTagsKey().checkCustomTags().addGlobalTags()
	}
	return ps
}

type OtelResourceMetric struct {
	Operation   string            `json:"operation"`   // metric.name
	Attributes  map[string]string `json:"attributes"`  // tags
	Service     string            `json:"service"`     // metric.resource.(service.name) -> 固定值 otel-service
	Resource    string            `json:"resource"`    // global.Meter name
	Description string            `json:"description"` // metric.Description
	StartTime   uint64            `json:"start_time"`  // start time
	UnitTime    uint64            `json:"unit_time"`   // end time
	ValueType   string            `json:"value_type"`  // double | int | histogram | ExponentialHistogram | summary
	Value       interface{}       `json:"value"`       // 上述五种类型所对应的值

	// Exemplar 可获取 spanid 等
}

func (s *SpansStorage) ToDatakitMetric(rss []*metricpb.ResourceMetrics) []*OtelResourceMetric {
	orms := make([]*OtelResourceMetric, 0)
	for _, resourceMetrics := range rss {
		dt := newEmptyTags(s.RegexpString, s.GlobalTags)
		tags := dt.setAttributesToTags(resourceMetrics.Resource.Attributes).tags
		LibraryMetrics := resourceMetrics.GetInstrumentationLibraryMetrics()
		for _, libraryMetric := range LibraryMetrics {
			resource := libraryMetric.InstrumentationLibrary.Name
			metrics := libraryMetric.GetMetrics()
			for _, metrice := range metrics {
				ps := s.getData(metrice)
				for _, p := range ps {
					orm := &OtelResourceMetric{
						Operation:   metrice.Name,
						Attributes:  tags,
						Service:     otelServiceName,
						Resource:    resource,
						Description: metrice.Description,
						ValueType:   p.typeName,
						StartTime:   p.startTime,
						UnitTime:    p.unitTime,
						Value:       p.val,
					}
					for k, v := range p.tags.resource() {
						orm.Attributes[k] = v
					}
					orms = append(orms, orm)
				}
			}
		}
	}
	return orms
}

func makePoints(orms []*OtelResourceMetric) []*dkio.Point {
	pts := make([]*dkio.Point, 0)
	for _, resourceMetric := range orms {
		tags := map[string]string{
			"description":          resourceMetric.Description,
			"instrumentation_name": resourceMetric.Resource,
		}
		for k, v := range resourceMetric.Attributes {
			tags[k] = v
		}
		fields := map[string]interface{}{
			resourceMetric.Operation: resourceMetric.Value,
		}
		UnitTime := time.Unix(0, int64(resourceMetric.UnitTime))
		if UnitTime.IsZero() {
			UnitTime = time.Now()
		}
		// 指标集名称定义：'instrumentationName'
		pt, err := dkio.NewPoint(resourceMetric.Service, tags, fields, &dkio.PointOption{
			Time:              UnitTime,
			Category:          datakit.Metric,
			DisableGlobalTags: false,
			Strict:            true,
		})
		if err != nil {
			l.Errorf("make point err=%v", err)
			continue
		}

		pts = append(pts, pt)
	}
	return pts
}
