// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package lsblk

import (
	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

type docMeasurement struct{}

//nolint:lll
func (*docMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   "lsblk",
		Cat:    point.Metric,
		Desc:   "Linux block device inventory and mounted filesystem capacity collected from `/dev/block`, `/sys`, and `statfs`.",
		DescZh: "从 `/dev/block`、`/sys` 和 `statfs` 采集的 Linux 块设备清单和已挂载文件系统容量指标。",
		Fields: map[string]interface{}{
			"fs_avail": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.SizeByte,
				Desc:     "Available bytes on the mounted filesystem from `statfs.Bavail * statfs.Bsize`.",
				Taggedby: []string{"name", "kname", "maj_min", "parent", "is_mounted", "label", "uuid", "serial", "model", "state", "type", "vendor", "owner", "group", "mountpoint"},
			},
			"fs_size": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.SizeByte,
				Desc:     "Total bytes on the mounted filesystem from `statfs.Blocks * statfs.Bsize`.",
				Taggedby: []string{"name", "kname", "maj_min", "parent", "is_mounted", "label", "uuid", "serial", "model", "state", "type", "vendor", "owner", "group", "mountpoint"},
			},
			"fs_used": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.SizeByte,
				Desc:     "Used bytes on the mounted filesystem from `(statfs.Blocks - statfs.Bfree) * statfs.Bsize`.",
				Taggedby: []string{"name", "kname", "maj_min", "parent", "is_mounted", "label", "uuid", "serial", "model", "state", "type", "vendor", "owner", "group", "mountpoint"},
			},
			"fs_used_percent": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.Percent,
				Desc:     "Used filesystem capacity percentage, calculated as `fs_used / fs_size * 100`.",
				Taggedby: []string{"name", "kname", "maj_min", "parent", "is_mounted", "label", "uuid", "serial", "model", "state", "type", "vendor", "owner", "group", "mountpoint"},
			},
			"rq_size": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.NCount,
				Desc:     "Maximum number of I/O requests allowed in the block device request queue from `/sys/.../queue/nr_requests`.",
				Taggedby: []string{"name", "kname", "maj_min", "parent", "is_mounted", "label", "uuid", "serial", "model", "state", "type", "vendor", "owner", "group"},
			},
			"size": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.SizeByte,
				Desc:     "Block device size in bytes, calculated from the `/sys` sector count multiplied by 512.",
				Taggedby: []string{"name", "kname", "maj_min", "parent", "is_mounted", "label", "uuid", "serial", "model", "state", "type", "vendor", "owner", "group"},
			},
		},
		Tags: map[string]interface{}{
			"name":       &inputs.TagInfo{Desc: "Device name."},
			"kname":      &inputs.TagInfo{Desc: "Internal kernel device name."},
			"parent":     &inputs.TagInfo{Desc: "Parent device name."},
			"maj_min":    &inputs.TagInfo{Desc: "Major:Minor device number."},
			"is_mounted": &inputs.TagInfo{Desc: "Whether the device has a mounted filesystem. Values are `yes` or `no`."},
			"mountpoint": &inputs.TagInfo{Desc: "Where the device is mounted."},
			"label":      &inputs.TagInfo{Desc: "Filesystem LABEL."},
			"uuid":       &inputs.TagInfo{Desc: "Filesystem UUID."},
			"model":      &inputs.TagInfo{Desc: "Device identifier."},
			"serial":     &inputs.TagInfo{Desc: "Disk serial number."},
			"state":      &inputs.TagInfo{Desc: "State of the device."},
			"type":       &inputs.TagInfo{Desc: "Device type."},
			"vendor":     &inputs.TagInfo{Desc: "Device vendor."},
			"owner":      &inputs.TagInfo{Desc: "User name."},
			"group":      &inputs.TagInfo{Desc: "Group name."},
		},
	}
}
