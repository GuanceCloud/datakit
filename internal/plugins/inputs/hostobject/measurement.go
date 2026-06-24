// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//nolint:lll,goconst,funlen // Measurement metadata contains intentionally long repeated literals.
package hostobject

import (
	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

type docMeasurement struct{}

//nolint:lll
func (*docMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   hostObjMeasurementName,
		Cat:    point.Object,
		Desc:   "Host object metrics",
		DescZh: "主机对象信息，记录主机身份、操作系统、DataKit 版本及主机资源概览字段。",
		Tags: map[string]interface{}{
			"host":       &inputs.TagInfo{Desc: "Hostname. Required."},
			"unicast_ip": &inputs.TagInfo{Desc: "Host unicast ip"},
			"name":       &inputs.TagInfo{Desc: "Hostname"},
			"os":         &inputs.TagInfo{Desc: "Host OS type"},
			"arch":       &inputs.TagInfo{Desc: "Host OS Arch"},
		},
		Fields: map[string]interface{}{
			"message":                    &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.NoUnit, Desc: "JSON summary of the collected host object fields for this host."},
			"start_time":                 &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.TimestampMS, Desc: "Host boot time as a Unix epoch timestamp in milliseconds."},
			"datakit_ver":                &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.NoUnit, Desc: "DataKit version string that produced this host object."},
			"cpu_usage":                  &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.Percent, Desc: "Current total CPU usage percentage for the host."},
			"num_cpu":                    &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.NCount, Desc: "Number of logical CPUs on the host."},
			"mem_used_percent":           &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.Percent, Desc: "Current host memory usage percentage."},
			"load":                       &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.NoUnit, Desc: "Current host system load value, collected from the host load average."},
			"disk_total":                 &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.SizeByte, Desc: "Total disk capacity on the host in bytes."},
			"disk_used_percent":          &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.Percent, Desc: "Current host disk usage percentage."},
			"diskio_read_bytes_per_sec":  &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.BytesPerSec, Desc: "Current host disk read throughput in bytes per second."},
			"diskio_write_bytes_per_sec": &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.BytesPerSec, Desc: "Current host disk write throughput in bytes per second."},
			"net_recv_bytes_per_sec":     &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.BytesPerSec, Desc: "Current host network receive throughput in bytes per second."},
			"net_send_bytes_per_sec":     &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.BytesPerSec, Desc: "Current host network send throughput in bytes per second."},
			"logging_level":              &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.NoUnit, Desc: "Current DataKit logging level string on this host."},
			"is_docker":                  &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.Bool, Desc: "Whether this DataKit instance is running in Docker: 1 means true and 0 means false."},
		},
	}
}
