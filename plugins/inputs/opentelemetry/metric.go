// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package opentelemetry

import (
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	itrace "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/trace"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/point"
	metricspb "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/opentelemetry/compiled/v1/metrics"
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

func parseResourceMetrics(resmcs []*metricspb.ResourceMetrics) []*OTELMetrics {
	var omcs []*OTELMetrics
	for _, resmc := range resmcs {
		restags, resfields := extractAtrribute(resmc.Resource.Attributes)
		for _, scopemetrics := range resmc.ScopeMetrics {
			scopetags, scopefields := extractAtrribute(scopemetrics.Scope.Attributes)
			scopetags = itrace.MergeTags(restags, scopetags)
			scopefields = itrace.MergeFields(resfields, scopefields)

			res := scopemetrics.Scope.Name
			for _, metric := range scopemetrics.Metrics {
				omc := &OTELMetrics{
					resource:    res,
					name:        metric.Name,
					description: metric.Description,
					points:      extractMetricPoints(metric, scopetags, scopefields),
				}
				omcs = append(omcs, omc)
			}
		}
	}

	return omcs
}

func extractMetricPoints(metric *metricspb.Metric, extratags map[string]string, extrafields map[string]interface{}) []*pointData {
	var points []*pointData
	switch t := metric.Data.(type) {
	case *metricspb.Metric_Gauge:
		for _, pt := range t.Gauge.DataPoints {
			tags, fields := extractAtrribute(pt.Attributes)
			tags = itrace.MergeTags(extratags, tags)
			fields = itrace.MergeFields(extrafields, fields)
			data := &pointData{tags: tags, fields: fields}
			if v, ok := pt.Value.(*metricspb.NumberDataPoint_AsDouble); ok {
				data.value = v.AsDouble
			} else if v, ok := pt.Value.(*metricspb.NumberDataPoint_AsInt); ok {
				data.value = v.AsInt
			}
			data.ts = int64(pt.TimeUnixNano)
			points = append(points, data)
		}
	case *metricspb.Metric_Sum:
		for _, pt := range t.Sum.DataPoints {
			tags, fields := extractAtrribute(pt.Attributes)
			tags = itrace.MergeTags(extratags, tags)
			fields = itrace.MergeFields(extrafields, fields)
			data := &pointData{tags: tags, fields: fields}
			if v, ok := pt.Value.(*metricspb.NumberDataPoint_AsDouble); ok {
				data.value = v.AsDouble
			} else if v, ok := pt.Value.(*metricspb.NumberDataPoint_AsInt); ok {
				data.value = v.AsInt
			}
			data.ts = int64(pt.TimeUnixNano)
			points = append(points, data)
		}
	case *metricspb.Metric_Histogram:
		for _, pt := range t.Histogram.DataPoints {
			tags, fields := extractAtrribute(pt.Attributes)
			tags = itrace.MergeTags(extratags, tags)
			fields = itrace.MergeFields(extrafields, fields)
			data := &pointData{tags: tags, fields: fields}
			data.value = pt.GetSum()
			data.ts = int64(pt.TimeUnixNano)
			points = append(points, data)
		}
	case *metricspb.Metric_ExponentialHistogram:
		for _, pt := range t.ExponentialHistogram.DataPoints {
			tags, fields := extractAtrribute(pt.Attributes)
			tags = itrace.MergeTags(extratags, tags)
			fields = itrace.MergeFields(extrafields, fields)
			data := &pointData{tags: tags, fields: fields}
			data.value = pt.GetSum()
			data.ts = int64(pt.TimeUnixNano)
			points = append(points, data)
		}
	case *metricspb.Metric_Summary:
		for _, pt := range t.Summary.DataPoints {
			tags, fields := extractAtrribute(pt.Attributes)
			tags = itrace.MergeTags(extratags, tags)
			fields = itrace.MergeFields(extrafields, fields)
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
