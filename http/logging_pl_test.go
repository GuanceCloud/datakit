package http

import (
	"testing"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/worker"
)

func TestLogDataCallback(t *testing.T) {
	task := &worker.Task{}
	pt, err := io.MakePoint(
		"abc",
		map[string]string{
			"a": "b",
		},
		map[string]interface{}{
			"d": 1,
			"f": 2.2,
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	data := &logTaskData{
		point: []*io.Point{pt},
	}

	task.Data = data

	r := []*pipeline.Result{pipeline.NewResult()}
	if _, _, err := data.callback(task, r); err != nil {
		t.Error(err)
	} else {
		t.Logf("result: %v", r)
	}
}
