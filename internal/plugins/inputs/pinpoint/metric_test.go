// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package pinpoint metrics.
package pinpoint

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/pinpoint/cache"

	"google.golang.org/grpc/metadata"

	ppv1 "github.com/GuanceCloud/tracing-protos/pinpoint-gen-go/v1"
)

var statBatchJson = `
[
	 {
		"timestamp": 1700034163398,
		"collectInterval": 5000,
		"gc": {
			"type": 1,
			"jvmMemoryHeapUsed": 34614400,
			"jvmMemoryHeapMax": 921174016,
			"jvmMemoryNonHeapUsed": 77776360,
			"jvmMemoryNonHeapMax": -1,
			"jvmGcOldCount": 3,
			"jvmGcOldTime": 261,
			"jvmGcDetailed": {
				"jvmGcNewCount": 12,
				"jvmGcNewTime": 166,
				"jvmPoolCodeCacheUsed": 0.04169387817382812,
				"jvmPoolNewGenUsed": 0.04419467192907675,
				"jvmPoolOldGenUsed": 0.02950394533472828,
				"jvmPoolPermGenUsed": -1,
				"jvmPoolMetaspaceUsed": 0.9600545592227224
			}
		},
		"cpuLoad": {
			"jvmCpuLoad": 0.024096385542168672,
			"systemCpuLoad": 0.028112449799196786
		},
		"transaction": {},
		"dataSourceList": {},
		"responseTime": {},
		"fileDescriptor": {
			"openFileDescriptorCount": 180
		},
		"directBuffer": {
			"directCount": 10,
			"directMemoryUsed": 270337
		},
		"totalThread": {
			"totalThreadCount": 32
		},
		"loadedClass": {
			"loadedClassCount": 11580
		}
	},
	{
		"timestamp": 1700034173397,
		"collectInterval": 5000,
		"gc": {
			"type": 1,
			"jvmMemoryHeapUsed": 34614400,
			"jvmMemoryHeapMax": 921174016,
			"jvmMemoryNonHeapUsed": 77776360,
			"jvmMemoryNonHeapMax": -1,
			"jvmGcOldCount": 3,
			"jvmGcOldTime": 261,
			"jvmGcDetailed": {
				"jvmGcNewCount": 12,
				"jvmGcNewTime": 166,
				"jvmPoolCodeCacheUsed": 0.04169387817382812,
				"jvmPoolNewGenUsed": 0.04419467192907675,
				"jvmPoolOldGenUsed": 0.02950394533472828,
				"jvmPoolPermGenUsed": -1,
				"jvmPoolMetaspaceUsed": 0.9600545592227224
			}
		},
		"cpuLoad": {
			"jvmCpuLoad": 0.024096385542168672,
			"systemCpuLoad": 0.028112449799196786
		},
		"transaction": {},
		"dataSourceList": {},
		"responseTime": {},
		"fileDescriptor": {
			"openFileDescriptorCount": 180
		},
		"directBuffer": {
			"directCount": 10,
			"directMemoryUsed": 270337
		},
		"totalThread": {
			"totalThreadCount": 32
		},
		"loadedClass": {
			"loadedClassCount": 11580
		}
	}
]`

func Test_statBatchToPoints(t *testing.T) {
	agentKey := "agentid"
	batch := &ppv1.PAgentStatBatch{}
	stats := make([]*ppv1.PAgentStat, 0)
	err := json.Unmarshal([]byte(statBatchJson), &stats)
	if err != nil {
		t.Errorf("Unmarshal err=%v", err)
		return
	}
	batch.AgentStat = stats
	t.Logf("batch len=%d", len(batch.GetAgentStat()))
	agentCache = &cache.AgentCache{
		Agents:    make(map[string]*ppv1.PAgentInfo),
		Metas:     make(map[string]*cache.MetaData),
		EventData: make(map[string]*cache.EventItem),
		SpanData:  make(map[int64]*cache.SpanItem),
	}
	agentCache.SetAgentInfo(agentKey, &ppv1.PAgentInfo{
		Hostname:     "testClient",
		Ip:           "10.200.10.10",
		Ports:        "8081,8080",
		ServiceType:  1202,
		Pid:          123,
		AgentVersion: "2.3.2",
	})

	MD := metadata.MD{agentKey: {"agentid"}}
	pts := statBatchToPoints(MD, batch)
	log.Debugf("pts.len=%d", len(pts))
	for _, pt := range pts {
		t.Logf("point:=%s", pt.LineProto())
		assert.Equal(t, pt.GetTag("hostname"), "testClient")
		assert.Equal(t, pt.GetTag("ip"), "10.200.10.10")
	}
}

func BenchmarkStatBatchToPoints(b *testing.B) {
	agentKey := "agentid"
	batch := &ppv1.PAgentStatBatch{}
	stats := make([]*ppv1.PAgentStat, 0)
	err := json.Unmarshal([]byte(statBatchJson), &stats)
	if err != nil {
		b.Errorf("Unmarshal err=%v", err)
		return
	}
	batch.AgentStat = stats
	agentCache = &cache.AgentCache{
		Agents:    make(map[string]*ppv1.PAgentInfo),
		Metas:     make(map[string]*cache.MetaData),
		EventData: make(map[string]*cache.EventItem),
		SpanData:  make(map[int64]*cache.SpanItem),
	}
	agentCache.SetAgentInfo(agentKey, &ppv1.PAgentInfo{
		Hostname:     "testClient",
		Ip:           "10.200.10.10",
		Ports:        "8081,8080",
		ServiceType:  1202,
		Pid:          123,
		AgentVersion: "2.3.2",
	})

	MD := metadata.MD{agentKey: {"agentid"}}
	for i := 0; i < b.N; i++ {
		_ = statBatchToPoints(MD, batch)
	}
	/*
		goos: linux
		goarch: amd64
		pkg: gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/pinpoint
		cpu: AMD Ryzen 7 7700X 8-Core Processor
		BenchmarkStatBatchToPoints-16             210639              5529 ns/op            7397 B/op        163 allocs/op
		PASS
		ok      gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/pinpoint   2.257s
	*/
}
