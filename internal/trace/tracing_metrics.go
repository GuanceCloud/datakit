// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package trace

import (
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	dto "github.com/prometheus/client_model/go"

	"github.com/prometheus/client_golang/prometheus"
)

const TracingMetricName = "tracing_metrics"

var DefaultLabelNames = []string{
	TagService,
	TagSource,
	TagOperation,
	TagEnv,
	TagSpanStatus,
	TagVersion,
	FieldResource,
	TagHost,
	TagHttpStatusCode,
	TagHttpStatusClass,
	TagRpcGrpcStatusCode,
	TagPodName,
	TagPodNamespace,
	TagProject,
	TagRemoteIP,
}

func sliceContain(labels []string, label string) bool {
	for _, v := range labels {
		if v == label {
			return true
		}
	}
	return false
}

func AddLabels(labels []string, tags []string) []string {
	newLabels := make([]string, 0)
	for _, label := range labels { //nolint:gosimple
		newLabels = append(newLabels, label)
	}
	for _, v := range tags {
		v = strings.ReplaceAll(v, ".", "_")
		if !sliceContain(newLabels, v) {
			newLabels = append(newLabels, v)
		}
	}
	return newLabels
}

func DelLabels(labels []string, tags []string) []string {
	for _, v := range tags {
		for i, label := range labels {
			if v == label {
				labels = append(labels[:i], labels[i+1:]...)
				break
			}
		}
	}
	return labels
}

// NewTraceHits 创建一个新的用于追踪命中次数的计数器向量 (每一个 span).
func NewTraceHits(labelsNames []string) *prometheus.CounterVec {
	// 使用prometheus库创建一个新的计数器向量
	// 计数器名称为"hits"
	// 帮助文本说明此指标用于任何APM服务
	return prometheus.NewCounterVec(
		prometheus.CounterOpts{ //nolint:promlinter
			Name: "hits",
			Help: "count of spans.", // 指标说明
		},
		labelsNames, // 应用于计数器的标签列表
	)
}

// NewTraceHitsByHTTPStatus 创建一个新的CounterVec用于按HTTP状态码追踪命中次数
// 仅仅针对 HTTP 服务的 span.
func NewTraceHitsByHTTPStatus(labelsNames []string) *prometheus.CounterVec {
	// 创建并返回一个新的CounterVec，用于追踪按HTTP状态码分类的命中次数
	// 计数器名称为"hits_by_http_status"
	// 帮助文本说明此指标适用于任何APM服务
	return prometheus.NewCounterVec(
		prometheus.CounterOpts{ //nolint:promlinter
			Name: "hits_by_http_status",                                                          // 计数器的名称
			Help: "Represent the count of hits for a given span break down by HTTP status code.", // 计数器的帮助描述
		},
		labelsNames, // 应用于计数器的标签名称列表
	)
}

// NewTraceLatency 创建一个用于追踪延迟的直方图指标向量
// 每一个 span 都会记录一个延迟时间。
func NewTraceLatency(labelsNames []string) *prometheus.HistogramVec {
	return prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "latency", // 指标名称
			Help: "Represent the latency distribution for all services, resources, and versions across different " +
				"environments and additional primary tags. Recommended for all latency measurement use cases.",
			Buckets: []float64{ // 直方图的桶，用于记录不同延迟范围的样本数
				100, // 100us
				500,
				1000, // 1ms 单位是微妙
				5000,
				10000,
				20000,
				30000,
				50000,
				100000, // 100ms
				500000,
				1000000, // 1s
				5000000,
				30000000,
				60000000,     // 1min
				60000000 * 5, // 5min
			},
		}, labelsNames)
}

func NewTraceErrors(labelsNames []string) *prometheus.CounterVec {
	return prometheus.NewCounterVec(
		prometheus.CounterOpts{ //nolint:promlinter
			Name: "errors",
			Help: "Represent the count of errors for a given span.",
		},
		labelsNames,
	)
}

func NewTraceErrorsByHTTPStatus(labelsNames []string) *prometheus.CounterVec {
	return prometheus.NewCounterVec(
		prometheus.CounterOpts{ //nolint:promlinter
			Name: "errors_by_http_status",
		},
		labelsNames,
	)
}

// NewTraceApdex  指标名称为 apdex.
// 计算方式：(请求时间阈值数+请求时间大于阈值的数/2)/请求总数
// 得到一个 0~1 的值.
// 阈值数: 500ms
// >0.9 满意
// >0.75 容忍.
func NewTraceApdex() *prometheus.HistogramVec {
	return prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "apdex",
			Buckets: []float64{
				2000000, // 2s 满意
				8000000, // 2-8s 容忍
				// 不用设置桶上限
			},
			Help: "Measures the Apdex score for each web service",
		},
		[]string{TagService, TagEnv, TagVersion, FieldResource, TagSource, TagRemoteIP},
	)
}

func GatherPoints(register *prometheus.Registry, inputTags map[string]string) []*point.Point {
	icm, err := register.Gather()
	if err != nil {
		log.Warnf("p8sToPoint gather metrics error:%v", err)
		return nil
	}
	// All gathered data should have the same timestamp, we enforce it.
	now := time.Now()
	var pts []*point.Point
	for _, mf := range icm {
		fieldName := mf.GetName()
		for _, m := range mf.Metric {
			switch *mf.Type {
			case dto.MetricType_COUNTER:
				var kvs point.KVs
				if m.GetCounter() != nil {
					kvs = point.NewTags(inputTags)
					for _, label := range m.GetLabel() {
						kvs = kvs.AddTag(label.GetName(), label.GetValue())
					}
					kvs = kvs.Add(fieldName, m.GetCounter().GetValue())
					pts = append(pts, point.NewPoint(TracingMetricName, kvs, point.WithTime(now)))
				}
			case dto.MetricType_SUMMARY:
				continue // TODO
			case dto.MetricType_GAUGE:
				continue // TODO
			case dto.MetricType_HISTOGRAM:
				if fieldName == "apdex" {
					if pt := apdexToPoint(m.GetHistogram(), now, fieldName); pt != nil {
						for k, v := range inputTags {
							pt.AddTag(k, v)
						}
						for _, label := range m.GetLabel() {
							pt.AddTag(label.GetName(), label.GetValue())
						}
						pts = append(pts, pt)
					}
					continue
				}
				if m.GetHistogram() != nil {
					pts1 := histogramToPoint(m.GetHistogram(), now, fieldName)
					for _, pt := range pts1 {
						for k, v := range inputTags {
							pt.AddTag(k, v)
						}
						for _, label := range m.GetLabel() {
							pt.AddTag(label.GetName(), label.GetValue())
						}
					}
					if len(pts1) > 0 {
						pts = append(pts, pts1...)
					}
				}
			case dto.MetricType_UNTYPED:
				continue // TODO
			case dto.MetricType_GAUGE_HISTOGRAM:
				continue // TODO
			case dto.MetricType_INFO:
				continue
			default:
				// pass
			}
		}
	}

	return pts
}

func apdexToPoint(his *dto.Histogram, t time.Time, hisName string) *point.Point {
	opts := point.DefaultMetricOptions()
	opts = append(opts, point.WithTime(t))

	// 先统计count 和 sum
	var kvs point.KVs
	count := his.GetSampleCount() // 总次数，作为分母。
	if count == 0 {
		return nil
	}
	satisfy, tolerate := uint64(0), uint64(0)
	for _, bucket := range his.Bucket {
		if bucket.GetUpperBound() == 2000000 { // 2s
			satisfy = bucket.GetCumulativeCount()
		}
		if bucket.GetUpperBound() == 8000000 { // 8s
			tolerate = bucket.GetCumulativeCount()
		}
	}
	tolerate -= satisfy // tolerate 是 2-8s 的次数，需要减去 2s 的次数.
	apdex := (float64(satisfy + tolerate/2)) / float64(count)

	apdex = math.Round(apdex*100) / 100

	kvs = kvs.Add(hisName, apdex)
	return point.NewPoint(TracingMetricName, kvs, opts...)
}

func histogramToPoint(his *dto.Histogram, t time.Time, hisName string) []*point.Point {
	opts := point.DefaultMetricOptions()
	opts = append(opts, point.WithTime(t))
	pts := make([]*point.Point, 0)
	// 先统计count 和 sum
	var kvs point.KVs
	count := his.GetSampleCount()
	sum := his.GetSampleSum()
	kvs = kvs.Add(hisName+"_count", count).Add(hisName+"_sum", sum)

	pts = append(pts, point.NewPoint(TracingMetricName, kvs, opts...))

	// 统计 bucket
	for i, bucket := range his.Bucket {
		var bkvs point.KVs
		bkvs = bkvs.Add(hisName+"_bucket", bucket.GetCumulativeCount()).
			AddTag("le", strconv.FormatFloat(bucket.GetUpperBound(), 'f', -1, 64))
		pts = append(pts, point.NewPoint(TracingMetricName, bkvs, opts...))

		if i == len(his.Bucket)-1 {
			// 最后一个 bucket 如果 cumulative_count 不等于 count，则说明有 +inf
			if bucket.GetCumulativeCount() != count {
				var infKvs point.KVs
				infKvs = infKvs.Add(hisName+"_bucket", count).
					AddTag("le", "+Inf")
				pts = append(pts, point.NewPoint(TracingMetricName, infKvs, opts...))
			}
		}
	}
	return pts
}
