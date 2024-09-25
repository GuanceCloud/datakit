// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package metrics

import (
	"testing"

	"github.com/grafana/jfr-parser/common/attributes"
	"github.com/grafana/jfr-parser/common/filters"
	"github.com/grafana/jfr-parser/common/types"
	"github.com/grafana/jfr-parser/parser"
)

var chunks = func() jfrChunks {
	chunks, err := parser.ParseFile("testdata/main.jfr")
	if err != nil {
		panic(err)
	}
	return chunks
}()

func TestResolveDDProfilerSetting(t *testing.T) {
	c := chunks.resolveDDProfilerSetting()

	t.Logf("%+#v", *c)
}

func TestCpuCores(t *testing.T) {
	maxReadTimeNS, maxBytesRead, totalReadTimeNS, totalBytesRead := chunks.socketIORead()
	t.Log(maxReadTimeNS, maxBytesRead, totalReadTimeNS, totalBytesRead)
}

func TestAllocWeight(t *testing.T) {
	for _, chunk := range chunks {
		for _, event := range chunk.Apply(filters.DatadogAllocationSample) {
			value, err := attributes.AllocWeight.GetValue(event)
			if err != nil {
				t.Fatal(err)
			}
			t.Log(value)
		}
	}
}

func TestCompilationDuration(t *testing.T) {
	durationNS := chunks.compilationDuration()
	t.Log(durationNS)
}

func TestAllocations(t *testing.T) {
	maxNS, totalPauseNS, count := chunks.gcPauseDuration()

	t.Log(maxNS, totalPauseNS, count)
}

func TestShowClassMeta(t *testing.T) {
	for _, chunk := range chunks {
		chunk.ShowClassMeta(types.SocketRead)
	}
}
