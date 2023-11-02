// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package disk

import (
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

type docMeasurement struct{}

//nolint:lll
func (*docMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "disk",
		Type: "metric",
		Fields: map[string]interface{}{
			"total": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.SizeByte,
				Desc: "Total disk size in bytes.",
			},
			"free": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.SizeByte,
				Desc: "Free disk size in bytes.",
			},
			"used": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.SizeByte,
				Desc: "Used disk size in bytes.",
			},
			"used_percent": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.Percent,
				Desc: "Used disk size in percent.",
			},
			"inodes_used_percent": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.Percent,
				Desc: "Inode used percent",
			},
			"inodes_total": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.NCount,
				Desc: "Total Inode(**DEPRECATED: use inodes_total_mb instead**).",
			},
			"inodes_total_mb": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.NCount,
				Desc: "Total Inode(need to multiply by 10^6).",
			},
			"inodes_free": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.NCount,
				Desc: "Free Inode(**DEPRECATED: use inodes_free_mb instead**).",
			},
			"inodes_free_mb": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.NCount,
				Desc: "Free Inode(need to multiply by 10^6).",
			},
			"inodes_used_mb": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.NCount,
				Desc: "Used Inode(need to multiply by 10^6).",
			},
			"inodes_used": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.NCount,
				Desc: "Used Inode(**DEPRECATED: use inodes_used_mb instead**).",
			},
		},
		Tags: map[string]interface{}{
			"host":        &inputs.TagInfo{Desc: "System hostname."},
			"device":      &inputs.TagInfo{Desc: "Disk device name. (on /dev/mapper return symbolic link, like `readlink /dev/mapper/*` result)"},
			"fstype":      &inputs.TagInfo{Desc: "File system name."},
			"mount_point": &inputs.TagInfo{Desc: "Mount point."},
		},
	}
}
