// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package opentelemetry

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	common "github.com/GuanceCloud/tracing-protos/opentelemetry-gen-go/common/v1"
	metrics "github.com/GuanceCloud/tracing-protos/opentelemetry-gen-go/metrics/v1"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
)

func (ipt *Input) parseResourceMetricsV2(resmcs []*metrics.ResourceMetrics) {
	start := time.Now()
	var pts []*point.Point
	for _, resmc := range resmcs {
		if resmc.GetResource() == nil {
			return
		}

		resourceTags := attributesToTag(resmc.Resource.GetAttributes())

		for _, scopeMetrics := range resmc.GetScopeMetrics() {
			var scopeTags map[string]string
			if scopeStats := scopeMetrics.GetScope(); scopeStats != nil {
				scopeTags = attributesToTag(scopeMetrics.GetScope().GetAttributes())
				scopeTags["scope_name"] = scopeMetrics.GetScope().GetName()
			}

			for _, metric := range scopeMetrics.GetMetrics() {
				switch t := metric.Data.(type) {
				case *metrics.Metric_Gauge:
					for _, dataPoint := range t.Gauge.GetDataPoints() {
						ptTags := attributesToTag(dataPoint.GetAttributes())
						kvs := mergeTags(resourceTags, scopeTags, ptTags)
						kvs = kvs.AddTag(unitTag, metric.GetUnit())
						pt := numberDataToPoint(kvs, dataPoint, metric.GetName())
						pts = append(pts, pt)
					}
				case *metrics.Metric_Sum:
					for _, dataPoint := range t.Sum.GetDataPoints() {
						ptTags := attributesToTag(dataPoint.GetAttributes())
						kvs := mergeTags(resourceTags, scopeTags, ptTags)
						kvs = kvs.AddTag(unitTag, metric.GetUnit())
						pt := numberDataToPoint(kvs, dataPoint, metric.GetName())
						pts = append(pts, pt)
					}
				case *metrics.Metric_Summary:
					for _, dataPoint := range t.Summary.GetDataPoints() {
						ptTags := attributesToTag(dataPoint.GetAttributes())
						kvs := mergeTags(resourceTags, scopeTags, ptTags)
						kvs = kvs.AddTag(unitTag, metric.GetUnit())
						pt := summaryToPoint(kvs, dataPoint, metric.GetName())
						pts = append(pts, pt)
					}
				case *metrics.Metric_Histogram:
					for _, his := range t.Histogram.GetDataPoints() {
						hisTags := attributesToTag(his.GetAttributes())
						kvs := mergeTags(resourceTags, scopeTags, hisTags)
						kvs = kvs.AddV2(metric.Name+minSuffix, his.GetMin(), false).
							AddV2(metric.Name+maxSuffix, his.GetMax(), false).
							AddV2(metric.Name+countSuffix, his.GetCount(), false).
							AddV2(metric.Name+sumSuffix, his.GetSum(), false).
							AddTag(unitTag, metric.GetUnit())

						ts := time.Unix(0, int64(his.GetTimeUnixNano()))
						opts := point.DefaultMetricOptions()
						opts = append(opts, point.WithTime(ts))
						pts = append(pts, point.NewPointV2(metricName, kvs, opts...))

						// bucket
						if len(his.GetBucketCounts()) > 1 && len(his.GetExplicitBounds()) > 0 {
							bucketSum := uint64(0)
							for i, bucket := range his.BucketCounts {
								bucketSum += bucket

								if len(his.GetExplicitBounds()) > i {
									bKvs := mergeTags(resourceTags, scopeTags, hisTags)
									bKvs = bKvs.AddV2(metric.Name+bucketSuffix, bucketSum, false).
										AddTag(leTag, strconv.FormatFloat(his.ExplicitBounds[i], 'f', -1, 64)).
										AddTag(unitTag, metric.GetUnit())
									pts = append(pts, point.NewPointV2(metricName, bKvs, opts...))
								} else {
									bKvs := mergeTags(resourceTags, scopeTags, hisTags)
									bKvs = bKvs.AddV2(metric.Name+bucketSuffix, bucketSum, false).
										AddTag(leTag, infSuffix).
										AddTag(unitTag, metric.GetUnit())
									pts = append(pts, point.NewPointV2(metricName, bKvs, opts...))
								}
							}
						}
					}
				case *metrics.Metric_ExponentialHistogram:
					for _, his := range t.ExponentialHistogram.GetDataPoints() {
						hisTags := attributesToTag(his.GetAttributes())
						kvs := mergeTags(resourceTags, scopeTags, hisTags)

						kvs = kvs.Add(metric.Name+minSuffix, his.GetMin(), false, false).
							Add(metric.Name+maxSuffix, his.GetMax(), false, false).
							Add(metric.Name+countSuffix, his.GetCount(), false, false).
							Add(metric.Name+sumSuffix, his.GetSum(), false, false).
							AddTag(unitTag, metric.GetUnit())
						if his.GetCount() > 0 {
							kvs = kvs.Add(metric.Name+avgSuffix,
								fmt.Sprintf("%.3f", his.GetSum()/float64(his.GetCount())), false, false)
						}
						ts := time.Unix(0, int64(his.GetTimeUnixNano()))
						opts := point.DefaultMetricOptions()
						opts = append(opts, point.WithTime(ts))
						pts = append(pts, point.NewPointV2(metricName, kvs, opts...))
					}
				}

				if len(pts) >= 100 {
					if err := ipt.feeder.Feed(point.Metric, pts,
						dkio.WithSource(inputName),
						dkio.WithCollectCost(time.Since(start)),
					); err != nil {
						log.Errorf("feed err=%s", err.Error())
					}
					pts = make([]*point.Point, 0, cap(pts))
				}
			}
		}
	}

	_ = ipt.feeder.Feed(point.Metric, pts, dkio.WithSource(inputName), dkio.WithCollectCost(time.Since(start)))
}

func attributesToTag(src []*common.KeyValue) map[string]string {
	shadowTags := make(map[string]string)
	for _, keyVal := range src {
		key := keyVal.GetKey()
		switch keyVal.GetValue().Value.(type) {
		case *common.AnyValue_BytesValue, *common.AnyValue_StringValue:
			if s := keyVal.Value.GetStringValue(); len(s) > maxLogMetricFiledLen {
				shadowTags[key] = s[:maxLogMetricFiledLen]
			} else {
				shadowTags[key] = s
			}
		case *common.AnyValue_DoubleValue:
			shadowTags[key] = fmt.Sprintf("%.3f", keyVal.Value.GetDoubleValue())
		case *common.AnyValue_IntValue:
			shadowTags[key] = fmt.Sprintf("%d", keyVal.Value.GetIntValue())
		case *common.AnyValue_KvlistValue:
			shadowTags[key] = keyVal.Value.GetKvlistValue().String()
		case *common.AnyValue_ArrayValue:
			shadowTags[key] = keyVal.Value.GetArrayValue().String()
		}
	}
	for _, s := range delMetricKey {
		delete(shadowTags, s)
	}
	log.Infof("shadowtags len=%d", len(shadowTags))
	return shadowTags
}

func mergeTags(resource, scope, pt map[string]string) point.KVs {
	var kv point.KVs
	for _, m := range []map[string]string{resource, scope, pt} {
		for k, v := range m {
			k = strings.ReplaceAll(k, ".", "_")
			kv = kv.AddTag(k, v)
		}
	}
	return kv
}

func mergeTagsToField(resource, scope, pt map[string]string) point.KVs {
	var kv point.KVs
	for _, m := range []map[string]string{resource, scope, pt} {
		for k, v := range m {
			k = strings.ReplaceAll(k, ".", "_")
			kv = kv.AddV2(k, v, false)
		}
	}
	return kv
}

func numberDataToPoint(kvs point.KVs, pt *metrics.NumberDataPoint, name string) *point.Point {
	if v, ok := pt.Value.(*metrics.NumberDataPoint_AsDouble); ok {
		kvs = kvs.Add(name, v.AsDouble, false, false)
	} else if v, ok := pt.Value.(*metrics.NumberDataPoint_AsInt); ok {
		kvs = kvs.Add(name, v.AsInt, false, false)
	}
	ts := time.Unix(0, int64(pt.GetTimeUnixNano()))
	opts := point.DefaultMetricOptions()
	opts = append(opts, point.WithTime(ts))

	return point.NewPointV2(metricName, kvs, opts...)
}

func summaryToPoint(kvs point.KVs, summary *metrics.SummaryDataPoint, name string) *point.Point {
	kvs = kvs.Add(name+countSuffix, summary.GetCount(), false, false).
		Add(name+sumSuffix, summary.GetSum(), false, false)
	ts := time.Unix(0, int64(summary.GetTimeUnixNano()))
	opts := point.DefaultMetricOptions()
	opts = append(opts, point.WithTime(ts))

	return point.NewPointV2(metricName, kvs, opts...)
}
