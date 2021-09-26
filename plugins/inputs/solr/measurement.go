package solr

import (
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

type measurement struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
	ts     time.Time
}

func (m *measurement) LineProto() (*io.Point, error) {
	return io.MakePoint(m.name, m.tags, m.fields, m.ts)
}

func (m *measurement) Info() *inputs.MeasurementInfo {
	return nil
}

// ----------------------- Solr v7.x + -----------------
// ---------------------- measurement ------------------

type SolrRequestTimes measurement

func (m *SolrRequestTimes) LineProto() (*io.Point, error) {
	return io.MakePoint(m.name, m.tags, m.fields, m.ts)
}

func (m *SolrRequestTimes) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: metricNameRequestTimes,
		Desc: "Request handler request times statistics.",
		Tags: map[string]interface{}{
			"host":     &inputs.TagInfo{Desc: "host name"},
			"core":     &inputs.TagInfo{Desc: "solr core"},
			"category": &inputs.TagInfo{Desc: "category name"},
			"handler":  &inputs.TagInfo{Desc: "request handler"},
			"group":    &inputs.TagInfo{Desc: "metric group"},
			"instance": &inputs.TagInfo{Desc: "instance name, generated based on server address"},
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
			"p95":        newFieldInfoFloatMS("Request processing time in milliseconds for the request which belongs to the 95th Percentile. "),
			"p99":        newFieldInfoFloatMS("Request processing time in milliseconds for the request which belongs to the 99th Percentile. "),
			"p999":       newFieldInfoFloatMS("Request processing time in milliseconds for the request which belongs to the 99.9th Percentile. "),
		},
	}
}

type SolrCache measurement

func (m *SolrCache) LineProto() (*io.Point, error) {
	return io.MakePoint(m.name, m.tags, m.fields, m.ts)
}

func (m *SolrCache) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: metricNameCache,
		Desc: "Cache statistics.",
		Tags: map[string]interface{}{
			"host":     &inputs.TagInfo{Desc: "host name"},
			"core":     &inputs.TagInfo{Desc: "solr core"},
			"name":     &inputs.TagInfo{Desc: "cache name"},
			"category": &inputs.TagInfo{Desc: "category name"},
			"group":    &inputs.TagInfo{Desc: "metric group"},
			"instance": &inputs.TagInfo{Desc: "instance name, generated based on server address"},
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

func (m *SolrSearcher) LineProto() (*io.Point, error) {
	return io.MakePoint(m.name, m.tags, m.fields, m.ts)
}

func (m *SolrSearcher) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: metricNameSearcher,
		Desc: "Searcher Statistics",
		Tags: map[string]interface{}{
			"host":     &inputs.TagInfo{Desc: "host name"},
			"core":     &inputs.TagInfo{Desc: "solr core"},
			"category": &inputs.TagInfo{Desc: "category name"},
			"group":    &inputs.TagInfo{Desc: "metric group"},
			"instance": &inputs.TagInfo{Desc: "instance name, generated based on server address"},
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
		Unit:     inputs.SizeMiB,
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
