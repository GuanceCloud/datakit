// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package nginx

import (
	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

type NginxMeasurement struct{}

//nolint:lll
func (m *NginxMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   measurementNginx,
		Cat:    point.Metric,
		Desc:   "NGINX instance-level connection and process metrics collected from stub_status, VTS, or the NGINX Plus API.",
		DescZh: "通过 stub_status、VTS 或 NGINX Plus API 采集的 NGINX 实例级连接和进程指标。",
		Fields: map[string]interface{}{
			"load_timestamp":      &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.TimestampMS, Desc: "NGINX VTS module load time as a Unix timestamp in milliseconds.", Taggedby: []string{"nginx_server", "nginx_port"}},
			"connection_active":   &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Current number of active client connections.", Taggedby: []string{"nginx_server", "nginx_port"}},
			"connection_reading":  &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Current number of client connections where NGINX is reading the request header.", Taggedby: []string{"nginx_server", "nginx_port"}},
			"connection_writing":  &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Current number of client connections where NGINX is writing the response.", Taggedby: []string{"nginx_server", "nginx_port"}},
			"connection_waiting":  &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Current number of idle keep-alive client connections waiting for a request.", Taggedby: []string{"nginx_server", "nginx_port"}},
			"connection_handled":  &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Total number of client connections handled by NGINX.", Taggedby: []string{"nginx_server", "nginx_port"}},
			"connection_requests": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Total number of client requests processed by NGINX.", Taggedby: []string{"nginx_server", "nginx_port"}},
			"connection_accepts":  &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Total number of accepted client connections.", Taggedby: []string{"nginx_server", "nginx_port"}},
			"connection_dropped":  &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Total number of accepted client connections that were not handled.", Taggedby: []string{"nginx_server", "nginx_port"}},
			"pid":                 &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NoUnit, Desc: "NGINX master process ID reported by the NGINX Plus API.", Taggedby: []string{"nginx_server", "nginx_port"}},
			"ppid":                &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NoUnit, Desc: "Parent process ID for the NGINX master process reported by the NGINX Plus API.", Taggedby: []string{"nginx_server", "nginx_port"}},
		},
		Tags: map[string]interface{}{
			"nginx_server":  inputs.NewTagInfo("Configured NGINX server host or URL host."),
			"nginx_port":    inputs.NewTagInfo("Configured NGINX server port."),
			"host":          inputs.NewTagInfo("Host name reported by VTS or the host running the NGINX collector."),
			"nginx_version": inputs.NewTagInfo("NGINX version reported by VTS or the NGINX Plus API."),
		},
	}
}
