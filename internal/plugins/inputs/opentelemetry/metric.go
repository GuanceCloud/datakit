// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package opentelemetry

import (
	"time"

	metrics "github.com/GuanceCloud/tracing-protos/opentelemetry-gen-go/metrics/v1"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/point"
	itrace "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/trace"
)

type pointData struct {
	value  interface{}
	tags   map[string]string
	fields map[string]interface{}
	ts     int64
}

type OTELMetrics struct {
	resource    string
	name        string
	description string
	tags        map[string]string
	fields      map[string]interface{}
	points      []*pointData
}

func (m *OTELMetrics) getPoints() []*point.Point {
	if m.tags == nil {
		m.tags = make(map[string]string)
	}
	m.tags["instrumentation_name"] = m.name
	m.tags["description"] = m.description

	var pts []*point.Point
	for i := range m.points {
		if m.points[i].fields == nil {
			m.points[i].fields = make(map[string]interface{})
		}
		m.points[i].fields[m.name] = m.points[i].value

		var tm time.Time
		if m.points[i].ts == 0 {
			tm = time.Now()
		} else {
			tm = time.Unix(0, m.points[i].ts)
		}
		if pt, err := point.NewPoint("otel-service",
			itrace.MergeTags(m.tags, m.points[i].tags),
			itrace.MergeFields(m.fields, m.points[i].fields),
			&point.PointOption{Time: tm, Category: datakit.Metric, Strict: true}); err != nil {
			log.Debugf(err.Error())
		} else {
			pts = append(pts, pt)
		}
	}

	return pts
}

func parseResourceMetrics(resmcs []*metrics.ResourceMetrics) []*OTELMetrics {
	var omcs []*OTELMetrics
	for _, resmc := range resmcs {
		resattrs := extractAtrributes(resmc.Resource.Attributes)
		for _, scopemetrics := range resmc.ScopeMetrics {
			scpattrs := extractAtrributes(scopemetrics.Scope.Attributes)

			res := scopemetrics.Scope.Name
			for _, metric := range scopemetrics.Metrics {
				omc := &OTELMetrics{
					resource:    res,
					name:        metric.Name,
					description: metric.Description,
					points:      extractMetricPoints(metric, newAttributes(resattrs).merge(scpattrs...)),
				}
				omcs = append(omcs, omc)
			}
		}
	}

	return omcs
}

func extractMetricPoints(metric *metrics.Metric, attrs *attributes) []*pointData {
	var points []*pointData
	switch t := metric.Data.(type) {
	case *metrics.Metric_Gauge:
		for _, pt := range t.Gauge.DataPoints {
			tags, fields := attrs.merge(extractAtrributes(pt.Attributes)...).splite()
			data := &pointData{tags: tags, fields: fields}
			if v, ok := pt.Value.(*metrics.NumberDataPoint_AsDouble); ok {
				data.value = v.AsDouble
			} else if v, ok := pt.Value.(*metrics.NumberDataPoint_AsInt); ok {
				data.value = v.AsInt
			}
			data.ts = int64(pt.TimeUnixNano)
			points = append(points, data)
		}
	case *metrics.Metric_Sum:
		for _, pt := range t.Sum.DataPoints {
			tags, fields := attrs.merge(extractAtrributes(pt.Attributes)...).splite()
			data := &pointData{tags: tags, fields: fields}
			if v, ok := pt.Value.(*metrics.NumberDataPoint_AsDouble); ok {
				data.value = v.AsDouble
			} else if v, ok := pt.Value.(*metrics.NumberDataPoint_AsInt); ok {
				data.value = v.AsInt
			}
			data.ts = int64(pt.TimeUnixNano)
			points = append(points, data)
		}
	case *metrics.Metric_Histogram:
		for _, pt := range t.Histogram.DataPoints {
			tags, fields := attrs.merge(extractAtrributes(pt.Attributes)...).splite()
			data := &pointData{tags: tags, fields: fields}
			data.value = pt.GetSum()
			data.ts = int64(pt.TimeUnixNano)
			points = append(points, data)
		}
	case *metrics.Metric_ExponentialHistogram:
		for _, pt := range t.ExponentialHistogram.DataPoints {
			tags, fields := attrs.merge(extractAtrributes(pt.Attributes)...).splite()
			data := &pointData{tags: tags, fields: fields}
			data.value = pt.GetSum()
			data.ts = int64(pt.TimeUnixNano)
			points = append(points, data)
		}
	case *metrics.Metric_Summary:
		for _, pt := range t.Summary.DataPoints {
			tags, fields := attrs.merge(extractAtrributes(pt.Attributes)...).splite()
			data := &pointData{tags: tags, fields: fields}
			data.value = pt.GetSum()
			data.ts = int64(pt.TimeUnixNano)
			points = append(points, data)
		}
	default:
		log.Warnf("unknown metric.Data type or deprecated Data type")
	}

	return points
}
