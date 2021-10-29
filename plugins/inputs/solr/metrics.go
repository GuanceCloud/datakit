package solr

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// --------------------- solr v6.6 + -------------------
// ------------------- stats struct --------------------

const (
	prefixSearcher          = "SEARCHER.searcher."
	prefixRegexRequesttimes = "(QUERY|UPDATE)\\./.*\\.requestTimes"
	prefixRegexCache        = "CACHE\\.searcher\\.(document|queryResult|filter)Cache"
)

// RequestTimesStats request times/errors/timeout
//    SEARCH(select) and UPDATE(update)
// Use map instead, fields need to be filtered.
type RequestTimesStats struct {
	Count     int64   `json:"count"`
	RateMean  float64 `json:"meanRate"`
	Rate1min  float64 `json:"1minRate"`
	Rate5min  float64 `json:"5minRate"`
	Rate15min float64 `json:"15minRate"`
	Min       float64 `json:"min_ms"`
	Max       float64 `json:"max_ms"`
	Mean      float64 `json:"mean_ms"`
	Median    float64 `json:"median_ms"`
	Stddev    float64 `json:"stddev_ms"`
	P75       float64 `json:"p75_ms"`
	P95       float64 `json:"p95_ms"`
	P99       float64 `json:"p99_ms"`
	P999      float64 `json:"p999_ms"`
}

type CacheStats struct {
	CumulativeEvictions int64   `json:"cumulative_evictions"`
	CumulativeHitratio  float64 `json:"cumulative_hitratio"`
	CumulativeHits      int64   `json:"cumulative_hits"`
	CumulativeInserts   int64   `json:"cumulative_inserts"`
	CumulativeLookups   int64   `json:"cumulative_lookups"`
	Evictions           int64   `json:"cevictions"`
	Hitratio            float64 `json:"hitratio"`
	Hits                int64   `json:"hits"`
	Inserts             int64   `json:"inserts"`
	Lookups             int64   `json:"lookups"`
	Size                int64   `json:"size"`
	Warmup              int64   `json:"warmupTime"`
	MaxRAMInMB          int64   `json:"maxRamMB"`
	RAMBytesUsed        int64   `json:"ramBytesUsed"`
}

type SearcherStats struct {
	DeletedDocs int64 `json:"SEARCHER.searcher.deletedDocs"`
	MaxDocs     int64 `json:"SEARCHER.searcher.maxDoc"`
	NumDocs     int64 `json:"SEARCHER.searcher.numDocs"`
	Warmup      int64 `json:"SEARCHER.searcher.warmupTime"`
}

// --------------------- solr v6.6 + -------------------
// ----------------------- response struct -------------

type ResponseHeader struct {
	QTime  int64 `json:"QTime"`
	Status int64 `json:"status"`
}

type Response struct {
	Resp    ResponseHeader                        `json:"responseHeader"`
	Metrics map[string]map[string]json.RawMessage `json:"metrics"`
}

type SearcherResp struct {
	Resp    ResponseHeader           `json:"responseHeader"`
	Metrics map[string]SearcherStats `json:"metrics"`
}

type CacheResp struct {
	Resp    ResponseHeader                   `json:"responseHeader"`
	Metrics map[string]map[string]CacheStats `json:"metrics"`
}

type RequestTimesResp struct {
	Resp    ResponseHeader                          `json:"responseHeader"`
	Metrics map[string]map[string]RequestTimesStats `json:"metrics"`
}

// --------------------- solr v6.6 + -------------------
// ----------------------- request url -----------------

const (
	adminMetric = "solr/admin/metrics"
)

func URLSearcher(server string) string {
	param := [][2]string{
		{"group", "core"},
		{"wt", "json"},
		{"prefix", prefixSearcher},
	}
	return urljoin(server, adminMetric, param)
}

func URLRequestTimes(server string) string {
	param := [][2]string{
		{"group", "core"},
		{"wt", "json"},
		{"regex", prefixRegexRequesttimes},
	}
	return urljoin(server, adminMetric, param)
}

func URLCache(server string) string {
	param := [][2]string{
		{"group", "core"},
		{"wt", "json"},
		{"regex", prefixRegexCache},
	}
	return urljoin(server, adminMetric, param)
}

func URLAll(server string) string {
	param := [][2]string{
		{"group", "core"},
		{"group", "node"},
		{"wt", "json"},
		{"prefix", prefixSearcher},
		{"regex", prefixRegexCache},
		{"regex", prefixRegexRequesttimes},
	}
	return urljoin(server, adminMetric, param)
}

// --------------------- solr v6.6 + -------------------
// ----------------------- gather ----------------------

func (i *Input) gatherSolrSearcher(k string, v json.RawMessage, fields map[string]interface{}) error {
	var err error
	kSplit := strings.Split(k, ".")
	if len(kSplit) != 3 {
		return fmt.Errorf("searcher stats key: %s", k)
	}
	var fieldValue int
	switch kSplit[2] {
	case "deletedDocs":
		err = json.Unmarshal(v, &fieldValue)
		fields["deleted_docs"] = fieldValue
	case "maxDoc":
		err = json.Unmarshal(v, &fieldValue)
		fields["max_docs"] = fieldValue
	case "numDocs":
		err = json.Unmarshal(v, &fieldValue)
		fields["num_docs"] = fieldValue
	case "warmupTime":
		err = json.Unmarshal(v, &fieldValue)
		fields["warmup"] = fieldValue
	}
	return err
}

func (i *Input) gatherSolrCache(k string, v json.RawMessage, commTags map[string]string, ts time.Time) error {
	var err error
	cacheStat := CacheStats{}
	if err = json.Unmarshal(v, &cacheStat); err != nil {
		return err
	}

	kSplit := strings.Split(k, ".")
	if len(kSplit) != 3 {
		return fmt.Errorf("cache stats key: %s", k)
	}

	tags := map[string]string{}
	for kTag, vTag := range commTags {
		tags[kTag] = vTag
	}

	tags["category"] = kSplit[0]
	tags["name"] = kSplit[2]

	fields := map[string]interface{}{
		"cumulative_evictions": cacheStat.CumulativeEvictions,
		"cumulative_hitratio":  cacheStat.CumulativeHitratio,
		"cumulative_hits":      cacheStat.CumulativeHits,
		"cumulative_inserts":   cacheStat.CumulativeInserts,
		"cumulative_lookups":   cacheStat.CumulativeLookups,
		"evictions":            cacheStat.Evictions,
		"hitratio":             cacheStat.Hitratio,
		"hits":                 cacheStat.Hits,
		"inserts":              cacheStat.Inserts,
		"lookups":              cacheStat.Lookups,
		"size":                 cacheStat.Size,
		"warmup":               cacheStat.Warmup,
		"max_ram":              cacheStat.MaxRAMInMB,
		"ram_bytes_used":       cacheStat.RAMBytesUsed,
	}
	i.appendM(&SolrCache{
		name:   metricNameCache,
		fields: fields,
		tags:   tags,
		ts:     ts,
	})

	return err
}

func (i *Input) gatherSolrRequestTimes(k string, v json.RawMessage, commTags map[string]string, ts time.Time) error {
	var err error
	rqtimes := RequestTimesStats{}
	if err = json.Unmarshal(v, &rqtimes); err != nil {
		return err
	}

	kSplit := strings.Split(k, ".")
	if len(kSplit) < 3 {
		return fmt.Errorf("request times stats key: %s", k)
	} else if len(kSplit) > 3 {
		return nil
	}
	tags := map[string]string{}
	for kTag, vTag := range commTags {
		tags[kTag] = vTag
	}

	tags["category"] = kSplit[0]
	tags["handler"] = kSplit[1]

	fields := map[string]interface{}{
		"count":      rqtimes.Count,
		"rate_mean":  rqtimes.RateMean,
		"rate_1min":  rqtimes.Rate1min,
		"rate_5min":  rqtimes.Rate5min,
		"rate_15min": rqtimes.Rate15min,
		"min":        rqtimes.Min,
		"max":        rqtimes.Max,
		"mean":       rqtimes.Mean,
		"median":     rqtimes.Median,
		"stddev":     rqtimes.Stddev,
		"p75":        rqtimes.P75,
		"p95":        rqtimes.P95,
		"p99":        rqtimes.P99,
		"p999":       rqtimes.P999,
	}
	i.appendM(&SolrRequestTimes{
		name:   metricNameRequestTimes,
		fields: fields,
		tags:   tags,
		ts:     ts,
	})

	return nil
}
