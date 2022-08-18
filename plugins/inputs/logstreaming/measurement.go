package logstreaming

import (
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

type logstreamingMeasurement struct{}

func (*logstreamingMeasurement) LineProto() (*point.Point, error) { return nil, nil }

//nolint:lll
func (*logstreamingMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Type: "logging",
		Name: "logstreaming 日志接收",
		Desc: "非行协议数据格式时，使用 URL 中的 `source` 参数，如果该值为空，则默认为 `default`",
		Tags: map[string]interface{}{
			"service":        inputs.NewTagInfo("service 名称，对应 URL 中的 `service` 参数"),
			"ip_or_hostname": inputs.NewTagInfo("request IP or hostname"),
		},
		Fields: map[string]interface{}{
			"message": &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "日志正文，默认存在，可以使用 pipeline 删除此字段"},
		},
	}
}
