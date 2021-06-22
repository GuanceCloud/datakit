package statsd

import (
	"fmt"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

func (s *input) gather() {
	l.Debugf("try locking...")
	s.Lock()
	defer s.Unlock()
	now := time.Now()

	for _, m := range s.distributions {
		fields := map[string]interface{}{
			defaultFieldName: m.value,
		}
		l.Debugf("[distributions] add %s, fields: %+#v, tags: %+#v", m.name, fields, m.tags)
		s.acc.addFields(m.name, fields, m.tags, now)
	}
	s.distributions = make([]cacheddistributions, 0)

	for _, m := range s.timings {
		// Defining a template to parse field names for timers allows us to split
		// out multiple fields per timer. In this case we prefix each stat with the
		// field name and store these all in a single measurement.
		fields := make(map[string]interface{})
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
			for _, percentile := range s.Percentiles {
				name := fmt.Sprintf("%s%v_percentile", prefix, percentile)
				fields[name] = stats.Percentile(percentile)
			}
		}

		l.Debugf("[timings] add %s, fields: %+#v, tags: %+#v", m.name, fields, m.tags)
		s.acc.addFields(m.name, fields, m.tags, now)
	}
	if s.DeleteTimings {
		s.timings = make(map[string]cachedtimings)
	}

	for _, m := range s.gauges {
		l.Debugf("[gauges] add %s, fields: %+#v, tags: %+#v", m.name, m.fields, m.tags)
		s.acc.addFields(m.name, m.fields, m.tags, now)
	}
	if s.DeleteGauges {
		s.gauges = make(map[string]cachedgauge)
	}

	for _, m := range s.counters {
		l.Debugf("[counters] add %s, fields: %+#v, tags: %+#v", m.name, m.fields, m.tags)
		s.acc.addFields(m.name, m.fields, m.tags, now)
	}
	if s.DeleteCounters {
		s.counters = make(map[string]cachedcounter)
	}

	for _, m := range s.sets {
		fields := make(map[string]interface{})
		for field, set := range m.fields {
			fields[field] = int64(len(set))
		}
		l.Debugf("[sets] add %s, fields: %+#v, tags: %+#v", m.name, m.fields, m.tags)
		s.acc.addFields(m.name, fields, m.tags, now)
	}

	if s.DeleteSets {
		s.sets = make(map[string]cachedset)
	}

	l.Debugf("feed %d points...", len(s.acc.points))
	if len(s.acc.points) > 0 {
		if err := inputs.FeedMeasurement(inputName,
			datakit.Metric,
			s.acc.points,
			&io.Option{CollectCost: time.Since(now)}); err != nil {
			l.Error(err)
		} else {
			s.acc.points = s.acc.points[:0]
		}
	}

	s.expireCachedMetrics()
}
