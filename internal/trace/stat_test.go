// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package trace

import (
	"testing"

	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/point"
)

func TestStatTracingInfo(t *testing.T) {
	dkioFeed = func(name, category string, pts []*point.Point, opt *dkio.Option) error { return nil }

	var traces DatakitTraces
	for i := 0; i < 100; i++ {
		trace := randDatakitTrace(t, 10, randService(_services...), randResource(_resources...))
		traces = append(traces, trace)
	}

	StartTracingStatistic()

	for i := range traces {
		StatTracingInfo(traces[i])
	}
}
