package tdengine

import (
	"reflect"
	"testing"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

func Test_makeMeasurements(t *testing.T) {
	type args struct {
		subMetricName string
		res           restResult
		sql           selectSQL
	}
	time1, err := time.Parse("2006-01-02 15:04:05", "2022-06-20 13:52:09")
	if err != nil {
		t.Logf("err=%v", err)
		time1 = time.Now()
	}
	tests := []struct {
		name             string
		args             args
		wantMeasurements []inputs.Measurement
	}{
		{
			name: "case_all_type",
			args: args{
				subMetricName: "metric_test_name",
				res: restResult{
					Status: "200",
					Head:   make([]string, 0),
					ColumnMeta: [][]interface{}{
						{"column_string", 8, 32},
						{"column_int", 4, 1},
						{"column_int64", 5, 4},
						{"column_bool", 1, 16},
						{"column_float", 6, 4},
						{"column_float64", 7, 4},
						{"ts", 9, 32},
					},
					Data: [][]interface{}{
						{"zhangSan", 10, int64(10), true, 12.14, 15.16, "2022-06-20 13:52:09"},
						{"liSi", 20, int64(20), false, 12.14, 15.16, "2022-06-20 13:52:19"},
						{"wangWu", 40, int64(40), true, 12.14, 15.16, "2022-06-20 13:52:29"},
						{"zhaoLiu", 50, int64(50), false, 12.14, 15.16, "2022-06-20 13:52:39"},
					},
					Rows: 4,
				},
				sql: selectSQL{
					desc:      "测试-查询名单",
					title:     "",
					sql:       "",
					unit:      "s",
					fields:    []string{"column_int", "column_int64", "column_float", "column_float64"},
					tags:      []string{"column_bool", "column_string"},
					plugInFun: nil,
				},
			},
			wantMeasurements: []inputs.Measurement{
				&Measurement{
					name: "metric_test_name",
					tags: map[string]string{
						"column_bool":   "true",
						"column_string": "zhangSan",
						"unit":          "s",
					},
					fields: map[string]interface{}{
						"column_int":     10,
						"column_int64":   int64(10),
						"column_float":   12.14,
						"column_float64": 15.16,
					},
					ts: time1,
				},

				&Measurement{
					name: "metric_test_name",
					tags: map[string]string{
						"column_bool":   "false",
						"column_string": "liSi",
						"unit":          "s",
					},
					fields: map[string]interface{}{
						"column_int":     20,
						"column_int64":   int64(20),
						"column_float":   12.14,
						"column_float64": 15.16,
					},
					ts: time1.Add(time.Second * 10),
				},

				&Measurement{
					name: "metric_test_name",
					tags: map[string]string{
						"column_bool":   "true",
						"column_string": "wangWu",
						"unit":          "s",
					},
					fields: map[string]interface{}{
						"column_int":     40,
						"column_int64":   int64(40),
						"column_float":   12.14,
						"column_float64": 15.16,
					},
					ts: time1.Add(time.Second * 20),
				},

				&Measurement{
					name: "metric_test_name",
					tags: map[string]string{
						"column_bool":   "false",
						"column_string": "zhaoLiu",
						"unit":          "s",
					},
					fields: map[string]interface{}{
						"column_int":     50,
						"column_int64":   int64(50),
						"column_float":   12.14,
						"column_float64": 15.16,
					},
					ts: time1.Add(time.Second * 30),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotMeasurements := makeMeasurements(tt.args.subMetricName, tt.args.res, tt.args.sql)
			for i := 0; i < len(gotMeasurements); i++ {
				if !reflect.DeepEqual(gotMeasurements[i], tt.wantMeasurements[i]) {
					t.Errorf("makeMeasurements()[%d] = %v, want %v", i, gotMeasurements[i], tt.wantMeasurements[i])
				}
			}
		})
	}
}
