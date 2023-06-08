// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package logstreaming

import (
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

type logstreamingMeasurement struct{}

func (*logstreamingMeasurement) LineProto() (*point.Point, error) { return nil, nil }

//nolint:lll
func (*logstreamingMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Type: "logging",
		Name: "logstreaming",
		Desc: "非行协议数据格式时，使用 URL 中的 `source` 参数，如果该值为空，则默认为 `default`",
		Tags: map[string]interface{}{
			"service":        inputs.NewTagInfo("service 名称，对应 URL 中的 `service` 参数"),
			"ip_or_hostname": inputs.NewTagInfo("request IP or hostname"),
		},
		Fields: map[string]interface{}{
			"message": &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "日志正文，默认存在，可以使用 Pipeline 删除此字段"},
		},
	}
}
