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
		Name: "lsblk",
		Cat:  point.Metric,
		Fields: map[string]interface{}{
			"fsavail": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.SizeByte,
				Desc: "Available space on the filesystem.",
			},
			"fssize": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.SizeByte,
				Desc: "Total size of the filesystem.",
			},
			"fsused": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.SizeByte,
				Desc: "Used space on the filesystem.",
			},
			"fs_used_percent": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.Percent,
				Desc: "Percentage of used space on the filesystem.",
			},
			"rq_size": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.SizeByte,
				Desc: "Request queue size.",
			},
			"size": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.SizeByte,
				Desc: "Size of the device.",
			},
		},
		Tags: map[string]interface{}{
			"name":       &inputs.TagInfo{Desc: "Device name."},
			"kname":      &inputs.TagInfo{Desc: "Internal kernel device name."},
			"parent":     &inputs.TagInfo{Desc: "Parent device name."},
			"maj_min":    &inputs.TagInfo{Desc: "Major:Minor device number."},
			"fstype":     &inputs.TagInfo{Desc: "Filesystem type."},
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
