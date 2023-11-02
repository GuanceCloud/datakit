// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package hostdir

import (
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

type Measurement struct{}

//nolint:lll
func (m *Measurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: inputName,
		Type: "metric",
		Fields: map[string]interface{}{
			"file_size":           &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.SizeByte, Desc: "The size of files."},
			"file_count":          &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.NCount, Desc: "The number of files."},
			"dir_count":           &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.NCount, Desc: "The number of Dir."},
			"total":               &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.SizeByte, Desc: "Total disk size in bytes."},
			"free":                &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.SizeByte, Desc: "Free disk size in bytes."},
			"used_percent":        &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.Percent, Desc: "Used disk size in percent(only this dir used in total size)."},
			"inodes_used_percent": &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.Percent, Desc: "Inode used percent(only this dir used in total inode)."},
			"inodes_total":        &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.NCount, Desc: "Total inode."},
			"inodes_free":         &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.NCount, Desc: "Free inode."},
			"inodes_used":         &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.NCount, Desc: "Used inode(only this dir used)."},
		},
		Tags: map[string]interface{}{
			"host_directory": inputs.NewTagInfo("the start Dir."),
			"file_ownership": inputs.NewTagInfo("file ownership."),
			"file_system":    inputs.NewTagInfo("file system type."),
			"mount_point":    &inputs.TagInfo{Desc: "Mount point."},
		},
	}
}
