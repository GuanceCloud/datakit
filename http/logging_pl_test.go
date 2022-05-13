// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

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
