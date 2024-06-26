// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package opentelemetry testing
package opentelemetry

import (
	"testing"

	v1 "github.com/GuanceCloud/tracing-protos/opentelemetry-gen-go/common/v1"
	metrics "github.com/GuanceCloud/tracing-protos/opentelemetry-gen-go/metrics/v1"
	resource "github.com/GuanceCloud/tracing-protos/opentelemetry-gen-go/resource/v1"
)

func Test_parseResourceMetricsV2(t *testing.T) {
	msource := []*metrics.ResourceMetrics{
		{
			Resource: &resource.Resource{
				Attributes: []*v1.KeyValue{
					{
						Key: "host.name", Value: &v1.AnyValue{
							Value: &v1.AnyValue_StringValue{
								StringValue: "myClientHost",
							},
						},
					},
					{
						Key: "agent.version", Value: &v1.AnyValue{
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
													Key: "spanProcessorType", Value: &v1.AnyValue{
														Value: &v1.AnyValue_StringValue{
															StringValue: "BatchSpanProcessor",
														},
													},
												},
												{
													Key: "dropped", Value: &v1.AnyValue{
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
													Key: "http.method", Value: &v1.AnyValue{
														Value: &v1.AnyValue_StringValue{
															StringValue: "Get",
														},
													},
												},
												{
													Key: "http.route", Value: &v1.AnyValue{
														Value: &v1.AnyValue_StringValue{
															StringValue: "/tmall/",
														},
													},
												},
											},
											StartTimeUnixNano: 1,
											TimeUnixNano:      1,
											Count:             68,
											Sum:               getPtr(221.49527399999997),
											BucketCounts: []uint64{
												0,
												2,
												4,
												1,
												16,
												11,
												7,
												27,
												0,
												0,
												0,
												0,
												0,
												0,
												0,
												0,
											},
											ExplicitBounds: []float64{
												0,
												5,
												10,
												25,
												50,
												75,
												100,
												250,
												500,
												750,
												1000,
												2500,
												5000,
												7500,
												10000,
											},
											Exemplars: nil,
											Flags:     0,
											Min:       getPtr(3.455694),
											Max:       getPtr(186.694506),
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

	pts := parseResourceMetricsV2(msource)
	if len(pts) == 0 {
		t.Errorf("parse otel metric to point.len==0")
	} else {
		for _, pt := range pts {
			t.Logf("point = %s ", pt.LineProto())
		}
	}
}

func getPtr(f float64) *float64 {
	return &f
}
