// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package pipeline

import (
	"testing"

	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/jit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/plval"
)

func TestUploadJITAggregateEventsUsesPipelineUploadContract(t *testing.T) {
	original, _ := plval.GetManager()
	t.Cleanup(func() { plval.SetManager(original) })
	type call struct {
		category point.Category
		bucket   string
		points   []*point.Point
	}
	var calls []call
	plval.SetManager(plval.NewScriptManager(func(category point.Category, bucket string, value any) error {
		points, ok := value.([]*point.Point)
		if !ok {
			t.Fatalf("upload value type=%T", value)
		}
		calls = append(calls, call{category: category, bucket: bucket, points: points})
		return nil
	}, nil))
	values := []jit.Point{
		{Version: 1, Category: "metric", Measurement: "cpu", TimeUnixNano: 1, Tags: map[string]string{"host": "a"}, Fields: map[string]any{"avg": float64(1.5)}},
		{Version: 1, Category: "metric", Measurement: "cpu", TimeUnixNano: 2, Tags: map[string]string{"host": "b"}, Fields: map[string]any{"avg": float64(2.5)}},
		{Version: 1, Category: "logging", Measurement: "errors", TimeUnixNano: 3, Fields: map[string]any{"count": int64(2)}},
	}
	if err := uploadJITAggregateEvents(values); err != nil {
		t.Fatal(err)
	}
	if len(calls) != 2 {
		t.Fatalf("calls=%#v", calls)
	}
	if calls[0].category != point.Metric || calls[0].bucket != "cpu" || len(calls[0].points) != 2 ||
		calls[0].points[0].Get("avg") != float64(1.5) || calls[0].points[1].Get("host") != "b" {
		t.Fatalf("metric call=%#v", calls[0])
	}
	if calls[1].category != point.Logging || calls[1].bucket != "errors" || len(calls[1].points) != 1 ||
		calls[1].points[0].Get("count") != int64(2) {
		t.Fatalf("logging call=%#v", calls[1])
	}
}
