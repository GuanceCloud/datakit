// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build !windows
// +build !windows

package dialtesting

import (
	"github.com/GuanceCloud/cliutils/point"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

type httpMeasurement struct{}

//nolint:lll
func (m *httpMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "http_dial_testing",
		Cat:  point.DialTesting,
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
		Fields: map[string]interface{}{
			"status_code": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NoUnit,
				Desc:     "The response code",
			},
			"message": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.Gauge,
				Unit:     inputs.NoUnit,
				Desc:     "The message string which includes the header and the body of the request or the response",
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
				Unit:     inputs.Count,
				Desc:     "The sequence number of the test",
			},
			"config_vars": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.Gauge,
				Unit:     inputs.NoUnit,
				Desc:     "The configuration variables of the task",
			},
		},
	}
}

type tcpMeasurement struct{}

//nolint:lll
func (m *tcpMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "tcp_dial_testing",
		Cat:  point.DialTesting,
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
		Fields: map[string]interface{}{
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
			"traceroute": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.Gauge,
				Unit:     inputs.NoUnit,
				Desc:     "The json string fo the `traceroute` result",
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
				Unit:     inputs.Count,
				Desc:     "The sequence number of the test",
			},
			"config_vars": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.Gauge,
				Unit:     inputs.NoUnit,
				Desc:     "The configuration variables of the task",
			},
		},
	}
}

type icmpMeasurement struct{}

//nolint:lll
func (m *icmpMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "icmp_dial_testing",
		Cat:  point.DialTesting,
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
		Fields: map[string]interface{}{
			"message": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.Gauge,
				Unit:     inputs.NoUnit,
				Desc:     "The message string includes the average time of the round trip or the failure reason",
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
			"traceroute": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.Gauge,
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
				Unit:     inputs.Count,
				Desc:     "The number of the packets received",
			},
			"packets_sent": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.Count,
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
				Unit:     inputs.Count,
				Desc:     "The sequence number of the test",
			},
			"config_vars": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.Gauge,
				Unit:     inputs.NoUnit,
				Desc:     "The configuration variables of the task",
			},
		},
	}
}

type websocketMeasurement struct{}

//nolint:lll
func (m *websocketMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "websocket_dial_testing",
		Cat:  point.DialTesting,
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
		Fields: map[string]interface{}{
			"message": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.Gauge,
				Unit:     inputs.NoUnit,
				Desc:     "The message string includes the response time or the failure reason",
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
			"response_message": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.Gauge,
				Unit:     inputs.NoUnit,
				Desc:     "The message of the response",
			},
			"sent_message": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.Gauge,
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
				Unit:     inputs.Count,
				Desc:     "The sequence number of the test",
			},
			"config_vars": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.Gauge,
				Unit:     inputs.NoUnit,
				Desc:     "The configuration variables of the task",
			},
		},
	}
}

type multiMeasurement struct{}

//nolint:lll
func (m *multiMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "multi_dial_testing",
		Cat:  point.DialTesting,
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
		Fields: map[string]interface{}{
			"last_step": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The last number of the task be executed",
			},
			"message": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.Gauge,
				Unit:     inputs.NoUnit,
				Desc:     "The message string which includes the header and the body of the request or the response",
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
			"steps": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.Gauge,
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
				Unit:     inputs.Count,
				Desc:     "The sequence number of the test",
			},
			"config_vars": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.Gauge,
				Unit:     inputs.NoUnit,
				Desc:     "The configuration variables of the task",
			},
		},
	}
}
