// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//nolint:lll,goconst,funlen // Measurement metadata contains intentionally long repeated literals.
package logstreaming

import (
	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

const defaultMeasurementName = "default"

type logstreamingMeasurement struct{}

//nolint:lll
func (*logstreamingMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Cat:            point.Logging,
		MetaDuplicated: true,
		Name:           defaultMeasurementName,
		Desc: "Remote log entries received through the logstreaming HTTP endpoint. " +
			"The point name comes from the `source` query parameter and defaults to `default`.",
		DescZh: "通过 logstreaming HTTP 接口接收的远程日志。指标集名称来自 `source` 查询参数，默认值为 `default`。",
		Tags: map[string]interface{}{
			"collector_source_ip": inputs.NewTagInfo("Remote client IP address resolved from the HTTP request."),
			"service":             inputs.NewTagInfo("Service name. Using the `service` parameter in the URL."),
			"ip_or_hostname":      inputs.NewTagInfo("Hostname part of the logstreaming request URL."),
		},
		Fields: map[string]interface{}{
			"message": &inputs.FieldInfo{
				DataType: inputs.String,
				Unit:     inputs.NoUnit,
				Desc:     "Log message text parsed from the request body or protocol-specific payload.",
			},
			"status": &inputs.FieldInfo{
				DataType: inputs.String,
				Unit:     inputs.NoUnit,
				Desc:     "Log status string assigned to the streamed log entry before pipeline processing.",
			},
		},
	}
}
