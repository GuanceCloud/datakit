package collector

import (
	"testing"
	"time"

	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
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
				typeName:  "double",
				startTime: 0,
				unitTime:  0,
				val:       1.1,
				tags:      &dkTags{tags: make(map[string]string), replaceTags: make(map[string]string)},
			},
		},
		{
			name: "case_sum_double",
			args: args{metric: &metricpb.Metric{
				Name:        "metric-double",
				Description: "Metric_Sum",
				Data: &metricpb.Metric_Gauge{
					Gauge: &metricpb.Gauge{
						DataPoints: []*metricpb.NumberDataPoint{
							{Value: &metricpb.NumberDataPoint_AsDouble{AsDouble: 1.1}},
						},
					},
				},
			}},
			want: &date{
				typeName:  "double",
				startTime: 0,
				unitTime:  0,
				val:       1.1,
				tags:      &dkTags{tags: make(map[string]string), replaceTags: make(map[string]string)},
			},
		},
		{
			name: "case_sum_int",
			args: args{metric: &metricpb.Metric{
				Name:        "metric-double",
				Description: "Metric_Gauge",
				Data: &metricpb.Metric_Gauge{
					Gauge: &metricpb.Gauge{
						DataPoints: []*metricpb.NumberDataPoint{
							{Value: &metricpb.NumberDataPoint_AsInt{AsInt: 10}},
						},
					},
				},
			}},
			want: &date{
				typeName:  "int",
				startTime: 0,
				unitTime:  0,
				val:       10,
				tags:      &dkTags{tags: make(map[string]string), replaceTags: make(map[string]string)},
			},
		},
		{
			name: "case_Metric_Histogram",
			args: args{metric: &metricpb.Metric{
				Name:        "metric-double",
				Description: "Metric_Gauge",
				Data: &metricpb.Metric_Histogram{
					Histogram: &metricpb.Histogram{
						DataPoints: []*metricpb.HistogramDataPoint{
							{Sum: 1.1},
						},
					},
				},
			}},
			want: &date{
				typeName:  "histogram",
				startTime: 0,
				unitTime:  0,
				val:       1.1,
				tags:      &dkTags{tags: make(map[string]string), replaceTags: make(map[string]string)},
			},
		},
		{
			name: "case_Metric_Histogram",
			args: args{metric: &metricpb.Metric{
				Name:        "metric-double",
				Description: "Metric_Gauge",
				Data: &metricpb.Metric_ExponentialHistogram{
					ExponentialHistogram: &metricpb.ExponentialHistogram{
						DataPoints: []*metricpb.ExponentialHistogramDataPoint{
							{Sum: 1.1},
						},
					},
				},
			}},
			want: &date{
				typeName:  "ExponentialHistogram",
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
					wantVal, ok := tt.want.val.(int)
					if !ok {
						t.Error("want type int")
					}
					if val != int64(wantVal) {
						t.Errorf("got val =%d want val=%d", val, wantVal)
					}
				}
				if gotdate.typeName == "double" {
					val, ok := gotdate.val.(float64)
					if !ok {
						t.Error("want type is float64")
						return
					}
					wantVal, ok := tt.want.val.(float64)
					if !ok {
						t.Error("want type is float64")
						return
					}
					if val != wantVal {
						t.Errorf("got val =%f want val=%f", val, wantVal)
					}
				}
			}
		})
	}
}

func Test_makePoints(t *testing.T) {
	type args struct {
		orms []*OtelResourceMetric
	}
	pt, err := dkio.MakePoint("service", map[string]string{"tagA": "a"}, map[string]interface{}{"a": 10}, time.Now())
	resourceMetric := &OtelResourceMetric{
		Operation:   "sample-span1",
		Attributes:  map[string]string{"tagA": "a"},
		Service:     "service",
		Resource:    "resource",
		Description: "test for span",
		StartTime:   uint64(time.Now().Unix()),
		UnitTime:    uint64(time.Now().Unix()),
		ValueType:   "int",
		Value:       10,
	}

	if err != nil {
		t.Errorf("makepoint err=%v", err)
		return
	}
	tests := []struct {
		name string
		args args
		want []*dkio.Point
	}{
		{
			name: "case1",
			args: args{orms: []*OtelResourceMetric{resourceMetric}},
			want: []*dkio.Point{pt},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := makePoints(tt.args.orms)
			rpt := got[0]

			tags := rpt.Tags()
			if len(tags) != 3 {
				t.Errorf("tags len not 3")
			}
			if rpt.Name() != resourceMetric.Service {
				t.Errorf("service not equal")
			}
			fields, err := rpt.Fields()
			if err != nil {
				t.Errorf("field is  nil and err=%v", err)
				return
			}
			for key, i := range fields {
				t.Logf("key=%s val=%v", key, i)
			}
			if val, ok := fields[resourceMetric.Operation].(int64); ok {
				if int(val) != resourceMetric.Value {
					t.Errorf("val not equal")
				}
			} else {
				t.Errorf("can find field: Operation")
			}
		})
	}
}
