// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package tdengine

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/GuanceCloud/cliutils/point"

	"github.com/GuanceCloud/cliutils"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
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

	opts := append(point.DefaultMetricOptions(), point.WithTime(time1))

	tests := []struct {
		name             string
		args             args
		wantMeasurements []*point.Point
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
						{"liSi", 20, int64(20), false, 12.14, 15.16, "2022-06-20 13:52:09"},
						{"wangWu", 40, int64(40), true, 12.14, 15.16, "2022-06-20 13:52:09"},
						{"zhaoLiu", 50, int64(50), false, 12.14, 15.16, "2022-06-20 13:52:09"},
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

			wantMeasurements: []*point.Point{
				point.NewPointV2("metric_test_name",
					append(point.NewTags(map[string]string{
						"column_bool":   "true",
						"column_string": "zhangSan",
						"unit":          "s",
					}), point.NewKVs(map[string]interface{}{
						"column_int":     10,
						"column_int64":   int64(10),
						"column_float":   12.14,
						"column_float64": 15.16,
					})...),
					opts...),

				point.NewPointV2(
					"metric_test_name",
					append(point.NewTags(map[string]string{
						"column_bool":   "false",
						"column_string": "liSi",
						"unit":          "s",
					}), point.NewKVs(map[string]interface{}{
						"column_int":     20,
						"column_int64":   int64(20),
						"column_float":   12.14,
						"column_float64": 15.16,
					})...),
					opts...),

				point.NewPointV2(
					"metric_test_name",
					append(point.NewTags(map[string]string{
						"column_bool":   "true",
						"column_string": "wangWu",
						"unit":          "s",
					}), point.NewKVs(map[string]interface{}{
						"column_int":     40,
						"column_int64":   int64(40),
						"column_float":   12.14,
						"column_float64": 15.16,
					})...),
					opts...),

				point.NewPointV2(
					"metric_test_name",
					append(point.NewTags(map[string]string{
						"column_bool":   "false",
						"column_string": "zhaoLiu",
						"unit":          "s",
					}), point.NewKVs(map[string]interface{}{
						"column_int":     50,
						"column_int64":   int64(50),
						"column_float":   12.14,
						"column_float64": 15.16,
					})...),
					opts...),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ipt := &Input{
				semStop: cliutils.NewSem(),
				feeder:  dkio.DefaultFeeder(),
				Tagger:  datakit.DefaultGlobalTagger(),
			}
			gotMeasurements := makeMeasurements(tt.args.subMetricName, tt.args.res, tt.args.sql, ipt)
			for i := 0; i < len(gotMeasurements); i++ {
				assert.Equal(t, gotMeasurements[i].LineProto(), tt.wantMeasurements[i].LineProto())
			}
		})
	}
}

func Test_query(t *testing.T) {
	tdURL := os.Getenv("TD_ADAPTER_URL")
	user := os.Getenv("TD_USER")
	pw := os.Getenv("TD_PW")
	if tdURL == "" || user == "" || pw == "" {
		t.Log("no env for TD")
		return
	}
	t.Logf("TD_ADAPTER_URL=%s  TD_USER=%s  TD_PW=%s", tdURL, user, pw)
	base := UserToBase64(user, pw)
	type args struct {
		url       string
		basicAuth string
		token     string
		reqBody   []byte
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name: "show_creat_table_sql", args: args{
				url:       tdURL,
				basicAuth: base,
				token:     "",
				reqBody:   []byte("show CREATE TABLE log.dnodes_info"),
			}, want: []byte{}, wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := query(tt.args.url, tt.args.basicAuth, tt.args.token, tt.args.reqBody)
			if (err != nil) != tt.wantErr {
				t.Errorf("query() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			var res restResult
			if err := json.Unmarshal(got, &res); err != nil {
				l.Error("parse json error: ", err)
				return
			}
			t.Logf("query() got = %s, ", res.Data[0])
		})
	}
}
