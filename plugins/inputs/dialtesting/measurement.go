package dialtesting

import (
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

type httpMeasurement struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
	ts     time.Time
}

func (m *httpMeasurement) LineProto() (*io.Point, error) {
	return io.MakePoint(m.name, m.tags, m.fields, m.ts)
}

//nolint:lll
func (m *httpMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "http_dial_testing",
		Tags: map[string]interface{}{
			"name":               &inputs.TagInfo{Desc: "示例：拨测名称,百度测试"},
			"url":                &inputs.TagInfo{Desc: "示例 http://wwww.baidu.com"},
			"country":            &inputs.TagInfo{Desc: "示例 中国"},
			"province":           &inputs.TagInfo{Desc: "示例 浙江"},
			"city":               &inputs.TagInfo{Desc: "示例 杭州"},
			"internal":           &inputs.TagInfo{Desc: "示例 true（国内 true /海外 false）"},
			"isp":                &inputs.TagInfo{Desc: "示例 电信/移动/联通"},
			"status":             &inputs.TagInfo{Desc: "示例 OK/FAIL 两种状态 "},
			"status_code_class":  &inputs.TagInfo{Desc: "示例 2xx"},
			"status_code_string": &inputs.TagInfo{Desc: "示例 200 OK"},
			"proto":              &inputs.TagInfo{Desc: "示例 HTTP/1.0"},
		},
		Fields: map[string]interface{}{
			"status_code": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.UnknownUnit,
				Desc:     "web page response code",
			},
			"message": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.Gauge,
				Unit:     inputs.UnknownUnit,
				Desc:     "包括请求头(request_header)/请求体(request_body)/返回头(response_header)/返回体(response_body)/fail_reason 冗余一份",
			},
			"fail_reason": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.Gauge,
				Unit:     inputs.UnknownUnit,
				Desc:     "拨测失败原因",
			},
			"response_time": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.DurationUS,
				Desc:     "HTTP 相应时间, 单位 ms",
			},
			"response_body_size": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.SizeByte,
				Desc:     "body 长度",
			},
			"success": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.UnknownUnit,
				Desc:     "只有 1/-1 两种状态, 1 表示成功, -1 表示失败",
			},
			"proto": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.Gauge,
				Unit:     inputs.UnknownUnit,
				Desc:     "示例 HTTP/1.0",
			},
		},
	}
}

type tcpMeasurement struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
	ts     time.Time
}

func (m *tcpMeasurement) LineProto() (*io.Point, error) {
	return io.MakePoint(m.name, m.tags, m.fields, m.ts)
}

//nolint:lll
func (m *tcpMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "tcp_dial_testing",
		Tags: map[string]interface{}{
			"name":      &inputs.TagInfo{Desc: "示例 拨测名称,百度测试"},
			"dest_host": &inputs.TagInfo{Desc: "示例 wwww.baidu.com"},
			"dest_port": &inputs.TagInfo{Desc: "示例 80"},
			"country":   &inputs.TagInfo{Desc: "示例 中国"},
			"province":  &inputs.TagInfo{Desc: "示例 浙江"},
			"city":      &inputs.TagInfo{Desc: "示例 杭州"},
			"internal":  &inputs.TagInfo{Desc: "示例 true（国内 true /海外 false）"},
			"isp":       &inputs.TagInfo{Desc: "示例 电信/移动/联通"},
			"status":    &inputs.TagInfo{Desc: "示例 OK/FAIL 两种状态 "},
		},
		Fields: map[string]interface{}{
			"message": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.Gauge,
				Unit:     inputs.UnknownUnit,
				Desc:     "包括请求头(request_header)/请求体(request_body)/返回头(response_header)/返回体(response_body)/fail_reason 冗余一份",
			},
			"fail_reason": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.Gauge,
				Unit:     inputs.UnknownUnit,
				Desc:     "拨测失败原因",
			},
			"response_time": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.DurationUS,
				Desc:     "TCP 连接时间, 单位 us",
			},
			"success": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.UnknownUnit,
				Desc:     "只有 1/-1 两种状态, 1 表示成功, -1 表示失败",
			},
		},
	}
}

type icmpMeasurement struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
	ts     time.Time
}

func (m *icmpMeasurement) LineProto() (*io.Point, error) {
	return io.MakePoint(m.name, m.tags, m.fields, m.ts)
}

//nolint:lll
func (m *icmpMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "icmp_dial_testing",
		Tags: map[string]interface{}{
			"name":      &inputs.TagInfo{Desc: "示例 拨测名称,百度测试"},
			"dest_host": &inputs.TagInfo{Desc: "示例 wwww.baidu.com"},
			"country":   &inputs.TagInfo{Desc: "示例 中国"},
			"province":  &inputs.TagInfo{Desc: "示例 浙江"},
			"city":      &inputs.TagInfo{Desc: "示例 杭州"},
			"internal":  &inputs.TagInfo{Desc: "示例 true（国内 true /海外 false）"},
			"isp":       &inputs.TagInfo{Desc: "示例 电信/移动/联通"},
			"status":    &inputs.TagInfo{Desc: "示例 OK/FAIL 两种状态 "},
		},
		Fields: map[string]interface{}{
			"message": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.Gauge,
				Unit:     inputs.UnknownUnit,
				Desc:     "包括请求头(request_header)/请求体(request_body)/返回头(response_header)/返回体(response_body)/fail_reason 冗余一份",
			},
			"fail_reason": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.Gauge,
				Unit:     inputs.UnknownUnit,
				Desc:     "拨测失败原因",
			},
			"average_round_trip_time_in_millis": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.DurationMS,
				Desc:     "平均往返时间(RTT), ms",
			},
			"min_round_trip_time_in_millis": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.DurationMS,
				Desc:     "最小往返时间(RTT), ms",
			},
			"max_round_trip_time_in_millis": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.DurationMS,
				Desc:     "最大往返时间(RTT), ms",
			},
			"packet_loss_percent": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.DurationMS,
				Desc:     "丢包率",
			},
			"success": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.UnknownUnit,
				Desc:     "只有 1/-1 两种状态, 1 表示成功, -1 表示失败",
			},
		},
	}
}
