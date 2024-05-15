// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package opentelemetry testing
package opentelemetry

import (
	"regexp"
	"testing"

	v1 "github.com/GuanceCloud/tracing-protos/opentelemetry-gen-go/common/v1"
	metrics "github.com/GuanceCloud/tracing-protos/opentelemetry-gen-go/metrics/v1"
	resource "github.com/GuanceCloud/tracing-protos/opentelemetry-gen-go/resource/v1"
	"github.com/stretchr/testify/assert"
)

func Test_parseResourceMetrics(t *testing.T) {
	type args struct {
		resmcs []*metrics.ResourceMetrics
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "string_value_queueSize",
			args: args{resmcs: []*metrics.ResourceMetrics{
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
					},
				},
			}},
		},
	}
	extractAtrributes = extractAttrsWrapper([]*regexp.Regexp{})
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// assert.Equalf(t, tt.want, parseResourceMetrics(tt.args.resmcs), "parseResourceMetrics(%v)", tt.args.resmcs)
			want := parseResourceMetrics(tt.args.resmcs)
			if len(want) > 0 {
				otelM := want[0]
				assert.Equal(t, otelM.resource, "io.opentelemetry.sdk.trace")
				assert.Equal(t, otelM.name, "processedSpans")
				assert.Equal(t, len(otelM.tags), 0)

				pts := otelM.getPoints()
				for _, pt := range pts {
					t.Log(pt.LPPoint())
					/*
						otel-service,agent_version=1.30,description=The\ number\ of\ spans\ processed\ by\ the\ BatchSpanProcessor.\ [dropped\=true\ if\ they\ were\ dropped\ due\ to\ high\ throughput],
						host_name=myClientHost,instrumentation_name=processedSpans,spanProcessorType=BatchSpanProcessor processedSpans=12
						1714128415609886964 <nil>
					*/
				}
			}
		})
	}
}
