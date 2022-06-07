// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package sinkotel

import (
	"testing"
	"time"

	client "github.com/influxdata/influxdb1-client/v2"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/sink/sinkcommon"
)

func Test_pointToTrace(t *testing.T) {
	type args struct {
		pts []sinkcommon.ISinkPoint
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "case1",
			args: args{
				[]sinkcommon.ISinkPoint{makePoints(t)},
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

func makePoints(t *testing.T) *testPoint {
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
	pt, err := client.NewPoint(
		"test_for_trace",
		tags,
		fields,
		time.Now(),
	)
	if err != nil {
		t.Errorf("make point err=%v", err)
		return nil
	}
	return &testPoint{pt}
}

type testPoint struct {
	*client.Point
}

var _ sinkcommon.ISinkPoint = new(testPoint)

func (p *testPoint) ToPoint() *client.Point {
	return p.Point
}

func (p *testPoint) String() string {
	return ""
}

func (p *testPoint) ToJSON() (*sinkcommon.JSONPoint, error) {
	fields, err := p.Point.Fields()
	if err != nil {
		return nil, err
	}
	return &sinkcommon.JSONPoint{
		Measurement: p.ToPoint().Name(),
		Tags:        p.Point.Tags(),
		Fields:      fields,
		Time:        p.Point.Time(),
	}, nil
}
