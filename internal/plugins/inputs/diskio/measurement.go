// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package diskio

import (
	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

type docMeasurement struct{}

// https://www.kernel.org/doc/Documentation/ABI/testing/procfs-diskstats
func (*docMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   "diskio",
		Cat:    point.Metric,
		Desc:   "Block device I/O counters and derived per-second throughput metrics collected from the operating system disk I/O statistics.",
		DescZh: "从操作系统磁盘 I/O 统计采集的块设备 I/O 累计计数，以及派生的每秒吞吐指标。",
		Fields: map[string]interface{}{
			"reads": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Count,
				Unit:     inputs.NCount,
				Desc:     "Cumulative number of read requests completed by the device.",
			},
			"writes": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Count,
				Unit:     inputs.NCount,
				Desc:     "Cumulative number of write requests completed by the device.",
			},
			"read_bytes": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Count,
				Unit:     inputs.SizeByte,
				Desc:     "Cumulative bytes read from the device.",
			},
			"read_bytes/sec": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.BytesPerSec,
				Desc:     "The number of bytes read on the device per second.",
			},
			"write_bytes": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Count,
				Unit:     inputs.SizeByte,
				Desc:     "Cumulative bytes written to the device.",
			},
			"write_bytes/sec": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.BytesPerSec,
				Desc:     "The number of bytes written to the device per second.",
			},
			"read_time": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Count,
				Unit:     inputs.DurationMS,
				Desc:     "Cumulative time spent reading from the device.",
			},
			"write_time": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Count,
				Unit:     inputs.DurationMS,
				Desc:     "Cumulative time spent writing to the device.",
			},
			"io_time": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Count,
				Unit:     inputs.DurationMS,
				Desc:     "Cumulative time spent doing I/O. Linux only.",
			},
			"weighted_io_time": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Count,
				Unit:     inputs.DurationMS,
				Desc:     "Cumulative weighted time spent doing I/O. Linux only.",
			},
			"await": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.DurationMS,
				Desc:     "Average time for I/O requests to be served (ms). Linux only",
			},
			"iops_in_progress": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "I/Os currently in progress. Linux only",
			},
			"merged_reads": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Count,
				Unit:     inputs.NCount,
				Desc:     "Cumulative number of merged read requests. Linux only.",
			},
			"merged_writes": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Count,
				Unit:     inputs.NCount,
				Desc:     "Cumulative number of merged write requests. Linux only.",
			},
		},
		Tags: map[string]interface{}{
			"host":   &inputs.TagInfo{Desc: "System hostname."},
			"name":   &inputs.TagInfo{Desc: "Device name after applying `name_templates`, or `/dev/<device>` when no template matches."},
			"serial": &inputs.TagInfo{Desc: "Device serial number, or `unknown` when unavailable. Omitted when `skip_serial_number` is enabled."},
		},
	}
}
