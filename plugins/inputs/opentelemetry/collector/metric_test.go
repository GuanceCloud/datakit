package collector

import (
	"testing"

	metricpb "go.opentelemetry.io/proto/otlp/metrics/v1"
)

func Test_getData(t *testing.T) {
	type args struct {
		metric *metricpb.Metric
	}
	tests := []struct {
		name string
		args args
		want *date
	}{
		{
			name: "case_Gauge_int",
			args: args{metric: &metricpb.Metric{
				Name:        "metric-int",
				Description: "Metric_Gauge",
				Data: &metricpb.Metric_Gauge{
					Gauge: &metricpb.Gauge{
						DataPoints: []*metricpb.NumberDataPoint{
							{Value: &metricpb.NumberDataPoint_AsInt{AsInt: 1}},
						},
					},
				},
			}},
			want: &date{
				typeName:  "int",
				startTime: 0,
				unitTime:  0,
				val:       1,
				tags:      &dkTags{tags: make(map[string]string), replaceTags: make(map[string]string)},
			},
		},
		{
			name: "case_Gauge_double",
			args: args{metric: &metricpb.Metric{
				Name:        "metric-double",
				Description: "Metric_Gauge",
				Data: &metricpb.Metric_Gauge{
					Gauge: &metricpb.Gauge{
						DataPoints: []*metricpb.NumberDataPoint{
							{Value: &metricpb.NumberDataPoint_AsDouble{AsDouble: 1.1}},
						},
					},
				},
			}},
			want: &date{
				typeName:  "int",
				startTime: 0,
				unitTime:  0,
				val:       1.1,
				tags:      &dkTags{tags: make(map[string]string), replaceTags: make(map[string]string)},
			},
		},
	}
	storage := NewSpansStorage()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := storage.getData(tt.args.metric)
			for _, gotdate := range got {
				if gotdate.typeName == "int" {
					val, ok := gotdate.val.(int64)
					if !ok {
						t.Error("want type is int64")
					}
					wantVal := tt.want.val.(int)
					if val != int64(wantVal) {
						t.Errorf("got val =%d want val=%d", val, wantVal)
					}
				}
				if gotdate.typeName == "double" {
					val, ok := gotdate.val.(float64)
					if !ok {
						t.Error("want type is int64")
					}
					wantVal := tt.want.val.(float64)
					if val != wantVal {
						t.Errorf("got val =%f want val=%f", val, wantVal)
					}
				}
			}
		})
	}
}
