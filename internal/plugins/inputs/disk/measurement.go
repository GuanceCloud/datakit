// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package disk

import (
	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

type docMeasurement struct{}

//nolint:lll
func (*docMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   "disk",
		Cat:    point.Metric,
		Desc:   "Per-mounted filesystem capacity and inode usage metrics. Some filesystems, such as FAT-like filesystems on Linux, do not expose inode metrics.",
		DescZh: "按挂载点采集的文件系统容量和 inode 使用情况指标。部分文件系统（例如 Linux 上的 FAT 类文件系统）不提供 inode 指标。",
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
				Desc: "Inode used percent. Linux only",
			},
			"inodes_total_mb": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.NCount,
				Desc: "Total inode count divided by 1,000,000. Linux and macOS only.",
			},
			"inodes_free_mb": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.NCount,
				Desc: "Free inode count divided by 1,000,000. Linux and macOS only.",
			},
			"inodes_used_mb": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.NCount,
				Desc: "Used inode count divided by 1,000,000. Linux and macOS only.",
			},
			"inodes_total": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.NCount,
				Desc: "Deprecated. Total inode count. Linux and macOS only.",
			},
			"inodes_free": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.NCount,
				Desc: "Deprecated. Free inode count. Linux and macOS only.",
			},
			"inodes_used": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.NCount,
				Desc: "Deprecated. Used inode count. Linux and macOS only.",
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
