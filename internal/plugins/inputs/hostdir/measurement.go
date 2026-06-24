// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package hostdir

import (
	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

type Measurement struct{}

//nolint:lll
func (m *Measurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   inputName,
		Cat:    point.Metric,
		Desc:   "Directory usage metrics for configured host directories, including scanned file size, entry counts, containing filesystem capacity, and inode usage.",
		DescZh: "配置的主机目录使用情况指标，包括扫描到的文件大小、目录项数量、所在文件系统容量和 inode 使用情况。",
		Fields: map[string]interface{}{
			"file_size":           &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.SizeByte, Desc: "Total size in bytes of files currently stored under the monitored directory."},
			"file_count":          &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.NCount, Desc: "Current number of files and subdirectories under the monitored directory."},
			"dir_count":           &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.NCount, Desc: "Current number of subdirectories under the monitored directory."},
			"total":               &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.SizeByte, Desc: "Total filesystem capacity in bytes for the mount that contains the monitored directory."},
			"free":                &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.SizeByte, Desc: "Free filesystem capacity in bytes for the mount that contains the monitored directory."},
			"used_percent":        &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.Percent, Desc: "Percentage of filesystem capacity used by files under the monitored directory relative to the containing filesystem."},
			"inodes_used_percent": &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.Percent, Desc: "Percentage of filesystem inodes used by files and subdirectories under the monitored directory relative to the containing filesystem."},
			"inodes_total":        &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.NCount, Desc: "Total number of inodes available on the filesystem that contains the monitored directory."},
			"inodes_free":         &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.NCount, Desc: "Number of free inodes available on the filesystem that contains the monitored directory."},
			"inodes_used":         &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.NCount, Desc: "Number of inodes currently used by files and subdirectories under the monitored directory."},
		},
		Tags: map[string]interface{}{
			"host_directory": inputs.NewTagInfo("The root directory path being scanned by the collector."),
			"file_mode":      inputs.NewTagInfo("Permission mode of the monitored directory."),
			"file_ownership": inputs.NewTagInfo("Ownership mode of files counted under the monitored directory."),
			"file_system":    inputs.NewTagInfo("Filesystem type of the mount that contains the monitored directory."),
			"mount_point":    &inputs.TagInfo{Desc: "Mount point that contains the monitored directory."},
		},
	}
}
