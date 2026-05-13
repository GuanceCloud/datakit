// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package otlp

import (
	"fmt"
	"strconv"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	common "github.com/GuanceCloud/tracing-protos/opentelemetry-gen-go/common/v1"
	metrics "github.com/GuanceCloud/tracing-protos/opentelemetry-gen-go/metrics/v1"
)

const (
	DefaultMetricMeasurement  = "otel_service"
	DefaultMetricTemporality  = "__temporality"
	DefaultCollectorSourceTag = "collector_source_ip"
	DefaultScopeNameKey       = "scope_name"
)

type MetricsParserOptions struct {
	Measurement          string
	CollectorSourceIP    string
	CollectorSourceIPTag string
	ScopeNameKey         string
	TemporalityTagKey    string

	ResourceStringOptions StringMapOptions
	ScopeStringOptions    StringMapOptions
	PointStringOptions    StringMapOptions

	FieldName                func(metric *metrics.Metric, pointAttrs map[string]string) string
	MeasurementFromPointAttr func(metric *metrics.Metric, pointAttrs map[string]string) string
	KVOptions                func(metric *metrics.Metric) []point.KVOption
	PointOptions             func(ts time.Time) []point.Option
	ExponentialHistogramAvg  func(sum float64, count uint64) any
}

func defaultMetricsParserOptions(opts MetricsParserOptions) MetricsParserOptions {
	if opts.Measurement == "" {
		opts.Measurement = DefaultMetricMeasurement
	}
	if opts.CollectorSourceIPTag == "" {
		opts.CollectorSourceIPTag = DefaultCollectorSourceTag
	}
	if opts.ScopeNameKey == "" {
		opts.ScopeNameKey = DefaultScopeNameKey
	}
	if opts.TemporalityTagKey == "" {
		opts.TemporalityTagKey = DefaultMetricTemporality
	}
	if opts.ResourceStringOptions.DropKeys == nil {
		opts.ResourceStringOptions.DropKeys = droppedMetricAttributeKeys
	}
	if opts.ScopeStringOptions.DropKeys == nil {
		opts.ScopeStringOptions.DropKeys = droppedMetricAttributeKeys
	}
	if opts.PointStringOptions.DropKeys == nil {
		opts.PointStringOptions.DropKeys = droppedMetricAttributeKeys
	}
	if opts.FieldName == nil {
		opts.FieldName = func(metric *metrics.Metric, _ map[string]string) string {
			return metric.GetName()
		}
	}
	if opts.KVOptions == nil {
		opts.KVOptions = func(metric *metrics.Metric) []point.KVOption {
			return []point.KVOption{
				point.WithKVDesc(metric.GetDescription()),
				point.WithKVUnit(metric.GetUnit()),
			}
		}
	}
	if opts.PointOptions == nil {
		opts.PointOptions = func(ts time.Time) []point.Option {
			opts := point.DefaultMetricOptions()
			return append(opts, point.WithTime(ts))
		}
	}
	if opts.ExponentialHistogramAvg == nil {
		opts.ExponentialHistogramAvg = func(sum float64, count uint64) any {
			return fmt.Sprintf("%.3f", sum/float64(count))
		}
	}

	return opts
}

func ParseResourceMetricsV2(resmcs []*metrics.ResourceMetrics, opts MetricsParserOptions) []*point.Point {
	opts = defaultMetricsParserOptions(opts)

	pts := make([]*point.Point, 0)
	for _, resmc := range resmcs {
		if resmc.GetResource() == nil {
			continue
		}

		resourceTags := AttributesToStringMap(resmc.GetResource().GetAttributes(), opts.ResourceStringOptions)
		if opts.CollectorSourceIP != "" {
			resourceTags[opts.CollectorSourceIPTag] = opts.CollectorSourceIP
		}

		for _, scopeMetrics := range resmc.GetScopeMetrics() {
			scopeTags := AttributesToStringMap(scopeMetrics.GetScope().GetAttributes(), opts.ScopeStringOptions)
			if scope := scopeMetrics.GetScope(); scope != nil && scope.GetName() != "" && opts.ScopeNameKey != "" {
				scopeTags[opts.ScopeNameKey] = scope.GetName()
			}

			for _, metric := range scopeMetrics.GetMetrics() {
				fieldOpts := opts.KVOptions(metric)

				if gauge := metric.GetGauge(); gauge != nil {
					for _, dataPoint := range gauge.GetDataPoints() {
						pointTags := AttributesToStringMap(dataPoint.GetAttributes(), opts.PointStringOptions)
						pts = append(pts, buildNumberMetricPoint(metric, dataPoint, resourceTags, scopeTags, pointTags, opts, fieldOpts))
					}
				}

				if sum := metric.GetSum(); sum != nil {
					for _, dataPoint := range sum.GetDataPoints() {
						pointTags := AttributesToStringMap(dataPoint.GetAttributes(), opts.PointStringOptions)
						pt := buildNumberMetricPoint(metric, dataPoint, resourceTags, scopeTags, pointTags, opts, fieldOpts)
						if opts.TemporalityTagKey != "" {
							pt.SetTag(opts.TemporalityTagKey, aggregationTemporalityName(sum.GetAggregationTemporality()))
						}
						pts = append(pts, pt)
					}
				}

				if summary := metric.GetSummary(); summary != nil {
					for _, dataPoint := range summary.GetDataPoints() {
						pointTags := AttributesToStringMap(dataPoint.GetAttributes(), opts.PointStringOptions)
						kvs := MergeStringMapsAsTags(resourceTags, scopeTags, pointTags).
							AddTag(MetricTagUnit, metric.GetUnit())
						name := opts.FieldName(metric, pointTags)
						measurement := metricMeasurement(metric, pointTags, opts)
						pts = append(pts, SummaryDataPointToPoint(measurement, name, kvs, dataPoint, fieldOpts))
					}
				}

				if histogram := metric.GetHistogram(); histogram != nil {
					for _, dataPoint := range histogram.GetDataPoints() {
						pointTags := AttributesToStringMap(dataPoint.GetAttributes(), opts.PointStringOptions)
						pts = append(pts, buildHistogramPoints(metric, dataPoint, resourceTags, scopeTags, pointTags, opts, fieldOpts)...)
					}
				}

				if histogram := metric.GetExponentialHistogram(); histogram != nil {
					for _, dataPoint := range histogram.GetDataPoints() {
						pointTags := AttributesToStringMap(dataPoint.GetAttributes(), opts.PointStringOptions)
						pts = append(pts, buildExponentialHistogramPoint(metric, dataPoint, resourceTags, scopeTags, pointTags, opts, fieldOpts))
					}
				}
			}
		}
	}

	return pts
}

func buildNumberMetricPoint(metric *metrics.Metric, dataPoint *metrics.NumberDataPoint,
	resourceTags, scopeTags, pointTags map[string]string,
	opts MetricsParserOptions, fieldOpts []point.KVOption,
) *point.Point {
	name := opts.FieldName(metric, pointTags)
	measurement := metricMeasurement(metric, pointTags, opts)
	kvs := MergeStringMapsAsTags(resourceTags, scopeTags, pointTags).
		AddTag(MetricTagUnit, metric.GetUnit())
	return NumberDataPointToPoint(measurement, name, kvs, dataPoint, fieldOpts)
}

func buildHistogramPoints(metric *metrics.Metric, dataPoint *metrics.HistogramDataPoint,
	resourceTags, scopeTags, pointTags map[string]string,
	opts MetricsParserOptions, fieldOpts []point.KVOption,
) []*point.Point {
	name := opts.FieldName(metric, pointTags)
	measurement := metricMeasurement(metric, pointTags, opts)
	kvs := MergeStringMapsAsTags(resourceTags, scopeTags, pointTags).
		Add(name+MetricSuffixMin, dataPoint.GetMin(), fieldOpts...).
		Add(name+MetricSuffixMax, dataPoint.GetMax(), fieldOpts...).
		Add(name+MetricSuffixCount, dataPoint.GetCount(), fieldOpts...).
		Add(name+MetricSuffixSum, dataPoint.GetSum(), fieldOpts...).
		AddTag(MetricTagUnit, metric.GetUnit())

	ts := time.Unix(0, int64(dataPoint.GetTimeUnixNano()))
	pts := []*point.Point{
		point.NewPoint(measurement, kvs, opts.PointOptions(ts)...),
	}

	if len(dataPoint.GetBucketCounts()) <= 1 || len(dataPoint.GetExplicitBounds()) == 0 {
		return pts
	}

	bucketSum := uint64(0)
	for idx, bucket := range dataPoint.GetBucketCounts() {
		bucketSum += bucket

		bucketKVs := MergeStringMapsAsTags(resourceTags, scopeTags, pointTags).
			Add(name+MetricSuffixBucket, bucketSum, fieldOpts...).
			AddTag(MetricTagUnit, metric.GetUnit())

		if len(dataPoint.GetExplicitBounds()) > idx {
			bucketKVs = bucketKVs.AddTag(MetricTagLE, strconv.FormatFloat(dataPoint.GetExplicitBounds()[idx], 'f', -1, 64))
		} else {
			bucketKVs = bucketKVs.AddTag(MetricTagLE, MetricSuffixInf)
		}

		pts = append(pts, point.NewPoint(measurement, bucketKVs, opts.PointOptions(ts)...))
	}

	return pts
}

func buildExponentialHistogramPoint(metric *metrics.Metric, dataPoint *metrics.ExponentialHistogramDataPoint,
	resourceTags, scopeTags, pointTags map[string]string,
	opts MetricsParserOptions, fieldOpts []point.KVOption,
) *point.Point {
	name := opts.FieldName(metric, pointTags)
	measurement := metricMeasurement(metric, pointTags, opts)
	kvs := MergeStringMapsAsTags(resourceTags, scopeTags, pointTags).
		Add(name+MetricSuffixMin, dataPoint.GetMin(), fieldOpts...).
		Add(name+MetricSuffixMax, dataPoint.GetMax(), fieldOpts...).
		Add(name+MetricSuffixCount, dataPoint.GetCount(), fieldOpts...).
		Add(name+MetricSuffixSum, dataPoint.GetSum(), fieldOpts...).
		AddTag(MetricTagUnit, metric.GetUnit())

	if dataPoint.GetCount() > 0 {
		kvs = kvs.Add(name+MetricSuffixAvg, opts.ExponentialHistogramAvg(dataPoint.GetSum(), dataPoint.GetCount()), fieldOpts...)
	}

	ts := time.Unix(0, int64(dataPoint.GetTimeUnixNano()))
	return point.NewPoint(measurement, kvs, opts.PointOptions(ts)...)
}

func metricMeasurement(metric *metrics.Metric, pointTags map[string]string, opts MetricsParserOptions) string {
	if opts.MeasurementFromPointAttr == nil {
		return opts.Measurement
	}
	if measurement := opts.MeasurementFromPointAttr(metric, pointTags); measurement != "" {
		return measurement
	}
	return opts.Measurement
}

func aggregationTemporalityName(temporality metrics.AggregationTemporality) string {
	switch temporality { //nolint
	case metrics.AggregationTemporality_AGGREGATION_TEMPORALITY_DELTA:
		return "delta"
	case metrics.AggregationTemporality_AGGREGATION_TEMPORALITY_CUMULATIVE:
		return "cumulative"
	default:
		return "unspecified"
	}
}

func DefaultMetricStringMapOptions() StringMapOptions {
	return StringMapOptions{
		MaxValueLen: 1024 * 32,
		DropKeys:    droppedMetricAttributeKeys,
	}
}

func MetricNameFromAttribute(pointAttrs map[string]string, fallback string) string {
	if name := pointAttrs["metric_name"]; name != "" {
		return name
	}
	return fallback
}

func MetricMeasurementFromAttribute(pointAttrs map[string]string, fallback string) string {
	if measurement := pointAttrs["namespace"]; measurement != "" {
		return measurement
	}
	return fallback
}

func AttributesToTagMap(attrs []*common.KeyValue, opts StringMapOptions) map[string]string {
	return AttributesToStringMap(attrs, opts)
}
