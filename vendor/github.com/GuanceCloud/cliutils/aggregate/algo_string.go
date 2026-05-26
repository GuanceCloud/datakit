package aggregate

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

func formatMetricBaseForCalc(mb *MetricBase) string {
	if mb == nil {
		return "base=<nil>"
	}

	tagParts := make([]string, 0, len(mb.aggrTags))
	for _, kv := range mb.aggrTags {
		tagParts = append(tagParts, fmt.Sprintf("%s=%s", kv[0], kv[1]))
	}

	nextWallTime := "<zero>"
	if mb.nextWallTime > 0 {
		nextWallTime = time.Unix(mb.nextWallTime, 0).UTC().Format(time.RFC3339)
	}

	return fmt.Sprintf(
		"base={name=%s key=%s hash=%d window=%s next_wall_time=%s heap_idx=%d tags=[%s]}",
		mb.name,
		mb.key,
		mb.hash,
		time.Duration(mb.window),
		nextWallTime,
		mb.heapIdx,
		strings.Join(tagParts, ", "),
	)
}

func formatFloat64Slice(values []float64) string {
	if len(values) == 0 {
		return "[]"
	}

	parts := make([]string, 0, len(values))
	for _, value := range values {
		parts = append(parts, fmt.Sprintf("%g", value))
	}

	return "[" + strings.Join(parts, ", ") + "]"
}

func formatHistogramBuckets(leBucket map[string]float64) string {
	if len(leBucket) == 0 {
		return "{}"
	}

	keys := make([]string, 0, len(leBucket))
	for key := range leBucket {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		parts = append(parts, fmt.Sprintf("%s=%g", key, leBucket[key]))
	}

	return "{" + strings.Join(parts, ", ") + "}"
}

func formatDistinctValues(values map[any]struct{}) string {
	if len(values) == 0 {
		return "[]"
	}

	parts := make([]string, 0, len(values))
	for value := range values {
		parts = append(parts, fmt.Sprintf("%T:%v", value, value))
	}
	sort.Strings(parts)

	return "[" + strings.Join(parts, ", ") + "]"
}

func (c *algoSum) ToString() string {
	return fmt.Sprintf(
		"algoSum{delta=%g count=%d max_time=%d %s}",
		c.delta,
		c.count,
		c.maxTime,
		formatMetricBaseForCalc(&c.MetricBase),
	)
}

func (c *algoAvg) ToString() string {
	return fmt.Sprintf(
		"algoAvg{delta=%g count=%d max_time=%d %s}",
		c.delta,
		c.count,
		c.maxTime,
		formatMetricBaseForCalc(&c.MetricBase),
	)
}

func (c *algoCount) ToString() string {
	return fmt.Sprintf(
		"algoCount{count=%d max_time=%d %s}",
		c.count,
		c.maxTime,
		formatMetricBaseForCalc(&c.MetricBase),
	)
}

func (c *algoMax) ToString() string {
	return fmt.Sprintf(
		"algoMax{max=%g count=%d max_time=%d %s}",
		c.max,
		c.count,
		c.maxTime,
		formatMetricBaseForCalc(&c.MetricBase),
	)
}

func (a *algoMin) ToString() string {
	return fmt.Sprintf(
		"algoMin{min=%g count=%d max_time=%d %s}",
		a.min,
		a.count,
		a.maxTime,
		formatMetricBaseForCalc(&a.MetricBase),
	)
}

func (c *algoHistogram) ToString() string {
	return fmt.Sprintf(
		"algoHistogram{count=%d val=%g max_time=%d buckets=%s %s}",
		c.count,
		c.val,
		c.maxTime,
		formatHistogramBuckets(c.leBucket),
		formatMetricBaseForCalc(&c.MetricBase),
	)
}

func (a *algoQuantiles) ToString() string {
	return fmt.Sprintf(
		"algoQuantiles{count=%d max_time=%d quantiles=%s all=%s %s}",
		a.count,
		a.maxTime,
		formatFloat64Slice(a.quantiles),
		formatFloat64Slice(a.all),
		formatMetricBaseForCalc(&a.MetricBase),
	)
}

func (c *algoStdev) ToString() string {
	return fmt.Sprintf(
		"algoStdev{count=%d max_time=%d data=%s %s}",
		len(c.data),
		c.maxTime,
		formatFloat64Slice(c.data),
		formatMetricBaseForCalc(&c.MetricBase),
	)
}

func (c *algoCountDistinct) ToString() string {
	return fmt.Sprintf(
		"algoCountDistinct{count=%d max_time=%d distinct_values=%s %s}",
		len(c.distinctValues),
		c.maxTime,
		formatDistinctValues(c.distinctValues),
		formatMetricBaseForCalc(&c.MetricBase),
	)
}

func (a *algoCountFirst) ToString() string {
	return fmt.Sprintf(
		"algoCountFirst{first=%g first_time=%d count=%d %s}",
		a.first,
		a.firstTime,
		a.count,
		formatMetricBaseForCalc(&a.MetricBase),
	)
}

func (a *algoCountLast) ToString() string {
	return fmt.Sprintf(
		"algoCountLast{last=%g last_time=%d count=%d %s}",
		a.last,
		a.lastTime,
		a.count,
		formatMetricBaseForCalc(&a.MetricBase),
	)
}
