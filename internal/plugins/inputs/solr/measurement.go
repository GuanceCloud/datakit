// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package solr

import (
	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

type measurement struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
	ts     int64
}

var (
	solrRequestTimesTaggedby = []string{"host", "core", "category", "handler", "group", "instance"}
	solrCacheTaggedby        = []string{"host", "core", "name", "category", "group", "instance"}
	solrSearcherTaggedby     = []string{"host", "core", "category", "group", "instance"}
)

// ----------------------- Solr v7.x + -----------------
// ---------------------- measurement ------------------

type SolrRequestTimes measurement

// Point implement MeasurementV2.
func (m *SolrRequestTimes) Point() *point.Point {
	opts := point.DefaultMetricOptions()
	opts = append(opts, point.WithTimestamp(m.ts))

	return point.NewPoint(m.name,
		append(point.NewTags(m.tags), point.NewKVs(m.fields)...),
		opts...)
}

//nolint:lll
func (m *SolrRequestTimes) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   metricNameRequestTimes,
		Cat:    point.Metric,
		Desc:   "Request-time statistics for Solr request handlers.",
		DescZh: "Solr 请求处理器的请求速率和处理耗时统计。",
		Tags: map[string]interface{}{
			"host":     &inputs.TagInfo{Desc: "Hostname of the Solr node."},
			"core":     &inputs.TagInfo{Desc: "Solr core that emitted the metric."},
			"category": &inputs.TagInfo{Desc: "Solr metric category that owns the metric."},
			"handler":  &inputs.TagInfo{Desc: "Solr request handler path or name."},
			"group":    &inputs.TagInfo{Desc: "Solr metric group that owns the metric."},
			"instance": &inputs.TagInfo{Desc: "Instance name, generated based on server address."},
		},
		Fields: map[string]interface{}{
			"count":      &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Total number of requests made since the Solr process was started.", Taggedby: solrRequestTimesTaggedby},
			"rate_mean":  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.RequestsPerSec, Desc: "Average number of requests per second received.", Taggedby: solrRequestTimesTaggedby},
			"rate_1min":  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.RequestsPerSec, Desc: "Requests per second received over the past 1 minute.", Taggedby: solrRequestTimesTaggedby},
			"rate_5min":  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.RequestsPerSec, Desc: "Requests per second received over the past 5 minutes.", Taggedby: solrRequestTimesTaggedby},
			"rate_15min": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.RequestsPerSec, Desc: "Requests per second received over the past 15 minutes.", Taggedby: solrRequestTimesTaggedby},
			"min":        &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: "Minimum request processing time.", Taggedby: solrRequestTimesTaggedby},
			"max":        &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: "Maximum request processing time.", Taggedby: solrRequestTimesTaggedby},
			"mean":       &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: "Mean request processing time.", Taggedby: solrRequestTimesTaggedby},
			"median":     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: "Median request processing time.", Taggedby: solrRequestTimesTaggedby},
			"stddev":     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: "Standard deviation of request processing time.", Taggedby: solrRequestTimesTaggedby},
			"p75":        &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: "75th percentile request processing time.", Taggedby: solrRequestTimesTaggedby},
			"p95":        &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: "95th percentile request processing time.", Taggedby: solrRequestTimesTaggedby},
			"p99":        &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: "99th percentile request processing time.", Taggedby: solrRequestTimesTaggedby},
			"p999":       &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: "99.9th percentile request processing time.", Taggedby: solrRequestTimesTaggedby},
		},
	}
}

type SolrCache measurement

// Point implement MeasurementV2.
func (m *SolrCache) Point() *point.Point {
	opts := point.DefaultMetricOptions()
	opts = append(opts, point.WithTimestamp(m.ts))

	return point.NewPoint(m.name,
		append(point.NewTags(m.tags), point.NewKVs(m.fields)...),
		opts...)
}

//nolint:lll
func (m *SolrCache) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   metricNameCache,
		Cat:    point.Metric,
		Desc:   "Solr cache performance and memory statistics.",
		DescZh: "Solr 缓存命中、查找、驱逐、容量和内存使用统计。",
		Tags: map[string]interface{}{
			"host":     &inputs.TagInfo{Desc: "Hostname of the Solr node."},
			"core":     &inputs.TagInfo{Desc: "Solr core that emitted the metric."},
			"name":     &inputs.TagInfo{Desc: "Name of the Solr cache."},
			"category": &inputs.TagInfo{Desc: "Solr metric category that owns the metric."},
			"group":    &inputs.TagInfo{Desc: "Solr metric group that owns the metric."},
			"instance": &inputs.TagInfo{Desc: "Instance name, generated based on server address."},
		},
		Fields: map[string]interface{}{
			"cumulative_evictions": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of cache evictions across all caches since this node has been running.", Taggedby: solrCacheTaggedby},
			"cumulative_hitratio":  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.Percent, Desc: "Ratio of cache hits to lookups across all the caches since this node has been running.", Taggedby: solrCacheTaggedby},
			"cumulative_hits":      &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of cache hits across all the caches since this node has been running.", Taggedby: solrCacheTaggedby},
			"cumulative_inserts":   &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of cache insertions across all the caches since this node has been running.", Taggedby: solrCacheTaggedby},
			"cumulative_lookups":   &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of cache lookups across all the caches since this node has been running.", Taggedby: solrCacheTaggedby},
			"evictions":            &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of cache evictions for the current index searcher.", Taggedby: solrCacheTaggedby},
			"hitratio":             &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.Percent, Desc: "Ratio of cache hits to lookups for the current index searcher.", Taggedby: solrCacheTaggedby},
			"hits":                 &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of hits for the current index searcher.", Taggedby: solrCacheTaggedby},
			"inserts":              &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of inserts into the cache.", Taggedby: solrCacheTaggedby},
			"lookups":              &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Number of lookups against the cache.", Taggedby: solrCacheTaggedby},
			"size":                 &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of entries in the cache at that particular instance.", Taggedby: solrCacheTaggedby},
			"warmup":               &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: "Warm-up time for the registered index searcher. This time is taken in account for the \"auto-warming\" of caches.", Taggedby: solrCacheTaggedby},
			"max_ram":              &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeMB, Desc: "Maximum heap that should be used by the cache beyond which keys will be evicted.", Taggedby: solrCacheTaggedby},
			"ram_bytes_used":       &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Actual heap usage of the cache at that particular instance.", Taggedby: solrCacheTaggedby},
		},
	}
}

type SolrSearcher measurement

// Point implement MeasurementV2.
func (m *SolrSearcher) Point() *point.Point {
	opts := point.DefaultMetricOptions()
	opts = append(opts, point.WithTimestamp(m.ts))

	return point.NewPoint(m.name,
		append(point.NewTags(m.tags), point.NewKVs(m.fields)...),
		opts...)
}

func (m *SolrSearcher) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   metricNameSearcher,
		Cat:    point.Metric,
		Desc:   "Searcher lifecycle and document-count statistics.",
		DescZh: "Solr searcher 生命周期和文档数量统计。",
		Tags: map[string]interface{}{
			"host":     &inputs.TagInfo{Desc: "Hostname of the Solr node."},
			"core":     &inputs.TagInfo{Desc: "Solr core that emitted the metric."},
			"category": &inputs.TagInfo{Desc: "Solr metric category that owns the metric."},
			"group":    &inputs.TagInfo{Desc: "Solr metric group that owns the metric."},
			"instance": &inputs.TagInfo{Desc: "Instance name, generated based on server address."},
		},
		Fields: map[string]interface{}{
			"deleted_docs": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The number of deleted documents.",
				Taggedby: solrSearcherTaggedby,
			},
			"max_docs": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The largest possible document number.",
				Taggedby: solrSearcherTaggedby,
			},
			"num_docs": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The total number of indexed documents.",
				Taggedby: solrSearcherTaggedby,
			},
			"warmup": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.DurationMS,
				Desc:     "The time spent warming up.",
				Taggedby: solrSearcherTaggedby,
			},
		},
	}
}
