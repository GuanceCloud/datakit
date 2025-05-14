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
		Name: measurementNginx,
		Cat:  point.Metric,
		Fields: map[string]interface{}{
			"load_timestamp":      newOtherFieldInfo(inputs.Int, inputs.Gauge, inputs.TimestampMS, "Nginx process load time in milliseconds, exist when using vts"),
			"connection_active":   newCountFieldInfo("The current number of active client connections"),
			"connection_reading":  newCountFieldInfo("The total number of reading client connections"),
			"connection_writing":  newCountFieldInfo("The total number of writing client connections"),
			"connection_waiting":  newCountFieldInfo("The total number of waiting client connections"),
			"connection_handled":  newCountFieldInfo("The total number of handled client connections"),
			"connection_requests": newCountFieldInfo("The total number of requests client connections"),
			"connection_accepts":  newCountFieldInfo("The total number of accepts client connections"),
			"connection_dropped":  newCountFieldInfo("The total number of dropped client connections"),
			"pid":                 newCountFieldInfo("The pid of nginx process (only for Nginx plus)"),
			"ppid":                newCountFieldInfo("The ppid of nginx process (only for Nginx plus)"),
		},
		Tags: map[string]interface{}{
			"nginx_server":  inputs.NewTagInfo("Nginx server host"),
			"nginx_port":    inputs.NewTagInfo("Nginx server port"),
			"host":          inputs.NewTagInfo("Host name which installed nginx"),
			"nginx_version": inputs.NewTagInfo("Nginx version, exist when using vts"),
		},
	}
}
