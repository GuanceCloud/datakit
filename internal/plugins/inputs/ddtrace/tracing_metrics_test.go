// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package ddtrace

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/GuanceCloud/cliutils/point"
	itrace "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/trace"
)

func Test_initP8SMetrics(t *testing.T) {
	labels = itrace.DefaultLabelNames
	initP8SMetrics(labels)

	pts := make([]*point.Point, 0)
	for i := 0; i < 10; i++ {
		var kvs point.KVs
		kvs = kvs.AddTag("service", "test").
			AddTag("source", "ddtrace").
			AddTag("operation", "name").
			AddTag("span_type", "span_type").
			AddTag("env", "prod").
			AddTag("version", "v1.0.0").
			AddTag("status", "ok").
			AddTag("host", "localhost").
			Add("resource", "selece * from db").
			Add("message", "xxxxxxxxxxxxxxxxxxxxxxxxxxxxx").
			Add("duration", 1000*i)
		if i == 0 {
			kvs = kvs.Del("duration").Add("duration", 60000000*60) // 1 hour
			pt := point.NewPoint(inputName, kvs, point.DefaultLoggingOptions()...)
			pts = append(pts, pt)
		}

		for _, pt := range pts {
			spanMetrics(pt, labels, []string{})
		}

		time.Sleep(time.Millisecond)

		metricPts := itrace.GatherPoints(reg, map[string]string{})

		t.Logf("pts len = %d", len(metricPts))
		for _, pt := range metricPts {
			assert.NotEmpty(t, pt.GetTag(itrace.TagService))
			assert.NotEmpty(t, pt.GetTag(itrace.TagEnv))
			assert.NotEmpty(t, pt.GetTag(itrace.TagVersion))
			assert.NotEmpty(t, pt.GetTag(itrace.FieldResource))
		}
	}
}
