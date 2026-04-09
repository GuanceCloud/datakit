// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package opentelemetry

import (
	"math"
	T "testing"

	"github.com/GuanceCloud/cliutils/point"
	v1 "github.com/GuanceCloud/tracing-protos/opentelemetry-gen-go/common/v1"
	metrics "github.com/GuanceCloud/tracing-protos/opentelemetry-gen-go/metrics/v1"
	resource "github.com/GuanceCloud/tracing-protos/opentelemetry-gen-go/resource/v1"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	dkMetrics "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/metrics"
)

func Test_parseResourceMetricsV2(t *T.T) {
	msource := []*metrics.ResourceMetrics{
		{
			Resource: &resource.Resource{
				Attributes: []*v1.KeyValue{
					{
						Key: "host.name",
						Value: &v1.AnyValue{
							Value: &v1.AnyValue_StringValue{
								StringValue: "myClientHost",
							},
						},
					},
					{
						Key: "agent.version",
						Value: &v1.AnyValue{
							Value: &v1.AnyValue_StringValue{
								StringValue: "1.30",
							},
						},
					},
				},
			},
			ScopeMetrics: []*metrics.ScopeMetrics{
				{
					Scope: &v1.InstrumentationScope{
						Name:                   "io.opentelemetry.sdk.trace",
						Version:                "1.30.0",
						DroppedAttributesCount: 0,
					},
					Metrics: []*metrics.Metric{
						{
							Name:        "processedSpans",
							Description: "The number of spans processed by the BatchSpanProcessor. [dropped=true if they were dropped due to high throughput]",
							Unit:        "1",
							Data: &metrics.Metric_Sum{
								Sum: &metrics.Sum{
									DataPoints: []*metrics.NumberDataPoint{
										{
											Attributes: []*v1.KeyValue{
												{
													Key: "spanProcessorType",
													Value: &v1.AnyValue{
														Value: &v1.AnyValue_StringValue{
															StringValue: "BatchSpanProcessor",
														},
													},
												},
												{
													Key: "dropped",
													Value: &v1.AnyValue{
														Value: &v1.AnyValue_BoolValue{
															BoolValue: false,
														},
													},
												},
											},
											StartTimeUnixNano: 0,
											TimeUnixNano:      0,
											Value:             &metrics.NumberDataPoint_AsDouble{AsDouble: 12},
											Exemplars:         nil,
											Flags:             0,
										},
									},
									AggregationTemporality: 0,
									IsMonotonic:            false,
								},
							},
						},
					},
					SchemaUrl: "1.30",
				},
				{
					Scope: &v1.InstrumentationScope{Name: "io.opentelemetry.tomcat-7.0"},
					Metrics: []*metrics.Metric{
						{
							Name:        "http.server.duration",
							Description: "The duration of the inbound HTTP request",
							Unit:        "ms",
							Data: &metrics.Metric_Histogram{
								Histogram: &metrics.Histogram{
									DataPoints: []*metrics.HistogramDataPoint{
										{
											Attributes: []*v1.KeyValue{
												{
													Key: "http.method",
													Value: &v1.AnyValue{
														Value: &v1.AnyValue_StringValue{
															StringValue: "Get",
														},
													},
												},
												{
													Key: "http.route",
													Value: &v1.AnyValue{
														Value: &v1.AnyValue_StringValue{
															StringValue: "/tmall/",
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	ipt := defaultInput()
	ipt.feeder = &feeder{t: t}
	ipt.parseResourceMetricsV2(msource, "")
}

func getPtr(f float64) *float64 {
	return &f
}

type feeder struct {
	t *T.T
}

func (f *feeder) Feed(category point.Category, pts []*point.Point, opts ...dkio.FeedOption) error {
	f.t.Logf("category = %s", category)
	if len(pts) == 0 {
		f.t.Errorf("parse otel metric to point.len==0")
	} else {
		for _, pt := range pts {
			f.t.Logf("%s ", pt.Pretty())
		}
	}
	return nil
}

func (f *feeder) FeedLastError(err string, opts ...dkMetrics.LastErrorOption) {
	f.t.Logf("not implement")
}

type expBucket struct {
	index  int
	lb, ub float64
}

func expHistoBuckets(vmin, vmax float64, scale int32) []expBucket {
	base := math.Pow(2, math.Pow(2, float64(-scale)))
	startIdx, endIdx := math.Floor(math.Log2(vmin)/math.Log2(base)), math.Floor(math.Log2(vmax)/math.Log2(base))

	log.Debugf("base: %f, start: %f, end: %f", base, startIdx, endIdx)

	var buckets []expBucket
	for i := startIdx; i <= endIdx; i++ {
		buckets = append(buckets, expBucket{
			index: int(i),
			lb:    math.Pow(base, i),
			ub:    math.Pow(base, i+1),
		})
	}

	return buckets
}

func Test_expHistoBuckets(t *T.T) {
	t.Run(`basic`, func(t *T.T) {
		buckets := expHistoBuckets(0.1, 30*1000.0, 5) // 30ms ~ 30s

		for _, b := range buckets {
			t.Logf("[%d]: [%f, %f), range: %f", b.index, b.lb, b.ub, b.ub-b.lb)
		}
	})
}
