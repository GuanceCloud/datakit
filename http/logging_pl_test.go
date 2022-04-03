package http

import (
	"testing"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline"
)

func TestLogDataCallback(t *testing.T) {
	task := &logTask{}
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

	task.point = []*io.Point{pt}

	r := []*pipeline.Result{pipeline.NewResult()}
	if _, err := task.callback(r); err != nil {
		t.Error(err)
	} else {
		t.Logf("result: %v", r)
	}
}
