// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package pinpoint is JVM metrics
package pinpoint

import (
	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

type Measurement struct{}

func (m *Measurement) Info() *inputs.MeasurementInfo {
	agentTags := []string{"agent_id", "hostname", "ip", "pid", "ports", "container", "agentVersion"}

	return &inputs.MeasurementInfo{
		Name:   agentStatsMeasurement,
		Desc:   "Pinpoint agent JVM runtime statistics converted from Pinpoint PAgentStat batches.",
		DescZh: "从 Pinpoint PAgentStat 批次转换的 Pinpoint Agent JVM 运行时统计指标。",
		Fields: map[string]interface{}{
			"SystemCpuLoad": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.PercentDecimal,
				Desc:     "Recent system CPU load ratio reported by the Pinpoint agent, in the range 0 to 1.",
				Taggedby: agentTags,
			},

			"JvmCpuLoad": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.PercentDecimal,
				Desc:     "Recent JVM process CPU load ratio reported by the Pinpoint agent, in the range 0 to 1.",
				Taggedby: agentTags,
			},

			"JvmMemoryHeapUsed": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.SizeByte,
				Desc:     "Current JVM heap memory used in bytes.",
				Taggedby: agentTags,
			},

			"JvmMemoryHeapMax": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.SizeByte,
				Desc:     "Maximum JVM heap memory in bytes.",
				Taggedby: agentTags,
			},

			"JvmMemoryNonHeapUsed": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.SizeByte,
				Desc:     "Current JVM non-heap memory used in bytes.",
				Taggedby: agentTags,
			},

			"JvmMemoryNonHeapMax": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.SizeByte,
				Desc:     "Maximum JVM non-heap memory in bytes; Pinpoint may report -1 when the maximum is undefined.",
				Taggedby: agentTags,
			},

			"JvmGcOldCount": &inputs.FieldInfo{
				Type: inputs.Count, DataType: inputs.Int, Unit: inputs.NCount,
				Desc:     "Cumulative old-generation garbage collection count reported by the Pinpoint agent.",
				Taggedby: agentTags,
			},

			"JvmGcOldTime": &inputs.FieldInfo{
				Type: inputs.Count, DataType: inputs.Int, Unit: inputs.DurationMS,
				Desc:     "Cumulative old-generation garbage collection time in milliseconds reported by the Pinpoint agent.",
				Taggedby: agentTags,
			},

			"GcNewCount": &inputs.FieldInfo{
				Type: inputs.Count, DataType: inputs.Int, Unit: inputs.NCount,
				Desc:     "Cumulative new-generation garbage collection count reported by the Pinpoint agent.",
				Taggedby: agentTags,
			},

			"GcNewTime": &inputs.FieldInfo{
				Type: inputs.Count, DataType: inputs.Int, Unit: inputs.DurationMS,
				Desc:     "Cumulative new-generation garbage collection time in milliseconds reported by the Pinpoint agent.",
				Taggedby: agentTags,
			},

			"PoolCodeCacheUsed": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.PercentDecimal,
				Desc:     "JVM code-cache memory pool usage ratio reported by the Pinpoint agent, in the range 0 to 1.",
				Taggedby: agentTags,
			},

			"PoolNewGenUsed": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.PercentDecimal,
				Desc:     "JVM new-generation memory pool usage ratio reported by the Pinpoint agent, in the range 0 to 1.",
				Taggedby: agentTags,
			},

			"PoolOldGenUsed": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.PercentDecimal,
				Desc:     "JVM old-generation memory pool usage ratio reported by the Pinpoint agent, in the range 0 to 1.",
				Taggedby: agentTags,
			},

			"PoolSurvivorSpaceUsed": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.PercentDecimal,
				Desc:     "JVM survivor-space memory pool usage ratio reported by the Pinpoint agent, in the range 0 to 1.",
				Taggedby: agentTags,
			},

			"PoolPermGenUsed": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.PercentDecimal,
				Desc:     "JVM permanent-generation memory pool usage ratio reported by the Pinpoint agent; -1 means unavailable on JVMs without PermGen.",
				Taggedby: agentTags,
			},

			"PoolMetaspaceUsed": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.PercentDecimal,
				Desc:     "JVM metaspace memory pool usage ratio reported by the Pinpoint agent, in the range 0 to 1.",
				Taggedby: agentTags,
			},
		},

		Tags: map[string]interface{}{
			"hostname":     &inputs.TagInfo{Desc: "Host name"},
			"agent_id":     &inputs.TagInfo{Desc: "Agent ID"},
			"ip":           &inputs.TagInfo{Desc: "Agent IP"},
			"pid":          &inputs.TagInfo{Desc: "Process ID"},
			"ports":        &inputs.TagInfo{Desc: "Open ports"},
			"container":    &inputs.TagInfo{Desc: "Whether it is a container"},
			"agentVersion": &inputs.TagInfo{Desc: "Pinpoint agent version"},
		},
		Cat: point.Metric,
	}
}
