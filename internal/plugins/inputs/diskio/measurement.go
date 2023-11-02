// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package diskio

import (
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

type docMeasurement struct{}

// https://www.kernel.org/doc/Documentation/ABI/testing/procfs-diskstats

func (*docMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "diskio",
		Type: "metric",
		Fields: map[string]interface{}{
			"reads":            newFieldsInfoCount("The number of read requests."),
			"writes":           newFieldsInfoCount("The number of write requests."),
			"read_bytes":       newFieldsInfoBytes("The number of bytes read from the device."),
			"read_bytes/sec":   newFieldsInfoBytesPerSec("The number of bytes read from the per second."),
			"write_bytes":      newFieldsInfoBytes("The number of bytes written to the device."),
			"write_bytes/sec":  newFieldsInfoBytesPerSec("The number of bytes written to the device per second."),
			"read_time":        newFieldsInfoMS("Time spent reading."),
			"write_time":       newFieldsInfoMS("Time spent writing."),
			"io_time":          newFieldsInfoMS("Time spent doing I/Os."),
			"weighted_io_time": newFieldsInfoMS("Weighted time spent doing I/Os."),
			"iops_in_progress": newFieldsInfoCount("I/Os currently in progress."),
			"merged_reads":     newFieldsInfoCount("The number of merged read requests."),
			"merged_writes":    newFieldsInfoCount("The number of merged write requests."),
		},
		Tags: map[string]interface{}{
			"host": &inputs.TagInfo{Desc: "System hostname."},
			"name": &inputs.TagInfo{Desc: "Device name."},
		},
	}
}

func newFieldsInfoBytes(desc string) *inputs.FieldInfo {
	return &inputs.FieldInfo{
		Type:     inputs.Gauge,
		DataType: inputs.Int,
		Unit:     inputs.SizeByte,
		Desc:     desc,
	}
}

func newFieldsInfoBytesPerSec(desc string) *inputs.FieldInfo {
	return &inputs.FieldInfo{
		Type:     inputs.Gauge,
		DataType: inputs.Int,
		Unit:     inputs.BytesPerSec,
		Desc:     desc,
	}
}

func newFieldsInfoCount(desc string) *inputs.FieldInfo {
	return &inputs.FieldInfo{
		Type:     inputs.Gauge,
		DataType: inputs.Int,
		Unit:     inputs.NCount,
		Desc:     desc,
	}
}

func newFieldsInfoMS(desc string) *inputs.FieldInfo {
	return &inputs.FieldInfo{
		Type:     inputs.Gauge,
		DataType: inputs.Int,
		Unit:     inputs.DurationMS,
		Desc:     desc,
	}
}
