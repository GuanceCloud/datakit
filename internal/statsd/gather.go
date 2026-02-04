// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package statsd

import (
	"fmt"
	"time"

	"github.com/GuanceCloud/cliutils/point"
)

func (col *Collector) GetLoggings() []*point.Point {
	if len(col.loggingPts) > 0 {
		var arr []*point.Point
		arr = append(arr, col.loggingPts...)

		col.loggingPts = col.loggingPts[:0]
		return arr
	}

	return nil
}

func (col *Collector) GetPoints() ([]*point.Point, error) {
	col.opts.l.Debugf("try locking...")
	col.Lock()
	defer col.Unlock()
	now := time.Now()

	for _, m := range col.distributions {
		fields := map[string]any{
			defaultFieldName: m.value,
		}
		col.opts.l.Debugf("[on distributions] add %s, fields: %+#v, tags: %+#v", m.name, fields, m.tags)
		col.acc.addFields(m.name, fields, m.tags, now)
	}

	col.distributions = make([]cacheddistributions, 0)

	for _, m := range col.timings {
		// Defining a template to parse field names for timers allows us to split
		// out multiple fields per timer. In this case we prefix each stat with the
		// field name and store these all in a single measurement.
		fields := make(map[string]any)
		for fieldName, stats := range m.fields {
			var prefix string
			if fieldName != defaultFieldName {
				prefix = fieldName + "_"
			}
			fields[prefix+"mean"] = stats.Mean()
			fields[prefix+"stddev"] = stats.Stddev()
			fields[prefix+"sum"] = stats.Sum()
			fields[prefix+"upper"] = stats.Upper()
			fields[prefix+"lower"] = stats.Lower()
			fields[prefix+"count"] = stats.Count()
			for _, percentile := range col.opts.percentiles {
				name := fmt.Sprintf("%s%v_percentile", prefix, percentile)
				fields[name] = stats.Percentile(percentile)
			}
		}

		col.opts.l.Debugf("[on timings] add %s, fields: %+#v, tags: %+#v", m.name, fields, m.tags)
		col.acc.addFields(m.name, fields, m.tags, now)
	}

	if col.opts.deleteTimings {
		col.timings = make(map[string]cachedtimings)
	}

	for _, m := range col.gauges {
		col.opts.l.Debugf("[on gauges] add %s, fields: %+#v, tags: %+#v", m.name, m.fields, m.tags)
		col.acc.addFields(m.name, m.fields, m.tags, now)
	}
	if col.opts.deleteGauges {
		col.gauges = make(map[string]cachedgauge)
	}

	for _, m := range col.counters {
		col.opts.l.Debugf("[on counters] add %s, fields: %+#v, tags: %+#v", m.name, m.fields, m.tags)
		col.acc.addFields(m.name, m.fields, m.tags, now)
	}
	if col.opts.deleteCounters {
		col.counters = make(map[string]cachedcounter)
	}

	for _, m := range col.sets {
		fields := make(map[string]any)
		for field, set := range m.fields {
			fields[field] = int64(len(set))
		}
		col.opts.l.Debugf("[on sets] add %s, fields: %+#v, tags: %+#v", m.name, m.fields, m.tags)
		col.acc.addFields(m.name, fields, m.tags, now)
	}

	if col.opts.deleteSets {
		col.sets = make(map[string]cachedset)
	}

	points := make([]*point.Point, 0)
	if len(col.acc.points) > 0 {
		points = append(points, col.acc.points...)
		col.acc.points = col.acc.points[:0]
	}
	col.expireCachedMetrics()

	collectPointsTotalVec.WithLabelValues().Observe(float64(len(points)))
	return points, nil
}

func (col *Collector) expireCachedMetrics() {
	// If Max TTL wasn't configured, skip expiration.
	if col.opts.maxTTL == 0 {
		return
	}

	now := time.Now()

	for key, cached := range col.gauges {
		if now.After(cached.expiresAt) {
			delete(col.gauges, key)
		}
	}

	for key, cached := range col.sets {
		if now.After(cached.expiresAt) {
			delete(col.sets, key)
		}
	}

	for key, cached := range col.timings {
		if now.After(cached.expiresAt) {
			delete(col.timings, key)
		}
	}

	for key, cached := range col.counters {
		if now.After(cached.expiresAt) {
			delete(col.counters, key)
		}
	}
}
