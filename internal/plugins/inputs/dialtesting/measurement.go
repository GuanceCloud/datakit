// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package dialtesting

import (
	"github.com/GuanceCloud/cliutils/point"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

type httpMeasurement struct{}

func withTaskIDField(fields map[string]interface{}) map[string]interface{} {
	fields["task_id"] = &inputs.FieldInfo{
		DataType: inputs.String,
		Type:     inputs.Gauge,
		Unit:     inputs.NoUnit,
		Desc:     "The dialtesting task external ID",
	}
	return fields
}

//nolint:lll
func (m *httpMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   "http_dial_testing",
		Cat:    point.DialTesting,
		Desc:   "HTTP synthetic test results, including response status, phase timings, payload size, certificate expiry, and failure details.",
		DescZh: "HTTP 拨测结果，包含响应状态、各阶段耗时、响应体大小、证书到期时间和失败详情。",
		Tags: map[string]interface{}{
			"name":               &inputs.TagInfo{Desc: "The name of the task"},
			"url":                &inputs.TagInfo{Desc: "The URL of the endpoint to be monitored"},
			"node_name":          &inputs.TagInfo{Desc: "The name of the node"},
			"dest_ip":            &inputs.TagInfo{Desc: "The IP address of the destination"},
			"country":            &inputs.TagInfo{Desc: "The name of the country"},
			"province":           &inputs.TagInfo{Desc: "The name of the province"},
			"city":               &inputs.TagInfo{Desc: "The name of the city"},
			"internal":           &inputs.TagInfo{Desc: "The boolean value, true for domestic and false for overseas"},
			"isp":                &inputs.TagInfo{Desc: "ISP, such as `chinamobile`, `chinaunicom`, `chinatelecom`"},
			"status":             &inputs.TagInfo{Desc: "The status of the task, either 'OK' or 'FAIL'"},
			"status_code_class":  &inputs.TagInfo{Desc: "The class of the status code, such as '2xx'"},
			"status_code_string": &inputs.TagInfo{Desc: "The status string, such as '200 OK'"},
			"proto":              &inputs.TagInfo{Desc: "The protocol of the HTTP, such as 'HTTP/1.1'"},
			"method":             &inputs.TagInfo{Desc: "HTTP method, such as `GET`"},
			"owner":              &inputs.TagInfo{Desc: "The owner name"}, // used for fees calculation
			"datakit_version":    &inputs.TagInfo{Desc: "The DataKit version"},
			LabelDF:              &inputs.TagInfo{Desc: "The label of the task"},
		},
		Fields: withTaskIDField(map[string]interface{}{
			"status_code": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NoUnit,
				Desc:     "The response code",
			},
			"message": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.String,
				Unit:     inputs.NoUnit,
				Desc:     "The message string which includes the header and the body of the request or the response",
			},
			"task": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.String,
				Unit:     inputs.NoUnit,
				Desc:     "The raw task string",
			},
			"fail_reason": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.String,
				Unit:     inputs.NoUnit,
				Desc:     "The reason that leads to the failure of the task",
			},
			"response_time": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.DurationUS,
				Desc:     "The time of the response",
			},
			"response_download": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.DurationUS,
				Desc:     "HTTP downloading time",
			},
			"response_ttfb": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.DurationUS,
				Desc:     "HTTP response `ttfb`",
			},
			"response_dns": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.DurationUS,
				Desc:     "HTTP DNS parsing time",
			},
			"response_ssl": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.DurationUS,
				Desc:     "HTTP ssl handshake time",
			},
			"response_connection": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.DurationUS,
				Desc:     "HTTP connection time",
			},
			"response_body_size": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.SizeByte,
				Desc:     "The length of the body of the response",
			},
			"success": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NoUnit,
				Desc:     "The number to specify whether is successful, 1 for success, -1 for failure",
			},
			"seq_number": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The sequence number of the test",
			},
			"config_vars": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.String,
				Unit:     inputs.NoUnit,
				Desc:     "The configuration variables of the task",
			},
			"ssl_cert_not_after": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.TimestampUS,
				Desc:     "The SSL certificate not after time",
			},
			"ssl_cert_expires_in_days": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.DurationDay,
				Desc:     "The SSL certificate expires in days",
			},
		}),
	}
}

type tcpMeasurement struct{}

//nolint:lll
func (m *tcpMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   "tcp_dial_testing",
		Cat:    point.DialTesting,
		Desc:   "TCP synthetic test results, including connection latency, DNS-inclusive latency, traceroute output, and failure details.",
		DescZh: "TCP 拨测结果，包含连接耗时、含 DNS 的总耗时、路由追踪结果和失败详情。",
		Tags: map[string]interface{}{
			"name":            &inputs.TagInfo{Desc: "The name of the task"},
			"dest_host":       &inputs.TagInfo{Desc: "The name of the host to be monitored"},
			"dest_port":       &inputs.TagInfo{Desc: "The port of the TCP connection"},
			"dest_ip":         &inputs.TagInfo{Desc: "The IP address"},
			"node_name":       &inputs.TagInfo{Desc: "The name of the node"},
			"country":         &inputs.TagInfo{Desc: "The name of the country"},
			"province":        &inputs.TagInfo{Desc: "The name of the province"},
			"city":            &inputs.TagInfo{Desc: "The name of the city"},
			"internal":        &inputs.TagInfo{Desc: "The boolean value, true for domestic and false for overseas"},
			"isp":             &inputs.TagInfo{Desc: "ISP, such as `chinamobile`, `chinaunicom`, `chinatelecom`"},
			"status":          &inputs.TagInfo{Desc: "The status of the task, either 'OK' or 'FAIL'"},
			"proto":           &inputs.TagInfo{Desc: "The protocol of the task"},
			"owner":           &inputs.TagInfo{Desc: "The owner name"}, // used for fees calculation
			"datakit_version": &inputs.TagInfo{Desc: "The DataKit version"},
			LabelDF:           &inputs.TagInfo{Desc: "The label of the task"},
		},
		Fields: withTaskIDField(map[string]interface{}{
			"message": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.String,
				Unit:     inputs.NoUnit,
				Desc:     "The message string includes the response time or fail reason",
			},
			"task": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.String,
				Unit:     inputs.NoUnit,
				Desc:     "The raw task string",
			},
			"traceroute": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.String,
				Unit:     inputs.NoUnit,
				Desc:     "The json string fo the `traceroute` result",
			},
			"fail_reason": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.String,
				Unit:     inputs.NoUnit,
				Desc:     "The reason that leads to the failure of the task",
			},
			"response_time": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.DurationUS,
				Desc:     "The time of the response ",
			},
			"response_time_with_dns": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.DurationUS,
				Desc:     "The time of the response, which contains DNS time",
			},
			"success": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NoUnit,
				Desc:     "The number to specify whether is successful, 1 for success, -1 for failure",
			},
			"seq_number": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The sequence number of the test",
			},
			"config_vars": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.String,
				Unit:     inputs.NoUnit,
				Desc:     "The configuration variables of the task",
			},
		}),
	}
}

type icmpMeasurement struct{}

//nolint:lll
func (m *icmpMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   "icmp_dial_testing",
		Cat:    point.DialTesting,
		Desc:   "ICMP synthetic test results, including packet latency, packet loss, routing information, and failure details.",
		DescZh: "ICMP 拨测结果，包含报文时延、丢包、路由信息和失败详情。",
		Tags: map[string]interface{}{
			"name":            &inputs.TagInfo{Desc: "The name of the task"},
			"dest_host":       &inputs.TagInfo{Desc: "The name of the host to be monitored"},
			"node_name":       &inputs.TagInfo{Desc: "The name of the node"},
			"country":         &inputs.TagInfo{Desc: "The name of the country"},
			"province":        &inputs.TagInfo{Desc: "The name of the province"},
			"city":            &inputs.TagInfo{Desc: "The name of the city"},
			"internal":        &inputs.TagInfo{Desc: "The boolean value, true for domestic and false for overseas"},
			"isp":             &inputs.TagInfo{Desc: "ISP, such as `chinamobile`, `chinaunicom`, `chinatelecom`"},
			"status":          &inputs.TagInfo{Desc: "The status of the task, either 'OK' or 'FAIL'"},
			"proto":           &inputs.TagInfo{Desc: "The protocol of the task"},
			"owner":           &inputs.TagInfo{Desc: "The owner name"}, // used for fees calculation
			"datakit_version": &inputs.TagInfo{Desc: "The DataKit version"},
			LabelDF:           &inputs.TagInfo{Desc: "The label of the task"},
		},
		Fields: withTaskIDField(map[string]interface{}{
			"message": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.String,
				Unit:     inputs.NoUnit,
				Desc:     "The message string includes the average time of the round trip or the failure reason",
			},
			"task": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.String,
				Unit:     inputs.NoUnit,
				Desc:     "The raw task string",
			},
			"fail_reason": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.String,
				Unit:     inputs.NoUnit,
				Desc:     "The reason that leads to the failure of the task",
			},
			"traceroute": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.String,
				Unit:     inputs.NoUnit,
				Desc:     "The `json` string fo the `traceroute` result",
			},
			"average_round_trip_time": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.DurationUS,
				Desc:     "The average time of the round trip(RTT)",
			},
			"average_round_trip_time_in_millis": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.DurationMS,
				Desc:     "The average time of the round trip(RTT), deprecated",
			},
			"min_round_trip_time": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.DurationUS,
				Desc:     "The minimum time of the round trip(RTT)",
			},
			"min_round_trip_time_in_millis": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.DurationMS,
				Desc:     "The minimum time of the round trip(RTT), deprecated",
			},
			"std_round_trip_time": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.DurationUS,
				Desc:     "The standard deviation of the round trip",
			},
			"std_round_trip_time_in_millis": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.DurationMS,
				Desc:     "The standard deviation of the round trip, deprecated",
			},
			"max_round_trip_time": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.DurationUS,
				Desc:     "The maximum time of the round trip(RTT)",
			},
			"max_round_trip_time_in_millis": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.DurationMS,
				Desc:     "The maximum time of the round trip(RTT), deprecated",
			},
			"packet_loss_percent": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.Percent,
				Desc:     "The loss percent of the packets",
			},
			"packets_received": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The number of the packets received",
			},
			"packets_sent": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The number of the packets sent",
			},
			"success": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NoUnit,
				Desc:     "The number to specify whether is successful, 1 for success, -1 for failure",
			},
			"seq_number": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The sequence number of the test",
			},
			"config_vars": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.String,
				Unit:     inputs.NoUnit,
				Desc:     "The configuration variables of the task",
			},
		}),
	}
}

type websocketMeasurement struct{}

//nolint:lll
func (m *websocketMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   "websocket_dial_testing",
		Cat:    point.DialTesting,
		Desc:   "WebSocket synthetic test results, including handshake status, response latency, and failure details.",
		DescZh: "WebSocket 拨测结果，包含握手状态、响应耗时和失败详情。",
		Tags: map[string]interface{}{
			"name":            &inputs.TagInfo{Desc: "The name of the task"},
			"url":             &inputs.TagInfo{Desc: "The URL string, such as `ws://www.abc.com`"},
			"node_name":       &inputs.TagInfo{Desc: "The name of the node"},
			"country":         &inputs.TagInfo{Desc: "The name of the country"},
			"province":        &inputs.TagInfo{Desc: "The name of the province"},
			"city":            &inputs.TagInfo{Desc: "The name of the city"},
			"internal":        &inputs.TagInfo{Desc: "The boolean value, true for domestic and false for overseas"},
			"isp":             &inputs.TagInfo{Desc: "ISP, such as `chinamobile`, `chinaunicom`, `chinatelecom`"},
			"status":          &inputs.TagInfo{Desc: "The status of the task, either 'OK' or 'FAIL'"},
			"proto":           &inputs.TagInfo{Desc: "The protocol of the task"},
			"owner":           &inputs.TagInfo{Desc: "The owner name"}, // used for fees calculation
			"datakit_version": &inputs.TagInfo{Desc: "The DataKit version"},
			LabelDF:           &inputs.TagInfo{Desc: "The label of the task"},
		},
		Fields: withTaskIDField(map[string]interface{}{
			"message": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.String,
				Unit:     inputs.NoUnit,
				Desc:     "The message string includes the response time or the failure reason",
			},
			"task": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.String,
				Unit:     inputs.NoUnit,
				Desc:     "The raw task string",
			},
			"fail_reason": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.String,
				Unit:     inputs.NoUnit,
				Desc:     "The reason that leads to the failure of the task",
			},
			"response_message": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.String,
				Unit:     inputs.NoUnit,
				Desc:     "The message of the response",
			},
			"sent_message": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.String,
				Unit:     inputs.NoUnit,
				Desc:     "The sent message ",
			},
			"response_time": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.DurationUS,
				Desc:     "The time of the response",
			},
			"response_time_with_dns": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.DurationUS,
				Desc:     "The time of the response, include DNS",
			},
			"success": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NoUnit,
				Desc:     "The number to specify whether is successful, 1 for success, -1 for failure",
			},
			"seq_number": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The sequence number of the test",
			},
			"config_vars": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.String,
				Unit:     inputs.NoUnit,
				Desc:     "The configuration variables of the task",
			},
			"ssl_cert_not_after": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.TimestampUS,
				Desc:     "The SSL certificate not after time",
			},
			"ssl_cert_expires_in_days": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.DurationDay,
				Desc:     "The SSL certificate expires in days",
			},
		}),
	}
}

type multiMeasurement struct{}

//nolint:lll
func (m *multiMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   "multi_dial_testing",
		Cat:    point.DialTesting,
		Desc:   "Multi-step synthetic test results, including step execution status, elapsed time, and failure details.",
		DescZh: "多步骤拨测结果，包含步骤执行状态、耗时和失败详情。",
		Tags: map[string]interface{}{
			"name":            &inputs.TagInfo{Desc: "The name of the task"},
			"node_name":       &inputs.TagInfo{Desc: "The name of the node"},
			"country":         &inputs.TagInfo{Desc: "The name of the country"},
			"province":        &inputs.TagInfo{Desc: "The name of the province"},
			"city":            &inputs.TagInfo{Desc: "The name of the city"},
			"internal":        &inputs.TagInfo{Desc: "The boolean value, true for domestic and false for overseas"},
			"isp":             &inputs.TagInfo{Desc: "ISP, such as `chinamobile`, `chinaunicom`, `chinatelecom`"},
			"status":          &inputs.TagInfo{Desc: "The status of the task, either 'OK' or 'FAIL'"},
			"owner":           &inputs.TagInfo{Desc: "The owner name"}, // used for fees calculation
			"datakit_version": &inputs.TagInfo{Desc: "The DataKit version"},
			LabelDF:           &inputs.TagInfo{Desc: "The label of the task"},
		},
		Fields: withTaskIDField(map[string]interface{}{
			"last_step": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The last number of the task be executed",
			},
			"message": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.String,
				Unit:     inputs.NoUnit,
				Desc:     "The message string which includes the header and the body of the request or the response",
			},
			"task": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.String,
				Unit:     inputs.NoUnit,
				Desc:     "The raw task string",
			},
			"fail_reason": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.String,
				Unit:     inputs.NoUnit,
				Desc:     "The reason that leads to the failure of the task",
			},
			"steps": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.String,
				Unit:     inputs.NoUnit,
				Desc:     "The result of each step",
			},
			"response_time": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.DurationUS,
				Desc:     "The time of the response",
			},
			"success": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NoUnit,
				Desc:     "The number to specify whether is successful, 1 for success, -1 for failure",
			},
			"seq_number": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The sequence number of the test",
			},
			"config_vars": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.String,
				Unit:     inputs.NoUnit,
				Desc:     "The configuration variables of the task",
			},
		}),
	}
}

type grpcMeasurement struct{}

//nolint:lll
func (m *grpcMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   "grpc_dial_testing",
		Cat:    point.DialTesting,
		Desc:   "gRPC synthetic test results, including RPC response status, latency, response data, and failure details.",
		DescZh: "gRPC 拨测结果，包含 RPC 响应状态、耗时、响应数据和失败详情。",
		Tags: map[string]interface{}{
			"name":            &inputs.TagInfo{Desc: "The name of the task"},
			"server":          &inputs.TagInfo{Desc: "The gRPC server address"},
			"dest_host":       &inputs.TagInfo{Desc: "The name of the host to be monitored"},
			"method":          &inputs.TagInfo{Desc: "The gRPC method name"},
			"node_name":       &inputs.TagInfo{Desc: "The name of the node"},
			"country":         &inputs.TagInfo{Desc: "The name of the country"},
			"province":        &inputs.TagInfo{Desc: "The name of the province"},
			"city":            &inputs.TagInfo{Desc: "The name of the city"},
			"internal":        &inputs.TagInfo{Desc: "The boolean value, true for domestic and false for overseas"},
			"isp":             &inputs.TagInfo{Desc: "ISP, such as `chinamobile`, `chinaunicom`, `chinatelecom`"},
			"status":          &inputs.TagInfo{Desc: "The status of the task, either 'OK' or 'FAIL'"},
			"proto":           &inputs.TagInfo{Desc: "The protocol of the task"},
			"owner":           &inputs.TagInfo{Desc: "The owner name"},
			"datakit_version": &inputs.TagInfo{Desc: "The DataKit version"},
			LabelDF:           &inputs.TagInfo{Desc: "The label of the task"},
		},
		Fields: withTaskIDField(map[string]interface{}{
			"message": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.String,
				Unit:     inputs.NoUnit,
				Desc:     "The message string includes the response time or the failure reason",
			},
			"task": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.String,
				Unit:     inputs.NoUnit,
				Desc:     "The raw task string",
			},
			"fail_reason": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.String,
				Unit:     inputs.NoUnit,
				Desc:     "The reason that leads to the failure of the task",
			},
			"response_time": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.DurationUS,
				Desc:     "The time of the response",
			},
			"success": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NoUnit,
				Desc:     "The number to specify whether is successful, 1 for success, -1 for failure",
			},
			"seq_number": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The sequence number of the test",
			},
			"config_vars": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.String,
				Unit:     inputs.NoUnit,
				Desc:     "The configuration variables of the task",
			},
			"ssl_cert_not_after": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.TimestampUS,
				Desc:     "The SSL certificate not after time",
			},
			"ssl_cert_expires_in_days": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.DurationDay,
				Desc:     "The SSL certificate expires in days",
			},
		}),
	}
}

type sslMeasurement struct{}

//nolint:lll
func (m *sslMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "ssl_dial_testing",
		Cat:  point.DialTesting,
		Tags: map[string]interface{}{
			"name":            &inputs.TagInfo{Desc: "The name of the task"},
			"dest_host":       &inputs.TagInfo{Desc: "The name of the host to be monitored"},
			"dest_port":       &inputs.TagInfo{Desc: "The port of the SSL connection"},
			"dest_ip":         &inputs.TagInfo{Desc: "The IP address"},
			"server_name":     &inputs.TagInfo{Desc: "The TLS server name"},
			"node_name":       &inputs.TagInfo{Desc: "The name of the node"},
			"country":         &inputs.TagInfo{Desc: "The name of the country"},
			"province":        &inputs.TagInfo{Desc: "The name of the province"},
			"city":            &inputs.TagInfo{Desc: "The name of the city"},
			"internal":        &inputs.TagInfo{Desc: "The boolean value, true for domestic and false for overseas"},
			"isp":             &inputs.TagInfo{Desc: "ISP, such as `chinamobile`, `chinaunicom`, `chinatelecom`"},
			"status":          &inputs.TagInfo{Desc: "The status of the task, either 'OK' or 'FAIL'"},
			"proto":           &inputs.TagInfo{Desc: "The protocol of the task"},
			"owner":           &inputs.TagInfo{Desc: "The owner name"}, // used for fees calculation
			"datakit_version": &inputs.TagInfo{Desc: "The DataKit version"},
			LabelDF:           &inputs.TagInfo{Desc: "The label of the task"},
		},
		Fields: withTaskIDField(map[string]interface{}{
			"message": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.Gauge,
				Unit:     inputs.NoUnit,
				Desc:     "The message string includes the response time or fail reason",
			},
			"task": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.Gauge,
				Unit:     inputs.NoUnit,
				Desc:     "The raw task string",
			},
			"fail_reason": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.Gauge,
				Unit:     inputs.NoUnit,
				Desc:     "The reason that leads to the failure of the task",
			},
			"response_time": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.DurationUS,
				Desc:     "The TCP connection and TLS handshake duration",
			},
			"tls_handshake_time": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.DurationUS,
				Desc:     "The TLS handshake duration",
			},
			"tls_version": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.Gauge,
				Unit:     inputs.NoUnit,
				Desc:     "The TLS protocol version",
			},
			"ssl_cert_subject": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.Gauge,
				Unit:     inputs.NoUnit,
				Desc:     "The SSL certificate subject",
			},
			"ssl_cert_issuer": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.Gauge,
				Unit:     inputs.NoUnit,
				Desc:     "The SSL certificate issuer",
			},
			"ssl_cert_not_before": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.TimestampUS,
				Desc:     "The SSL certificate not before time",
			},
			"ssl_cert_not_after": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.TimestampUS,
				Desc:     "The SSL certificate not after time",
			},
			"ssl_cert_expires_in_days": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.DurationDay,
				Desc:     "The SSL certificate expires in days",
			},
			"success": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NoUnit,
				Desc:     "The number to specify whether is successful, 1 for success, -1 for failure",
			},
			"seq_number": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.Count,
				Desc:     "The sequence number of the test",
			},
			"config_vars": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.Gauge,
				Unit:     inputs.NoUnit,
				Desc:     "The configuration variables of the task",
			},
		}),
	}
}

type browserMeasurement struct{}

//nolint:lll
func (m *browserMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "browser_dial_testing",
		Cat:  point.DialTesting,
		Tags: map[string]interface{}{
			"name":            &inputs.TagInfo{Desc: "The name of the task"},
			"url":             &inputs.TagInfo{Desc: "The URL of the page to be monitored"},
			"node_name":       &inputs.TagInfo{Desc: "The name of the node"},
			"country":         &inputs.TagInfo{Desc: "The name of the country"},
			"province":        &inputs.TagInfo{Desc: "The name of the province"},
			"city":            &inputs.TagInfo{Desc: "The name of the city"},
			"internal":        &inputs.TagInfo{Desc: "The boolean value, true for domestic and false for overseas"},
			"isp":             &inputs.TagInfo{Desc: "ISP, such as `chinamobile`, `chinaunicom`, `chinatelecom`"},
			"status":          &inputs.TagInfo{Desc: "The status of the task, either 'OK' or 'FAIL'"},
			"browser_engine":  &inputs.TagInfo{Desc: "The browser engine used to run the task"},
			"viewport":        &inputs.TagInfo{Desc: "The browser viewport size, such as `1920x1080`"},
			"owner":           &inputs.TagInfo{Desc: "The owner name"},
			"datakit_version": &inputs.TagInfo{Desc: "The DataKit version"},
			LabelDF:           &inputs.TagInfo{Desc: "The label of the task"},
		},
		Fields: withTaskIDField(map[string]interface{}{
			"message": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.Gauge,
				Unit:     inputs.NoUnit,
				Desc:     "The message string includes success message or failure reason",
			},
			"fail_reason": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.Gauge,
				Unit:     inputs.NoUnit,
				Desc:     "The reason that leads to the failure of the task",
			},
			"response_time": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.DurationUS,
				Desc:     "The browser run duration",
			},
			"success": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NoUnit,
				Desc:     "The number to specify whether is successful, 1 for success, -1 for failure",
			},
			"seq_number": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.Count,
				Desc:     "The sequence number of the test",
			},
			"last_step": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The last browser step sequence number",
			},
			"steps": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.Gauge,
				Unit:     inputs.NoUnit,
				Desc:     "The JSON string of browser step results",
			},
			"browser_run_id": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.Gauge,
				Unit:     inputs.NoUnit,
				Desc:     "The browser run ID",
			},
			"viewport_width": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NoUnit,
				Desc:     "The browser viewport width",
			},
			"viewport_height": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NoUnit,
				Desc:     "The browser viewport height",
			},
			"retry_count": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.Count,
				Desc:     "The retry count of the browser run",
			},
			"retry_records": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.Gauge,
				Unit:     inputs.NoUnit,
				Desc:     "The JSON string of browser retry attempt records",
			},
			"trace_id": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.Gauge,
				Unit:     inputs.NoUnit,
				Desc:     "The first trace ID captured during the browser run",
			},
			"browser_config_vars": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.Gauge,
				Unit:     inputs.NoUnit,
				Desc:     "The JSON string of variables defined in browser_config",
			},
			"has_screenshot": &inputs.FieldInfo{
				DataType: inputs.Bool,
				Type:     inputs.Gauge,
				Unit:     inputs.NoUnit,
				Desc:     "Whether the browser run has uploaded screenshots",
			},
			"screenshot_upload_error": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.Gauge,
				Unit:     inputs.NoUnit,
				Desc:     "The browser screenshot upload error",
			},
			"config_vars": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.Gauge,
				Unit:     inputs.NoUnit,
				Desc:     "The configuration variables of the task",
			},
		}),
	}
}
