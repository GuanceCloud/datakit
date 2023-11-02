// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package solr

import (
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

type measurement struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
	ts     time.Time
}

// ----------------------- Solr v7.x + -----------------
// ---------------------- measurement ------------------

type SolrRequestTimes measurement

// Point implement MeasurementV2.
func (m *SolrRequestTimes) Point() *point.Point {
	opts := point.DefaultMetricOptions()
	opts = append(opts, point.WithTime(m.ts))

	return point.NewPointV2(m.name,
		append(point.NewTags(m.tags), point.NewKVs(m.fields)...),
		opts...)
}

//nolint:lll
func (m *SolrRequestTimes) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: metricNameRequestTimes,
		Type: "metric",
		Desc: "Request handler request times statistics.",
		Tags: map[string]interface{}{
			"host":     &inputs.TagInfo{Desc: "System hostname."},
			"core":     &inputs.TagInfo{Desc: "Solr core."},
			"category": &inputs.TagInfo{Desc: "Category name."},
			"handler":  &inputs.TagInfo{Desc: "Request handler."},
			"group":    &inputs.TagInfo{Desc: "Metric group."},
			"instance": &inputs.TagInfo{Desc: "Instance name, generated based on server address."},
		},
		Fields: map[string]interface{}{
			"count":      newFieldInfoCount("Total number of requests made since the Solr process was started."),
			"rate_mean":  newFieldInfoRPS("Average number of requests per second received"),
			"rate_1min":  newFieldInfoRPS("Requests per second received over the past 1 minutes."),
			"rate_5min":  newFieldInfoRPS("Requests per second received over the past 5 minutes."),
			"rate_15min": newFieldInfoRPS("Requests per second received over the past 15 minutes."),
			"min":        newFieldInfoFloatMS("Min of all the request processing time."),
			"max":        newFieldInfoFloatMS("Max of all the request processing time."),
			"mean":       newFieldInfoFloatMS("Mean of all the request processing time."),
			"median":     newFieldInfoFloatMS("Median of all the request processing time."),
			"stddev":     newFieldInfoFloatMS("Stddev of all the request processing time."),
			"p75":        newFieldInfoFloatMS("Request processing time for the request which belongs to the 75th Percentile."),
			"p95":        newFieldInfoFloatMS("Request processing time in milliseconds for the request which belongs to the 95th Percentile."),
			"p99":        newFieldInfoFloatMS("Request processing time in milliseconds for the request which belongs to the 99th Percentile."),
			"p999":       newFieldInfoFloatMS("Request processing time in milliseconds for the request which belongs to the 99.9th Percentile."),
		},
	}
}

type SolrCache measurement

// Point implement MeasurementV2.
func (m *SolrCache) Point() *point.Point {
	opts := point.DefaultMetricOptions()
	opts = append(opts, point.WithTime(m.ts))

	return point.NewPointV2(m.name,
		append(point.NewTags(m.tags), point.NewKVs(m.fields)...),
		opts...)
}

//nolint:lll
func (m *SolrCache) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: metricNameCache,
		Type: "metric",
		Desc: "Cache statistics.",
		Tags: map[string]interface{}{
			"host":     &inputs.TagInfo{Desc: "System hostname."},
			"core":     &inputs.TagInfo{Desc: "Solr core."},
			"name":     &inputs.TagInfo{Desc: "Cache name."},
			"category": &inputs.TagInfo{Desc: "Category name."},
			"group":    &inputs.TagInfo{Desc: "Metric group."},
			"instance": &inputs.TagInfo{Desc: "Instance name, generated based on server address."},
		},
		Fields: map[string]interface{}{
			"cumulative_evictions": newFieldInfoCount("Number of cache evictions across all caches since this node has been running."),
			"cumulative_hitratio":  newFieldInfoPercent("Ratio of cache hits to lookups across all the caches since this node has been running."),
			"cumulative_hits":      newFieldInfoCount("Number of cache hits across all the caches since this node has been running."),
			"cumulative_inserts":   newFieldInfoCount("Number of cache insertions across all the caches since this node has been running."),
			"cumulative_lookups":   newFieldInfoCount("Number of cache lookups across all the caches since this node has been running."),
			"evictions":            newFieldInfoCount("Number of cache evictions for the current index searcher."),
			"hitratio":             newFieldInfoPercent("Ratio of cache hits to lookups for the current index searcher."),
			"hits":                 newFieldInfoCount("Number of hits for the current index searcher."),
			"inserts":              newFieldInfoCount("Number of inserts into the cache."),
			"lookups":              newFieldInfoCount("Number of lookups against the cache."),
			"size":                 newFieldInfoCount("Number of entries in the cache at that particular instance."),
			"warmup":               newFieldInfoIntMS("Warm-up time for the registered index searcher. This time is taken in account for the \"auto-warming\" of caches."),
			"max_ram":              newFieldInfoMiB("Maximum heap that should be used by the cache beyond which keys will be evicted."),
			"ram_bytes_used":       newFieldInfoByte("Actual heap usage of the cache at that particular instance."),
		},
	}
}

type SolrSearcher measurement

// Point implement MeasurementV2.
func (m *SolrSearcher) Point() *point.Point {
	opts := point.DefaultMetricOptions()
	opts = append(opts, point.WithTime(m.ts))

	return point.NewPointV2(m.name,
		append(point.NewTags(m.tags), point.NewKVs(m.fields)...),
		opts...)
}

func (m *SolrSearcher) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: metricNameSearcher,
		Type: "metric",
		Desc: "Searcher Statistics",
		Tags: map[string]interface{}{
			"host":     &inputs.TagInfo{Desc: "System hostname."},
			"core":     &inputs.TagInfo{Desc: "Solr core."},
			"category": &inputs.TagInfo{Desc: "Category name."},
			"group":    &inputs.TagInfo{Desc: "Metric group."},
			"instance": &inputs.TagInfo{Desc: "Instance name, generated based on server address."},
		},
		Fields: map[string]interface{}{
			"deleted_docs": newFieldInfoCount("The number of deleted documents."),
			"max_docs":     newFieldInfoCount("The largest possible document number."),
			"num_docs":     newFieldInfoCount("The total number of indexed documents."),
			"warmup":       newFieldInfoIntMS("The time spent warming up."),
		},
	}
}

// ----------------------- newFieldInfo ----------------

// count.
func newFieldInfoCount(desc string) *inputs.FieldInfo {
	return &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.NCount,
		Desc:     desc,
	}
}

// 吞吐量.
func newFieldInfoRPS(desc string) *inputs.FieldInfo {
	return &inputs.FieldInfo{
		DataType: inputs.Float,
		Type:     inputs.Gauge,
		Unit:     inputs.RequestsPerSec,
		Desc:     desc,
	}
}

// 时间 ms.
func newFieldInfoFloatMS(desc string) *inputs.FieldInfo {
	return &inputs.FieldInfo{
		DataType: inputs.Float,
		Type:     inputs.Gauge,
		Unit:     inputs.DurationMS,
		Desc:     desc,
	}
}

// 时间 ms.
func newFieldInfoIntMS(desc string) *inputs.FieldInfo {
	return &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.DurationMS,
		Desc:     desc,
	}
}

func newFieldInfoMiB(desc string) *inputs.FieldInfo {
	return &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.SizeMB,
		Desc:     desc,
	}
}

func newFieldInfoByte(desc string) *inputs.FieldInfo {
	return &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.SizeByte,
		Desc:     desc,
	}
}

// percent %.
func newFieldInfoPercent(desc string) *inputs.FieldInfo {
	return &inputs.FieldInfo{
		DataType: inputs.Float,
		Type:     inputs.Gauge,
		Unit:     inputs.Percent,
		Desc:     desc,
	}
}
