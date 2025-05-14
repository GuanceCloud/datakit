// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package hostobject

import (
	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

type docMeasurement struct{}

//nolint:lll
func (*docMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: hostObjMeasurementName,
		Cat:  point.Object,
		Desc: "Host object metrics",
		Tags: map[string]interface{}{
			"host":       &inputs.TagInfo{Desc: "Hostname. Required."},
			"unicast_ip": &inputs.TagInfo{Desc: "Host unicast ip"},
			"name":       &inputs.TagInfo{Desc: "Hostname"},
			"os":         &inputs.TagInfo{Desc: "Host OS type"},
			"arch":       &inputs.TagInfo{Desc: "Host OS Arch"},
		},
		Fields: map[string]interface{}{
			"message":                    &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "Summary of all host information"},
			"start_time":                 &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.DurationMS, Desc: "Host startup time (Unix timestamp)"},
			"datakit_ver":                &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "Collector version"},
			"cpu_usage":                  &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.Percent, Desc: "CPU usage"},
			"num_cpu":                    &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.NCount, Desc: "CPU numbers"},
			"mem_used_percent":           &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.Percent, Desc: "Memory usage"},
			"load":                       &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.UnknownUnit, Desc: "System load"},
			"disk_total":                 &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.SizeByte, Desc: "Disk total"},
			"disk_used_percent":          &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.Percent, Desc: "Disk usage"},
			"diskio_read_bytes_per_sec":  &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.BytesPerSec, Desc: "Disk read rate"},
			"diskio_write_bytes_per_sec": &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.BytesPerSec, Desc: "Disk write rate"},
			"net_recv_bytes_per_sec":     &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.BytesPerSec, Desc: "Network receive rate"},
			"net_send_bytes_per_sec":     &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.BytesPerSec, Desc: "Network send rate"},
			"logging_level":              &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "Log level"},
			"is_docker":                  &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.UnknownUnit, Desc: "Docker mode"},
			"dk_upgrader":                &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "Upgrade's host and port"},
		},
	}
}
