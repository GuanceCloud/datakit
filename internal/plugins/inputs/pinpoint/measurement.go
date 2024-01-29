// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package pinpoint is JVM metrics
package pinpoint

import (
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

type Measurement struct{}

func (m *Measurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: inputName,
		Fields: map[string]interface{}{
			"SystemCpuLoad": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.Percent,
				Desc: "system CPU load",
			},

			"JvmCpuLoad": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.Percent,
				Desc: "Jvm CPU load",
			},

			"JvmMemoryHeapUsed": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.SizeByte,
				Desc: "Jvm Memory Heap Used",
			},

			"JvmMemoryHeapMax": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.SizeByte,
				Desc: "Jvm Memory Heap Max",
			},

			"JvmMemoryNonHeapUsed": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.SizeByte,
				Desc: "Jvm Memory NonHeap Used",
			},

			"JvmMemoryNonHeapMax": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.SizeByte,
				Desc: "Jvm Memory NonHeap Max",
			},

			"JvmGcOldCount": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.NCount,
				Desc: "Jvm Gc Old Count",
			},

			"JvmGcOldTime": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.TimestampMS,
				Desc: "Jvm Gc Old Time",
			},

			"GcNewCount": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.NCount,
				Desc: "Jvm Gc NewCount",
			},

			"GcNewTime": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.TimestampMS,
				Desc: "Jvm Gc NewTime",
			},

			"PoolCodeCacheUsed": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.SizeByte,
				Desc: "Jvm Pool Code Cache Used",
			},

			"PoolNewGenUsed": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.SizeByte,
				Desc: "Jvm Pool New GenUsed",
			},

			"PoolOldGenUsed": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.SizeByte,
				Desc: "Duration of Jvm garbage collection actions",
			},

			"PoolSurvivorSpaceUsed": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.SizeByte,
				Desc: "Jvm Pool Survivor SpaceUsed",
			},

			"PoolPermGenUsed": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.NCount,
				Desc: "The maximum file descriptor count",
			},

			"PoolMetaspaceUsed": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.NCount,
				Desc: "Jvm Pool meta space used",
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
		Type: "metric",
	}
}
