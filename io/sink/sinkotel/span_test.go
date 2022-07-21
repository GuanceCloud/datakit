// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package sinkotel

import (
	"testing"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/point"
)

func Test_pointToTrace(t *testing.T) {
	type args struct {
		pts []*point.Point
	}

	tests := []struct {
		name string
		args args
	}{
		{
			name: "case1",
			args: args{
				[]*point.Point{makePoints(t)},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.args.pts == nil || len(tt.args.pts) == 0 {
				t.Errorf("point len ==0")
				return
			}
			gotRoSpans := pointToTrace(tt.args.pts)
			for _, span := range gotRoSpans {
				t.Logf("%+v \n", span)
				t.Logf("%s \n", span.SpanContext().SpanID().String())
				t.Logf("%s \n", span.Parent().SpanID().String())
			}
		})
	}
}

func makePoints(t *testing.T) *point.Point {
	t.Helper()
	startTime := time.Date(2020, time.December, 8, 19, 15, 0, 0, time.UTC)

	tags := map[string]string{
		TAG_OPERATION:   "span_name",
		TAG_PROJECT:     "project",
		TAG_SERVICE:     "datakit",
		TAG_SPAN_STATUS: STATUS_OK,
	}
	fields := map[string]interface{}{
		FIELD_DURATION: int64(time.Second),
		FIELD_PARENTID: "0",
		FIELD_RESOURCE: "/get",
		FIELD_SPANID:   "b94a9cc92458f08d",
		FIELD_START:    startTime.UnixNano() / int64(time.Millisecond),
		FIELD_TRACEID:  "b94a9cc92458f08d",
	}

	pt, err := point.NewPoint("test_for_trace", tags, fields, nil)
	if err != nil {
		t.Errorf("make point err=%v", err)
		return nil
	}

	return pt
}
